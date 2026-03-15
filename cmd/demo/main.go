package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"liteclaw/pkg/agent"
	"liteclaw/pkg/api"
	"liteclaw/pkg/im"
	"liteclaw/pkg/integration"
	"liteclaw/pkg/provider"
	"liteclaw/pkg/team"
)

func main() {
	log.Println("╔══════════════════════════════════════════════════════════════╗")
	log.Println("║                                                              ║")
	log.Println("║     🚀 LiteClaw AI员工系统 v2.0 🚀                         ║")
	log.Println("║                                                              ║")
	log.Println("╚══════════════════════════════════════════════════════════════╝")
	log.Println()

	// 1. 初始化Provider注册表
	registry := provider.NewRegistry()
	
	// 注册OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		registry.Register("openai", provider.NewOpenAI(apiKey))
		log.Println("✅ OpenAI Provider 已注册")
	}

	// 2. 创建AI员工团队
	videoTeam := team.NewTeam(team.TeamConfig{
		ID:          "video-production-team",
		Name:        "短视频制作团队",
		Description: "AI员工团队，自动完成短视频制作全流程",
	}, registry)

	// 3. 创建AI员工
	agents := []*agent.Agent{
		createAgent("director", "运营总监", "负责选题策划、团队协调", "openai"),
		createAgent("researcher", "素材研究员", "收集素材、数据分析", "openai"),
		createAgent("writer", "内容创作者", "撰写脚本、文案优化", "openai"),
		createAgent("narrator", "配音师", "TTS配音、时间戳提取", "openai"),
		createAgent("editor", "视频剪辑师", "渲染视频、字幕特效", "openai"),
		createAgent("publisher", "发布专员", "多平台发布、数据监控", "openai"),
	}

	// 添加到团队
	for _, a := range agents {
		if err := videoTeam.AddAgent(a); err != nil {
			log.Printf("❌ 添加Agent失败: %v", err)
		}
	}
	log.Printf("✅ 已添加 %d 个AI员工", len(agents))

	// 4. 定义工作流
	workflow := &team.WorkflowDefinition{
		ID:          "video-production-workflow",
		Name:        "短视频制作流程",
		Description: "从选题到发布的完整自动化流程",
		Steps: []team.WorkflowStep{
			{ID: "step1", Name: "选题推送", AgentID: "director", Task: "分析今日热点，推送5个选题建议"},
			{ID: "step2", Name: "素材收集", AgentID: "researcher", Task: "根据选题收集相关素材和数据", DependsOn: []string{"step1"}},
			{ID: "step3", Name: "脚本创作", AgentID: "writer", Task: "撰写60秒短视频脚本，包含旁白文案", DependsOn: []string{"step2"}},
			{ID: "step4", Name: "配音生成", AgentID: "narrator", Task: "使用TTS生成配音，提取时间戳", DependsOn: []string{"step3"}},
			{ID: "step5", Name: "视频渲染", AgentID: "editor", Task: "渲染视频，添加字幕和特效", DependsOn: []string{"step4"}},
			{ID: "step6", Name: "自动发布", AgentID: "publisher", Task: "发布到抖音、快手、视频号、小红书", DependsOn: []string{"step5"}},
		},
	}

	videoTeam.GetWorkflowEngine().RegisterWorkflow(workflow)
	log.Println("✅ 工作流已注册")

	// 5. 设置IM通知
	notifier := integration.NewTeamNotifier(videoTeam)
	
	// 飞书通知
	if webhook := os.Getenv("FEISHU_WEBHOOK"); webhook != "" {
		feishuClient := im.NewIMClient(im.IMConfig{
			Type:       im.IMTypeFeishu,
			WebhookURL: webhook,
		})
		notifier.AddIMChannel("feishu", feishuClient)
		log.Println("✅ 飞书通知已启用")
	}

	// Telegram通知
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		telegramClient := im.NewIMClient(im.IMConfig{
			Type:     im.IMTypeTelegram,
			BotToken: token,
		})
		notifier.AddIMChannel("telegram", telegramClient)
		log.Println("✅ Telegram通知已启用")
	}

	// 6. 设置定时任务
	videoTeam.ScheduleTask("director", "开始今日短视频制作", "0 30 9 * * *")
	log.Println("✅ 定时任务已设置 (每日9:30执行)")

	// 7. 启动团队
	videoTeam.GetScheduler().Start()
	log.Println("✅ 团队已启动")

	// 8. 创建Web服务器
	server := api.NewServer()
	server.AddTeam(videoTeam)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 9. 启动HTTP服务器
	go func() {
		log.Printf("🌐 Web服务器启动: http://localhost:%s", port)
		if err := http.ListenAndServe(":"+port, server); err != nil {
			log.Printf("❌ HTTP服务器错误: %v", err)
		}
	}()

	// 10. 演示工作流执行
	log.Println()
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("📊 系统状态")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("团队名称: %s", videoTeam.GetName())
	log.Printf("团队ID: %s", videoTeam.GetID())
	log.Printf("AI员工: %v", videoTeam.ListAgents())
	log.Println()

	// 演示执行工作流
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🎬 开始演示工作流执行")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	ctx := context.Background()
	
	// 发送开始通知
	notifier.NotifyWorkflowStart(ctx, "feishu", "video-production-workflow")

	result, err := videoTeam.ExecuteWorkflow(ctx, "video-production-workflow", map[string]interface{}{
		"topic": "AI技术发展趋势",
		"style": "科普向",
	})

	if err != nil {
		log.Printf("❌ 工作流执行失败: %v", err)
		notifier.NotifyWorkflowError(ctx, "feishu", "video-production-workflow", err)
	} else {
		log.Printf("✅ 工作流执行成功！状态: %s, 耗时: %v", result.Status, result.Duration)
		
		// 发送完成通知
		var stepNames []string
		for _, step := range result.Steps {
			stepNames = append(stepNames, fmt.Sprintf("✅ %s", step.StepID))
		}
		notifier.NotifyWorkflowComplete(ctx, "feishu", "video-production-workflow", stepNames)
	}

	log.Println()
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🎊 系统运行中...")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("💡 访问 http://localhost:8080 查看Dashboard")
	log.Println("💡 按Ctrl+C停止服务")
	log.Println()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println()
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("👋 正在关闭服务...")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	videoTeam.GetScheduler().Stop()
	log.Println("✅ 服务已停止")
}

func createAgent(id, name, desc, providerName string) *agent.Agent {
	return agent.NewAgent(
		id,
		name,
		desc,
		agent.WithProvider(providerName),
		agent.WithModel("gpt-4"),
	)
}
