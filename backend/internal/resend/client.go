package resend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	APIKey     string
	FromEmail  string
	HTTPClient *http.Client
}

func NewClient(apiKey, fromEmail string) *Client {
	return &Client{
		APIKey:     apiKey,
		FromEmail:  fromEmail,
		HTTPClient: &http.Client{},
	}
}

// SendEmail sends an HTML email via the Resend API.
func (c *Client) SendEmail(to, subject, htmlBody string) error {
	// Prepare request body
	body := map[string]interface{}{
		"from":    c.FromEmail,
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resend API error: status code %d", resp.StatusCode)
	}

	return nil
}
