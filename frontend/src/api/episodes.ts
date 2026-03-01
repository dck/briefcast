import { apiFetch } from "./client"
import type { Episode } from "./types"

export const getEpisode = (id: number) => apiFetch<Episode>(`/episodes/${id}`)
export const markRead = (id: number) =>
  apiFetch<void>(`/episodes/${id}/read`, { method: "POST" })
export const toggleBookmark = (id: number) =>
  apiFetch<{ bookmarked: boolean }>(`/episodes/${id}/bookmark`, { method: "POST" })
export const shareEpisode = (id: number) =>
  apiFetch<{ shareUrl: string }>(`/episodes/${id}/share`, { method: "POST" })
