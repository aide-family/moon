import { ShieldCheck } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useParams } from "react-router-dom"

import { getSharedInspectionReport } from "@/api/inspection"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

/** 公开巡检报告：只凭 token 访问，不挂登录态与租户头。 */
export function PublicInspectionReportPage() {
  const { t } = useTranslation()
  const { token = "" } = useParams()
  const [title, setTitle] = useState("")
  const [format, setFormat] = useState("")
  const [content, setContent] = useState("")
  const [password, setPassword] = useState("")
  const [loading, setLoading] = useState(Boolean(token))
  const [failed, setFailed] = useState(!token)
  const [retrying, setRetrying] = useState(false)

  async function load(withPassword: string) {
    try {
      const result = await getSharedInspectionReport(token, withPassword)
      setTitle(result.title)
      setFormat(result.format)
      setContent(result.content)
      setFailed(false)
    } catch {
      setFailed(true)
    } finally {
      setRetrying(false)
      setLoading(false)
    }
  }

  useEffect(() => {
    if (!token) {
      return
    }
    // 首次加载是异步请求，不产生同步级联渲染。
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load("")
    // token 变化时重新请求；password 由用户显式提交。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token])

  return (
    <div className="min-h-svh bg-muted/40">
      <header className="sticky top-0 z-10 flex items-center gap-3 border-b border-border bg-card/90 px-4 py-3 backdrop-blur sm:px-6">
        <span className="grid size-8 place-items-center rounded-lg bg-primary font-bold text-primary-foreground">
          A
        </span>
        <div className="grid gap-0.5">
          <span className="font-heading text-sm font-semibold">{t("inspection.publicReport")}</span>
          <span className="text-xs text-muted-foreground">{t("inspection.publicReadOnly")}</span>
        </div>
        <ShieldCheck className="ml-auto size-4 text-muted-foreground" aria-hidden="true" />
      </header>
      <main className="mx-auto w-full max-w-[860px] px-4 py-6 sm:px-6 sm:py-8">
        {loading ? (
          <p className="py-16 text-center text-sm text-muted-foreground">{t("state.loading")}</p>
        ) : null}
        {!loading && failed ? (
          <div className="mx-auto grid max-w-md gap-3 rounded-xl border border-border bg-card p-6">
            <p className="text-sm font-medium">{t("inspection.publicExpired")}</p>
            <p className="text-sm text-muted-foreground">{t("inspection.reportUnavailable")}</p>
            <form
              className="flex gap-2"
              onSubmit={(event) => {
                event.preventDefault()
                setRetrying(true)
                void load(password)
              }}
            >
              <Input
                type="password"
                value={password}
                placeholder={t("inspection.passwordPlaceholder")}
                aria-label={t("inspection.sharePassword")}
                onChange={(event) => setPassword(event.target.value)}
              />
              <Button type="submit" disabled={retrying}>
                {t("inspection.unlock")}
              </Button>
            </form>
          </div>
        ) : null}
        {!loading && !failed ? (
          <>
            <h1 className="mb-4 text-center font-heading text-xl font-semibold">{title}</h1>
            {format === "HTML" ? (
              <iframe
                title={title}
                sandbox=""
                srcDoc={content}
                className="h-[80vh] w-full rounded-lg border border-border bg-white shadow-sm"
              />
            ) : (
              <pre className="overflow-auto rounded-lg border border-border bg-white p-5 font-mono text-xs whitespace-pre-wrap text-foreground">
                {content}
              </pre>
            )}
          </>
        ) : null}
      </main>
      <footer className="pb-6 text-center text-xs text-muted-foreground">
        {t("inspection.generatedBy")}
      </footer>
    </div>
  )
}
