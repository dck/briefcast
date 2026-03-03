import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import { Check } from "lucide-react"
import { BriefcastLogo, BriefcastWordmark } from "@/components/briefcast-logo"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { useAuth } from "@/src/auth"

const waveformHeights = [30, 60, 85, 70, 45, 35, 75, 90, 50, 28, 65, 42, 80, 38, 70]

const processingSteps = [
  "Whisper transcription",
  "Show notes matched",
  "AI summarization",
]

const summaryBullets = [
  "LLMs shifting to autonomous multi-step agents.",
  "Architecture matters more than raw compute.",
  "Open-source closing gap faster than expected.",
]

const features = [
  {
    icon: "🎙️",
    title: "Whisper Transcription",
    description:
      "Episodes are transcribed with high accuracy, speaker labels, and timestamps - then matched against show notes for richer context.",
  },
  {
    icon: "⚡",
    title: "AI Summaries",
    description:
      "Key insights, quotes, and topics distilled into a structured 3-minute read. No sponsor segments, no fluff.",
  },
  {
    icon: "📡",
    title: "Automatic Tracking",
    description:
      "Subscribe to any podcast feed. The moment a new episode publishes, Briefcast picks it up and starts processing.",
    emphasized: true,
  },
]

const timeline = [
  {
    title: "Add a podcast",
    description:
      "Paste any podcast URL or search our directory. Briefcast starts watching immediately.",
  },
  {
    title: "New episode detected",
    description:
      "Our engine polls feeds continuously. The moment a new episode publishes, we grab the audio.",
  },
  {
    title: "Transcribe & summarize",
    description:
      "Whisper transcribes the audio, we match it with show notes, and AI extracts key insights.",
  },
  {
    title: "Brief delivered to you",
    description:
      "The summary lands in your Telegram, inbox, or RSS reader - ready to read in 3 minutes.",
  },
]

const channels = [
  {
    icon: "✈️",
    title: "Telegram",
    description: "Formatted briefs straight to your chat.",
    status: "Live",
    statusClass: "bg-primary/10 text-primary",
  },
  {
    icon: "📧",
    title: "Email",
    description: "HTML digests delivered to your inbox.",
    status: "Live",
    statusClass: "bg-primary/10 text-primary",
  },
  {
    icon: "📡",
    title: "RSS Feed",
    description: "Personal feed delivery for Reeder, Feedly, and any RSS reader.",
    status: "Coming soon",
    statusClass: "bg-[rgba(139,94,66,0.12)] text-[var(--brand-bark)]",
  },
  {
    icon: "🔗",
    title: "Read-it-later friendly",
    description:
      "Each brief will include a public link so read-it-later tools can fetch content without direct integrations.",
    status: "Planned",
    statusClass: "bg-[rgba(139,94,66,0.12)] text-[var(--brand-bark)]",
  },
]

const authProviders = [
  { name: "Google", href: "/api/auth/google", Icon: GoogleIcon },
  { name: "GitHub", href: "/api/auth/github", Icon: GitHubIcon },
  { name: "Yandex", href: "/api/auth/yandex", Icon: YandexIcon },
]

function GoogleIcon() {
  return (
    <svg viewBox="0 0 18 18" fill="none" className="h-4 w-4" aria-hidden="true">
      <path d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844a4.14 4.14 0 0 1-1.796 2.716v2.259h2.908C16.658 14.115 17.64 11.807 17.64 9.2Z" fill="#4285F4" />
      <path d="M9 18c2.43 0 4.467-.806 5.956-2.184l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 0 0 9 18Z" fill="#34A853" />
      <path d="M3.964 10.706A5.41 5.41 0 0 1 3.682 9c0-.593.102-1.17.282-1.706V4.962H.957A8.996 8.996 0 0 0 0 9c0 1.452.348 2.827.957 4.038l3.007-2.332Z" fill="#FBBC05" />
      <path d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 0 0 .957 4.962L3.964 7.294C4.672 5.163 6.656 3.58 9 3.58Z" fill="#EA4335" />
    </svg>
  )
}

function GitHubIcon() {
  return (
    <svg viewBox="0 0 18 18" fill="currentColor" className="h-4 w-4" aria-hidden="true">
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M9 0C4.03 0 0 4.03 0 9c0 3.98 2.584 7.354 6.168 8.546.45.082.616-.195.616-.433 0-.214-.008-.78-.012-1.53-2.504.544-3.032-1.207-3.032-1.207-.41-1.04-1-1.316-1-1.316-.816-.558.062-.547.062-.547.903.063 1.378.927 1.378.927.8 1.373 2.1.976 2.612.747.08-.58.313-.977.569-1.201-1.998-.227-4.1-1-4.1-4.44 0-.98.35-1.782.927-2.41-.093-.228-.402-1.14.088-2.374 0 0 .756-.242 2.476.923A8.63 8.63 0 0 1 9 4.133c.765.003 1.536.103 2.257.302 1.719-1.165 2.474-.923 2.474-.923.491 1.233.182 2.146.09 2.374.577.628.925 1.43.925 2.41 0 3.449-2.105 4.21-4.11 4.434.324.279.612.828.612 1.669 0 1.205-.011 2.176-.011 2.472 0 .24.163.52.62.432C15.418 16.35 18 12.978 18 9c0-4.97-4.03-9-9-9Z"
      />
    </svg>
  )
}

function YandexIcon() {
  return (
    <svg viewBox="0 0 18 18" fill="none" className="h-4 w-4" aria-hidden="true">
      <rect width="18" height="18" rx="4" fill="#FC3F1D" />
      <path d="M10.27 14H8.73V9.08H7.9C6.71 9.08 6.09 9.72 6.09 10.74c0 1.14.49 1.67 1.46 2.34L8.6 14H6.94l-1.2-1.08c-1.2-.87-1.81-1.75-1.81-3.08 0-1.77 1.15-2.9 3.08-2.9h1V4h1.54v10H10.27ZM10.27 7.08V5.08H9.44c-1.01 0-1.54.47-1.54 1.34 0 .9.5 1.35 1.48 1.35h.89V7.08Z" fill="white" />
    </svg>
  )
}

export function LandingPage() {
  const { user, loading } = useAuth()
  const navigate = useNavigate()
  const [signupOpen, setSignupOpen] = useState(false)

  useEffect(() => {
    if (!loading && user) navigate("/feed", { replace: true })
  }, [user, loading, navigate])

  return (
    <Dialog open={signupOpen} onOpenChange={setSignupOpen}>
      <div className="min-h-screen bg-background text-foreground">
        <header className="sticky top-0 z-40 border-b border-border/90 bg-background/90 backdrop-blur-xl">
          <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4 lg:px-12">
            <a href="#" className="flex items-center gap-2.5">
              <BriefcastLogo size={34} />
              <BriefcastWordmark className="text-lg" />
            </a>

            <div className="flex items-center gap-8">
              <nav className="hidden items-center gap-7 text-sm font-medium text-muted-foreground md:flex">
                <a href="#features" className="transition-colors hover:text-foreground">
                  Features
                </a>
                <a href="#how" className="transition-colors hover:text-foreground">
                  How it works
                </a>
              </nav>
              <DialogTrigger asChild>
                <button
                  type="button"
                  className="inline-flex h-9 items-center rounded-md bg-primary px-5 text-sm font-semibold text-primary-foreground transition-all hover:-translate-y-px hover:bg-primary/90 hover:shadow-[0_10px_24px_rgba(193,68,14,0.25)]"
                >
                  Sign Up
                </button>
              </DialogTrigger>
            </div>
          </div>
        </header>

        <main>
          <section className="mx-auto max-w-6xl px-6 pb-20 pt-16 text-center lg:px-12 lg:pt-24">
            <h1 className="mx-auto max-w-4xl font-display text-5xl font-bold leading-[1.05] tracking-[-0.04em] text-foreground sm:text-6xl lg:text-[4.2rem]">
              Your podcasts,
              <br />
              <span className="text-primary">condensed</span> to 3 minutes.
            </h1>

            <p className="mx-auto mt-5 max-w-2xl text-base leading-relaxed text-muted-foreground lg:text-[1.05rem]">
              Subscribe to any show. Briefcast transcribes every new episode and delivers an AI summary straight to Telegram, email, or RSS.
            </p>

            <div className="mx-auto mb-11 mt-10 max-w-5xl overflow-hidden rounded-[20px] border border-border bg-secondary shadow-[0_24px_64px_rgba(139,94,66,0.13)]">
              <div className="grid gap-4 px-5 py-6 md:grid-cols-[1fr_auto_1fr_auto_1fr] md:items-center md:gap-0 lg:px-7">
                <div className="text-left">
                  <p className="mb-2 text-[0.62rem] font-bold uppercase tracking-[0.08em] text-muted-foreground">
                    Episode · 2h 47m
                  </p>
                  <div className="rounded-xl border border-border bg-card p-3">
                    <div className="flex items-center gap-2.5">
                      <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[linear-gradient(135deg,#2C1A0E,#C1440E,#E07A30)] text-base">
                        🎙️
                      </div>
                      <div>
                        <p className="text-xs font-bold text-foreground">Lex Fridman Podcast</p>
                        <p className="text-[0.68rem] text-muted-foreground">2 hr 47 min</p>
                      </div>
                    </div>
                    <div className="mt-3 flex h-5 items-end gap-[3px]">
                      {waveformHeights.map((height, index) => (
                        <span
                          key={`${height}-${index}`}
                          className={`w-[3px] rounded-full ${
                            index % 3 === 1 ? "bg-[var(--brand-amber)]" : "bg-[var(--brand-parchment)]"
                          }`}
                          style={{ height: `${height}%` }}
                        />
                      ))}
                    </div>
                  </div>
                </div>

                <span className="hidden px-3 text-lg leading-none text-[var(--brand-parchment)] md:block">
                  →
                </span>

                <div className="text-left">
                  <p className="mb-2 text-[0.62rem] font-bold uppercase tracking-[0.08em] text-muted-foreground">
                    Briefcast processes
                  </p>
                  <div className="rounded-xl border border-border bg-card p-3">
                    <div className="flex flex-col gap-2">
                      {processingSteps.map((step) => (
                        <div key={step} className="flex items-center gap-2 text-xs font-medium text-foreground/85">
                          <span>{step}</span>
                          <span className="ml-auto inline-flex h-4 w-4 items-center justify-center rounded-full bg-emerald-600/15 text-emerald-700">
                            <Check className="h-2.5 w-2.5" />
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

                <span className="hidden px-3 text-lg leading-none text-[var(--brand-parchment)] md:block">
                  →
                </span>

                <div className="text-left">
                  <p className="mb-2 text-[0.62rem] font-bold uppercase tracking-[0.08em] text-muted-foreground">
                    Your brief · 3 min read
                  </p>
                  <div className="rounded-xl border border-border bg-card p-3">
                    <div className="flex flex-col gap-2">
                      {summaryBullets.map((item) => (
                        <div key={item} className="flex items-start gap-2 text-xs leading-relaxed text-[var(--brand-bark)]">
                          <span className="mt-1 h-1 w-1 shrink-0 rounded-full bg-[var(--brand-amber)]" />
                          <span>{item}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div id="signup" className="flex flex-wrap items-center justify-center gap-2.5">
              {authProviders.map(({ name, href, Icon }) => (
                <a
                  key={name}
                  href={href}
                  className="inline-flex h-11 items-center gap-2.5 rounded-lg border border-border bg-card px-5 text-sm font-semibold text-foreground transition-all hover:-translate-y-px hover:border-[var(--brand-sand)] hover:bg-secondary hover:shadow-[0_6px_18px_rgba(139,94,66,0.1)]"
                >
                  <Icon />
                  Continue with {name}
                </a>
              ))}
            </div>
            <p className="mt-4 text-xs text-muted-foreground">
              Or continue instantly with Google, GitHub, or Yandex to start your Briefcast account.
            </p>
          </section>

          <section id="features" className="px-6 py-20 lg:px-12">
            <div className="mx-auto max-w-6xl">
              <p className="mb-2 text-xs font-bold uppercase tracking-[0.12em] text-primary">
                Features
              </p>
              <h2 className="font-display text-4xl font-bold tracking-[-0.03em] text-foreground">
                How it saves you time
              </h2>
              <p className="mt-3 max-w-xl text-sm leading-relaxed text-muted-foreground">
                Everything is automatic. Add a show, and Briefcast does the rest.
              </p>

              <div className="mt-12 grid gap-4 md:grid-cols-3">
                {features.map((feature) => (
                  <article
                    key={feature.title}
                    className={`rounded-2xl border p-6 transition-all ${
                      feature.emphasized
                        ? "border-primary bg-primary text-primary-foreground shadow-[0_18px_40px_rgba(193,68,14,0.2)]"
                        : "border-border bg-card text-foreground hover:-translate-y-1 hover:shadow-[0_16px_36px_rgba(139,94,66,0.1)]"
                    }`}
                  >
                    <span
                      className={`mb-4 inline-flex h-11 w-11 items-center justify-center rounded-xl text-xl ${
                        feature.emphasized
                          ? "bg-white/15"
                          : "bg-[rgba(224,122,48,0.12)]"
                      }`}
                    >
                      {feature.icon}
                    </span>
                    <h3 className="text-sm font-bold">{feature.title}</h3>
                    <p
                      className={`mt-2 text-sm leading-relaxed ${
                        feature.emphasized ? "text-white/75" : "text-muted-foreground"
                      }`}
                    >
                      {feature.description}
                    </p>
                  </article>
                ))}
              </div>
            </div>
          </section>

          <section id="how" className="border-y border-border bg-secondary/80 px-6 py-20 lg:px-12">
            <div className="mx-auto max-w-6xl">
              <p className="mb-2 text-xs font-bold uppercase tracking-[0.12em] text-primary">
                How it works
              </p>
              <h2 className="font-display text-4xl font-bold tracking-[-0.03em] text-foreground">
                Subscribe once.
                <br />
                Read forever.
              </h2>
              <p className="mt-3 max-w-2xl text-sm leading-relaxed text-muted-foreground">
                Briefcast runs in the background and delivers summaries before you even know a new episode is out.
              </p>

              <div className="mt-12 grid gap-10 lg:grid-cols-[1.2fr_1fr]">
                <div className="space-y-2">
                  {timeline.map((step, index) => (
                    <div key={step.title} className="flex gap-4">
                      <div className="flex w-10 shrink-0 flex-col items-center">
                        <span className="flex h-9 w-9 items-center justify-center rounded-full bg-primary text-sm font-bold text-primary-foreground shadow-[0_6px_18px_rgba(193,68,14,0.28)]">
                          {index + 1}
                        </span>
                        {index < timeline.length - 1 && (
                          <span className="mt-1 h-12 w-px bg-gradient-to-b from-[var(--brand-parchment)] to-transparent" />
                        )}
                      </div>
                      <div className="pb-7">
                        <h3 className="font-display text-base font-semibold text-foreground">
                          {step.title}
                        </h3>
                        <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
                          {step.description}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>

                <div>
                  <p className="mb-4 text-[0.72rem] font-bold uppercase tracking-[0.08em] text-muted-foreground">
                    Deliver to where you already read
                  </p>
                  <div className="grid gap-3 sm:grid-cols-2">
                    {channels.map((channel) => (
                      <article key={channel.title} className="rounded-xl border border-border bg-card p-4">
                        <span className="mb-2 inline-flex h-9 w-9 items-center justify-center rounded-lg bg-[rgba(224,122,48,0.12)] text-base">
                          {channel.icon}
                        </span>
                        <h3 className="text-sm font-bold text-foreground">{channel.title}</h3>
                        <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
                          {channel.description}
                        </p>
                        <span
                          className={`mt-2 inline-block rounded px-2 py-0.5 text-[0.62rem] font-bold uppercase tracking-[0.06em] ${channel.statusClass}`}
                        >
                          {channel.status}
                        </span>
                      </article>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </section>
        </main>

        <footer className="flex flex-wrap items-center justify-between gap-3 bg-foreground px-6 py-7 lg:px-12">
          <a href="#" className="flex items-center gap-2">
            <BriefcastLogo size={28} />
            <BriefcastWordmark
              className="text-base text-white/90"
              accentClassName="text-[#E07A30]"
            />
          </a>
          <span className="text-xs text-white/35">
            © 2026 Briefcast · Built with ☕ and curiosity
          </span>
        </footer>
      </div>

      <DialogContent className="max-w-md border-border bg-card">
        <DialogHeader>
          <DialogTitle className="font-display text-xl">Sign UP</DialogTitle>
          <DialogDescription>
            Continue with your preferred provider and start receiving concise podcast briefs.
          </DialogDescription>
        </DialogHeader>
        <div className="mt-2 flex flex-col gap-2">
          {authProviders.map(({ name, href, Icon }) => (
            <Button key={name} variant="outline" className="h-11 justify-start gap-2.5 border-border bg-card" asChild>
              <a href={href} onClick={() => setSignupOpen(false)}>
                <Icon />
                Continue with {name}
              </a>
            </Button>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  )
}
