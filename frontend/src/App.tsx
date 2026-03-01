import { Routes, Route, Navigate } from "react-router-dom"
import { LandingPage } from "@/components/pages/landing-page"
import { DashboardFeed } from "@/components/pages/dashboard-feed"
import { EpisodeSummary } from "@/components/pages/episode-summary"
import { AdminDashboard } from "@/components/pages/admin-dashboard"
import { AdminSettings } from "@/components/pages/admin-settings"

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/feed" element={<DashboardFeed />} />
      <Route path="/episodes/:id" element={<EpisodeSummary />} />
      <Route path="/admin" element={<AdminDashboard />} />
      <Route path="/admin/settings" element={<AdminSettings />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
