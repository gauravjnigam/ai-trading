package notify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// TelegramNotifier sends messages via the Telegram Bot API.
type TelegramNotifier struct {
	botToken string
	chatID   string
}

// NewTelegram returns a TelegramNotifier. Returns nil if botToken is empty.
func NewTelegram(botToken, chatID string) *TelegramNotifier {
	if botToken == "" {
		return nil
	}
	return &TelegramNotifier{botToken: botToken, chatID: chatID}
}

func (t *TelegramNotifier) Name() string { return "telegram" }

func (t *TelegramNotifier) Send(msg Message) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	params := url.Values{}
	params.Set("chat_id", t.chatID)
	params.Set("text", "```\n"+msg.Text+"\n```")
	params.Set("parse_mode", "Markdown")

	resp, err := http.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("telegram decode: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("telegram API error: %s", result.Description)
	}
	return nil
}
