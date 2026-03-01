package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const telegramAPIBaseURL = "https://api.telegram.org"

type Client struct {
	BotToken   string
	HTTPClient *http.Client
}

// NewClient creates a new Telegram Bot API client with the given bot token.
func NewClient(botToken string) *Client {
	return &Client{
		BotToken:   botToken,
		HTTPClient: &http.Client{},
	}
}

// sendMessageRequest represents the payload for the sendMessage API call.
type sendMessageRequest struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

// sendMessageResponse represents the API response from Telegram.
type sendMessageResponse struct {
	OK    bool                   `json:"ok"`
	Error string                 `json:"description,omitempty"`
	Result map[string]interface{} `json:"result,omitempty"`
}

// SendMessage sends a text message to a Telegram chat using HTML parse mode.
func (c *Client) SendMessage(chatID, text string) error {
	return c.sendMessage(chatID, text, "HTML")
}

// SendMessageMarkdown sends a text message to a Telegram chat using MarkdownV2 parse mode.
func (c *Client) SendMessageMarkdown(chatID, text string) error {
	return c.sendMessage(chatID, text, "MarkdownV2")
}

// sendMessage is the internal method that handles sending messages with the specified parse mode.
func (c *Client) sendMessage(chatID, text, parseMode string) error {
	payload := sendMessageRequest{
		ChatID:                chatID,
		Text:                  text,
		ParseMode:             parseMode,
		DisableWebPagePreview: true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", telegramAPIBaseURL, c.BotToken)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status code %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp sendMessageResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram API error: %s", apiResp.Error)
	}

	return nil
}
