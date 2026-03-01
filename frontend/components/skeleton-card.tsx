export function SkeletonCard() {
  return (
    <div className="flex gap-4 rounded-lg border border-border bg-card p-4">
      {/* Cover art skeleton */}
      <div className="h-14 w-14 shrink-0 animate-pulse rounded-lg bg-muted" />

      {/* Content skeleton */}
      <div className="flex min-w-0 flex-1 flex-col gap-2.5">
        {/* Podcast name + date */}
        <div className="flex items-center gap-2">
          <div className="h-3 w-20 animate-pulse rounded bg-muted" />
          <div className="h-3 w-16 animate-pulse rounded bg-muted" />
        </div>
        {/* Episode title */}
        <div className="h-4 w-3/4 animate-pulse rounded bg-muted" />
        {/* Teaser line 1 */}
        <div className="h-3 w-full animate-pulse rounded bg-muted" />
        {/* Teaser line 2 */}
        <div className="h-3 w-5/6 animate-pulse rounded bg-muted" />
        {/* Action row */}
        <div className="mt-1 flex items-center gap-2">
          <div className="h-7 w-24 animate-pulse rounded-md bg-muted" />
          <div className="h-7 w-7 animate-pulse rounded-md bg-muted" />
        </div>
      </div>
    </div>
  )
}
