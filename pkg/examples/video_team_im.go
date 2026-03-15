package examples

import (
	"context"
	"log"
	"os"

	"liteclaw/pkg/agent"
	"liteclaw/pkg/im"
	"liteclaw/pkg/integration"
	"liteclaw/pkg/provider"
	"liteclaw/pkg/team"
)

// VideoProductionTeamWithNotification creates a video production team with IM notifications
func VideoProductionTeamWithNotification() {
	// Initialize provider registry
	registry := provider.NewRegistry()
	
	// Register providers (assuming API keys are set in environment)
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		registry.Register("openai", provider.NewOpenAI(apiKey))
	}

	// Create team
	videoTeam := team.NewTeam(team.TeamConfig{
		ID:          "video-production-team",
		Name:        "短视频制作团队",
		Description: "AI员工团队，自动完成短视频制作全流程",
	}, registry)

	// Create AI employees
	director := createAgent("director", "运营总监", "负责选题策划、团队协调")
	researcher := createAgent("researcher", "素材研究员", "收集素材、数据分析")
	writer := createAgent("writer", "内容创作者", "撰写脚本、文案优化")
	narrator := createAgent("narrator", "配音师", "TTS配音、时间戳提取")
	editor := createAgent("editor", "视频剪辑师", "渲染视频、字幕特效")
	publisher := createAgent("publisher", "发布专员", "多平台发布、数据监控")

	// Add agents to team
	videoTeam.AddAgent(director)
	videoTeam.AddAgent(researcher)
	videoTeam.AddAgent(writer)
	videoTeam.AddAgent(narrator)
	videoTeam.AddAgent(editor)
	videoTeam.AddAgent(publisher)

	// Create IM clients
	feishuWebhook := os.Getenv("FEISHU_WEBHOOK")
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	notifier := integration.NewTeamNotifier(videoTeam)

	// Add Feishu notification
	if feishuWebhook != "" {
		feishuClient := im.NewIMClient(im.IMConfig{
			Type:       im.IMTypeFeishu,
			WebhookURL: feishuWebhook,
		})
		notifier.AddIMChannel("feishu", feishuClient)
		log.Println("✅ 飞书通知已启用")
	}

	// Add Telegram notification
	if telegramToken != "" {
		telegramClient := im.NewIMClient(im.IMConfig{
			Type:     im.IMTypeTelegram,
			BotToken: telegramToken,
		})
		notifier.AddIMChannel("telegram", telegramClient)
		log.Println("✅ Telegram通知已启用")
	}

	// Define workflow
	workflow := &team.WorkflowDefinition{
		ID:          "video-production-workflow",
		Name:        "短视频制作流程",
		Description: "从选题到发布的完整自动化流程",
		Steps: []team.WorkflowStep{
			{
				ID:      "step1",
				Name:    "选题推送",
				AgentID: "director",
				Task:    "分析今日热点，推送5个选题建议",
			},
			{
				ID:        "step2",
				Name:      "素材收集",
				AgentID:   "researcher",
				Task:      "根据选题收集相关素材和数据",
				DependsOn: []string{"step1"},
			},
			{
				ID:        "step3",
				Name:      "脚本创作",
				AgentID:   "writer",
				Task:      "撰写60秒短视频脚本，包含旁白文案",
				DependsOn: []string{"step2"},
			},
			{
				ID:        "step4",
				Name:      "配音生成",
				AgentID:   "narrator",
				Task:      "使用TTS生成配音，提取时间戳",
				DependsOn: []string{"step3"},
			},
			{
				ID:        "step5",
				Name:      "视频渲染",
				AgentID:   "editor",
				Task:      "渲染视频，添加字幕和特效",
				DependsOn: []string{"step4"},
			},
			{
				ID:        "step6",
				Name:      "自动发布",
				AgentID:   "publisher",
				Task:      "发布到抖音、快手、视频号、小红书",
				DependsOn: []string{"step5"},
			},
		},
	}

	// Register workflow
	videoTeam.GetWorkflowEngine().RegisterWorkflow(workflow)

	// Schedule daily video production at 9:30 AM
	videoTeam.ScheduleTask("director", "开始今日短视频制作", "0 30 9 * * *")

	// Start team
	videoTeam.GetScheduler().Start()

	log.Printf("🎬 AI员工团队 [%s] 已启动！", videoTeam.GetName())
	log.Printf("👥 团队成员: %v", videoTeam.ListAgents())
	log.Printf("⏰ 每日9:30自动执行视频制作流程")

	// Execute workflow with notifications
	ctx := context.Background()
	
	// Notify start
	notifier.NotifyWorkflowStart(ctx, "feishu", "video-production-workflow")

	result, err := videoTeam.ExecuteWorkflow(ctx, "video-production-workflow", map[string]interface{}{
		"topic": "AI技术发展趋势",
		"style": "科普向",
	})

	if err != nil {
		log.Printf("工作流执行失败: %v", err)
		notifier.NotifyWorkflowError(ctx, "feishu", "video-production-workflow", err)
		return
	}

	// Notify completion
	var stepNames []string
	for _, step := range result.Steps {
		stepNames = append(stepNames, fmt.Sprintf("✅ %s (%s)", step.StepID, step.AgentID))
	}
	notifier.NotifyWorkflowComplete(ctx, "feishu", "video-production-workflow", stepNames)

	log.Printf("✅ 工作流执行完成！状态: %s, 耗时: %v", result.Status, result.Duration)
}

func createAgent(id, name, desc string) *agent.Agent {
	return agent.NewAgent(
		id,
		name,
		desc,
		agent.WithProvider("openai"),
		agent.WithModel("gpt-4"),
	)
}
