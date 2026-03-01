import { useState } from "react"
import {
  ListMusic,
  Users,
  Cpu,
  Activity,
  RotateCcw,
} from "lucide-react"
import { AdminSidebar } from "@/components/admin-sidebar"
import { StatsCard } from "@/components/stats-card"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
} from "@/components/ui/card"

type QueueRow = {
  podcastName: string
  podcastCover: string
  episodeTitle: string
  status: "pending" | "processing" | "done" | "failed"
  retries: number
  elapsed: string
}

const queueRows: QueueRow[] = [
  {
    podcastName: "Syntax FM",
    podcastCover: "https://picsum.photos/seed/syntax/32/32",
    episodeTitle: "Server Components are Finally Good",
    status: "done",
    retries: 0,
    elapsed: "4m 12s",
  },
  {
    podcastName: "Changelog",
    podcastCover: "https://picsum.photos/seed/changelog/32/32",
    episodeTitle: "Open Source Funding in 2026",
    status: "processing",
    retries: 0,
    elapsed: "1m 34s",
  },
  {
    podcastName: "JS Party",
    podcastCover: "https://picsum.photos/seed/jsparty/32/32",
    episodeTitle: "Bun 2.0 Deep Dive",
    status: "processing",
    retries: 0,
    elapsed: "0m 48s",
  },
  {
    podcastName: "Syntax FM",
    podcastCover: "https://picsum.photos/seed/syntax/32/32",
    episodeTitle: "CSS Strategy Deep Dive",
    status: "failed",
    retries: 2,
    elapsed: "6m 01s",
  },
  {
    podcastName: "Changelog",
    podcastCover: "https://picsum.photos/seed/changelog/32/32",
    episodeTitle: "Building in Public: Lessons Learned",
    status: "pending",
    retries: 0,
    elapsed: "--",
  },
]

// Small inline bar chart for system health
function MiniBar({ values, max }: { values: number[]; max: number }) {
  return (
    <div className="flex items-end gap-0.5" style={{ height: 32 }}>
      {values.map((v, i) => (
        <div
          key={i}
          className="w-2 rounded-t bg-primary/60"
          style={{ height: `${(v / max) * 100}%` }}
        />
      ))}
    </div>
  )
}

export function AdminDashboard() {
  const [activeSection, setActiveSection] = useState("overview")

  return (
    <div className="flex min-h-screen bg-background">
      <AdminSidebar
        activeSection={activeSection}
        onSectionChange={setActiveSection}
      />

      <main className="ml-60 flex-1 px-8 py-8">
        <h1 className="mb-6 text-xl font-semibold tracking-tight text-foreground">
          Admin Overview
        </h1>

        {/* Stats row -- first two are skeleton, last two are real */}
        <div className="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {/* Skeleton cards */}
          <div className="rounded-lg border border-border bg-card p-5 shadow-sm">
            <div className="flex items-start justify-between">
              <div className="flex flex-col gap-2">
                <div className="h-8 w-16 animate-pulse rounded bg-muted" />
                <div className="h-4 w-24 animate-pulse rounded bg-muted" />
              </div>
              <div className="h-5 w-5 animate-pulse rounded bg-muted" />
            </div>
          </div>
          <div className="rounded-lg border border-border bg-card p-5 shadow-sm">
            <div className="flex items-start justify-between">
              <div className="flex flex-col gap-2">
                <div className="h-8 w-16 animate-pulse rounded bg-muted" />
                <div className="h-4 w-24 animate-pulse rounded bg-muted" />
              </div>
              <div className="h-5 w-5 animate-pulse rounded bg-muted" />
            </div>
          </div>
          <StatsCard
            value="14,320"
            label="Groq tokens today"
            icon={<Cpu className="h-5 w-5" />}
          />
          <StatsCard
            value="42"
            label="Active users"
            icon={<Users className="h-5 w-5" />}
          />
        </div>

        {/* Main content: table + right panel */}
        <div className="flex gap-6">
          {/* Episode queue table (2/3) */}
          <div className="flex-[2] overflow-hidden rounded-lg border border-border bg-card">
            <div className="border-b border-border px-5 py-3">
              <h2 className="text-sm font-semibold text-card-foreground">
                Episode Queue
              </h2>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-left text-xs font-medium text-muted-foreground">
                    <th className="px-5 py-2.5">Podcast</th>
                    <th className="px-5 py-2.5">Episode</th>
                    <th className="px-5 py-2.5">Status</th>
                    <th className="px-5 py-2.5 text-center">Retries</th>
                    <th className="px-5 py-2.5">Elapsed</th>
                    <th className="px-5 py-2.5">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {queueRows.map((row, i) => (
                    <tr
                      key={i}
                      className={`border-b border-border last:border-0 ${
                        row.status === "failed" ? "bg-destructive/5" : ""
                      }`}
                    >
                      <td className="px-5 py-3">
                        <div className="flex items-center gap-2">
                          <img
                            src={row.podcastCover}
                            alt={row.podcastName}
                            width={24}
                            height={24}
                            className="h-6 w-6 rounded object-cover"
                          />
                          <span className="text-foreground">{row.podcastName}</span>
                        </div>
                      </td>
                      <td className="max-w-[200px] truncate px-5 py-3 text-foreground">
                        {row.episodeTitle}
                      </td>
                      <td className="px-5 py-3">
                        <StatusBadge status={row.status} />
                      </td>
                      <td className="px-5 py-3 text-center text-muted-foreground">
                        {row.retries}
                      </td>
                      <td className="px-5 py-3 text-muted-foreground">
                        {row.elapsed}
                      </td>
                      <td className="px-5 py-3">
                        {row.status === "failed" && (
                          <Button variant="outline" size="sm" className="h-7 gap-1 text-xs">
                            <RotateCcw className="h-3 w-3" />
                            Retry
                          </Button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Right panel: System health (1/3) */}
          <div className="flex flex-[1] flex-col gap-4">
            {/* Worker Status */}
            <Card className="gap-4 py-4">
              <CardHeader className="px-5 py-0">
                <CardTitle className="text-sm">Worker Status</CardTitle>
              </CardHeader>
              <CardContent className="px-5 py-0">
                <div className="flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-green-500" />
                  <span className="text-sm text-foreground">Running</span>
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  Last heartbeat: 12s ago
                </p>
              </CardContent>
            </Card>

            {/* RSS Poller */}
            <Card className="gap-4 py-4">
              <CardHeader className="px-5 py-0">
                <CardTitle className="text-sm">RSS Poller</CardTitle>
              </CardHeader>
              <CardContent className="px-5 py-0">
                <div className="flex flex-col gap-1 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Last run</span>
                    <span className="text-foreground">3 min ago</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Next run</span>
                    <span className="text-foreground">in 2 min</span>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Groq API */}
            <Card className="gap-4 py-4">
              <CardHeader className="px-5 py-0">
                <CardTitle className="text-sm">Groq API</CardTitle>
              </CardHeader>
              <CardContent className="px-5 py-0">
                <div className="flex flex-col gap-1 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Requests today</span>
                    <span className="text-foreground">128</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Tokens</span>
                    <span className="text-foreground">14,320</span>
                  </div>
                </div>
                <div className="mt-3">
                  <MiniBar
                    values={[4, 8, 6, 12, 9, 15, 11, 14, 7, 10, 13, 8]}
                    max={16}
                  />
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </div>
  )
}
