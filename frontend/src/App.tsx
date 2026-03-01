import { Routes, Route, Navigate } from "react-router-dom"
import { AuthProvider, ProtectedRoute, AdminRoute } from "./auth"
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
    <AuthProvider>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/feed" element={<ProtectedRoute><DashboardFeed /></ProtectedRoute>} />
        <Route path="/saved" element={<ProtectedRoute><SavedPage /></ProtectedRoute>} />
        <Route path="/episodes/:id" element={<ProtectedRoute><EpisodeSummary /></ProtectedRoute>} />
        <Route path="/settings" element={<ProtectedRoute><SettingsPage /></ProtectedRoute>} />
        <Route path="/admin" element={<AdminRoute><AdminDashboard /></AdminRoute>} />
        <Route path="/admin/episodes" element={<AdminRoute><AdminEpisodes /></AdminRoute>} />
        <Route path="/admin/users" element={<AdminRoute><AdminUsers /></AdminRoute>} />
        <Route path="/admin/sessions" element={<AdminRoute><AdminSessions /></AdminRoute>} />
        <Route path="/admin/settings" element={<AdminRoute><AdminSettings /></AdminRoute>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AuthProvider>
  )
}
