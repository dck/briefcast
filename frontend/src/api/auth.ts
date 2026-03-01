import { apiFetch } from "./client"
import type { User } from "./types"

export const getMe = () => apiFetch<User>("/auth/me")
export const logout = () => apiFetch<void>("/auth/logout", { method: "POST" })
