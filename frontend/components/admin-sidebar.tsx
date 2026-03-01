import {
  LayoutDashboard,
  ListMusic,
  Users,
  Activity,
  Settings,
  LogOut,
} from "lucide-react"
import { Link, useLocation } from "react-router-dom"
import { BriefcastLogo } from "@/components/briefcast-logo"
import { useAuth } from "@/src/auth"

const adminNavItems = [
  { label: "Overview", icon: LayoutDashboard, href: "/admin" },
  { label: "Episodes", icon: ListMusic, href: "/admin/episodes" },
  { label: "Users", icon: Users, href: "/admin/users" },
  { label: "Sessions", icon: Activity, href: "/admin/sessions" },
  { label: "Settings", icon: Settings, href: "/admin/settings" },
]

export function AdminSidebar() {
  const location = useLocation()

  return (
    <aside className="fixed left-0 top-0 z-30 flex h-screen w-60 flex-col border-r border-border bg-sidebar">
      {/* Logo */}
      <div className="flex items-center gap-2.5 px-5 py-5">
        <BriefcastLogo size={22} className="text-primary" />
        <span className="text-base font-semibold tracking-tight text-sidebar-foreground">
          Briefcast
        </span>
      </div>

      {/* Admin nav */}
      <nav className="flex flex-col gap-0.5 px-3">
        <span className="mb-2 px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Admin
        </span>
        {adminNavItems.map((item) => {
          const isActive = location.pathname === item.href
          return (
            <Link
              key={item.href}
              to={item.href}
              className={`flex items-center gap-3 rounded-md px-3 py-2 text-left text-sm font-medium transition-colors ${
                isActive
                  ? "bg-accent text-accent-foreground"
                  : "text-muted-foreground hover:bg-accent/60 hover:text-sidebar-foreground"
              }`}
            >
              <item.icon className="h-4 w-4" />
              {item.label}
            </Link>
          )
        })}
      </nav>

      {/* Spacer */}
      <div className="flex-1" />

      {/* User profile */}
      <AdminUserProfile />
    </aside>
  )
}

function AdminUserProfile() {
  const { user, logout } = useAuth()

  return (
    <div className="border-t border-border px-3 py-4">
      <div className="flex items-center gap-3 px-3">
        <img
          src={user?.avatarUrl || "https://picsum.photos/seed/admin/32/32"}
          alt="Admin avatar"
          width={32}
          height={32}
          className="h-8 w-8 rounded-full object-cover"
        />
        <div className="flex min-w-0 flex-1 flex-col">
          <span className="truncate text-sm font-medium text-sidebar-foreground">
            {user?.name || "Admin"}
          </span>
          <button
            type="button"
            onClick={logout}
            className="flex items-center gap-1 text-xs text-muted-foreground transition-colors hover:text-destructive"
          >
            <LogOut className="h-3 w-3" />
            Logout
          </button>
        </div>
      </div>
    </div>
  )
}
