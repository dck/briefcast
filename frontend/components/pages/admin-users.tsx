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

type UserRow = {
  id: number
  name: string
  email: string
  oauthProvider: string
  joined: string
  podcasts: number
  telegram: string
  lastSeen: string
  isActive: boolean
}

const userRows: UserRow[] = [
  {
    id: 1,
    name: "Alex Johnson",
    email: "alex@example.com",
    oauthProvider: "Google",
    joined: "Jan 15, 2026",
    podcasts: 3,
    telegram: "@alexj",
    lastSeen: "2 min ago",
    isActive: true,
  },
  {
    id: 2,
    name: "Sarah Chen",
    email: "sarah@example.com",
    oauthProvider: "GitHub",
    joined: "Jan 22, 2026",
    podcasts: 5,
    telegram: "--",
    lastSeen: "1 hour ago",
    isActive: true,
  },
  {
    id: 3,
    name: "Mike Torres",
    email: "mike@example.com",
    oauthProvider: "Google",
    joined: "Feb 1, 2026",
    podcasts: 2,
    telegram: "@miket",
    lastSeen: "3 days ago",
    isActive: true,
  },
  {
    id: 4,
    name: "Emma Wilson",
    email: "emma@example.com",
    oauthProvider: "Yandex",
    joined: "Feb 10, 2026",
    podcasts: 1,
    telegram: "--",
    lastSeen: "2 weeks ago",
    isActive: false,
  },
  {
    id: 5,
    name: "David Kim",
    email: "david@example.com",
    oauthProvider: "GitHub",
    joined: "Feb 18, 2026",
    podcasts: 4,
    telegram: "@davidk",
    lastSeen: "5 min ago",
    isActive: true,
  },
]

export function AdminUsers() {
  return (
    <div className="flex min-h-screen bg-background">
      <AdminSidebar />

      <main className="ml-60 flex-1 px-8 py-8">
        <h1 className="mb-6 text-xl font-semibold tracking-tight text-foreground">
          Users
        </h1>

        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <Table>
            <TableHeader>
              <TableRow className="text-xs font-medium text-muted-foreground">
                <TableHead className="px-5">Name</TableHead>
                <TableHead className="px-5">Email</TableHead>
                <TableHead className="px-5">OAuth Provider</TableHead>
                <TableHead className="px-5">Joined</TableHead>
                <TableHead className="px-5 text-center">Podcasts</TableHead>
                <TableHead className="px-5">Telegram</TableHead>
                <TableHead className="px-5">Last Seen</TableHead>
                <TableHead className="px-5">Status</TableHead>
                <TableHead className="px-5">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {userRows.map((user) => (
                <TableRow key={user.id}>
                  <TableCell className="px-5 font-medium text-foreground">
                    {user.name}
                  </TableCell>
                  <TableCell className="px-5 text-muted-foreground">
                    {user.email}
                  </TableCell>
                  <TableCell className="px-5 text-muted-foreground">
                    {user.oauthProvider}
                  </TableCell>
                  <TableCell className="px-5 text-muted-foreground">
                    {user.joined}
                  </TableCell>
                  <TableCell className="px-5 text-center text-muted-foreground">
                    {user.podcasts}
                  </TableCell>
                  <TableCell className="px-5 text-muted-foreground">
                    {user.telegram}
                  </TableCell>
                  <TableCell className="px-5 text-muted-foreground">
                    {user.lastSeen}
                  </TableCell>
                  <TableCell className="px-5">
                    <span
                      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                        user.isActive
                          ? "border border-primary/30 bg-primary/10 text-primary"
                          : "border border-border bg-muted text-muted-foreground"
                      }`}
                    >
                      {user.isActive ? "Active" : "Inactive"}
                    </span>
                  </TableCell>
                  <TableCell className="px-5">
                    {user.isActive && (
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-7 text-xs"
                      >
                        Deactivate
                      </Button>
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
