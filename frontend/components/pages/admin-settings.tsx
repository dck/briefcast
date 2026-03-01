import { useState } from "react"
import { AdminSidebar } from "@/components/admin-sidebar"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card"

const settingsNav = [
  { id: "general", label: "General" },
  { id: "models", label: "Models" },
  { id: "processing", label: "Processing" },
]

export function AdminSettings() {
  const [activeSettingsTab, setActiveSettingsTab] = useState("processing")

  return (
    <div className="flex min-h-screen bg-background">
      <AdminSidebar />

      <main className="ml-60 flex-1 px-8 py-8">
        <h1 className="mb-6 text-xl font-semibold tracking-tight text-foreground">
          Settings
        </h1>

        <div className="flex gap-8">
          {/* Left vertical nav (1/4) */}
          <nav className="flex w-44 shrink-0 flex-col gap-0.5">
            {settingsNav.map((item) => {
              const isActive = activeSettingsTab === item.id
              return (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => setActiveSettingsTab(item.id)}
                  className={`rounded-md px-3 py-2 text-left text-sm font-medium transition-colors ${
                    isActive
                      ? "bg-accent text-accent-foreground"
                      : "text-muted-foreground hover:bg-accent/60 hover:text-foreground"
                  }`}
                >
                  {item.label}
                </button>
              )
            })}
          </nav>

          {/* Right form area (3/4) */}
          <div className="flex min-w-0 flex-1 flex-col gap-6">
            {/* Models card */}
            <Card>
              <CardHeader>
                <CardTitle>AI Models</CardTitle>
                <CardDescription>
                  Configure which models Briefcast uses for transcription and
                  summarization.
                </CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-5">
                {/* Whisper Model */}
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="whisper-model"
                    className="text-sm font-medium text-foreground"
                  >
                    Whisper Model
                  </label>
                  <Input
                    id="whisper-model"
                    defaultValue="whisper-large-v3"
                  />
                  <p className="text-xs text-muted-foreground">
                    The Groq-hosted Whisper model used for audio transcription.{" "}
                    <a
                      href="https://console.groq.com/docs"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="font-medium text-primary hover:underline"
                    >
                      View Groq docs &rarr;
                    </a>
                  </p>
                </div>

                {/* LLM Model */}
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="llm-model"
                    className="text-sm font-medium text-foreground"
                  >
                    LLM Model
                  </label>
                  <Input
                    id="llm-model"
                    defaultValue="llama-3.3-70b-versatile"
                  />
                  <p className="text-xs text-muted-foreground">
                    The language model used to generate structured summaries from
                    transcripts.
                  </p>
                </div>

                <div>
                  <Button size="sm">Save Models</Button>
                </div>
              </CardContent>
            </Card>

            {/* Processing card */}
            <Card>
              <CardHeader>
                <CardTitle>Processing</CardTitle>
                <CardDescription>
                  Configure RSS polling intervals and retry behavior.
                </CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-5">
                {/* RSS Poll Interval */}
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="poll-interval"
                    className="text-sm font-medium text-foreground"
                  >
                    RSS Poll Interval
                  </label>
                  <div className="flex items-center gap-2">
                    <Input
                      id="poll-interval"
                      type="number"
                      defaultValue={5}
                      className="w-24"
                    />
                    <span className="text-sm text-muted-foreground">
                      minutes
                    </span>
                  </div>
                </div>

                {/* Max Retries */}
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="max-retries"
                    className="text-sm font-medium text-foreground"
                  >
                    Max Retries
                  </label>
                  <Input
                    id="max-retries"
                    type="number"
                    defaultValue={3}
                    className="w-24"
                  />
                </div>

                {/* Retry Backoff */}
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-foreground">
                    Retry Backoff
                  </label>
                  <div className="flex items-center gap-3">
                    <div className="flex items-center gap-1.5">
                      <Input
                        type="number"
                        defaultValue={1}
                        className="w-20"
                        aria-label="1st retry delay"
                      />
                      <span className="text-xs text-muted-foreground">min</span>
                    </div>
                    <div className="flex items-center gap-1.5">
                      <Input
                        type="number"
                        defaultValue={5}
                        className="w-20"
                        aria-label="2nd retry delay"
                      />
                      <span className="text-xs text-muted-foreground">min</span>
                    </div>
                    <div className="flex items-center gap-1.5">
                      <Input
                        type="number"
                        defaultValue={15}
                        className="w-20"
                        aria-label="3rd retry delay"
                      />
                      <span className="text-xs text-muted-foreground">min</span>
                    </div>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Delay before each consecutive retry attempt (1st, 2nd, 3rd).
                  </p>
                </div>

                <div>
                  <Button size="sm">Save Processing</Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </div>
  )
}
