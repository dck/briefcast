package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Specific error types for different failure modes
var (
	ErrRateLimit        = fmt.Errorf("groq: rate limited")
	ErrModelUnavailable = fmt.Errorf("groq: model not available")
)

type Client struct {
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 10 * time.Minute},
	}
}

// TranscribeResult holds the transcription response
type TranscribeResult struct {
	Text string
}

// Transcribe sends an audio file to Groq Whisper API and returns the transcript.
// audioPath: path to the audio file on disk.
// model: whisper model name (e.g., "whisper-large-v3")
func (c *Client) Transcribe(ctx context.Context, audioPath, model string) (*TranscribeResult, error) {
	// Open the audio file
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add file field
	fileField, err := writer.CreateFormFile("file", audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(fileField, file); err != nil {
		return nil, fmt.Errorf("failed to copy file to form: %w", err)
	}

	// Add model field
	if err := writer.WriteField("model", model); err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}

	// Add response_format field
	if err := writer.WriteField("response_format", "text"); err != nil {
		return nil, fmt.Errorf("failed to write response_format field: %w", err)
	}

	// Close the writer to finalize multipart body
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/audio/transcriptions", &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, context.DeadlineExceeded
		}
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, c.handleErrorResponse(resp.StatusCode, string(respBody), resp.Header)
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Response is plain text
	return &TranscribeResult{
		Text: string(respBody),
	}, nil
}

// SummarizeResult holds the summarization response
type SummarizeResult struct {
	Text       string
	TokensUsed int
}

// Summarize sends text to Groq Llama API for summarization.
// model: LLM model name (e.g., "llama-3.3-70b-versatile")
func (c *Client) Summarize(ctx context.Context, prompt, model string) (*SummarizeResult, error) {
	// Create request body
	requestBody := map[string]interface{}{
		"model":       model,
		"temperature": 0.3,
		"max_tokens":  8192,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, context.DeadlineExceeded
		}
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, c.handleErrorResponse(resp.StatusCode, string(respBody), resp.Header)
	}

	// Parse response
	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &SummarizeResult{
		Text:       chatResp.Choices[0].Message.Content,
		TokensUsed: chatResp.Usage.TotalTokens,
	}, nil
}

// handleErrorResponse processes HTTP error responses
func (c *Client) handleErrorResponse(statusCode int, body string, headers http.Header) error {
	// Handle rate limit (429)
	if statusCode == http.StatusTooManyRequests {
		retryAfter := time.Duration(0)
		if retryAfterHeader := headers.Get("Retry-After"); retryAfterHeader != "" {
			if seconds, err := strconv.Atoi(retryAfterHeader); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		return &RateLimitError{RetryAfter: retryAfter}
	}

	// Handle model not available (404)
	if statusCode == http.StatusNotFound {
		return ErrModelUnavailable
	}

	// Handle server errors
	if statusCode >= 500 {
		return fmt.Errorf("groq server error: status %d", statusCode)
	}

	// Try to parse error response for specific error messages
	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &errResp); err == nil {
		if strings.Contains(strings.ToLower(errResp.Error.Message), "model") ||
			strings.Contains(strings.ToLower(errResp.Error.Message), "not available") {
			return ErrModelUnavailable
		}
		return fmt.Errorf("groq api error: %s", errResp.Error.Message)
	}

	return fmt.Errorf("groq api error: status %d", statusCode)
}

// RateLimitError represents a rate limit error with optional retry-after information
type RateLimitError struct {
	RetryAfter time.Duration
}

// Error implements the error interface
func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("groq: rate limited (retry after %v)", e.RetryAfter)
	}
	return "groq: rate limited"
}

// Is implements error matching for use with errors.Is()
func (e *RateLimitError) Is(target error) bool {
	return target == ErrRateLimit
}
