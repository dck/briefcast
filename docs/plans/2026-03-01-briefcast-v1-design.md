# Briefcast v1 Design

## Overview

Briefcast is a podcast summary platform. It monitors RSS feeds, downloads new episodes, transcribes them via Groq Whisper, generates structured article-style summaries via Groq Llama, and notifies users via Telegram/email.

## Architecture

Follows PRD section 7 exactly:

- **4 Docker containers**: Caddy (reverse proxy), frontend (static-web-server), server (Go Chi API), worker (Go gocron)
- **Single SQLite database** (WAL mode) shared between server and worker via Docker named volume
- **Caddy routes**: `/api/*` and `/e/*` → Go server, everything else → static frontend

## Key Design Decisions (Deviations from PRD)

1. **Frontend framework**: Vite + react-router + shadcn/ui (not Next.js/daisyUI). v0-generated components kept, Next.js stripped.
2. **DB layer**: Hand-written `database/sql` queries (not sqlc). Simpler, no code gen.
3. **SQLite driver**: `modernc.org/sqlite` — pure Go, no CGO, simpler Docker builds.

## Backend Structure

```
backend/
├── cmd/server/main.go       # Chi router, migrations, listen :8080
├── cmd/worker/main.go        # gocron jobs, start
├── internal/
│   ├── config/               # godotenv, typed Config struct
│   ├── db/                   # SQLite connection, query helpers
│   ├── handler/              # auth, feed, episode, podcast, admin
│   ├── middleware/            # auth, admin-only, session sliding
│   ├── worker/               # poller, processor, chunker, notifier
│   ├── groq/                 # Whisper + Llama API client
│   ├── resend/               # Email client
│   ├── telegram/             # Bot API client
│   ├── oauth/                # Google, GitHub, Yandex
│   └── settings/             # DB settings hot-reload
├── templates/share.html      # Public share page (goldmark SSR)
├── migrations/               # Goose SQL files
└── Dockerfile
```

## Frontend Structure

```
frontend/
├── src/
│   ├── main.tsx              # React root + router
│   ├── api/                  # Fetch wrappers for /api/*
│   ├── pages/                # Route-level components
│   │   ├── Landing.tsx
│   │   ├── Feed.tsx
│   │   ├── Saved.tsx
│   │   ├── Episode.tsx
│   │   ├── Settings.tsx
│   │   └── admin/
│   │       ├── Dashboard.tsx
│   │       ├── Episodes.tsx
│   │       ├── Users.tsx
│   │       ├── Sessions.tsx
│   │       └── Settings.tsx
│   ├── components/           # Reusable UI components (from v0)
│   ├── hooks/
│   └── lib/
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

## Data Model

Exactly as PRD section 5. 12 tables: users, sessions, podcasts, episodes, subscriptions, share_links, episode_reads, bookmarks, notifications, episode_logs, api_logs, worker_heartbeats, settings.

## API Routes

Exactly as PRD section 9. Public, auth, user, and admin route groups.
