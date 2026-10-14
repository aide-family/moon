import { useEffect, useState } from "react"

import { checkHealth } from "@/api/health"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export function HomePage() {
  const [attempt, setAttempt] = useState(0)
  const [loading, setLoading] = useState(true)
  const [status, setStatus] = useState<string | null>(null)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    let active = true

    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const health = await checkHealth()
        if (!active) {
          return
        }
        setStatus(health.status)
      } catch {
        if (!active) {
          return
        }
        setStatus(null)
        setFailed(true)
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }

    void load()
    return () => {
      active = false
    }
  }, [attempt])

  return (
    <main className="flex min-h-svh items-center justify-center p-6">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Alioth</CardTitle>
          <CardDescription>/v1/health</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <p className="font-mono text-sm">
            {loading ? "..." : failed || status == null ? "—" : status}
          </p>
          <Button
            type="button"
            variant="outline"
            disabled={loading}
            onClick={() => setAttempt((value) => value + 1)}
          >
            Refresh
          </Button>
        </CardContent>
      </Card>
    </main>
  )
}
