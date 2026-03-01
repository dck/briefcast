package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/briefcast/briefcast/internal/config"
	"github.com/briefcast/briefcast/internal/db"
	"github.com/briefcast/briefcast/internal/groq"
	"github.com/briefcast/briefcast/internal/resend"
	"github.com/briefcast/briefcast/internal/settings"
	"github.com/briefcast/briefcast/internal/telegram"
	"github.com/briefcast/briefcast/internal/worker"
	"github.com/briefcast/briefcast/migrations"
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

	// Initialize clients
	groqClient := groq.NewClient(cfg.GroqAPIKey)
	tgClient := telegram.NewClient(cfg.TelegramBotToken)
	resendClient := resend.NewClient(cfg.ResendAPIKey, cfg.ResendFromEmail)

	// Initialize workers
	poller := worker.NewPoller(database)
	processor := worker.NewProcessor(database, groqClient, tgClient, cfg.TelegramAdminChatID, cfg.AudioTmpDir)
	notifier := worker.NewNotifier(database, tgClient, resendClient, cfg.BaseURL)

	// Create scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal("creating scheduler: ", err)
	}

	// RSS poller job - reads interval from settings, default 60 min
	_, err = s.NewJob(
		gocron.DurationJob(1*time.Minute),
		gocron.NewTask(func() {
			writeHeartbeat(database, "rss_poller")

			interval, _ := settings.GetInt(database, "rss_poll_interval_minutes")
			if interval <= 0 {
				interval = 60
			}

			var lastCheck sql.NullString
			database.QueryRow("SELECT MAX(last_checked_at) FROM podcasts").Scan(&lastCheck)
			if lastCheck.Valid {
				t, err := time.Parse("2006-01-02 15:04:05", lastCheck.String)
				if err == nil && time.Since(t) < time.Duration(interval)*time.Minute {
					return
				}
			}

			log.Println("starting RSS poll")
			if err := poller.Poll(); err != nil {
				log.Printf("RSS poll error: %v", err)
			}
		}),
	)
	if err != nil {
		log.Fatal("scheduling poller: ", err)
	}

	// Episode processor job - runs every minute
	_, err = s.NewJob(
		gocron.DurationJob(1*time.Minute),
		gocron.NewTask(func() {
			writeHeartbeat(database, "episode_processor")
			if err := processor.ProcessPending(); err != nil {
				log.Printf("processor error: %v", err)
			}
		}),
	)
	if err != nil {
		log.Fatal("scheduling processor: ", err)
	}

	// Notification job - check for done episodes needing notification
	_, err = s.NewJob(
		gocron.DurationJob(30*time.Second),
		gocron.NewTask(func() {
			rows, err := database.Query(`
				SELECT id FROM episodes
				WHERE status = 'done' AND current_step = 'notify'
			`)
			if err != nil {
				log.Printf("notification query error: %v", err)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var id int
				if err := rows.Scan(&id); err != nil {
					continue
				}
				if err := notifier.NotifyForEpisode(id); err != nil {
					log.Printf("notification error for episode %d: %v", id, err)
				} else {
					database.Exec("UPDATE episodes SET current_step = 'complete' WHERE id = ?", id)
				}
			}
		}),
	)
	if err != nil {
		log.Fatal("scheduling notifier: ", err)
	}

	s.Start()
	log.Println("worker started")

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("shutting down worker...")
	s.Shutdown()
}

func writeHeartbeat(db *sql.DB, name string) {
	_, _ = db.Exec(`
		INSERT INTO worker_heartbeats (worker_name, last_beat_at)
		VALUES (?, datetime('now'))
		ON CONFLICT(worker_name) DO UPDATE SET last_beat_at = datetime('now')
	`, name)
}
