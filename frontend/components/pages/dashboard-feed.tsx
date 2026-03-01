import { useState, useMemo } from "react"
import { Search } from "lucide-react"
import { AppSidebar } from "@/components/app-sidebar"
import { EpisodeCard } from "@/components/episode-card"
import { SkeletonCard } from "@/components/skeleton-card"
import { Button } from "@/components/ui/button"

type Episode = {
  podcastId: string
  podcastName: string
  podcastCover: string
  episodeTitle: string
  publishedDate: string
  teaser: string
  unread: boolean
}

const episodes: Episode[] = [
  {
    podcastId: "syntax",
    podcastName: "Syntax FM",
    podcastCover: "https://picsum.photos/seed/syntax/64/64",
    episodeTitle: "Server Components are Finally Good - Here's What Changed",
    publishedDate: "Feb 27, 2026",
    teaser:
      "Wes and Scott break down the latest React Server Components improvements, including streaming, caching strategies, and why the DX is finally living up to the hype.",
    unread: true,
  },
  {
    podcastId: "changelog",
    podcastName: "Changelog",
    podcastCover: "https://picsum.photos/seed/changelog/64/64",
    episodeTitle: "The State of Open Source Funding in 2026",
    publishedDate: "Feb 25, 2026",
    teaser:
      "Adam and Jerod sit down with the maintainers of three major open source projects to discuss sustainability, corporate sponsorship, and what's actually working.",
    unread: true,
  },
  {
    podcastId: "jsparty",
    podcastName: "JS Party",
    podcastCover: "https://picsum.photos/seed/jsparty/64/64",
    episodeTitle: "Bun 2.0 Deep Dive: Is Node Dead Yet?",
    publishedDate: "Feb 22, 2026",
    teaser:
      "The panel takes Bun 2.0 for a spin, benchmarks it against Node and Deno, and debates whether it's ready for production workloads at scale.",
    unread: true,
  },
  {
    podcastId: "syntax",
    podcastName: "Syntax FM",
    podcastCover: "https://picsum.photos/seed/syntax/64/64",
    episodeTitle: "Why Your CSS Strategy is Wrong",
    publishedDate: "Feb 20, 2026",
    teaser:
      "A deep dive into modern CSS architecture patterns including cascade layers, container queries, and when you should and shouldn't reach for Tailwind.",
    unread: false,
  },
  {
    podcastId: "changelog",
    podcastName: "Changelog",
    podcastCover: "https://picsum.photos/seed/changelog/64/64",
    episodeTitle: "Building in Public: Lessons from 10 Indie Hackers",
    publishedDate: "Feb 18, 2026",
    teaser:
      "Ten indie hackers share their hard-won lessons on building products in public, marketing on a budget, and finding that elusive product-market fit.",
    unread: true,
  },
  {
    podcastId: "jsparty",
    podcastName: "JS Party",
    podcastCover: "https://picsum.photos/seed/jsparty/64/64",
    episodeTitle: "TypeScript 6.0 - Was It Worth the Wait?",
    publishedDate: "Feb 14, 2026",
    teaser:
      "A breakdown of every major feature in TypeScript 6.0 and whether the long development cycle was justified by the outcome.",
    unread: false,
  },
]

export function DashboardFeed() {
  const [selectedPodcast, setSelectedPodcast] = useState<string | null>(null)
  const [search, setSearch] = useState("")

  const filtered = useMemo(() => {
    let result = episodes
    if (selectedPodcast) {
      result = result.filter((ep) => ep.podcastId === selectedPodcast)
    }
    if (search.trim()) {
      const q = search.toLowerCase()
      result = result.filter(
        (ep) =>
          ep.episodeTitle.toLowerCase().includes(q) ||
          ep.podcastName.toLowerCase().includes(q) ||
          ep.teaser.toLowerCase().includes(q)
      )
    }
    return result
  }, [selectedPodcast, search])

  return (
    <div className="flex min-h-screen bg-background">
      <AppSidebar
        selectedPodcast={selectedPodcast}
        onSelectPodcast={setSelectedPodcast}
      />

      <main className="ml-60 flex-1 px-8 py-8">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight text-foreground">
            {selectedPodcast
              ? episodes.find((e) => e.podcastId === selectedPodcast)?.podcastName ?? "Feed"
              : "Your Feed"}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {selectedPodcast
              ? `Showing episodes from ${episodes.find((e) => e.podcastId === selectedPodcast)?.podcastName ?? "this podcast"}`
              : "Latest summaries from all your podcasts"}
          </p>
        </div>

        {/* Search */}
        <div className="relative mb-6 w-full max-w-md">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search episodes..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-9 w-full rounded-md border border-input bg-card pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground/60 focus:border-ring focus:outline-none focus:ring-2 focus:ring-ring/20"
          />
        </div>

        {/* Timeline episode list */}
        {filtered.length > 0 ? (
          <div className="flex flex-col gap-3">
            {filtered.map((ep) => (
              <EpisodeCard key={ep.episodeTitle} {...ep} />
            ))}

            {/* Skeleton for loading state demo */}
            <SkeletonCard />

            {/* Load more */}
            <Button variant="outline" className="mt-4 w-full">
              Load more
            </Button>
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <p className="text-sm text-muted-foreground">
              {search.trim()
                ? "No episodes match your search."
                : "No episodes yet."}
            </p>
          </div>
        )}
      </main>
    </div>
  )
}
