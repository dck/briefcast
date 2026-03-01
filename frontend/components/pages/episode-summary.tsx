import { useState } from "react"
import {
  Bookmark,
  Share2,
  ExternalLink,
  Copy,
  Check,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { BriefcastLogo } from "@/components/briefcast-logo"

const topicSections = [
  {
    title: "The Evolution of Server Components",
    body: "Server Components have undergone a significant transformation since their initial release. The panel discusses how the mental model has shifted from \"components that run on the server\" to \"components that define the data boundary.\" Wes points out that the biggest change is not technical but conceptual -- developers now think about data flow first and rendering second. Scott adds that the new streaming primitives make progressive loading feel native rather than bolted on.",
  },
  {
    title: "Caching Strategies That Actually Work",
    body: "The conversation shifts to caching, where both hosts agree the ecosystem has finally converged on sensible defaults. The new `use cache` directive paired with `cacheLife` profiles removes much of the guesswork. They walk through a real-world example of caching a product catalog page with stale-while-revalidate semantics, showing how a single annotation replaces what used to be dozens of lines of manual cache management.",
  },
  {
    title: "Developer Experience Improvements",
    body: "Perhaps the most celebrated change is the dramatic improvement in error messages and dev tooling. The new error overlay not only shows what went wrong but suggests fixes. Hot module replacement now works seamlessly across the server-client boundary, and the built-in profiler highlights which components are server-rendered vs. client-rendered with color-coded overlays.",
  },
]

const quotes = [
  {
    speaker: "Wes Bos",
    text: "I used to tell people to avoid Server Components until they matured. I can't say that anymore -- they're genuinely good now, and the DX gap with client components has basically closed.",
  },
  {
    speaker: "Scott Tolinski",
    text: "The caching story was the missing piece. Once you don't have to think about caching, you stop thinking about Server Components as \"different\" and they just become... components.",
  },
]

const references = [
  {
    name: "Next.js 16 Cache Components RFC",
    url: "https://nextjs.org",
    description: "Official RFC detailing the new caching architecture",
  },
  {
    name: "React Server Components Spec",
    url: "https://react.dev",
    description: "Updated specification for RSC wire format and streaming protocol",
  },
  {
    name: "Vercel Caching Deep Dive",
    url: "https://vercel.com",
    description: "Blog post walking through real-world caching patterns",
  },
]

export function EpisodeSummary() {
  const [copied, setCopied] = useState(false)
  const shareUrl = "https://briefcast.app/episodes/rsc-finally-good"

  function handleCopy() {
    navigator.clipboard.writeText(shareUrl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Top bar */}
      <header className="flex items-center justify-between border-b border-border px-6 py-3">
        <div className="flex items-center gap-2.5">
          <BriefcastLogo size={20} className="text-primary" />
          <span className="text-base font-semibold tracking-tight text-foreground">
            Briefcast
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm">
            <Bookmark className="h-4 w-4" />
            Bookmark
          </Button>
          <Button variant="ghost" size="sm">
            <Share2 className="h-4 w-4" />
            Share
          </Button>
        </div>
      </header>

      {/* Article content */}
      <article className="mx-auto max-w-[680px] px-6 py-10">
        {/* Episode header */}
        <div className="mb-8">
          <div className="mb-3 flex items-center gap-3">
            <img
              src="https://picsum.photos/seed/syntax/40/40"
              alt="Syntax FM cover"
              width={40}
              height={40}
              className="h-10 w-10 rounded-lg object-cover"
            />
            <span className="text-sm text-muted-foreground">Syntax FM</span>
          </div>
          <h1 className="text-balance text-2xl font-bold leading-snug tracking-tight text-foreground lg:text-3xl">
            Server Components are Finally Good - Here&apos;s What Changed
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">Feb 27, 2026</p>
          <Separator className="mt-6" />
        </div>

        {/* Overview */}
        <p className="mb-10 text-lg leading-relaxed text-foreground/90">
          In this episode, Wes and Scott revisit React Server Components after
          months of rapid iteration. They cover the three biggest changes --
          streaming primitives, the new caching layer, and dramatically improved
          developer tooling -- and explain why they believe RSC is now ready for
          mainstream adoption.
        </p>

        {/* Topic sections */}
        {topicSections.map((section) => (
          <section key={section.title} className="mb-10">
            <h2 className="mb-3 text-lg font-semibold text-primary">
              {section.title}
            </h2>
            <p className="text-base leading-relaxed text-foreground/85">
              {section.body}
            </p>
          </section>
        ))}

        {/* Key Opinions & Takes */}
        <section className="mb-10">
          <h2 className="mb-4 text-lg font-semibold text-primary">
            Key Opinions & Takes
          </h2>
          <div className="flex flex-col gap-4">
            {quotes.map((q) => (
              <blockquote
                key={q.speaker}
                className="rounded-r-lg border-l-4 border-primary/40 bg-accent/60 py-4 pl-5 pr-4"
              >
                <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-foreground/70">
                  {q.speaker}
                </p>
                <p className="text-base italic leading-relaxed text-foreground/85">
                  &ldquo;{q.text}&rdquo;
                </p>
              </blockquote>
            ))}
          </div>
        </section>

        {/* References */}
        <section className="mb-10">
          <h2 className="mb-4 text-lg font-semibold text-primary">
            References
          </h2>
          <ul className="flex flex-col gap-3">
            {references.map((ref) => (
              <li key={ref.name} className="flex items-start gap-2">
                <ExternalLink className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
                <div>
                  <a
                    href={ref.url}
                    className="text-sm font-medium text-primary hover:underline"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    {ref.name}
                  </a>
                  <p className="text-sm text-muted-foreground">
                    {ref.description}
                  </p>
                </div>
              </li>
            ))}
          </ul>
        </section>

        <Separator className="mb-8" />

        {/* Bottom actions */}
        <div className="flex flex-col gap-4">
          <a
            href="#"
            className="text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            Listen to full episode &rarr;
          </a>
          <div className="flex items-center gap-2">
            <input
              type="text"
              readOnly
              value={shareUrl}
              className="h-9 flex-1 rounded-md border border-input bg-muted/50 px-3 text-sm text-foreground"
            />
            <Button variant="outline" size="sm" onClick={handleCopy}>
              {copied ? (
                <Check className="h-4 w-4" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
              {copied ? "Copied" : "Copy"}
            </Button>
          </div>
        </div>
      </article>
    </div>
  )
}
