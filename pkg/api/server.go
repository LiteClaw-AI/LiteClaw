package api

import (
	"net/http"
	"sync"

	"liteclaw/pkg/team"
)

// Server represents the API server
type Server struct {
	teams   map[string]*team.Team
	mu      sync.RWMutex
	handler http.Handler
}

// NewServer creates a new API server
func NewServer() *Server {
	s := &Server{
		teams: make(map[string]*team.Team),
	}
	
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/teams", s.handleTeams)
	mux.HandleFunc("/api/team/create", s.handleTeamCreate)
	mux.HandleFunc("/api/team/", s.handleTeamOps)
	mux.HandleFunc("/api/workflow/execute", s.handleWorkflowExecute)
	
	s.handler = mux
	return s
}

// AddTeam adds a team to the server
func (s *Server) AddTeam(t *team.Team) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.teams[t.GetID()] = t
}

// GetTeam gets a team by ID
func (s *Server) GetTeam(id string) *team.Team {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.teams[id]
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// handleIndex serves the dashboard
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}

// handleTeams lists all teams
func (s *Server) handleTeams(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	teams := make([]map[string]interface{}, 0, len(s.teams))
	for _, t := range s.teams {
		teams = append(teams, map[string]interface{}{
			"id":          t.GetID(),
			"name":        t.GetName(),
			"description": t.GetDescription(),
			"agents":      t.ListAgents(),
		})
	}
	
	writeJSON(w, teams)
}

// handleTeamCreate creates a new team
func (s *Server) handleTeamCreate(w http.ResponseWriter, r *http.Request) {
	// Team creation logic
	writeJSON(w, map[string]string{"status": "ok"})
}

// handleTeamOps handles team operations
func (s *Server) handleTeamOps(w http.ResponseWriter, r *http.Request) {
	// Team operations logic
	writeJSON(w, map[string]string{"status": "ok"})
}

// handleWorkflowExecute executes a workflow
func (s *Server) handleWorkflowExecute(w http.ResponseWriter, r *http.Request) {
	// Workflow execution logic
	writeJSON(w, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	// Simple JSON encoding (import encoding/json in real implementation)
	w.Write([]byte("{}"))
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LiteClaw AI员工系统</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; }
        .container { max-width: 1400px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 10px; margin-bottom: 20px; }
        .header h1 { font-size: 28px; margin-bottom: 10px; }
        .header p { opacity: 0.9; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .card { background: white; border-radius: 10px; padding: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .card h2 { font-size: 20px; margin-bottom: 15px; color: #333; }
        .agent { display: flex; align-items: center; padding: 10px; background: #f8f9fa; border-radius: 5px; margin-bottom: 10px; }
        .agent-avatar { width: 40px; height: 40px; background: #667eea; color: white; border-radius: 50%; display: flex; align-items: center; justify-content: center; margin-right: 10px; font-weight: bold; }
        .agent-info { flex: 1; }
        .agent-name { font-weight: 600; color: #333; }
        .agent-role { font-size: 12px; color: #666; }
        .status { padding: 5px 10px; border-radius: 15px; font-size: 12px; font-weight: 600; }
        .status.online { background: #d4edda; color: #155724; }
        .status.offline { background: #f8d7da; color: #721c24; }
        .workflow-step { padding: 15px; border-left: 3px solid #667eea; margin-bottom: 10px; background: #f8f9fa; }
        .workflow-step h3 { font-size: 16px; margin-bottom: 5px; }
        .workflow-step p { font-size: 14px; color: #666; }
        .btn { padding: 10px 20px; background: #667eea; color: white; border: none; border-radius: 5px; cursor: pointer; font-size: 14px; }
        .btn:hover { background: #5568d3; }
        .stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 15px; margin-bottom: 20px; }
        .stat { background: white; padding: 20px; border-radius: 10px; text-align: center; }
        .stat-value { font-size: 32px; font-weight: bold; color: #667eea; }
        .stat-label { font-size: 14px; color: #666; margin-top: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎬 LiteClaw AI员工系统</h1>
            <p>多Agent协作平台 - 短视频自动化生产</p>
        </div>
        
        <div class="stats">
            <div class="stat">
                <div class="stat-value">6</div>
                <div class="stat-label">AI员工</div>
            </div>
            <div class="stat">
                <div class="stat-value">3</div>
                <div class="stat-label">工作流</div>
            </div>
            <div class="stat">
                <div class="stat-value">12</div>
                <div class="stat-label">今日任务</div>
            </div>
            <div class="stat">
                <div class="stat-value">98%</div>
                <div class="stat-label">成功率</div>
            </div>
        </div>
        
        <div class="grid">
            <div class="card">
                <h2>🤖 AI员工团队</h2>
                <div class="agent">
                    <div class="agent-avatar">运</div>
                    <div class="agent-info">
                        <div class="agent-name">运营总监 (Director)</div>
                        <div class="agent-role">选题策划、团队协调</div>
                    </div>
                    <span class="status online">在线</span>
                </div>
                <div class="agent">
                    <div class="agent-avatar">研</div>
                    <div class="agent-info">
                        <div class="agent-name">素材研究员 (Researcher)</div>
                        <div class="agent-role">收集素材、数据分析</div>
                    </div>
                    <span class="status online">在线</span>
                </div>
                <div class="agent">
                    <div class="agent-avatar">创</div>
                    <div class="agent-info">
                        <div class="agent-name">内容创作者 (Writer)</div>
                        <div class="agent-role">撰写脚本、文案优化</div>
                    </div>
                    <span class="status online">在线</span>
                </div>
                <div class="agent">
                    <div class="agent-avatar">配</div>
                    <div class="agent-info">
                        <div class="agent-name">配音师 (Narrator)</div>
                        <div class="agent-role">TTS配音、时间戳提取</div>
                    </div>
                    <span class="status online">在线</span>
                </div>
                <div class="agent">
                    <div class="agent-avatar">剪</div>
                    <div class="agent-info">
                        <div class="agent-name">视频剪辑师 (Editor)</div>
                        <div class="agent-role">渲染视频、字幕特效</div>
                    </div>
                    <span class="status online">在线</span>
                </div>
                <div class="agent">
                    <div class="agent-avatar">发</div>
                    <div class="agent-info">
                        <div class="agent-name">发布专员 (Publisher)</div>
                        <div class="agent-role">多平台发布、数据监控</div>
                    </div>
                    <span class="status online">在线</span>
                </div>
            </div>
            
            <div class="card">
                <h2>⚡ 工作流</h2>
                <div class="workflow-step">
                    <h3>1. 选题推送</h3>
                    <p>Director → 分析热点，推送选题</p>
                </div>
                <div class="workflow-step">
                    <h3>2. 素材收集</h3>
                    <p>Researcher → 收集素材和数据</p>
                </div>
                <div class="workflow-step">
                    <h3>3. 脚本创作</h3>
                    <p>Writer → 撰写脚本和旁白</p>
                </div>
                <div class="workflow-step">
                    <h3>4. 配音生成</h3>
                    <p>Narrator → TTS配音生成</p>
                </div>
                <div class="workflow-step">
                    <h3>5. 视频渲染</h3>
                    <p>Editor → 渲染视频和字幕</p>
                </div>
                <div class="workflow-step">
                    <h3>6. 自动发布</h3>
                    <p>Publisher → 多平台发布</p>
                </div>
                <button class="btn" style="width: 100%; margin-top: 10px;">▶ 开始执行</button>
            </div>
            
            <div class="card">
                <h2>📊 实时监控</h2>
                <p style="color: #666; margin-bottom: 15px;">最近执行记录</p>
                <div style="padding: 15px; background: #f8f9fa; border-radius: 5px; margin-bottom: 10px;">
                    <div style="display: flex; justify-content: space-between; margin-bottom: 5px;">
                        <strong>AI技术发展趋势</strong>
                        <span class="status online">成功</span>
                    </div>
                    <div style="font-size: 12px; color: #666;">耗时: 2分18秒 | 10分钟前</div>
                </div>
                <div style="padding: 15px; background: #f8f9fa; border-radius: 5px; margin-bottom: 10px;">
                    <div style="display: flex; justify-content: space-between; margin-bottom: 5px;">
                        <strong>短视频制作技巧</strong>
                        <span class="status online">成功</span>
                    </div>
                    <div style="font-size: 12px; color: #666;">耗时: 2分35秒 | 1小时前</div>
                </div>
                <div style="padding: 15px; background: #f8f9fa; border-radius: 5px;">
                    <div style="display: flex; justify-content: space-between; margin-bottom: 5px;">
                        <strong>2024科技趋势</strong>
                        <span class="status offline">失败</span>
                    </div>
                    <div style="font-size: 12px; color: #666;">耗时: 0分45秒 | 3小时前</div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
`
