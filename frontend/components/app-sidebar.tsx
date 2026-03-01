import {
  Rss,
  Settings,
  Shield,
  LogOut,
  Plus,
} from "lucide-react"
import { BriefcastLogo } from "@/components/briefcast-logo"

const navItems = [
  { label: "Feed", icon: Rss, href: "/" },
  { label: "Settings", icon: Settings, href: "/settings" },
  { label: "Admin", icon: Shield, href: "/admin" },
]

export type Podcast = {
  id: string
  name: string
  cover: string
  episodeCount: number
}

const podcasts: Podcast[] = [
  {
    id: "syntax",
    name: "Syntax FM",
    cover: "https://picsum.photos/seed/syntax/40/40",
    episodeCount: 12,
  },
  {
    id: "changelog",
    name: "Changelog",
    cover: "https://picsum.photos/seed/changelog/40/40",
    episodeCount: 8,
  },
  {
    id: "jsparty",
    name: "JS Party",
    cover: "https://picsum.photos/seed/jsparty/40/40",
    episodeCount: 5,
  },
]

export function AppSidebar({
  selectedPodcast,
  onSelectPodcast,
}: {
  selectedPodcast: string | null
  onSelectPodcast: (id: string | null) => void
}) {
  return (
    <aside className="fixed left-0 top-0 z-30 flex h-screen w-60 flex-col border-r border-border bg-sidebar">
      {/* Logo */}
      <div className="flex items-center gap-2.5 px-5 py-5">
        <BriefcastLogo size={22} className="text-primary" />
        <span className="text-base font-semibold tracking-tight text-sidebar-foreground">
          Briefcast
        </span>
      </div>

      {/* Navigation */}
      <nav className="flex flex-col gap-0.5 px-3">
        {navItems.map((item) => (
          <a
            key={item.label}
            href={item.href}
            className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
          >
            <item.icon className="h-4 w-4" />
            {item.label}
          </a>
        ))}
      </nav>

      {/* Podcasts section */}
      <div className="mt-6 flex flex-col px-3">
        <div className="mb-2 flex items-center justify-between px-3">
          <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Podcasts
          </span>
          <button
            type="button"
            className="inline-flex h-6 w-6 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
            aria-label="Add podcast"
          >
            <Plus className="h-3.5 w-3.5" />
          </button>
        </div>

        <div className="flex flex-col gap-0.5">
          {/* "All podcasts" option */}
          <button
            type="button"
            onClick={() => onSelectPodcast(null)}
            className={`flex items-center gap-3 rounded-md px-3 py-2 text-left transition-colors ${
              selectedPodcast === null
                ? "bg-accent text-accent-foreground"
                : "text-muted-foreground hover:bg-accent/60 hover:text-sidebar-foreground"
            }`}
          >
            <Rss className="h-4 w-4 shrink-0" />
            <span className="truncate text-sm font-medium">All podcasts</span>
          </button>

          {podcasts.map((podcast) => {
            const isActive = selectedPodcast === podcast.id
            return (
              <button
                key={podcast.id}
                type="button"
                onClick={() => onSelectPodcast(podcast.id)}
                className={`flex items-center gap-3 rounded-md px-3 py-2 text-left transition-colors ${
                  isActive
                    ? "bg-accent text-accent-foreground"
                    : "text-sidebar-foreground hover:bg-accent/60"
                }`}
              >
                <img
                  src={podcast.cover}
                  alt={`${podcast.name} cover`}
                  width={24}
                  height={24}
                  className="h-6 w-6 shrink-0 rounded object-cover"
                />
                <span className="flex-1 truncate text-sm font-medium">
                  {podcast.name}
                </span>
                <span className="text-xs text-muted-foreground">
                  {podcast.episodeCount}
                </span>
              </button>
            )
          })}
        </div>
      </div>

      {/* Spacer */}
      <div className="flex-1" />

      {/* User profile */}
      <div className="border-t border-border px-3 py-4">
        <div className="flex items-center gap-3 px-3">
          <img
            src="https://picsum.photos/seed/user/32/32"
            alt="User avatar"
            width={32}
            height={32}
            className="h-8 w-8 rounded-full object-cover"
          />
          <div className="flex min-w-0 flex-1 flex-col">
            <span className="truncate text-sm font-medium text-sidebar-foreground">
              Alex Johnson
            </span>
            <a
              href="/logout"
              className="flex items-center gap-1 text-xs text-muted-foreground transition-colors hover:text-destructive"
            >
              <LogOut className="h-3 w-3" />
              Logout
            </a>
          </div>
        </div>
      </div>
    </aside>
  )
}
