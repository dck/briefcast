import type { ReactNode } from "react"

export function StatsCard({
  value,
  label,
  icon,
}: {
  value: string
  label: string
  icon: ReactNode
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-5 shadow-sm">
      <div className="flex items-start justify-between">
        <div className="flex flex-col gap-1">
          <span className="text-3xl font-bold tracking-tight text-card-foreground">
            {value}
          </span>
          <span className="text-sm text-muted-foreground">{label}</span>
        </div>
        <div className="text-primary">{icon}</div>
      </div>
    </div>
  )
}
