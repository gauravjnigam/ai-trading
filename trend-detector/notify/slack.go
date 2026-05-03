package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackNotifier posts messages to a Slack incoming webhook.
type SlackNotifier struct {
	webhookURL string
	channel    string
}

// NewSlack returns a SlackNotifier. Returns nil if webhookURL is empty.
func NewSlack(webhookURL, channel string) *SlackNotifier {
	if webhookURL == "" {
		return nil
	}
	return &SlackNotifier{webhookURL: webhookURL, channel: channel}
}

func (s *SlackNotifier) Name() string { return "slack" }

func (s *SlackNotifier) Send(msg Message) error {
	payload := map[string]any{
		"channel": s.channel,
		"text":    "```\n" + msg.Text + "\n```",
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack send: unexpected status %d", resp.StatusCode)
	}
	return nil
}
