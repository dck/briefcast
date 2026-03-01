import { Bookmark } from "lucide-react"

export function EpisodeCard({
  podcastName,
  podcastCover,
  episodeTitle,
  publishedDate,
  teaser,
  unread = false,
}: {
  podcastName: string
  podcastCover: string
  episodeTitle: string
  publishedDate: string
  teaser: string
  unread?: boolean
}) {
  return (
    <article className="group relative flex gap-4 rounded-lg border border-border bg-card p-4 transition-colors hover:border-primary/20 hover:bg-accent/40">
      {/* Cover art */}
      <img
        src={podcastCover}
        alt={`${podcastName} cover art`}
        width={56}
        height={56}
        className="h-14 w-14 shrink-0 rounded-lg object-cover"
      />

      {/* Content */}
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <div className="flex items-center gap-2">
          {unread && (
            <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-primary" aria-label="Unread" />
          )}
          <span className="text-xs font-medium text-muted-foreground">
            {podcastName}
          </span>
          <span className="text-xs text-muted-foreground/60">{publishedDate}</span>
        </div>

        <h3 className="text-sm font-semibold leading-snug text-card-foreground">
          {episodeTitle}
        </h3>

        <p className="mt-0.5 line-clamp-2 text-sm leading-relaxed text-muted-foreground">
          {teaser}
        </p>

        {/* Action row */}
        <div className="mt-2 flex items-center gap-2">
          <button
            type="button"
            className="inline-flex h-7 items-center rounded-md bg-primary/10 px-3 text-xs font-medium text-primary transition-colors hover:bg-primary/20"
          >
            Read summary
          </button>
          <button
            type="button"
            className="inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-card-foreground"
            aria-label="Bookmark episode"
          >
            <Bookmark className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
    </article>
  )
}
