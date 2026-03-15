package im

import "context"

// IMClient defines the interface for IM (Instant Messaging) clients
type IMClient interface {
	// SendMessage sends a generic message
	SendMessage(ctx context.Context, chatID string, message string) error
	
	// SendWorkflowNotification sends workflow execution notification
	SendWorkflowNotification(ctx context.Context, chatID string, workflowID string, status string, steps []string) error
	
	// SendAgentMessage sends a message from an agent
	SendAgentMessage(ctx context.Context, chatID string, agentName string, message string) error
}

// IMType represents the type of IM platform
type IMType string

const (
	IMTypeFeishu   IMType = "feishu"
	IMTypeTelegram IMType = "telegram"
	IMTypeDingTalk IMType = "dingtalk"
	IMTypeWeChat   IMType = "wechat"
)

// IMConfig represents IM configuration
type IMConfig struct {
	Type     IMType
	WebhookURL string
	BotToken   string
	AppID      string
	AppSecret  string
}

// NewIMClient creates a new IM client based on type
func NewIMClient(config IMConfig) IMClient {
	switch config.Type {
	case IMTypeFeishu:
		return NewFeishuClient(FeishuConfig{
			WebhookURL: config.WebhookURL,
			AppID:      config.AppID,
			AppSecret:  config.AppSecret,
		})
	case IMTypeTelegram:
		return NewTelegramClient(TelegramConfig{
			BotToken: config.BotToken,
		})
	default:
		return nil
	}
}
