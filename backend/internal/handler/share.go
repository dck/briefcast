package handler

import (
	"bytes"
	"database/sql"
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/yuin/goldmark"
)

type SharePageData struct {
	EpisodeTitle    string
	PodcastTitle    string
	PodcastImageURL string
	Teaser          string
	SummaryHTML     template.HTML
	PublishedAt     string
	AudioURL        string
}

// SharePage handles GET /e/{token}
// Looks up the share_link by token, fetches episode and podcast data,
// renders Markdown summary to HTML via goldmark, renders the template.
// No authentication required.
func SharePage(db *sql.DB, tmplFS embed.FS) http.HandlerFunc {
	// Parse template from embed.FS at startup
	tmpl := template.Must(template.ParseFS(tmplFS, "templates/share.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")

		// Query share_link → episode → podcast
		var data SharePageData
		var summary sql.NullString
		var audioURL, imageURL, publishedAt sql.NullString

		err := db.QueryRow(`
			SELECT e.title, p.title, p.image_url, e.summary, e.audio_url, e.published_at
			FROM share_links sl
			JOIN episodes e ON sl.episode_id = e.id
			JOIN podcasts p ON e.podcast_id = p.id
			WHERE sl.token = ? AND e.status = 'done'
		`, token).Scan(
			&data.EpisodeTitle, &data.PodcastTitle,
			&imageURL, &summary, &audioURL, &publishedAt,
		)

		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, "Internal server error", 500)
			return
		}

		data.PodcastImageURL = imageURL.String
		data.AudioURL = audioURL.String
		data.PublishedAt = publishedAt.String

		// Convert Markdown to HTML via goldmark
		if summary.Valid {
			var buf bytes.Buffer
			if err := goldmark.Convert([]byte(summary.String), &buf); err == nil {
				data.SummaryHTML = template.HTML(buf.String())
			}

			// Extract teaser: first ~200 chars of the plain summary text
			teaser := summary.String
			if idx := strings.Index(teaser, "\n\n"); idx > 0 && idx < 300 {
				teaser = teaser[:idx]
			}
			if len(teaser) > 200 {
				teaser = teaser[:200] + "..."
			}
			data.Teaser = teaser
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	}
}
