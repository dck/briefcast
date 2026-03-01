import { apiFetch } from "./client"
import type { User } from "./types"

export const getSettings = () => apiFetch<User>("/settings")
export const updateSettings = (data: {
  telegramChatId?: string
  email?: string
  notifyTelegram?: boolean
  notifyEmail?: boolean
}) => apiFetch<void>("/settings", { method: "PUT", body: JSON.stringify(data) })
