# Briefcast v1 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a complete podcast summary platform — Go backend (API + worker), React Vite frontend, Docker deployment.

**Architecture:** Go Chi HTTP server + gocron worker sharing a SQLite (WAL) database. React Vite SPA served by static-web-server. Caddy reverse proxy in front. Episode pipeline: RSS poll → download → Groq Whisper transcription → Groq Llama summarization → Telegram/email notification.

**Tech Stack:** Go 1.23 (Chi v5, goose, gocron, zerolog, goldmark, modernc.org/sqlite), React 19 (Vite, react-router, shadcn/ui, Tailwind CSS), Docker Compose, Caddy.

**Reference:** `prd.md` (full PRD), `docs/plans/2026-03-01-briefcast-v1-design.md` (design doc).

---

## Phase 1: Frontend — Convert to Vite SPA & Fix Bugs

### Task 1: Fix episode-summary.tsx bug

**Files:**
- Fix: `frontend/components/pages/episode-summary.tsx:74`

**Step 1:** In `episode-summary.tsx`, line 74 uses `<AudioWaveform>` which is not imported. Replace with `<BriefcastLogo>` which is imported (line 13) but unused:

```tsx
// Line 74 — change FROM:
<AudioWaveform className="h-5 w-5 text-primary" />
// TO:
<BriefcastLogo size={20} className="text-primary" />
```

**Step 2:** Commit: `fix: replace missing AudioWaveform with BriefcastLogo in episode-summary`

---

### Task 2: Convert frontend from Next.js to Vite

**Files:**
- Delete: `frontend/app/` directory (Next.js app router)
- Delete: `frontend/next.config.mjs`
- Delete: `frontend/pnpm-lock.yaml`
- Create: `frontend/index.html`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`
- Create: `frontend/vite.config.ts`
- Modify: `frontend/package.json`
- Modify: `frontend/tsconfig.json`

**Step 1:** Remove Next.js dependencies and add Vite + react-router to `package.json`:

```json
{
  "name": "briefcast-frontend",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc -b && vite build",
    "preview": "vite preview",
    "lint": "eslint ."
  }
}
```

Remove: `next`, `@next/*`, `next-themes` deps.
Add: `vite`, `@vitejs/plugin-react`, `react-router-dom`.
Keep: all `@radix-ui/*`, `tailwind-merge`, `clsx`, `lucide-react`, `recharts`, `react-hook-form`, `zod`, shadcn deps.

**Step 2:** Create `frontend/vite.config.ts`:

```ts
import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import path from "path"

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "."),
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:8080",
      "/e": "http://localhost:8080",
    },
  },
})
```

**Step 3:** Create `frontend/index.html`:

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Briefcast</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

**Step 4:** Create `frontend/src/main.tsx`:

```tsx
import React from "react"
import ReactDOM from "react-dom/client"
import { BrowserRouter } from "react-router-dom"
import App from "./App"
import "../app/globals.css"

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>
)
```

**Step 5:** Create `frontend/src/App.tsx` with routes:

```tsx
import { Routes, Route, Navigate } from "react-router-dom"
import { LandingPage } from "@/components/pages/landing-page"
import { DashboardFeed } from "@/components/pages/dashboard-feed"
import { EpisodeSummary } from "@/components/pages/episode-summary"
import { SavedPage } from "@/components/pages/saved-page"
import { SettingsPage } from "@/components/pages/settings-page"
import { AdminDashboard } from "@/components/pages/admin-dashboard"
import { AdminEpisodes } from "@/components/pages/admin-episodes"
import { AdminUsers } from "@/components/pages/admin-users"
import { AdminSessions } from "@/components/pages/admin-sessions"
import { AdminSettings } from "@/components/pages/admin-settings"

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/feed" element={<DashboardFeed />} />
      <Route path="/saved" element={<SavedPage />} />
      <Route path="/episodes/:id" element={<EpisodeSummary />} />
      <Route path="/settings" element={<SettingsPage />} />
      <Route path="/admin" element={<AdminDashboard />} />
      <Route path="/admin/episodes" element={<AdminEpisodes />} />
      <Route path="/admin/users" element={<AdminUsers />} />
      <Route path="/admin/sessions" element={<AdminSessions />} />
      <Route path="/admin/settings" element={<AdminSettings />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
```

**Step 6:** Update `frontend/tsconfig.json` for Vite (remove Next.js-specific config, add Vite types).

**Step 7:** Remove `frontend/app/` directory and `frontend/next.config.mjs`.

**Step 8:** Update all page components to remove any `"use client"` directives (not needed in Vite).

**Step 9:** Replace `next-themes` ThemeProvider with a simple CSS-based dark mode or remove theme switching for now.

**Step 10:** Run `npm install` and `npm run dev` — verify the app loads at localhost:5173.

**Step 11:** Commit: `feat: convert frontend from Next.js to Vite SPA with react-router`

---

### Task 3: Add missing frontend pages

**Files:**
- Create: `frontend/components/pages/saved-page.tsx`
- Create: `frontend/components/pages/settings-page.tsx`
- Create: `frontend/components/pages/admin-episodes.tsx`
- Create: `frontend/components/pages/admin-users.tsx`
- Create: `frontend/components/pages/admin-sessions.tsx`

**Step 1:** Create `saved-page.tsx` — Same layout as `dashboard-feed.tsx` but shows bookmarked episodes only. Use AppSidebar, EpisodeCard components.

**Step 2:** Create `settings-page.tsx` — User profile settings: Telegram chat ID, email, notification toggles (Telegram on/off, Email on/off). Use AppSidebar + form with Input, Switch components.

**Step 3:** Create `admin-episodes.tsx` — Full episode queue (all statuses). Table with columns: Podcast, Title, Status (StatusBadge), Current Step, Retry Count, Last Error, Time Elapsed, Actions (Retry/Skip buttons). Status filter tabs: All/Pending/Processing/Done/Failed/Skipped.

**Step 4:** Create `admin-users.tsx` — Users table: Name, Email, OAuth Provider, Joined, Podcast Count, Telegram Connected, Last Seen, Active status. Deactivate button per user.

**Step 5:** Create `admin-sessions.tsx` — Sessions table: User, Created At, Last Seen, Expires At, Revoke button.

**Step 6:** Update sidebar navigation links to use `react-router-dom` `Link` components instead of static items.

**Step 7:** Verify all pages render correctly at their routes.

**Step 8:** Commit: `feat: add missing pages (saved, settings, admin episodes/users/sessions)`

---

### Task 4: Create frontend API layer

**Files:**
- Create: `frontend/src/api/client.ts`
- Create: `frontend/src/api/auth.ts`
- Create: `frontend/src/api/feed.ts`
- Create: `frontend/src/api/episodes.ts`
- Create: `frontend/src/api/podcasts.ts`
- Create: `frontend/src/api/settings.ts`
- Create: `frontend/src/api/admin.ts`
- Create: `frontend/src/api/types.ts`

**Step 1:** Create `types.ts` with all shared TypeScript types matching the backend data model:

```ts
export type User = {
  id: number
  name: string
  email: string
  avatarUrl: string
  oauthProvider: string
  telegramChatId: string
  notifyTelegram: boolean
  notifyEmail: boolean
  isAdmin: boolean
  createdAt: string
  lastSeenAt: string
}

export type Podcast = {
  id: number
  title: string
  description: string
  imageUrl: string
  rssUrl: string
  episodeCount: number
}

export type Episode = {
  id: number
  podcastId: number
  podcastTitle: string
  podcastImageUrl: string
  title: string
  description: string
  audioUrl: string
  summary: string
  status: "pending" | "processing" | "done" | "failed" | "skipped"
  currentStep: string
  retryCount: number
  lastError: string
  publishedAt: string
  processedAt: string
  isRead: boolean
  isBookmarked: boolean
}

export type AdminStats = {
  pending: number
  processing: number
  done: number
  failed: number
  skipped: number
  groqRequestsToday: number
  groqTokensToday: number
  workerLastBeat: string
  rssLastRun: string
  rssNextRun: string
  processingPaused: boolean
}

export type Session = {
  token: string
  userId: number
  userName: string
  createdAt: string
  lastSeenAt: string
  expiresAt: string
}

export type Settings = {
  [key: string]: string
}
```

**Step 2:** Create `client.ts` — base fetch wrapper:

```ts
const BASE = "/api"

export async function apiFetch<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...opts?.headers },
    ...opts,
  })
  if (res.status === 401) {
    window.location.href = "/"
    throw new Error("Unauthorized")
  }
  if (!res.ok) {
    const body = await res.text()
    throw new Error(body || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}
```

**Step 3:** Create route-specific API modules (`auth.ts`, `feed.ts`, `episodes.ts`, `podcasts.ts`, `settings.ts`, `admin.ts`) — each exports functions matching the PRD's API routes:

```ts
// auth.ts
export const getMe = () => apiFetch<User>("/auth/me")
export const logout = () => apiFetch<void>("/auth/logout", { method: "POST" })

// feed.ts
export const getFeed = (page: number) => apiFetch<{ episodes: Episode[]; hasMore: boolean }>(`/feed?page=${page}`)
export const getSaved = (page: number) => apiFetch<{ episodes: Episode[]; hasMore: boolean }>(`/saved?page=${page}`)

// episodes.ts
export const getEpisode = (id: number) => apiFetch<Episode>(`/episodes/${id}`)
export const markRead = (id: number) => apiFetch<void>(`/episodes/${id}/read`, { method: "POST" })
export const toggleBookmark = (id: number) => apiFetch<{ bookmarked: boolean }>(`/episodes/${id}/bookmark`, { method: "POST" })
export const shareEpisode = (id: number) => apiFetch<{ shareUrl: string }>(`/episodes/${id}/share`, { method: "POST" })

// podcasts.ts
export const getPodcasts = () => apiFetch<Podcast[]>("/podcasts")
export const addPodcast = (rssUrl: string) => apiFetch<Podcast>("/podcasts", { method: "POST", body: JSON.stringify({ rssUrl }) })
export const removePodcast = (id: number) => apiFetch<void>(`/podcasts/${id}`, { method: "DELETE" })

// settings.ts
export const getSettings = () => apiFetch<User>("/settings")
export const updateSettings = (s: Partial<User>) => apiFetch<void>("/settings", { method: "PUT", body: JSON.stringify(s) })

// admin.ts — all /admin/* routes
```

**Step 4:** Commit: `feat: add frontend API layer with typed fetch wrappers`

---

## Phase 2: Backend — Foundation

### Task 5: Initialize Go module and project structure

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go` (skeleton)
- Create: `backend/cmd/worker/main.go` (skeleton)
- Create: `backend/internal/config/config.go`
- Create: `backend/.env.example`

**Step 1:** Initialize Go module:

```bash
cd backend && go mod init github.com/user/briefcast
```

**Step 2:** Create `internal/config/config.go` — typed Config struct loaded from `.env` via `godotenv`:

```go
package config

import (
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    ServerPort          string
    BaseURL             string
    DatabasePath        string
    AudioTmpDir         string
    SessionSecret       string
    GoogleClientID      string
    GoogleClientSecret  string
    GitHubClientID      string
    GitHubClientSecret  string
    YandexClientID      string
    YandexClientSecret  string
    GroqAPIKey          string
    ResendAPIKey        string
    ResendFromEmail     string
    TelegramBotToken    string
    TelegramAdminChatID string
}

func Load() (*Config, error) {
    godotenv.Load() // ignore error — env vars may be set directly
    return &Config{
        ServerPort:          getEnv("SERVER_PORT", "8080"),
        BaseURL:             getEnv("BASE_URL", "http://localhost:5173"),
        DatabasePath:        getEnv("DATABASE_PATH", "./data/briefcast.db"),
        AudioTmpDir:         getEnv("AUDIO_TMP_DIR", "./data/audio"),
        SessionSecret:       os.Getenv("SESSION_SECRET"),
        // ... all other fields from .env.example
    }, nil
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" { return v }
    return fallback
}
```

**Step 3:** Create `.env.example` matching PRD section 10.

**Step 4:** Create skeleton `cmd/server/main.go` and `cmd/worker/main.go` — just loads config and prints startup message.

**Step 5:** Run `go mod tidy`, verify `go build ./...` succeeds.

**Step 6:** Commit: `feat: initialize Go module with config and project structure`

---

### Task 6: Database layer — migrations and connection

**Files:**
- Create: `backend/internal/db/db.go`
- Create: `backend/migrations/001_init.sql`

**Step 1:** Create `migrations/001_init.sql` — full schema from PRD section 5:

```sql
-- +goose Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    oauth_provider TEXT NOT NULL,
    oauth_id TEXT NOT NULL,
    email TEXT,
    name TEXT,
    avatar_url TEXT,
    telegram_chat_id TEXT,
    notify_telegram BOOLEAN DEFAULT false,
    notify_email BOOLEAN DEFAULT false,
    is_admin BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TEXT NOT NULL,
    last_seen_at TEXT,
    UNIQUE(oauth_provider, oauth_id)
);

CREATE TABLE sessions ( ... );  -- full schema per PRD
CREATE TABLE podcasts ( ... );
CREATE TABLE episodes ( ... );
CREATE TABLE subscriptions ( ... );
CREATE TABLE share_links ( ... );
CREATE TABLE episode_reads ( ... );
CREATE TABLE bookmarks ( ... );
CREATE TABLE notifications ( ... );
CREATE TABLE episode_logs ( ... );
CREATE TABLE api_logs ( ... );
CREATE TABLE worker_heartbeats ( ... );
CREATE TABLE settings ( ... );

-- Seed default settings
INSERT INTO settings (key, value, updated_at) VALUES
    ('groq_whisper_model', 'whisper-large-v3', datetime('now')),
    ('groq_llm_model', 'llama-3.3-70b-versatile', datetime('now')),
    ('rss_poll_interval_minutes', '60', datetime('now')),
    ('max_retries', '3', datetime('now')),
    ('retry_backoff_minutes', '1,5,15', datetime('now')),
    ('audio_max_size_mb', '500', datetime('now')),
    ('chunk_size_tokens', '4000', datetime('now')),
    ('processing_paused', 'false', datetime('now'));

-- +goose Down
DROP TABLE IF EXISTS settings;
-- ... drop all tables in reverse order
```

**Step 2:** Create `internal/db/db.go` — SQLite connection with WAL, embedded migrations via goose:

```go
package db

import (
    "database/sql"
    "embed"
    "github.com/pressly/goose/v3"
    _ "modernc.org/sqlite"
)

//go:embed ../../migrations/*.sql
var migrations embed.FS

func Open(path string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
    if err != nil { return nil, err }
    if err := db.Ping(); err != nil { return nil, err }
    goose.SetBaseFS(migrations)
    if err := goose.SetDialect("sqlite3"); err != nil { return nil, err }
    if err := goose.Up(db, "migrations"); err != nil { return nil, err }
    return db, nil
}
```

Note: The embed path may need adjustment based on the final package location. The migrations directory should be embedded relative to the db package or from the cmd packages.

**Step 3:** Update `cmd/server/main.go` to open DB on startup.

**Step 4:** Verify: `go run ./cmd/server` creates the SQLite file with all tables.

**Step 5:** Commit: `feat: add SQLite database layer with goose migrations`

---

### Task 7: Auth — OAuth flows and session management

**Files:**
- Create: `backend/internal/oauth/oauth.go`
- Create: `backend/internal/middleware/auth.go`
- Create: `backend/internal/handler/auth.go`

**Step 1:** Create `internal/oauth/oauth.go` — configures `golang.org/x/oauth2` for Google, GitHub, Yandex. Each provider returns an `oauth2.Config` with proper endpoints and scopes.

**Step 2:** Create `internal/handler/auth.go`:
- `GET /api/auth/{provider}` — generates state token, stores in cookie, redirects to OAuth provider
- `GET /api/auth/callback` — exchanges code for token, fetches user info from provider API, upserts user in DB, creates session (UUID token), sets signed cookie, redirects to `/feed`
- `POST /api/auth/logout` — deletes session from DB, clears cookie
- `GET /api/auth/me` — returns current user from session

**Step 3:** Create `internal/middleware/auth.go`:
- `RequireAuth` — reads session cookie, validates session exists and not expired, extends sliding expiry (if within last 24h of expiry, extend by 30 days), sets user in request context
- `RequireAdmin` — checks `is_admin = true` on user from context
- `UpdateLastSeen` — updates `users.last_seen_at` and `sessions.last_seen_at`

**Step 4:** Wire auth routes into Chi router in `cmd/server/main.go`.

**Step 5:** Commit: `feat: add OAuth authentication (Google, GitHub, Yandex) with session management`

---

### Task 8: API handlers — Feed, Episodes, Podcasts, Settings

**Files:**
- Create: `backend/internal/handler/feed.go`
- Create: `backend/internal/handler/episode.go`
- Create: `backend/internal/handler/podcast.go`
- Create: `backend/internal/handler/settings.go`

**Step 1:** Create `internal/handler/feed.go`:
- `GET /api/feed?page=N` — returns paginated episodes for user's subscriptions where `status=done`, grouped by podcast, sorted by `published_at DESC`. Includes `is_read` and `is_bookmarked` per user. 20 per page.
- `GET /api/saved?page=N` — same but only bookmarked episodes.

**Step 2:** Create `internal/handler/episode.go`:
- `GET /api/episodes/{id}` — full episode with summary (only if `status=done` for non-admin)
- `POST /api/episodes/{id}/read` — upsert into `episode_reads`
- `POST /api/episodes/{id}/bookmark` — toggle: insert or delete from `bookmarks`
- `POST /api/episodes/{id}/share` — get or create share_link, return public URL

**Step 3:** Create `internal/handler/podcast.go`:
- `GET /api/podcasts` — user's subscriptions with episode counts
- `POST /api/podcasts` — add podcast by RSS URL. Parse RSS to get title/description/image. Reuse existing podcast if same RSS URL. Create subscription. Only future episodes processed.
- `DELETE /api/podcasts/{id}` — soft delete subscription (`active=false`)

**Step 4:** Create `internal/handler/settings.go`:
- `GET /api/settings` — user notification preferences
- `PUT /api/settings` — update telegram_chat_id, email, notify_telegram, notify_email

**Step 5:** Wire all routes into Chi router with `RequireAuth` middleware.

**Step 6:** Commit: `feat: add user API handlers (feed, episodes, podcasts, settings)`

---

### Task 9: Admin API handlers

**Files:**
- Create: `backend/internal/handler/admin.go`

**Step 1:** Implement all admin routes from PRD section 9:
- `GET /api/admin/stats` — queue depth counts, Groq usage today (from api_logs), worker heartbeats, processing_paused setting
- `GET /api/admin/episodes` — all episodes with logs, filterable by status
- `POST /api/admin/episodes/{id}/retry` — retry from current_step, reset retry_count
- `POST /api/admin/episodes/{id}/retry-all` — reset to step=download, status=pending
- `POST /api/admin/episodes/{id}/skip` — mark as skipped
- `GET /api/admin/users` — all users with podcast counts
- `POST /api/admin/users/{id}/deactivate` — set is_active=false, delete all sessions
- `GET /api/admin/sessions` — all active sessions with user info
- `DELETE /api/admin/sessions/{token}` — revoke session
- `GET /api/admin/settings` — all settings key-value pairs
- `PUT /api/admin/settings` — update settings
- `POST /api/admin/processing/resume` — set processing_paused=false

**Step 2:** Wire admin routes with `RequireAuth` + `RequireAdmin` middleware.

**Step 3:** Commit: `feat: add admin API handlers`

---

### Task 10: Public share page (SSR)

**Files:**
- Create: `backend/templates/share.html`
- Add handler in: `backend/internal/handler/episode.go`

**Step 1:** Create `templates/share.html` — Go `html/template` rendering Markdown via goldmark. Clean semantic HTML with proper OG meta tags:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.EpisodeTitle}} — {{.PodcastTitle}} | Briefcast</title>
    <meta name="description" content="{{.Teaser}}">
    <meta property="og:title" content="{{.EpisodeTitle}}">
    <meta property="og:description" content="{{.Teaser}}">
    <meta property="og:image" content="{{.PodcastImageURL}}">
    <style>/* Clean typography CSS */</style>
</head>
<body>
    <article>
        <h1>{{.EpisodeTitle}}</h1>
        <p class="meta">{{.PodcastTitle}} · {{.PublishedAt}}</p>
        {{.SummaryHTML}}
    </article>
</body>
</html>
```

**Step 2:** Add `GET /e/{token}` handler — looks up share_link, renders template with goldmark-converted Markdown.

**Step 3:** Commit: `feat: add server-side rendered public share page`

---

## Phase 3: Backend — Worker

### Task 11: RSS poller

**Files:**
- Create: `backend/internal/worker/poller.go`

**Step 1:** Implement RSS poller:
- Fetches all podcasts from DB
- Parses each RSS feed (use `gofeed` library)
- For each new episode (GUID not in DB): insert with `status=pending`, `current_step=download`
- Skip episodes with no audio enclosure (`status=skipped`, reason logged)
- Only insert episodes published after the podcast's `created_at` (no backfill)
- Updates `podcasts.last_checked_at`
- Logs warnings on fetch failures, continues to next feed

**Step 2:** Commit: `feat: add RSS feed poller`

---

### Task 12: Episode processor pipeline

**Files:**
- Create: `backend/internal/worker/processor.go`
- Create: `backend/internal/worker/chunker.go`
- Create: `backend/internal/groq/client.go`

**Step 1:** Create `internal/groq/client.go` — Groq API client:
- `Transcribe(audioPath string) (string, error)` — sends audio to Whisper API
- `Summarize(transcript, showNotes string) (string, error)` — sends to Llama API with structured prompt
- `MergeSummaries(summaries []string) (string, error)` — merges chunk summaries
- Handles rate limits (429 → wait for Retry-After), 5xx retries, model unavailable → pause processing

**Step 2:** Create `internal/worker/chunker.go`:
- Splits transcript into ~4000 token chunks with 200-token overlap
- Simple token estimation: split by whitespace, ~1 token per word

**Step 3:** Create `internal/worker/processor.go` — processes episodes step by step:
- Picks up episodes with `status=pending` or `status=processing` (resume after crash)
- **Download step**: HTTP GET audio URL → save to `$AUDIO_TMP_DIR/{episode_id}.mp3`. Handle 404, corrupt, too large, unsupported format (ffmpeg conversion attempt).
- **Transcribe step**: Send audio to Groq Whisper → save transcript to DB → delete audio file
- **Summarize step**: Send transcript + show notes to Groq Llama. If too long, use chunker. Save summary to DB. Set `status=done`.
- **Retry logic**: Exponential backoff per step (configurable intervals from settings). Max retries from settings. After max → `status=failed`, notify admin.
- Each step logs to `episode_logs` table with duration_ms and status.

**Step 4:** Commit: `feat: add episode processing pipeline (download, transcribe, summarize)`

---

### Task 13: Notification system

**Files:**
- Create: `backend/internal/worker/notifier.go`
- Create: `backend/internal/telegram/client.go`
- Create: `backend/internal/resend/client.go`

**Step 1:** Create `internal/telegram/client.go` — Telegram Bot API:
- `SendMessage(chatID, text string) error` — sends message via HTTP POST to Telegram API
- Used for user notifications and admin alerts

**Step 2:** Create `internal/resend/client.go` — Resend email API:
- `SendEmail(to, subject, htmlBody string) error`

**Step 3:** Create `internal/worker/notifier.go`:
- After `status=done`, find all subscribers with active notification preferences
- For each subscriber: send Telegram and/or email based on their settings
- Telegram format per PRD (emoji, podcast name, title, teaser, link)
- Email: same content as HTML
- Log results to `notifications` table
- No retry on notification failure in v1

**Step 4:** Commit: `feat: add notification system (Telegram + email)`

---

### Task 14: Worker main with gocron

**Files:**
- Modify: `backend/cmd/worker/main.go`
- Create: `backend/internal/settings/settings.go`

**Step 1:** Create `internal/settings/settings.go` — reads settings from DB on every call (hot-reload):

```go
func Get(db *sql.DB, key string) (string, error)
func GetInt(db *sql.DB, key string) (int, error)
func Set(db *sql.DB, key, value string) error
```

**Step 2:** Complete `cmd/worker/main.go`:
- Opens DB connection
- Creates gocron scheduler
- Registers RSS poller job (interval from settings, default 1h)
- Registers episode processor job (runs every 1 minute, picks up pending episodes)
- Writes worker heartbeat every minute
- Graceful shutdown on SIGINT/SIGTERM

**Step 3:** Commit: `feat: complete worker with gocron scheduling and hot-reload settings`

---

## Phase 4: Backend — Server Main

### Task 15: Complete Chi server with all routes

**Files:**
- Modify: `backend/cmd/server/main.go`

**Step 1:** Wire everything together:

```go
func main() {
    cfg := config.Load()
    db := db.Open(cfg.DatabasePath)
    defer db.Close()

    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // Public
    r.Get("/api/health", handler.Health)
    r.Get("/e/{token}", handler.SharePage(db))

    // Auth
    r.Get("/api/auth/{provider}", handler.AuthRedirect(cfg))
    r.Get("/api/auth/callback", handler.AuthCallback(cfg, db))

    // Authenticated routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireAuth(db, cfg))
        r.Post("/api/auth/logout", handler.Logout(db))
        r.Get("/api/auth/me", handler.Me)
        r.Get("/api/feed", handler.Feed(db))
        r.Get("/api/saved", handler.Saved(db))
        // ... all user routes
    })

    // Admin routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireAuth(db, cfg))
        r.Use(middleware.RequireAdmin)
        // ... all admin routes
    })

    log.Info().Str("port", cfg.ServerPort).Msg("server starting")
    http.ListenAndServe(":"+cfg.ServerPort, r)
}
```

**Step 2:** Add zerolog structured logging.

**Step 3:** Verify: `go build ./cmd/server && go build ./cmd/worker` both compile.

**Step 4:** Commit: `feat: complete Chi server with all routes wired`

---

## Phase 5: Infrastructure

### Task 16: Docker setup

**Files:**
- Create: `backend/Dockerfile`
- Create: `frontend/Dockerfile`
- Create: `docker-compose.yml`
- Create: `Caddyfile`

**Step 1:** Create `backend/Dockerfile` (multi-stage):

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server
RUN go build -o worker ./cmd/worker

FROM alpine:latest
RUN apk add --no-cache ca-certificates ffmpeg
WORKDIR /app
COPY --from=builder /app/server /app/worker ./
COPY --from=builder /app/templates ./templates
```

**Step 2:** Create `frontend/Dockerfile`:

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM joseluisq/static-web-server:2-alpine
COPY --from=builder /app/dist /public
ENV SERVER_FALLBACK_PAGE=/public/index.html
```

**Step 3:** Create `docker-compose.yml` per PRD section 11.

**Step 4:** Create `Caddyfile` per PRD section 11.

**Step 5:** Commit: `feat: add Docker setup (Compose, Caddyfile, Dockerfiles)`

---

### Task 17: Makefile and scripts

**Files:**
- Create: `Makefile`
- Create: `scripts/bootstrap-vps.sh`
- Create: `scripts/backup.sh`

**Step 1:** Create `Makefile` with targets per PRD: `dev`, `build`, `test`, `lint`, `docker`, `deploy`, `backup`.

**Step 2:** Create `scripts/bootstrap-vps.sh` per PRD section 11.

**Step 3:** Create `scripts/backup.sh` per PRD section 11.

**Step 4:** Commit: `feat: add Makefile and deployment scripts`

---

### Task 18: GitHub Actions CI/CD

**Files:**
- Create: `.github/workflows/deploy.yml`

**Step 1:** Create deploy workflow per PRD section 11:
- Trigger on push to main
- Run frontend lint + build
- Run Go tests + golangci-lint
- Build Docker images
- SSH deploy to VPS

**Step 2:** Commit: `feat: add GitHub Actions CI/CD pipeline`

---

## Phase 6: Integration & Polish

### Task 19: Wire frontend to real API

**Files:**
- Modify: all page components to use API layer instead of mock data

**Step 1:** Update `dashboard-feed.tsx` — replace hardcoded episodes array with `useEffect` + `getFeed()` API call. Add loading states with SkeletonCard.

**Step 2:** Update `episode-summary.tsx` — fetch episode by ID from URL params, render real Markdown summary.

**Step 3:** Update `saved-page.tsx` — use `getSaved()` API call.

**Step 4:** Update `settings-page.tsx` — load/save via settings API.

**Step 5:** Update admin pages — use admin API calls.

**Step 6:** Update sidebars — load podcasts from API, show real user info.

**Step 7:** Add auth context — check `/api/auth/me` on app load, redirect to landing if not logged in.

**Step 8:** Commit: `feat: wire all frontend pages to backend API`

---

### Task 20: Add health check endpoint

**Files:**
- Add to: `backend/internal/handler/health.go`

**Step 1:** Simple handler:

```go
func Health(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}
```

**Step 2:** Commit: `feat: add /api/health endpoint`

---

## Execution Order

Tasks 1-4 (frontend) and Tasks 5-6 (backend foundation) can run in parallel.
Tasks 7-10 depend on Task 6.
Tasks 11-14 depend on Task 7.
Task 15 depends on Tasks 7-14.
Task 16-18 can run in parallel after Task 15.
Task 19 depends on Tasks 1-4 and Task 15.
Task 20 is independent.
