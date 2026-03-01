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
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    last_seen_at TEXT,
    UNIQUE(oauth_provider, oauth_id)
);

CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    last_seen_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT NOT NULL
);

CREATE TABLE podcasts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rss_url TEXT NOT NULL UNIQUE,
    title TEXT,
    description TEXT,
    image_url TEXT,
    last_checked_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE episodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    podcast_id INTEGER NOT NULL REFERENCES podcasts(id),
    guid TEXT NOT NULL,
    title TEXT,
    description TEXT,
    audio_url TEXT,
    show_notes TEXT,
    transcript TEXT,
    summary TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    current_step TEXT DEFAULT 'download',
    retry_count INTEGER DEFAULT 0,
    last_error TEXT,
    skip_reason TEXT,
    published_at TEXT,
    processed_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(podcast_id, guid)
);

CREATE TABLE subscriptions (
    user_id INTEGER NOT NULL REFERENCES users(id),
    podcast_id INTEGER NOT NULL REFERENCES podcasts(id),
    active BOOLEAN DEFAULT true,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY(user_id, podcast_id)
);

CREATE TABLE share_links (
    token TEXT PRIMARY KEY,
    episode_id INTEGER NOT NULL UNIQUE REFERENCES episodes(id),
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE episode_reads (
    user_id INTEGER NOT NULL REFERENCES users(id),
    episode_id INTEGER NOT NULL REFERENCES episodes(id),
    read_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY(user_id, episode_id)
);

CREATE TABLE bookmarks (
    user_id INTEGER NOT NULL REFERENCES users(id),
    episode_id INTEGER NOT NULL REFERENCES episodes(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY(user_id, episode_id)
);

CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    episode_id INTEGER NOT NULL REFERENCES episodes(id),
    sent_via_telegram BOOLEAN DEFAULT false,
    sent_via_email BOOLEAN DEFAULT false,
    telegram_error TEXT,
    email_error TEXT,
    sent_at TEXT
);

CREATE TABLE episode_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    episode_id INTEGER NOT NULL REFERENCES episodes(id),
    step TEXT NOT NULL,
    status TEXT NOT NULL,
    message TEXT,
    duration_ms INTEGER,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE api_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service TEXT NOT NULL,
    episode_id INTEGER REFERENCES episodes(id),
    status TEXT NOT NULL,
    duration_ms INTEGER,
    tokens_used INTEGER,
    error TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE worker_heartbeats (
    worker_name TEXT PRIMARY KEY,
    last_beat_at TEXT NOT NULL
);

CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Seed default settings
INSERT INTO settings (key, value) VALUES
    ('groq_whisper_model', 'whisper-large-v3'),
    ('groq_llm_model', 'llama-3.3-70b-versatile'),
    ('rss_poll_interval_minutes', '60'),
    ('max_retries', '3'),
    ('retry_backoff_minutes', '1,5,15'),
    ('audio_max_size_mb', '500'),
    ('chunk_size_tokens', '4000'),
    ('processing_paused', 'false');

-- Indexes for common queries
CREATE INDEX idx_episodes_status ON episodes(status);
CREATE INDEX idx_episodes_podcast_id ON episodes(podcast_id);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_podcast_id ON subscriptions(podcast_id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_episode_logs_episode_id ON episode_logs(episode_id);
CREATE INDEX idx_api_logs_created_at ON api_logs(created_at);
CREATE INDEX idx_notifications_episode_id ON notifications(episode_id);
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_episode_reads_user_id ON episode_reads(user_id);

-- +goose Down
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS worker_heartbeats;
DROP TABLE IF EXISTS api_logs;
DROP TABLE IF EXISTS episode_logs;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS bookmarks;
DROP TABLE IF EXISTS episode_reads;
DROP TABLE IF EXISTS share_links;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS episodes;
DROP TABLE IF EXISTS podcasts;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
