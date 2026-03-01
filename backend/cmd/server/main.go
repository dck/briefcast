package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/briefcast/briefcast/internal/config"
	"github.com/briefcast/briefcast/internal/db"
	"github.com/briefcast/briefcast/internal/handler"
	"github.com/briefcast/briefcast/internal/middleware"
	"github.com/briefcast/briefcast/internal/oauth"
	"github.com/briefcast/briefcast/migrations"
	"github.com/briefcast/briefcast/templates"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("loading config: ", err)
	}

	database, err := db.Open(cfg.DatabasePath, migrations.FS)
	if err != nil {
		log.Fatal("opening database: ", err)
	}
	defer database.Close()

	providers := oauth.Providers(cfg)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)

	// Public routes
	r.Get("/api/health", handler.Health)
	r.Get("/e/{token}", handler.SharePage(database, templates.FS))

	// Auth routes (no auth required)
	r.Get("/api/auth/{provider}", handler.AuthRedirect(cfg, providers))
	r.Get("/api/auth/callback", handler.AuthCallback(cfg, database, providers))

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(database, cfg.SessionSecret))

		r.Post("/api/auth/logout", handler.Logout(database))
		r.Get("/api/auth/me", handler.Me())

		r.Get("/api/feed", handler.Feed(database))
		r.Get("/api/saved", handler.Saved(database))

		r.Get("/api/episodes/{id}", handler.GetEpisode(database))
		r.Post("/api/episodes/{id}/read", handler.MarkRead(database))
		r.Post("/api/episodes/{id}/bookmark", handler.ToggleBookmark(database))
		r.Post("/api/episodes/{id}/share", handler.ShareEpisode(database))

		r.Get("/api/podcasts", handler.ListPodcasts(database))
		r.Post("/api/podcasts", handler.AddPodcast(database))
		r.Delete("/api/podcasts/{id}", handler.RemovePodcast(database))

		r.Get("/api/settings", handler.GetSettings())
		r.Put("/api/settings", handler.UpdateSettings(database))
	})

	// Admin routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(database, cfg.SessionSecret))
		r.Use(middleware.RequireAdmin)

		r.Get("/api/admin/stats", handler.AdminStats(database))
		r.Get("/api/admin/episodes", handler.AdminEpisodes(database))
		r.Post("/api/admin/episodes/{id}/retry", handler.AdminRetryEpisode(database))
		r.Post("/api/admin/episodes/{id}/retry-all", handler.AdminRetryAllEpisode(database))
		r.Post("/api/admin/episodes/{id}/skip", handler.AdminSkipEpisode(database))
		r.Get("/api/admin/users", handler.AdminUsers(database))
		r.Post("/api/admin/users/{id}/deactivate", handler.AdminDeactivateUser(database))
		r.Get("/api/admin/sessions", handler.AdminSessions(database))
		r.Delete("/api/admin/sessions/{token}", handler.AdminRevokeSession(database))
		r.Get("/api/admin/settings", handler.AdminGetSettings(database))
		r.Put("/api/admin/settings", handler.AdminUpdateSettings(database))
		r.Post("/api/admin/processing/resume", handler.AdminResumeProcessing(database))
	})

	log.Printf("server starting on :%s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		log.Fatal("server error: ", err)
	}
}
