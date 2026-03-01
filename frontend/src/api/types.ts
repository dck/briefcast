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

export type AdminUser = User & {
  podcastCount: number
  isActive: boolean
}

export type EpisodeLog = {
  id: number
  step: string
  status: string
  message: string
  durationMs: number
  createdAt: string
}

export type AdminEpisode = Episode & {
  skipReason: string
  logs: EpisodeLog[]
}
