package integration

import (
	"context"
	"fmt"

	"liteclaw/pkg/im"
	"liteclaw/pkg/team"
)

// TeamNotifier provides IM notification for teams
type TeamNotifier struct {
	team    *team.Team
	clients map[string]im.IMClient // channel -> client
}

// NewTeamNotifier creates a new team notifier
func NewTeamNotifier(t *team.Team) *TeamNotifier {
	return &TeamNotifier{
		team:    t,
		clients: make(map[string]im.IMClient),
	}
}

// AddIMChannel adds an IM channel
func (n *TeamNotifier) AddIMChannel(channelID string, client im.IMClient) {
	n.clients[channelID] = client
}

// NotifyWorkflowStart notifies workflow start
func (n *TeamNotifier) NotifyWorkflowStart(ctx context.Context, channelID string, workflowID string) error {
	client, ok := n.clients[channelID]
	if !ok {
		return fmt.Errorf("channel not found: %s", channelID)
	}

	return client.SendWorkflowNotification(ctx, channelID, workflowID, "started", []string{
		"🚀 工作流开始执行",
	})
}

// NotifyWorkflowComplete notifies workflow completion
func (n *TeamNotifier) NotifyWorkflowComplete(ctx context.Context, channelID string, workflowID string, steps []string) error {
	client, ok := n.clients[channelID]
	if !ok {
		return fmt.Errorf("channel not found: %s", channelID)
	}

	allSteps := append([]string{"✅ 工作流执行完成"}, steps...)
	return client.SendWorkflowNotification(ctx, channelID, workflowID, "completed", allSteps)
}

// NotifyWorkflowError notifies workflow error
func (n *TeamNotifier) NotifyWorkflowError(ctx context.Context, channelID string, workflowID string, err error) error {
	client, ok := n.clients[channelID]
	if !ok {
		return fmt.Errorf("channel not found: %s", channelID)
	}

	return client.SendWorkflowNotification(ctx, channelID, workflowID, "failed", []string{
		"❌ 工作流执行失败",
		fmt.Sprintf("错误: %v", err),
	})
}

// NotifyAgentMessage notifies a message from an agent
func (n *TeamNotifier) NotifyAgentMessage(ctx context.Context, channelID string, agentName string, message string) error {
	client, ok := n.clients[channelID]
	if !ok {
		return fmt.Errorf("channel not found: %s", channelID)
	}

	return client.SendAgentMessage(ctx, channelID, agentName, message)
}

// BroadcastToAll sends a message to all channels
func (n *TeamNotifier) BroadcastToAll(ctx context.Context, message string) error {
	for channelID, client := range n.clients {
		if err := client.SendMessage(ctx, channelID, message); err != nil {
			return fmt.Errorf("send to channel %s: %w", channelID, err)
		}
	}
	return nil
}
