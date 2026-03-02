package handler

import (
	"bytes"
	"database/sql"
	"embed"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/dck/briefcast/internal/repository"
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
	repo := repository.NewShareRepository(db)

	// Parse template from embed.FS at startup
	tmpl := template.Must(template.ParseFS(tmplFS, "share.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")

		shared, err := repo.GetSharedEpisode(r.Context(), token)
		if errors.Is(err, repository.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, "Internal server error", 500)
			return
		}

		data := SharePageData{
			EpisodeTitle:    shared.EpisodeTitle,
			PodcastTitle:    shared.PodcastTitle,
			PodcastImageURL: shared.PodcastImageURL,
			PublishedAt:     shared.PublishedAt,
			AudioURL:        shared.AudioURL,
		}

		// Convert Markdown to HTML via goldmark
		if shared.HasSummary {
			var buf bytes.Buffer
			if err := goldmark.Convert([]byte(shared.Summary), &buf); err == nil {
				data.SummaryHTML = template.HTML(buf.String())
			}

			// Extract teaser: first ~200 chars of the plain summary text
			teaser := shared.Summary
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
