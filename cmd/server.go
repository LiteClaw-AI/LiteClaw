package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start OpenAI-compatible API server",
	Long: `Start an HTTP server with OpenAI-compatible API endpoints.

Examples:
  # Start server on default port
  liteclaw server

  # Custom host and port
  liteclaw server --host 0.0.0.0 --port 8080

  # With CORS enabled
  liteclaw server --cors`,
	RunE: runServer,
}

var (
	serverHost string
	serverPort int
	serverCORS bool
)

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVarP(&serverHost, "host", "H", "127.0.0.1", "Host to bind")
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Port to listen")
	serverCmd.Flags().BoolVar(&serverCORS, "cors", false, "Enable CORS")
}

func runServer(cmd *cobra.Command, args []string) error {
	fmt.Printf("🌐 Starting server at http://%s:%d\n", serverHost, serverPort)
	fmt.Printf("   CORS: %v\n", serverCORS)
	fmt.Println()

	// TODO: Implement actual server
	return fmt.Errorf("server not yet implemented")
}
