import { Routes, Route, Navigate } from "react-router-dom"
import { LandingPage } from "@/components/pages/landing-page"
import { DashboardFeed } from "@/components/pages/dashboard-feed"
import { SavedPage } from "@/components/pages/saved-page"
import { SettingsPage } from "@/components/pages/settings-page"
import { EpisodeSummary } from "@/components/pages/episode-summary"
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
      <Route path="/settings" element={<SettingsPage />} />
      <Route path="/episodes/:id" element={<EpisodeSummary />} />
      <Route path="/admin" element={<AdminDashboard />} />
      <Route path="/admin/episodes" element={<AdminEpisodes />} />
      <Route path="/admin/users" element={<AdminUsers />} />
      <Route path="/admin/sessions" element={<AdminSessions />} />
      <Route path="/admin/settings" element={<AdminSettings />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
