import { apiFetch } from "./client"
import type { Podcast } from "./types"

export const getPodcasts = () => apiFetch<Podcast[]>("/podcasts")
export const addPodcast = (rssUrl: string) =>
  apiFetch<Podcast>("/podcasts", { method: "POST", body: JSON.stringify({ rssUrl }) })
export const removePodcast = (id: number) =>
  apiFetch<void>(`/podcasts/${id}`, { method: "DELETE" })
