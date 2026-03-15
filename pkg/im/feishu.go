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

// FeishuClient represents a Feishu (Lark) bot client
type FeishuClient struct {
	webhookURL string
	appID      string
	appSecret  string
	httpClient *http.Client
	accessToken string
	tokenExpire time.Time
}

// FeishuConfig represents Feishu bot configuration
type FeishuConfig struct {
	WebhookURL string // Custom robot webhook
	AppID      string // App ID for API access
	AppSecret  string // App Secret for API access
}

// FeishuMessage represents a Feishu message
type FeishuMessage struct {
	MsgType string                 `json:"msg_type"`
	Content map[string]interface{} `json:"content"`
}

// FeishuTextContent represents text content
type FeishuTextContent struct {
	Text string `json:"text"`
}

// FeishuPostContent represents rich text content
type FeishuPostContent struct {
	Post FeishuPostBody `json:"post"`
}

// FeishuPostBody represents post body
type FeishuPostBody struct {
	ZhCN FeishuPostLanguage `json:"zh_cn"`
}

// FeishuPostLanguage represents language-specific content
type FeishuPostLanguage struct {
	Title   string            `json:"title"`
	Content [][]FeishuPostTag `json:"content"`
}

// FeishuPostTag represents a post tag
type FeishuPostTag struct {
	Tag   string `json:"tag"`
	Text  string `json:"text,omitempty"`
	Href  string `json:"href,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

// FeishuCard represents an interactive card
type FeishuCard struct {
	Config   FeishuCardConfig   `json:"config,omitempty"`
	Elements []FeishuCardElement `json:"elements"`
}

// FeishuCardConfig represents card config
type FeishuCardConfig struct {
	WideScreenMode bool `json:"wide_screen_mode"`
	EnableForward  bool `json:"enable_forward"`
}

// FeishuCardElement represents a card element
type FeishuCardElement struct {
	Tag    string                 `json:"tag"`
	Text   *FeishuCardText        `json:"text,omitempty"`
	Actions []FeishuCardAction    `json:"actions,omitempty"`
	Fields []FeishuCardField      `json:"fields,omitempty"`
	Extra  map[string]interface{} `json:"extra,omitempty"`
}

// FeishuCardText represents card text
type FeishuCardText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// FeishuCardField represents a card field
type FeishuCardField struct {
	IsShort bool            `json:"is_short"`
	Text    FeishuCardText  `json:"text"`
}

// FeishuCardAction represents a card action
type FeishuCardAction struct {
	Tag    string            `json:"tag"`
	Text   FeishuCardText    `json:"text"`
	URL    string            `json:"url,omitempty"`
	Type   string            `json:"type,omitempty"`
	Value  map[string]string `json:"value,omitempty"`
}

// NewFeishuClient creates a new Feishu client
func NewFeishuClient(config FeishuConfig) *FeishuClient {
	return &FeishuClient{
		webhookURL: config.WebhookURL,
		appID:      config.AppID,
		appSecret:  config.AppSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendText sends a text message
func (c *FeishuClient) SendText(ctx context.Context, text string) error {
	msg := FeishuMessage{
		MsgType: "text",
		Content: map[string]interface{}{
			"text": text,
		},
	}
	return c.sendWebhook(ctx, msg)
}

// SendPost sends a rich text message
func (c *FeishuClient) SendPost(ctx context.Context, title string, content [][]FeishuPostTag) error {
	msg := FeishuMessage{
		MsgType: "post",
		Content: map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title":   title,
					"content": content,
				},
			},
		},
	}
	return c.sendWebhook(ctx, msg)
}

// SendCard sends an interactive card
func (c *FeishuClient) SendCard(ctx context.Context, card *FeishuCard) error {
	msg := FeishuMessage{
		MsgType: "interactive",
		Content: map[string]interface{}{
			"card": card,
		},
	}
	return c.sendWebhook(ctx, msg)
}

// SendWorkflowNotification sends a workflow execution notification
func (c *FeishuClient) SendWorkflowNotification(ctx context.Context, workflowID, status string, steps []string) error {
	// Build card elements
	var elements []FeishuCardElement

	// Title
	elements = append(elements, FeishuCardElement{
		Tag: "div",
		Text: &FeishuCardText{
			Tag:     "lark_md",
			Content: fmt.Sprintf("**工作流执行通知**\n**ID**: %s\n**状态**: %s", workflowID, status),
		},
	})

	// Steps
	var stepFields []FeishuCardField
	for _, step := range steps {
		stepFields = append(stepFields, FeishuCardField{
			IsShort: false,
			Text: FeishuCardText{
				Tag:     "lark_md",
				Content: step,
			},
		})
	}

	if len(stepFields) > 0 {
		elements = append(elements, FeishuCardElement{
			Tag:    "div",
			Fields: stepFields,
		})
	}

	card := &FeishuCard{
		Config: FeishuCardConfig{
			WideScreenMode: true,
			EnableForward:  true,
		},
		Elements: elements,
	}

	return c.SendCard(ctx, card)
}

// SendAgentMessage sends a message from an agent
func (c *FeishuClient) SendAgentMessage(ctx context.Context, agentName, message string) error {
	return c.SendText(ctx, fmt.Sprintf("🤖 **%s**: %s", agentName, message))
}

// sendWebhook sends a message via webhook
func (c *FeishuClient) sendWebhook(ctx context.Context, msg FeishuMessage) error {
	if c.webhookURL == "" {
		return fmt.Errorf("webhook URL is not configured")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.webhookURL, bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if code, ok := result["code"].(float64); ok && code != 0 {
		return fmt.Errorf("feishu API error: code=%v, msg=%v", code, result["msg"])
	}

	return nil
}

// getAccessToken gets tenant access token (for API access)
func (c *FeishuClient) getAccessToken(ctx context.Context) (string, error) {
	if c.accessToken != "" && time.Now().Before(c.tokenExpire) {
		return c.accessToken, nil
	}

	url := "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
	body := map[string]string{
		"app_id":     c.appID,
		"app_secret": c.appSecret,
	}

	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code          int    `json:"code"`
		Msg           string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire        int    `json:"expire"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", fmt.Errorf("get token failed: %s", result.Msg)
	}

	c.accessToken = result.TenantAccessToken
	c.tokenExpire = time.Now().Add(time.Duration(result.Expire-300) * time.Second)

	return c.accessToken, nil
}
