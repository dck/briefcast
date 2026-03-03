type Status = "pending" | "processing" | "done" | "failed" | "skipped"

const statusStyles: Record<Status, string> = {
  pending: "border border-border bg-muted text-muted-foreground",
  processing: "border border-primary/30 bg-primary/10 text-primary",
  done: "border border-emerald-700/25 bg-emerald-700/10 text-emerald-800",
  failed: "border border-destructive/35 bg-destructive/10 text-destructive",
  skipped: "border border-amber-700/30 bg-amber-600/10 text-amber-800",
}

const statusLabels: Record<Status, string> = {
  pending: "Pending",
  processing: "Processing",
  done: "Done",
  failed: "Failed",
  skipped: "Skipped",
}

export function StatusBadge({ status }: { status: Status }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${statusStyles[status]}`}
    >
      {statusLabels[status]}
    </span>
  )
}
