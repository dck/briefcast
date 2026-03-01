import { apiFetch } from "./client"
import type { Episode } from "./types"

export const getFeed = (page: number) =>
  apiFetch<{ episodes: Episode[]; hasMore: boolean }>(`/feed?page=${page}`)

export const getSaved = (page: number) =>
  apiFetch<{ episodes: Episode[]; hasMore: boolean }>(`/saved?page=${page}`)
