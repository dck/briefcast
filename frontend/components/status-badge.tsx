type Status = "pending" | "processing" | "done" | "failed" | "skipped"

const statusStyles: Record<Status, string> = {
  pending: "bg-gray-500 text-white",
  processing: "bg-blue-500 text-white animate-pulse",
  done: "bg-green-600 text-white",
  failed: "bg-red-600 text-white",
  skipped: "bg-yellow-500 text-white",
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
