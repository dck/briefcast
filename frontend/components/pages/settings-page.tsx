import { useState } from "react"
import { AppSidebar } from "@/components/app-sidebar"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card"

export function SettingsPage() {
  const [selectedPodcast, setSelectedPodcast] = useState<string | null>(null)
  const [telegramChatId, setTelegramChatId] = useState("")
  const [email, setEmail] = useState("alex@example.com")
  const [notifyTelegram, setNotifyTelegram] = useState(false)
  const [notifyEmail, setNotifyEmail] = useState(true)

  return (
    <div className="flex min-h-screen bg-background">
      <AppSidebar
        selectedPodcast={selectedPodcast}
        onSelectPodcast={setSelectedPodcast}
      />

      <main className="ml-60 flex-1 px-8 py-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight text-foreground">
            Settings
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Manage your notification preferences
          </p>
        </div>

        <div className="flex max-w-2xl flex-col gap-6">
          {/* Notification Channels */}
          <Card>
            <CardHeader>
              <CardTitle>Notification Channels</CardTitle>
              <CardDescription>
                Choose how you want to be notified when new summaries are ready.
              </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-5">
              {/* Telegram */}
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="telegram-chat-id">Telegram Chat ID</Label>
                <Input
                  id="telegram-chat-id"
                  placeholder="e.g. 123456789"
                  value={telegramChatId}
                  onChange={(e) => setTelegramChatId(e.target.value)}
                />
                <p className="text-xs text-muted-foreground">
                  Send <code className="rounded bg-muted px-1 py-0.5 text-xs">/start</code> to{" "}
                  <span className="font-medium">@BriefcastBot</span> on Telegram to get your
                  Chat ID.
                </p>
              </div>

              <div className="flex items-center justify-between rounded-md border border-border px-4 py-3">
                <div className="flex flex-col gap-0.5">
                  <Label htmlFor="notify-telegram" className="cursor-pointer">
                    Telegram notifications
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Receive a Telegram message for each new summary
                  </p>
                </div>
                <Switch
                  id="notify-telegram"
                  checked={notifyTelegram}
                  onCheckedChange={setNotifyTelegram}
                />
              </div>

              {/* Email */}
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>

              <div className="flex items-center justify-between rounded-md border border-border px-4 py-3">
                <div className="flex flex-col gap-0.5">
                  <Label htmlFor="notify-email" className="cursor-pointer">
                    Email notifications
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Receive an email digest for new summaries
                  </p>
                </div>
                <Switch
                  id="notify-email"
                  checked={notifyEmail}
                  onCheckedChange={setNotifyEmail}
                />
              </div>

              <div>
                <Button size="sm">Save Settings</Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </main>
    </div>
  )
}
