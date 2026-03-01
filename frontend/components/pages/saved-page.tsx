import { useState } from "react"
import { BookmarkCheck } from "lucide-react"
import { AppSidebar } from "@/components/app-sidebar"
import { EpisodeCard } from "@/components/episode-card"
import { Button } from "@/components/ui/button"

type SavedEpisode = {
  podcastId: string
  podcastName: string
  podcastCover: string
  episodeTitle: string
  publishedDate: string
  teaser: string
  unread: boolean
}

const savedEpisodes: SavedEpisode[] = [
  {
    podcastId: "syntax",
    podcastName: "Syntax FM",
    podcastCover: "https://picsum.photos/seed/syntax/64/64",
    episodeTitle: "Server Components are Finally Good - Here's What Changed",
    publishedDate: "Feb 27, 2026",
    teaser:
      "Wes and Scott break down the latest React Server Components improvements, including streaming, caching strategies, and why the DX is finally living up to the hype.",
    unread: false,
  },
  {
    podcastId: "changelog",
    podcastName: "Changelog",
    podcastCover: "https://picsum.photos/seed/changelog/64/64",
    episodeTitle: "The State of Open Source Funding in 2026",
    publishedDate: "Feb 25, 2026",
    teaser:
      "Adam and Jerod sit down with the maintainers of three major open source projects to discuss sustainability, corporate sponsorship, and what's actually working.",
    unread: false,
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

export function SavedPage() {
  const [selectedPodcast, setSelectedPodcast] = useState<string | null>(null)

  const filtered = selectedPodcast
    ? savedEpisodes.filter((ep) => ep.podcastId === selectedPodcast)
    : savedEpisodes

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
            Saved Episodes
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Episodes you've bookmarked for later
          </p>
        </div>

        {/* Bookmarked episode list */}
        {filtered.length > 0 ? (
          <div className="flex flex-col gap-3">
            {filtered.map((ep) => (
              <EpisodeCard key={ep.episodeTitle} {...ep} />
            ))}

            {/* Load more */}
            <Button variant="outline" className="mt-4 w-full">
              Load more
            </Button>
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <BookmarkCheck className="mb-3 h-10 w-10 text-muted-foreground/40" />
            <p className="text-sm font-medium text-foreground">
              No saved episodes yet
            </p>
            <p className="mt-1 text-sm text-muted-foreground">
              Bookmark episodes from your feed to find them here later.
            </p>
          </div>
        )}
      </main>
    </div>
  )
}
