package im

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TelegramClient represents a Telegram bot client
type TelegramClient struct {
	botToken   string
	httpClient *http.Client
	baseURL    string
}

// TelegramConfig represents Telegram bot configuration
type TelegramConfig struct {
	BotToken string
}

// TelegramMessage represents a Telegram message
type TelegramMessage struct {
	ChatID                string                 `json:"chat_id"`
	Text                  string                 `json:"text,omitempty"`
	ParseMode             string                 `json:"parse_mode,omitempty"`
	DisableNotification   bool                   `json:"disable_notification,omitempty"`
	ReplyToMessageID      int                    `json:"reply_to_message_id,omitempty"`
	ReplyMarkup           map[string]interface{} `json:"reply_markup,omitempty"`
}

// TelegramInlineKeyboard represents an inline keyboard
type TelegramInlineKeyboard struct {
	InlineKeyboard [][]TelegramInlineKeyboardButton `json:"inline_keyboard"`
}

// TelegramInlineKeyboardButton represents an inline keyboard button
type TelegramInlineKeyboardButton struct {
	Text         string `json:"text"`
	URL          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

// TelegramUpdate represents a Telegram update
type TelegramUpdate struct {
	UpdateID int                 `json:"update_id"`
	Message  *TelegramBotMessage `json:"message,omitempty"`
}

// TelegramBotMessage represents a bot message
type TelegramBotMessage struct {
	MessageID int             `json:"message_id"`
	From      *TelegramUser   `json:"from"`
	Chat      *TelegramChat   `json:"chat"`
	Date      int64           `json:"date"`
	Text      string          `json:"text"`
}

// TelegramUser represents a Telegram user
type TelegramUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// TelegramChat represents a Telegram chat
type TelegramChat struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title,omitempty"`
	Username string `json:"username,omitempty"`
}

// NewTelegramClient creates a new Telegram client
func NewTelegramClient(config TelegramConfig) *TelegramClient {
	return &TelegramClient{
		botToken: config.BotToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.telegram.org",
	}
}

// SendMessage sends a text message
func (c *TelegramClient) SendMessage(ctx context.Context, msg *TelegramMessage) error {
	return c.apiCall(ctx, "sendMessage", msg, nil)
}

// SendText sends a simple text message
func (c *TelegramClient) SendText(ctx context.Context, chatID, text string) error {
	return c.SendMessage(ctx, &TelegramMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	})
}

// SendHTML sends an HTML-formatted message
func (c *TelegramClient) SendHTML(ctx context.Context, chatID, html string) error {
	return c.SendMessage(ctx, &TelegramMessage{
		ChatID:    chatID,
		Text:      html,
		ParseMode: "HTML",
	})
}

// SendMarkdown sends a Markdown-formatted message
func (c *TelegramClient) SendMarkdown(ctx context.Context, chatID, markdown string) error {
	return c.SendMessage(ctx, &TelegramMessage{
		ChatID:    chatID,
		Text:      markdown,
		ParseMode: "MarkdownV2",
	})
}

// SendInlineKeyboard sends a message with inline keyboard
func (c *TelegramClient) SendInlineKeyboard(ctx context.Context, chatID, text string, keyboard *TelegramInlineKeyboard) error {
	return c.SendMessage(ctx, &TelegramMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
		ReplyMarkup: map[string]interface{}{
			"inline_keyboard": keyboard.InlineKeyboard,
		},
	})
}

// SendWorkflowNotification sends a workflow execution notification
func (c *TelegramClient) SendWorkflowNotification(ctx context.Context, chatID, workflowID, status string, steps []string) error {
	var text = fmt.Sprintf(
		"🔔 *工作流执行通知*\n\n"+
			"📋 **ID**: `%s`\n"+
			"📊 **状态**: %s\n\n"+
			"**执行步骤**:\n",
		workflowID,
		status,
	)

	for i, step := range steps {
		text += fmt.Sprintf("%d. %s\n", i+1, step)
	}

	return c.SendText(ctx, chatID, text)
}

// SendAgentMessage sends a message from an agent
func (c *TelegramClient) SendAgentMessage(ctx context.Context, chatID, agentName, message string) error {
	text := fmt.Sprintf("🤖 *%s*\n\n%s", agentName, message)
	return c.SendText(ctx, chatID, text)
}

// GetUpdates gets updates from Telegram
func (c *TelegramClient) GetUpdates(ctx context.Context, offset int, timeout int) ([]TelegramUpdate, error) {
	params := map[string]interface{}{
		"offset":  offset,
		"timeout": timeout,
	}

	var result struct {
		OK     bool             `json:"ok"`
		Result []TelegramUpdate `json:"result"`
	}

	if err := c.apiCall(ctx, "getUpdates", params, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("get updates failed")
	}

	return result.Result, nil
}

// apiCall makes an API call to Telegram
func (c *TelegramClient) apiCall(ctx context.Context, method string, params interface{}, result interface{}) error {
	url := fmt.Sprintf("%s/bot%s/%s", c.baseURL, c.botToken, method)

	bodyBytes, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
	} else {
		var apiResult struct {
			OK          bool   `json:"ok"`
			ErrorCode   int    `json:"error_code,omitempty"`
			Description string `json:"description,omitempty"`
		}
		if err := json.Unmarshal(respBody, &apiResult); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
		if !apiResult.OK {
			return fmt.Errorf("telegram API error: code=%d, description=%s", apiResult.ErrorCode, apiResult.Description)
		}
	}

	return nil
}
