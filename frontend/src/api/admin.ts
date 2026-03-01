import { apiFetch } from "./client"
import type { AdminStats, AdminEpisode, AdminUser, Session } from "./types"

export const getStats = () => apiFetch<AdminStats>("/admin/stats")
export const getAdminEpisodes = (status?: string) =>
  apiFetch<AdminEpisode[]>(`/admin/episodes${status ? `?status=${status}` : ""}`)
export const retryEpisode = (id: number) =>
  apiFetch<void>(`/admin/episodes/${id}/retry`, { method: "POST" })
export const retryEpisodeAll = (id: number) =>
  apiFetch<void>(`/admin/episodes/${id}/retry-all`, { method: "POST" })
export const skipEpisode = (id: number) =>
  apiFetch<void>(`/admin/episodes/${id}/skip`, { method: "POST" })
export const getUsers = () => apiFetch<AdminUser[]>("/admin/users")
export const deactivateUser = (id: number) =>
  apiFetch<void>(`/admin/users/${id}/deactivate`, { method: "POST" })
export const getSessions = () => apiFetch<Session[]>("/admin/sessions")
export const revokeSession = (token: string) =>
  apiFetch<void>(`/admin/sessions/${token}`, { method: "DELETE" })
export const getAdminSettings = () => apiFetch<Record<string, string>>("/admin/settings")
export const updateAdminSettings = (settings: Record<string, string>) =>
  apiFetch<void>("/admin/settings", { method: "PUT", body: JSON.stringify(settings) })
export const resumeProcessing = () =>
  apiFetch<void>("/admin/processing/resume", { method: "POST" })
