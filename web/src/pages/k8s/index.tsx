import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { formMessage } from "@/lib/form-error"

import { listAgents } from "@/api/agent"
import type { AgentItem } from "@/api/agent/types"
import { proxyK8s } from "@/api/k8s"
import { useSession } from "@/auth/session"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"

export function K8sPage() {
  const { t } = useTranslation()
  const session = useSession()
  const [agents, setAgents] = useState<AgentItem[]>([])
  const [agentUid, setAgentUid] = useState("")
  const [method, setMethod] = useState("GET")
  const [path, setPath] = useState("/api/v1/namespaces")
  const [result, setResult] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const canProxy = session.can("k8s:proxy")

  useEffect(() => {
    let active = true
    void listAgents({ page: 1, pageSize: 100 }).then((res) => {
      if (!active) return
      setAgents((res.items ?? []).filter((item) => item.online))
    })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  async function onProxy() {
    if (!agentUid || !canProxy) return
    setPending(true)
    setError("")
    try {
      const reply = await proxyK8s({ agentUid, method, path })
      const body = reply.body ? atob(reply.body) : ""
      setResult(`${reply.statusCode}\n${body}`)
    } catch (err) {
      setError(formMessage(err, t("state.failed")))
      setResult("")
    } finally {
      setPending(false)
    }
  }

  if (!canProxy && !session.can("k8s:read")) {
    return <p className="text-sm text-muted-foreground">{t("state.failed")}</p>
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("k8s.title")} description={t("k8s.description")} />
      <div className="grid max-w-3xl gap-4">
        <FormField label={t("k8s.agent")} htmlFor="k8s-agent">
          <Select value={agentUid} onValueChange={setAgentUid}>
            <SelectTrigger id="k8s-agent">
              <SelectValue placeholder={t("k8s.pickAgent")} />
            </SelectTrigger>
            <SelectContent>
              {agents.map((agent) => (
                <SelectItem key={agent.uid} value={agent.uid}>
                  {agent.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </FormField>
        <FormField label={t("k8s.method")} htmlFor="k8s-method">
          <Input id="k8s-method" value={method} onChange={(e) => setMethod(e.target.value)} disabled={!canProxy} />
        </FormField>
        <FormField label={t("k8s.path")} htmlFor="k8s-path">
          <Input id="k8s-path" value={path} onChange={(e) => setPath(e.target.value)} disabled={!canProxy} />
        </FormField>
        {canProxy ? (
          <Button disabled={!agentUid || pending} onClick={() => void onProxy()}>
            {t("k8s.send")}
          </Button>
        ) : null}
        {error ? <p className="text-sm text-destructive">{error}</p> : null}
        <Textarea className="min-h-48 font-mono text-xs" readOnly value={result} />
      </div>
    </section>
  )
}
