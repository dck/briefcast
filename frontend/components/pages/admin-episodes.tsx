import { useState } from "react"
import { RotateCcw, SkipForward } from "lucide-react"
import { AdminSidebar } from "@/components/admin-sidebar"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from "@/components/ui/table"

type EpisodeRow = {
  id: number
  podcastName: string
  podcastCover: string
  episodeTitle: string
  status: "pending" | "processing" | "done" | "failed" | "skipped"
  currentStep: string
  retries: number
  lastError: string
  elapsed: string
}

const episodeRows: EpisodeRow[] = [
  {
    id: 1,
    podcastName: "Syntax FM",
    podcastCover: "https://picsum.photos/seed/syntax/32/32",
    episodeTitle: "Server Components are Finally Good",
    status: "done",
    currentStep: "complete",
    retries: 0,
    lastError: "",
    elapsed: "4m 12s",
  },
  {
    id: 2,
    podcastName: "Changelog",
    podcastCover: "https://picsum.photos/seed/changelog/32/32",
    episodeTitle: "Open Source Funding in 2026",
    status: "processing",
    currentStep: "transcribe",
    retries: 0,
    lastError: "",
    elapsed: "1m 34s",
  },
  {
    id: 3,
    podcastName: "JS Party",
    podcastCover: "https://picsum.photos/seed/jsparty/32/32",
    episodeTitle: "Bun 2.0 Deep Dive: Is Node Dead Yet?",
    status: "processing",
    currentStep: "summarize",
    retries: 0,
    lastError: "",
    elapsed: "0m 48s",
  },
  {
    id: 4,
    podcastName: "Syntax FM",
    podcastCover: "https://picsum.photos/seed/syntax/32/32",
    episodeTitle: "CSS Strategy Deep Dive",
    status: "failed",
    currentStep: "transcribe",
    retries: 2,
    lastError: "Groq API rate limit exceeded: 429 Too Many Requests",
    elapsed: "6m 01s",
  },
  {
    id: 5,
    podcastName: "Changelog",
    podcastCover: "https://picsum.photos/seed/changelog/32/32",
    episodeTitle: "Building in Public: Lessons Learned",
    status: "pending",
    currentStep: "--",
    retries: 0,
    lastError: "",
    elapsed: "--",
  },
  {
    id: 6,
    podcastName: "JS Party",
    podcastCover: "https://picsum.photos/seed/jsparty/32/32",
    episodeTitle: "TypeScript 6.0 - Was It Worth the Wait?",
    status: "done",
    currentStep: "complete",
    retries: 1,
    lastError: "",
    elapsed: "5m 22s",
  },
  {
    id: 7,
    podcastName: "Syntax FM",
    podcastCover: "https://picsum.photos/seed/syntax/32/32",
    episodeTitle: "Why Your CSS Strategy is Wrong",
    status: "skipped",
    currentStep: "--",
    retries: 0,
    lastError: "Episode is a trailer, too short to process",
    elapsed: "--",
  },
  {
    id: 8,
    podcastName: "Changelog",
    podcastCover: "https://picsum.photos/seed/changelog/32/32",
    episodeTitle: "Postgres vs. SQLite: The Great Debate",
    status: "failed",
    currentStep: "download",
    retries: 3,
    lastError: "Connection timeout after 30s: ETIMEDOUT",
    elapsed: "2m 15s",
  },
]

const statusFilters = [
  { label: "All", value: "all" },
  { label: "Pending", value: "pending" },
  { label: "Processing", value: "processing" },
  { label: "Done", value: "done" },
  { label: "Failed", value: "failed" },
  { label: "Skipped", value: "skipped" },
] as const

export function AdminEpisodes() {
  const [activeFilter, setActiveFilter] = useState<string>("all")

  const filtered =
    activeFilter === "all"
      ? episodeRows
      : episodeRows.filter((row) => row.status === activeFilter)

  return (
    <div className="flex min-h-screen bg-background">
      <AdminSidebar />

      <main className="ml-60 flex-1 px-8 py-8">
        <h1 className="mb-6 text-xl font-semibold tracking-tight text-foreground">
          Episodes
        </h1>

        {/* Status filter tabs */}
        <div className="mb-4 flex gap-1">
          {statusFilters.map((filter) => (
            <button
              key={filter.value}
              type="button"
              onClick={() => setActiveFilter(filter.value)}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                activeFilter === filter.value
                  ? "bg-accent text-accent-foreground"
                  : "text-muted-foreground hover:bg-accent/60 hover:text-foreground"
              }`}
            >
              {filter.label}
            </button>
          ))}
        </div>

        {/* Episodes table */}
        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <Table>
            <TableHeader>
              <TableRow className="text-xs font-medium text-muted-foreground">
                <TableHead className="px-5">Podcast</TableHead>
                <TableHead className="px-5">Title</TableHead>
                <TableHead className="px-5">Status</TableHead>
                <TableHead className="px-5">Current Step</TableHead>
                <TableHead className="px-5 text-center">Retries</TableHead>
                <TableHead className="px-5">Last Error</TableHead>
                <TableHead className="px-5">Elapsed</TableHead>
                <TableHead className="px-5">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((row) => (
                <TableRow
                  key={row.id}
                  className={
                    row.status === "failed" ? "bg-destructive/5" : ""
                  }
                >
                  <TableCell className="px-5">
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
                  </TableCell>
                  <TableCell className="max-w-[200px] truncate px-5 text-foreground">
                    {row.episodeTitle}
                  </TableCell>
                  <TableCell className="px-5">
                    <StatusBadge status={row.status} />
                  </TableCell>
                  <TableCell className="px-5 text-muted-foreground">
                    {row.currentStep}
                  </TableCell>
                  <TableCell className="px-5 text-center text-muted-foreground">
                    {row.retries}
                  </TableCell>
                  <TableCell className="max-w-[180px] truncate px-5 text-muted-foreground">
                    {row.lastError || "--"}
                  </TableCell>
                  <TableCell className="px-5 text-muted-foreground">
                    {row.elapsed}
                  </TableCell>
                  <TableCell className="px-5">
                    {row.status === "failed" && (
                      <div className="flex items-center gap-1">
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-7 gap-1 text-xs"
                        >
                          <RotateCcw className="h-3 w-3" />
                          Retry
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-7 gap-1 text-xs"
                        >
                          <SkipForward className="h-3 w-3" />
                          Skip
                        </Button>
                      </div>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </main>
    </div>
  )
}
