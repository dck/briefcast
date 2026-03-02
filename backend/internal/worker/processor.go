package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dck/briefcast/internal/groq"
	"github.com/dck/briefcast/internal/settings"
	"github.com/dck/briefcast/internal/telegram"
)

const summaryPrompt = `You are an expert podcast summarizer. Create a comprehensive, structured article-style summary of the following podcast episode transcript.

Episode: %s
Podcast: %s

Requirements:
1. **Episode Overview** — 2-3 sentence description of what the episode covers
2. **Topic Breakdown** — each major topic discussed, in order, with an h2 heading and a full paragraph per topic
3. **Key Opinions & Takes** — notable positions expressed, attributed by speaker name where identifiable from context
4. **Concrete References** — tools, books, companies, links, and other resources explicitly mentioned

Important:
- Write in the SAME LANGUAGE as the transcript (do not translate)
- Use Markdown formatting (## for headings, **bold** for emphasis, - for lists)
- Be comprehensive — this replaces listening to the full episode
- Attribute opinions to specific speakers when you can identify them from context

Show notes for additional context:
%s

Transcript:
%s`

const mergePrompt = `Merge the following partial summaries of a podcast episode into one coherent structured article. Remove any duplicate content and ensure smooth transitions between sections.

Requirements:
- Maintain the same structure: Episode Overview, Topic Breakdown, Key Opinions & Takes, Concrete References
- Write in the SAME LANGUAGE as the summaries
- Use Markdown formatting

Partial summaries:
%s`

type Processor struct {
	DB             *sql.DB
	GroqClient     *groq.Client
	TelegramClient *telegram.Client
	AdminChatID    string
	AudioTmpDir    string
}

func NewProcessor(db *sql.DB, groqClient *groq.Client, tgClient *telegram.Client, adminChatID, audioTmpDir string) *Processor {
	return &Processor{
		DB:             db,
		GroqClient:     groqClient,
		TelegramClient: tgClient,
		AdminChatID:    adminChatID,
		AudioTmpDir:    audioTmpDir,
	}
}

// ProcessPending picks up pending/processing episodes and processes them through the pipeline.
// Called periodically by the worker scheduler.
func (p *Processor) ProcessPending() error {
	paused, err := settings.GetBool(p.DB, "processing_paused")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check processing_paused: %w", err)
	}
	if paused {
		return nil
	}

	var (
		episodeID   int
		podcastID   int
		title       string
		audioURL    string
		showNotes   sql.NullString
		currentStep string
		retryCount  int
	)
	err = p.DB.QueryRow(`
		SELECT id, podcast_id, title, audio_url, show_notes, current_step, retry_count
		FROM episodes
		WHERE status IN ('pending', 'processing')
		ORDER BY created_at ASC
		LIMIT 1`,
	).Scan(&episodeID, &podcastID, &title, &audioURL, &showNotes, &currentStep, &retryCount)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("query pending episode: %w", err)
	}

	_, err = p.DB.Exec(`UPDATE episodes SET status = 'processing' WHERE id = ?`, episodeID)
	if err != nil {
		return fmt.Errorf("set processing status: %w", err)
	}

	// Fetch podcast name for the summary prompt.
	var podcastName string
	_ = p.DB.QueryRow(`SELECT name FROM podcasts WHERE id = ?`, podcastID).Scan(&podcastName)

	switch currentStep {
	case "", "download":
		if err := p.stepDownload(episodeID, audioURL); err != nil {
			p.handleFailure(episodeID, "download", retryCount, err)
			return nil
		}
		return p.stepTranscribe(episodeID, title, podcastName, showNotes.String, retryCount)
	case "transcribe":
		return p.stepTranscribe(episodeID, title, podcastName, showNotes.String, retryCount)
	case "summarize":
		return p.stepSummarize(episodeID, title, podcastName, showNotes.String, retryCount)
	default:
		return nil
	}
}

// stepDownload downloads the episode audio file to the temp directory.
func (p *Processor) stepDownload(episodeID int, audioURL string) error {
	start := time.Now()

	resp, err := http.Get(audioURL)
	if err != nil {
		p.logStep(episodeID, "download", "error", err.Error(), time.Since(start).Milliseconds())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		p.logStep(episodeID, "download", "error", "audio URL returned 404", time.Since(start).Milliseconds())
		// Permanent failure — mark as failed directly.
		_, _ = p.DB.Exec(`UPDATE episodes SET status = 'failed', last_error = ? WHERE id = ?`,
			"audio URL returned 404", episodeID)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unexpected HTTP status %d", resp.StatusCode)
		p.logStep(episodeID, "download", "error", msg, time.Since(start).Milliseconds())
		return fmt.Errorf("%s", msg)
	}

	audioPath := filepath.Join(p.AudioTmpDir, fmt.Sprintf("%d.mp3", episodeID))
	f, err := os.Create(audioPath)
	if err != nil {
		p.logStep(episodeID, "download", "error", err.Error(), time.Since(start).Milliseconds())
		return err
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		p.logStep(episodeID, "download", "error", err.Error(), time.Since(start).Milliseconds())
		return err
	}

	maxSizeMB, err := settings.GetInt(p.DB, "audio_max_size_mb")
	if err == nil && maxSizeMB > 0 {
		maxBytes := int64(maxSizeMB) * 1024 * 1024
		if written > maxBytes {
			msg := fmt.Sprintf("audio file too large: %d bytes (max %d MB)", written, maxSizeMB)
			p.logStep(episodeID, "download", "error", msg, time.Since(start).Milliseconds())
			os.Remove(audioPath)
			_, _ = p.DB.Exec(`UPDATE episodes SET status = 'skipped', skip_reason = ? WHERE id = ?`, msg, episodeID)
			return nil
		}
	}

	p.logStep(episodeID, "download", "ok", fmt.Sprintf("downloaded %d bytes", written), time.Since(start).Milliseconds())
	_, _ = p.DB.Exec(`UPDATE episodes SET current_step = 'transcribe' WHERE id = ?`, episodeID)
	return nil
}

// stepTranscribe transcribes the episode audio using the Groq Whisper API.
func (p *Processor) stepTranscribe(episodeID int, title, podcastName, showNotes string, retryCount int) error {
	audioPath := filepath.Join(p.AudioTmpDir, fmt.Sprintf("%d.mp3", episodeID))
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		// Audio file missing — need to re-download.
		_, _ = p.DB.Exec(`UPDATE episodes SET current_step = 'download' WHERE id = ?`, episodeID)
		return nil
	}

	whisperModel, err := settings.Get(p.DB, "groq_whisper_model")
	if err != nil || whisperModel == "" {
		whisperModel = "whisper-large-v3"
	}

	start := time.Now()
	ctx := context.Background()
	result, err := p.GroqClient.Transcribe(ctx, audioPath, whisperModel)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		var rlErr *groq.RateLimitError
		if errors.As(err, &rlErr) {
			p.logStep(episodeID, "transcribe", "rate_limited", err.Error(), elapsed)
			p.logAPI("groq", episodeID, "rate_limited", elapsed, 0, err.Error())
			log.Printf("episode %d: rate limited, will retry next cycle", episodeID)
			return nil
		}
		if errors.Is(err, groq.ErrModelUnavailable) {
			p.logStep(episodeID, "transcribe", "error", "model unavailable", elapsed)
			p.logAPI("groq", episodeID, "error", elapsed, 0, err.Error())
			_ = settings.Set(p.DB, "processing_paused", "true")
			p.notifyAdmin(fmt.Sprintf("⚠️ Processing paused: Groq model %q unavailable for episode %d (%s)", whisperModel, episodeID, title))
			return nil
		}
		p.logStep(episodeID, "transcribe", "error", err.Error(), elapsed)
		p.logAPI("groq", episodeID, "error", elapsed, 0, err.Error())
		p.handleFailure(episodeID, "transcribe", retryCount, err)
		return nil
	}

	p.logStep(episodeID, "transcribe", "ok", fmt.Sprintf("transcribed %d chars", len(result.Text)), elapsed)
	p.logAPI("groq", episodeID, "ok", elapsed, 0, "")

	_, _ = p.DB.Exec(`UPDATE episodes SET transcript = ?, current_step = 'summarize' WHERE id = ?`,
		result.Text, episodeID)
	os.Remove(audioPath)

	return p.stepSummarize(episodeID, title, podcastName, showNotes, retryCount)
}

// stepSummarize generates a summary from the episode transcript.
func (p *Processor) stepSummarize(episodeID int, title, podcastName, showNotes string, retryCount int) error {
	var transcript string
	err := p.DB.QueryRow(`SELECT transcript FROM episodes WHERE id = ?`, episodeID).Scan(&transcript)
	if err != nil {
		return fmt.Errorf("fetch transcript: %w", err)
	}

	llmModel, err := settings.Get(p.DB, "groq_llm_model")
	if err != nil || llmModel == "" {
		llmModel = "llama-3.3-70b-versatile"
	}

	chunkSize, err := settings.GetInt(p.DB, "chunk_size_tokens")
	if err != nil || chunkSize <= 0 {
		chunkSize = 3000
	}

	wordCount := len(strings.Fields(transcript))
	ctx := context.Background()

	var summary string
	if wordCount > chunkSize {
		summary, err = p.summarizeChunked(ctx, episodeID, title, podcastName, showNotes, transcript, llmModel, chunkSize, retryCount)
	} else {
		summary, err = p.summarizeSingle(ctx, episodeID, title, podcastName, showNotes, transcript, llmModel, retryCount)
	}
	if err != nil {
		return nil // error already handled inside helpers
	}

	_, _ = p.DB.Exec(`UPDATE episodes SET summary = ?, status = 'done', processed_at = ?, current_step = 'notify' WHERE id = ?`,
		summary, time.Now().UTC(), episodeID)
	p.logStep(episodeID, "summarize", "ok", fmt.Sprintf("summary %d chars", len(summary)), 0)
	return nil
}

func (p *Processor) summarizeSingle(ctx context.Context, episodeID int, title, podcastName, showNotes, transcript, model string, retryCount int) (string, error) {
	prompt := fmt.Sprintf(summaryPrompt, title, podcastName, showNotes, transcript)
	start := time.Now()
	result, err := p.GroqClient.Summarize(ctx, prompt, model)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return "", p.handleSummarizeError(episodeID, model, title, retryCount, elapsed, err)
	}

	p.logAPI("groq", episodeID, "ok", elapsed, result.TokensUsed, "")
	return result.Text, nil
}

func (p *Processor) summarizeChunked(ctx context.Context, episodeID int, title, podcastName, showNotes, transcript, model string, chunkSize, retryCount int) (string, error) {
	overlap := chunkSize / 10
	chunks := ChunkTranscript(transcript, chunkSize, overlap)

	var partialSummaries []string
	for i, chunk := range chunks {
		prompt := fmt.Sprintf(summaryPrompt, title, podcastName, showNotes, chunk)
		start := time.Now()
		result, err := p.GroqClient.Summarize(ctx, prompt, model)
		elapsed := time.Since(start).Milliseconds()

		if err != nil {
			return "", p.handleSummarizeError(episodeID, model, title, retryCount, elapsed, err)
		}

		p.logAPI("groq", episodeID, "ok", elapsed, result.TokensUsed, "")
		p.logStep(episodeID, "summarize", "ok", fmt.Sprintf("chunk %d/%d summarized", i+1, len(chunks)), elapsed)
		partialSummaries = append(partialSummaries, result.Text)
	}

	// Merge partial summaries.
	mergeInput := strings.Join(partialSummaries, "\n\n---\n\n")
	prompt := fmt.Sprintf(mergePrompt, mergeInput)
	start := time.Now()
	result, err := p.GroqClient.Summarize(ctx, prompt, model)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return "", p.handleSummarizeError(episodeID, model, title, retryCount, elapsed, err)
	}

	p.logAPI("groq", episodeID, "ok", elapsed, result.TokensUsed, "")
	return result.Text, nil
}

func (p *Processor) handleSummarizeError(episodeID int, model, title string, retryCount int, elapsed int64, err error) error {
	var rlErr *groq.RateLimitError
	if errors.As(err, &rlErr) {
		p.logStep(episodeID, "summarize", "rate_limited", err.Error(), elapsed)
		p.logAPI("groq", episodeID, "rate_limited", elapsed, 0, err.Error())
		log.Printf("episode %d: rate limited during summarization, will retry next cycle", episodeID)
		return err
	}
	if errors.Is(err, groq.ErrModelUnavailable) {
		p.logStep(episodeID, "summarize", "error", "model unavailable", elapsed)
		p.logAPI("groq", episodeID, "error", elapsed, 0, err.Error())
		_ = settings.Set(p.DB, "processing_paused", "true")
		p.notifyAdmin(fmt.Sprintf("⚠️ Processing paused: Groq model %q unavailable for episode %d (%s)", model, episodeID, title))
		return err
	}
	p.logStep(episodeID, "summarize", "error", err.Error(), elapsed)
	p.logAPI("groq", episodeID, "error", elapsed, 0, err.Error())
	p.handleFailure(episodeID, "summarize", retryCount, err)
	return err
}

// handleFailure increments retry count and either retries or marks as permanently failed.
func (p *Processor) handleFailure(episodeID int, step string, retryCount int, err error) {
	retryCount++
	maxRetries, settErr := settings.GetInt(p.DB, "max_retries")
	if settErr != nil || maxRetries <= 0 {
		maxRetries = 3
	}

	if retryCount >= maxRetries {
		_, _ = p.DB.Exec(`UPDATE episodes SET status = 'failed', retry_count = ?, last_error = ? WHERE id = ?`,
			retryCount, err.Error(), episodeID)
		p.logStep(episodeID, step, "failed", fmt.Sprintf("max retries reached: %s", err.Error()), 0)

		var title string
		_ = p.DB.QueryRow(`SELECT title FROM episodes WHERE id = ?`, episodeID).Scan(&title)
		p.notifyAdmin(fmt.Sprintf("❌ Episode %d (%s) failed at step %q after %d retries: %s", episodeID, title, step, retryCount, err.Error()))
		return
	}

	_, _ = p.DB.Exec(`UPDATE episodes SET status = 'pending', retry_count = ?, last_error = ? WHERE id = ?`,
		retryCount, err.Error(), episodeID)
	log.Printf("episode %d: step %s failed (retry %d/%d): %v", episodeID, step, retryCount, maxRetries, err)
}

func (p *Processor) logStep(episodeID int, step, status, message string, durationMs int64) {
	_, err := p.DB.Exec(`INSERT INTO episode_logs (episode_id, step, status, message, duration_ms, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		episodeID, step, status, message, durationMs, time.Now().UTC())
	if err != nil {
		log.Printf("failed to log step: %v", err)
	}
}

func (p *Processor) logAPI(service string, episodeID int, status string, durationMs int64, tokensUsed int, errMsg string) {
	_, err := p.DB.Exec(`INSERT INTO api_logs (service, episode_id, status, duration_ms, tokens_used, error, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		service, episodeID, status, durationMs, tokensUsed, errMsg, time.Now().UTC())
	if err != nil {
		log.Printf("failed to log API call: %v", err)
	}
}

func (p *Processor) notifyAdmin(message string) {
	if p.TelegramClient == nil || p.AdminChatID == "" {
		log.Printf("admin notification (no telegram): %s", message)
		return
	}
	if err := p.TelegramClient.SendMessage(p.AdminChatID, message); err != nil {
		log.Printf("failed to send admin notification: %v", err)
	}
}
