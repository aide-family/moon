import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import {
  listAlertPages,
  listMyRealtimeTabs,
  setMyRealtimeTabs,
} from "@/api/alert-page"
import type { AlertPageItem, RealtimeAlertTab } from "@/api/alert-page/types"
import { useSession } from "@/auth/session"
import { PageHeader } from "@/components/page-header"
import { AlertEventList } from "@/pages/alert/list"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { cn } from "@/lib/utils"

export function AlertRealtimePage() {
  const { t } = useTranslation()
  const session = useSession()
  const [tabs, setTabs] = useState<RealtimeAlertTab[]>([{ id: "mine", name: "mine", kind: "mine" }])
  const [active, setActive] = useState("mine")
  const [pickerOpen, setPickerOpen] = useState(false)
  const [selectable, setSelectable] = useState<AlertPageItem[]>([])
  const [selected, setSelected] = useState<string[]>([])
  const [pending, setPending] = useState(false)
  const [error, setError] = useState("")
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    let alive = true
    async function load() {
      try {
        const result = await listMyRealtimeTabs()
        if (!alive) {
          return
        }
        const next = result.tabs?.length ? result.tabs : [{ id: "mine", name: "mine", kind: "mine" }]
        setTabs(next)
        setActive((current) => (next.some((tab) => tab.id === current) ? current : "mine"))
      } catch {
        if (alive) {
          setTabs([{ id: "mine", name: "mine", kind: "mine" }])
          setActive("mine")
        }
      }
    }
    void load()
    return () => {
      alive = false
    }
  }, [attempt, session.namespaceUid])

  async function openPicker() {
    setError("")
    setPending(true)
    try {
      const [selectableResult, tabResult] = await Promise.all([
        listAlertPages({ scope: "selectable", page: 1, pageSize: 100 }),
        listMyRealtimeTabs(),
      ])
      setSelectable(selectableResult.items ?? [])
      setSelected((tabResult.tabs ?? []).filter((tab) => tab.kind === "page").map((tab) => tab.id))
      setPickerOpen(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  async function saveTabs() {
    setPending(true)
    setError("")
    try {
      const result = await setMyRealtimeTabs(selected)
      setTabs(result.tabs?.length ? result.tabs : [{ id: "mine", name: "mine", kind: "mine" }])
      setPickerOpen(false)
      setAttempt((value) => value + 1)
      setActive((current) => {
        const next = result.tabs ?? []
        return next.some((tab) => tab.id === current) ? current : "mine"
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  function tabLabel(tab: RealtimeAlertTab) {
    return tab.kind === "mine" ? t("alertEvent.mineTitle") : tab.name
  }

  const mine = active === "mine"
  const pageUid = mine ? "" : active

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("alertEvent.title")} description={t("alertEvent.description")}>
        <Button variant="outline" onClick={() => void openPicker()} disabled={pending}>
          {t("alertEvent.choosePages")}
        </Button>
      </PageHeader>
      <div className="flex flex-wrap gap-2 border-b border-border/60 pb-2">
        {tabs.map((tab) => (
          <Button
            key={tab.id}
            type="button"
            size="sm"
            variant={active === tab.id ? "default" : "ghost"}
            className={cn(active === tab.id ? "" : "text-muted-foreground")}
            onClick={() => setActive(tab.id)}
          >
            {tabLabel(tab)}
          </Button>
        ))}
      </div>
      <AlertEventList
        key={`${active}-${attempt}`}
        view="REALTIME"
        mine={mine}
        pageUid={pageUid}
        hideHeader
      />
      <Dialog open={pickerOpen} onOpenChange={setPickerOpen}>
        <DialogContent className="max-h-[80vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{t("alertEvent.choosePages")}</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">{t("alertEvent.choosePagesHint")}</p>
          <div className="grid gap-3">
            {selectable.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("alertPage.emptySelectable")}</p>
            ) : (
              selectable.map((item) => {
                const checked = selected.includes(item.uid)
                return (
                  <label key={item.uid} className="flex items-start gap-3 rounded-lg border p-3">
                    <Checkbox
                      checked={checked}
                      onCheckedChange={(value) => {
                        setSelected((current) =>
                          value === true
                            ? [...current, item.uid]
                            : current.filter((uid) => uid !== item.uid),
                        )
                      }}
                    />
                    <div className="grid min-w-0 flex-1 gap-1">
                      <span className="font-medium">{item.name}</span>
                      <span className="truncate text-xs text-muted-foreground">
                        {item.public ? t("alertPage.public") : t("alertPage.private")} · {item.labels || "{}"}
                      </span>
                    </div>
                  </label>
                )
              })
            )}
          </div>
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setPickerOpen(false)}>
              {t("action.cancel")}
            </Button>
            <Button type="button" disabled={pending} onClick={() => void saveTabs()}>
              {t("action.save")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  )
}
