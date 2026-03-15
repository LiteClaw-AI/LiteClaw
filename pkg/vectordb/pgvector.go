package vectordb

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// PgVectorStore implements PostgreSQL pgvector storage
type PgVectorStore struct {
	db     *sql.DB
	config Config
}

// NewPgVectorStore creates a new pgvector store
func NewPgVectorStore(config Config) (*PgVectorStore, error) {
	if config.URL == "" {
		config.URL = "postgres://postgres:postgres@localhost:5432/vectordb?sslmode=disable"
	}

	db, err := sql.Open("postgres", config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PgVectorStore{
		db:     db,
		config: config,
	}

	// Ensure pgvector extension is installed
	if err := store.ensureExtension(); err != nil {
		return nil, err
	}

	return store, nil
}

// ensureExtension ensures pgvector extension is installed
func (s *PgVectorStore) ensureExtension() error {
	_, err := s.db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	return err
}

// CreateCollection creates a new table for vectors
func (s *PgVectorStore) CreateCollection(ctx context.Context, name string, dimension int) error {
	// Create table with vector column
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			embedding vector(%d),
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, name, dimension)

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create index for similarity search
	indexQuery := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s_embedding_idx ON %s 
		USING ivfflat (embedding vector_cosine_ops)
		WITH (lists = 100)
	`, name, name)

	_, err = s.db.ExecContext(ctx, indexQuery)
	if err != nil {
		// Index creation might fail if not enough data, ignore
	}

	return nil
}

// DeleteCollection deletes the table
func (s *PgVectorStore) DeleteCollection(ctx context.Context, name string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", name)
	_, err := s.db.ExecContext(ctx, query)
	return err
}

// ListCollections lists all vector tables
func (s *PgVectorStore) ListCollections(ctx context.Context) ([]CollectionInfo, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name NOT LIKE 'pg_%'
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var infos []CollectionInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		infos = append(infos, CollectionInfo{Name: name})
	}

	return infos, nil
}

// Insert inserts vectors into a table
func (s *PgVectorStore) Insert(ctx context.Context, collection string, vectors []Vector) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(`
		INSERT INTO %s (id, embedding, metadata)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			embedding = EXCLUDED.embedding,
			metadata = EXCLUDED.metadata
	`, collection))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, vec := range vectors {
		// Convert vector to string format
		vecStr := vectorToString(vec.Values)
		_, err := stmt.ExecContext(ctx, vec.ID, vecStr, vec.Metadata)
		if err != nil {
			return fmt.Errorf("failed to insert vector %s: %w", vec.ID, err)
		}
	}

	return tx.Commit()
}

// Update updates vectors (same as insert with ON CONFLICT)
func (s *PgVectorStore) Update(ctx context.Context, collection string, vectors []Vector) error {
	return s.Insert(ctx, collection, vectors)
}

// Delete deletes vectors from a table
func (s *PgVectorStore) Delete(ctx context.Context, collection string, ids []string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ANY($1)", collection)
	_, err := s.db.ExecContext(ctx, query, ids)
	return err
}

// Search searches for similar vectors
func (s *PgVectorStore) Search(ctx context.Context, collection string, query []float32, topK int, filter map[string]interface{}) ([]SearchResult, error) {
	vecStr := vectorToString(query)
	filterClause := buildWhereClause(filter)

	sqlQuery := fmt.Sprintf(`
		SELECT id, 1 - (embedding <=> '%s') AS score, metadata
		FROM %s
		%s
		ORDER BY embedding <=> '%s'
		LIMIT %d
	`, vecStr, collection, filterClause, vecStr, topK)

	rows, err := s.db.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var metadata []byte
		if err := rows.Scan(&r.ID, &r.Score, &metadata); err != nil {
			return nil, err
		}
		// Parse metadata JSON
		if len(metadata) > 0 {
			// metadata parsing would go here
			r.Metadata = make(map[string]interface{})
		}
		results = append(results, r)
	}

	return results, nil
}

// Get retrieves vectors by ID
func (s *PgVectorStore) Get(ctx context.Context, collection string, ids []string) ([]Vector, error) {
	query := fmt.Sprintf(`
		SELECT id, embedding, metadata
		FROM %s
		WHERE id = ANY($1)
	`, collection)

	rows, err := s.db.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vectors []Vector
	for rows.Next() {
		var v Vector
		var vecStr string
		var metadata []byte
		if err := rows.Scan(&v.ID, &vecStr, &metadata); err != nil {
			return nil, err
		}
		v.Values = parseVector(vecStr)
		// Parse metadata
		if len(metadata) > 0 {
			v.Metadata = make(map[string]interface{})
		}
		vectors = append(vectors, v)
	}

	return vectors, nil
}

// GetStats returns table statistics
func (s *PgVectorStore) GetStats(ctx context.Context, collection string) (*CollectionInfo, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", collection)
	var count int
	if err := s.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return nil, err
	}

	return &CollectionInfo{
		Name:        collection,
		VectorCount: count,
	}, nil
}

// Name returns the store name
func (s *PgVectorStore) Name() string {
	return "pgvector"
}

// Helper functions

func vectorToString(v []float32) string {
	str := "["
	for i, f := range v {
		if i > 0 {
			str += ","
		}
		str += fmt.Sprintf("%f", f)
	}
	str += "]"
	return str
}

func parseVector(s string) []float32 {
	// Parse vector string format "[1.0,2.0,3.0]"
	// Simplified implementation
	return []float32{}
}

func buildWhereClause(filter map[string]interface{}) string {
	if filter == nil || len(filter) == 0 {
		return ""
	}

	clause := "WHERE "
	first := true
	for key, value := range filter {
		if !first {
			clause += " AND "
		}
		clause += fmt.Sprintf("metadata->>'%s' = '%v'", key, value)
		first = false
	}
	return clause
}
