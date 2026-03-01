import { AdminSidebar } from "@/components/admin-sidebar"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from "@/components/ui/table"

type SessionRow = {
  token: string
  userName: string
  createdAt: string
  lastSeen: string
  expiresAt: string
}

const sessionRows: SessionRow[] = [
  {
    token: "sess_a1b2c3d4",
    userName: "Alex Johnson",
    createdAt: "Feb 28, 2026 09:12",
    lastSeen: "2 min ago",
    expiresAt: "Mar 7, 2026 09:12",
  },
  {
    token: "sess_e5f6g7h8",
    userName: "Sarah Chen",
    createdAt: "Feb 27, 2026 14:30",
    lastSeen: "1 hour ago",
    expiresAt: "Mar 6, 2026 14:30",
  },
  {
    token: "sess_i9j0k1l2",
    userName: "Mike Torres",
    createdAt: "Feb 25, 2026 08:45",
    lastSeen: "3 days ago",
    expiresAt: "Mar 4, 2026 08:45",
  },
  {
    token: "sess_m3n4o5p6",
    userName: "David Kim",
    createdAt: "Feb 28, 2026 11:00",
    lastSeen: "5 min ago",
    expiresAt: "Mar 7, 2026 11:00",
  },
  {
    token: "sess_q7r8s9t0",
    userName: "Alex Johnson",
    createdAt: "Feb 26, 2026 20:15",
    lastSeen: "1 day ago",
    expiresAt: "Mar 5, 2026 20:15",
  },
  {
    token: "sess_u1v2w3x4",
    userName: "Emma Wilson",
    createdAt: "Feb 14, 2026 10:00",
    lastSeen: "2 weeks ago",
    expiresAt: "Feb 21, 2026 10:00",
  },
]

export function AdminSessions() {
  return (
    <div className="flex min-h-screen bg-background">
      <AdminSidebar />

      <main className="ml-60 flex-1 px-8 py-8">
        <h1 className="mb-6 text-xl font-semibold tracking-tight text-foreground">
          Sessions
        </h1>

        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <Table>
            <TableHeader>
              <TableRow className="text-xs font-medium text-muted-foreground">
                <TableHead className="px-5">User</TableHead>
                <TableHead className="px-5">Created At</TableHead>
                <TableHead className="px-5">Last Seen</TableHead>
                <TableHead className="px-5">Expires At</TableHead>
                <TableHead className="px-5">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {sessionRows.map((session) => (
                <TableRow key={session.token}>
                  <TableCell className="px-5 font-medium text-foreground">
                    {session.userName}
                  </TableCell>
                  <TableCell className="px-5 text-muted-foreground">
                    {session.createdAt}
                  </TableCell>
                  <TableCell className="px-5 text-muted-foreground">
                    {session.lastSeen}
                  </TableCell>
                  <TableCell className="px-5 text-muted-foreground">
                    {session.expiresAt}
                  </TableCell>
                  <TableCell className="px-5">
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-7 text-xs"
                    >
                      Revoke
                    </Button>
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
