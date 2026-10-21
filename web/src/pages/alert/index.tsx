import { GripVertical } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { formMessage } from "@/lib/form-error"

import { alertEventRealtimeStats } from "@/api/alert-event"
import type { AlertEventStats } from "@/api/alert-event/types"
import {
  listAlertPages,
  listMyRealtimeTabs,
  setMyRealtimeTabs,
} from "@/api/alert-page"
import type { AlertPageItem, RealtimeAlertTab } from "@/api/alert-page/types"
import { useSession } from "@/auth/session"
import { PageHeader } from "@/components/page-header"
import { AlertEventList } from "@/pages/alert/list"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

// 实时告警页的自动刷新节奏：每次 tick 只发两个请求，数量统计和当前 Tab 的列表。
const REALTIME_REFRESH_MS = 30000
// 实时告警页最多同时显示的告警页面数量，和后端 SetMyRealtimeTabs 的上限一致。
const MAX_PAGE_TABS = 10
// 「我的告警」固定排在第一，不参与拖动排序。
const MINE_TAB_ID = "mine"
const MINE_TAB: RealtimeAlertTab = {
  id: MINE_TAB_ID,
  name: MINE_TAB_ID,
  kind: "mine",
}

export function AlertRealtimePage() {
  const { t } = useTranslation()
  const session = useSession()
  const [tabs, setTabs] = useState<RealtimeAlertTab[]>([MINE_TAB])
  const [active, setActive] = useState(MINE_TAB_ID)
  const [pickerOpen, setPickerOpen] = useState(false)
  const [selectable, setSelectable] = useState<AlertPageItem[]>([])
  const [selected, setSelected] = useState<string[]>([])
  const [pickerFilter, setPickerFilter] = useState("")
  const [pending, setPending] = useState(false)
  const [error, setError] = useState("")
  const [dragging, setDragging] = useState("")
  const [dropTarget, setDropTarget] = useState("")
  const [ordering, setOrdering] = useState(false)
  const [orderError, setOrderError] = useState("")
  const [attempt, setAttempt] = useState(0)
  const [refreshKey, setRefreshKey] = useState(0)
  const [counts, setCounts] = useState<Record<string, number>>({})
  const [stats, setStats] = useState<AlertEventStats>({
    firing: 0,
    suppressed: 0,
    recovered: 0,
  })

  useEffect(() => {
    let alive = true
    async function load() {
      try {
        const result = await listMyRealtimeTabs()
        if (!alive) {
          return
        }
        const next = result.tabs?.length ? result.tabs : [MINE_TAB]
        setTabs(next)
        setActive((current) =>
          next.some((tab) => tab.id === current) ? current : MINE_TAB_ID
        )
      } catch {
        if (alive) {
          setTabs([MINE_TAB])
          setActive(MINE_TAB_ID)
        }
      }
    }
    void load()
    return () => {
      alive = false
    }
  }, [attempt, session.namespaceUid])

  useEffect(() => {
    const timer = window.setInterval(
      () => setRefreshKey((value) => value + 1),
      REALTIME_REFRESH_MS
    )
    return () => window.clearInterval(timer)
  }, [])

  // 数量统计走独立接口：一次返回已选 Tab 的数量和统计条，列表接口不再为每个 Tab 单独查询。
  useEffect(() => {
    let alive = true
    async function loadStats() {
      try {
        const result = await alertEventRealtimeStats()
        if (!alive) {
          return
        }
        const next: Record<string, number> = {}
        for (const tab of result.tabs ?? []) {
          next[tab.id] = tab.count ?? 0
        }
        setCounts(next)
        setStats({
          firing: result.firing ?? 0,
          suppressed: result.suppressed ?? 0,
          recovered: result.recovered ?? 0,
        })
      } catch {
        // 保留上一次成功统计，避免瞬时失败把数量清零。
      }
    }
    void loadStats()
    return () => {
      alive = false
    }
  }, [attempt, session.namespaceUid, refreshKey])

  async function openPicker() {
    setError("")
    setPending(true)
    try {
      const [selectableResult, tabResult] = await Promise.all([
        listAlertPages({ scope: "selectable", page: 1, pageSize: 100 }),
        listMyRealtimeTabs(),
      ])
      setSelectable(selectableResult.items ?? [])
      setSelected(
        (tabResult.tabs ?? [])
          .filter((tab) => tab.kind === "page")
          .map((tab) => tab.id)
      )
      setPickerFilter("")
      setPickerOpen(true)
    } catch (err) {
      setError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function saveTabs() {
    if (selected.length > MAX_PAGE_TABS) {
      setError(t("alertEvent.pickerLimitReached", { max: MAX_PAGE_TABS }))
      return
    }
    setPending(true)
    setError("")
    try {
      const result = await setMyRealtimeTabs(selected)
      setTabs(result.tabs?.length ? result.tabs : [MINE_TAB])
      setPickerOpen(false)
      setAttempt((value) => value + 1)
      setActive((current) => {
        const next = result.tabs ?? []
        return next.some((tab) => tab.id === current) ? current : MINE_TAB_ID
      })
    } catch (err) {
      setError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  // 拖动只调整页面 Tab 的顺序：「我的告警」始终第一。顺序先本地生效，服务端失败再回滚。
  async function moveTab(sourceID: string, targetID: string) {
    if (
      ordering ||
      !sourceID ||
      sourceID === MINE_TAB_ID ||
      sourceID === targetID
    ) {
      return
    }
    const pageTabs = tabs.filter((tab) => tab.id !== MINE_TAB_ID)
    const from = pageTabs.findIndex((tab) => tab.id === sourceID)
    const to = pageTabs.findIndex((tab) => tab.id === targetID)
    if (from < 0 || to < 0) {
      return
    }
    const reordered = [...pageTabs]
    const [moved] = reordered.splice(from, 1)
    reordered.splice(to, 0, moved)
    const previous = tabs
    setTabs([MINE_TAB, ...reordered])
    setOrdering(true)
    setOrderError("")
    try {
      const result = await setMyRealtimeTabs(reordered.map((tab) => tab.id))
      if (result.tabs?.length) {
        setTabs(result.tabs)
      }
    } catch (err) {
      setTabs(previous)
      setOrderError(formMessage(err, t("state.failed")))
    } finally {
      setOrdering(false)
    }
  }

  function tabLabel(tab: RealtimeAlertTab) {
    return tab.kind === "mine" ? t("alertEvent.mineTitle") : tab.name
  }

  function selectTab(id: string) {
    // 切 Tab 只让列表按当前页面重新查询；其他页面不发列表请求，数量等下一次刷新。
    setActive(id)
  }

  const mine = active === MINE_TAB_ID
  const pageUid = mine ? "" : active
  // 本地筛选只影响可见项；批量操作作用于当前筛选结果，避免误选被筛掉的页面。
  const filteredPages = useMemo(() => {
    const keyword = pickerFilter.trim().toLowerCase()
    if (!keyword) {
      return selectable
    }
    return selectable.filter((item) =>
      [item.name, item.labels, item.remark].some((value) =>
        (value || "").toLowerCase().includes(keyword)
      )
    )
  }, [pickerFilter, selectable])

  function selectPages(uids: string[]) {
    // 批量选择同样不能越过上限，只补到剩余名额为止。
    setSelected((current) => {
      const remaining = MAX_PAGE_TABS - current.length
      if (remaining <= 0) {
        return current
      }
      return Array.from(new Set([...current, ...uids])).slice(0, MAX_PAGE_TABS)
    })
  }

  function togglePage(uid: string, checked: boolean) {
    setSelected((current) =>
      checked
        ? Array.from(new Set([...current, uid]))
        : current.filter((item) => item !== uid)
    )
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("alertEvent.title")}
        description={t("alertEvent.description")}
      >
        <Button
          variant="outline"
          onClick={() => void openPicker()}
          disabled={pending}
        >
          {t("alertEvent.choosePages")}
        </Button>
      </PageHeader>
      <div className="grid gap-2">
        <div className="flex flex-wrap items-center gap-2 border-b border-border/60 pb-2">
          {tabs.map((tab) => {
            const draggable = tab.id !== MINE_TAB_ID
            const dropping = dropTarget === tab.id && dragging !== tab.id
            return (
              <Button
                key={tab.id}
                type="button"
                size="sm"
                variant={active === tab.id ? "secondary" : "ghost"}
                className={cn(
                  "gap-2",
                  active === tab.id
                    ? "ring-1 ring-border"
                    : "text-muted-foreground",
                  // 拖动只用于告警页面之间的排序，「我的告警」固定第一，不显示拖动手柄。
                  draggable && "cursor-grab active:cursor-grabbing",
                  dropping && "ring-2 ring-primary/60"
                )}
                title={draggable ? t("alertEvent.dragHint") : undefined}
                draggable={draggable}
                onDragStart={(event) => {
                  if (!draggable) {
                    return
                  }
                  event.dataTransfer.effectAllowed = "move"
                  event.dataTransfer.setData("text/plain", tab.id)
                  setDragging(tab.id)
                  setDropTarget("")
                }}
                onDragOver={(event) => {
                  // 用 dataTransfer 的 dropEffect 告诉浏览器这里可以放下；
                  // 不依赖 dragging 状态，避免快速拖动时 React 尚未重渲染导致无法落下。
                  if (!draggable) {
                    return
                  }
                  event.preventDefault()
                  event.dataTransfer.dropEffect = "move"
                  setDropTarget((current) =>
                    current === tab.id ? current : tab.id
                  )
                }}
                onDragLeave={() => {
                  setDropTarget((current) =>
                    current === tab.id ? "" : current
                  )
                }}
                onDrop={(event) => {
                  if (!draggable) {
                    return
                  }
                  event.preventDefault()
                  const sourceID =
                    dragging || event.dataTransfer.getData("text/plain")
                  setDragging("")
                  setDropTarget("")
                  void moveTab(sourceID, tab.id)
                }}
                onDragEnd={() => {
                  setDragging("")
                  setDropTarget("")
                }}
                onClick={() => selectTab(tab.id)}
              >
                {draggable ? (
                  <GripVertical aria-hidden className="opacity-40" />
                ) : null}
                <span className="max-w-[14rem] truncate">{tabLabel(tab)}</span>
                <Badge
                  variant={
                    (counts[tab.id] ?? 0) > 0 ? "destructive" : "secondary"
                  }
                  className="font-mono tabular-nums"
                >
                  {counts[tab.id] ?? 0}
                </Badge>
              </Button>
            )
          })}
        </div>
        {orderError ? (
          <p className="text-sm text-destructive">{orderError}</p>
        ) : null}
      </div>
      <AlertEventList
        key={`${active}-${attempt}`}
        view="REALTIME"
        mine={mine}
        pageUid={pageUid}
        hideHeader
        refreshKey={refreshKey}
        stats={stats}
      />
      <Dialog open={pickerOpen} onOpenChange={setPickerOpen}>
        <DialogContent className="flex max-h-[80vh] flex-col gap-4 overflow-hidden sm:max-w-2xl">
          <DialogHeader className="shrink-0">
            <DialogTitle>{t("alertEvent.choosePages")}</DialogTitle>
          </DialogHeader>
          {/* 标题、批量操作和筛选固定在顶部，只有下面的页面列表滚动。 */}
          <div className="grid shrink-0 gap-3">
            <p className="text-sm text-muted-foreground">
              {t("alertEvent.choosePagesHint")}
            </p>
            <p className="text-xs text-muted-foreground">
              {t("alertEvent.choosePagesLimit", { max: MAX_PAGE_TABS })}
            </p>
            <div className="flex flex-nowrap items-center gap-2 overflow-x-auto pb-1">
              <Button
                type="button"
                size="sm"
                variant="outline"
                className="shrink-0"
                disabled={
                  filteredPages.length === 0 || selected.length >= MAX_PAGE_TABS
                }
                onClick={() =>
                  selectPages(filteredPages.map((item) => item.uid))
                }
              >
                {t("alertEvent.pickerSelectAll")}
              </Button>
              <Button
                type="button"
                size="sm"
                variant="outline"
                className="shrink-0"
                disabled={selected.length === 0}
                onClick={() => setSelected([])}
              >
                {t("alertEvent.pickerClearAll")}
              </Button>
              <Button
                type="button"
                size="sm"
                variant="outline"
                className="shrink-0"
                disabled={
                  filteredPages.every((item) => item.public) ||
                  selected.length >= MAX_PAGE_TABS
                }
                onClick={() =>
                  selectPages(
                    filteredPages
                      .filter((item) => !item.public)
                      .map((item) => item.uid)
                  )
                }
              >
                {t("alertEvent.pickerSelectPrivate")}
              </Button>
              <span
                className={cn(
                  "ml-auto shrink-0 font-mono text-xs text-muted-foreground",
                  selected.length >= MAX_PAGE_TABS && "text-destructive"
                )}
              >
                {t("alertEvent.pickerSelected", {
                  selected: selected.length,
                  total: selectable.length,
                })}
              </span>
            </div>
            <Input
              value={pickerFilter}
              placeholder={t("alertEvent.pickerFilterPlaceholder")}
              aria-label={t("alertEvent.pickerFilterPlaceholder")}
              onChange={(event) => setPickerFilter(event.target.value)}
            />
          </div>
          <div className="grid min-h-0 flex-1 content-start gap-3 overflow-y-auto pr-1">
            {selectable.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                {t("alertPage.emptySelectable")}
              </p>
            ) : filteredPages.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                {t("alertEvent.pickerNoMatch")}
              </p>
            ) : (
              filteredPages.map((item) => {
                const checked = selected.includes(item.uid)
                // 达到上限后只锁住未选中的项，已选中的仍可取消。
                const disabled = !checked && selected.length >= MAX_PAGE_TABS
                return (
                  <label
                    key={item.uid}
                    className={cn(
                      "flex items-start gap-3 rounded-lg border p-3",
                      disabled && "opacity-60"
                    )}
                    title={
                      disabled
                        ? t("alertEvent.pickerLimitReached", {
                            max: MAX_PAGE_TABS,
                          })
                        : undefined
                    }
                  >
                    <Checkbox
                      checked={checked}
                      disabled={disabled}
                      onCheckedChange={(value) =>
                        togglePage(item.uid, value === true)
                      }
                    />
                    <div className="grid min-w-0 flex-1 gap-1">
                      <span className="flex min-w-0 items-center gap-2">
                        <span className="min-w-0 truncate font-medium">
                          {item.name}
                        </span>
                        <Badge
                          variant={item.public ? "secondary" : "outline"}
                          className="shrink-0"
                        >
                          {item.public
                            ? t("alertPage.public")
                            : t("alertPage.private")}
                        </Badge>
                      </span>
                      <span className="truncate font-mono text-xs text-muted-foreground">
                        {item.labels || "{}"}
                      </span>
                    </div>
                  </label>
                )
              })
            )}
          </div>
          {error ? (
            <p className="shrink-0 text-sm text-destructive">{error}</p>
          ) : null}
          <DialogFooter className="shrink-0">
            <Button
              type="button"
              variant="outline"
              onClick={() => setPickerOpen(false)}
            >
              {t("action.cancel")}
            </Button>
            <Button
              type="button"
              disabled={pending}
              onClick={() => void saveTabs()}
            >
              {t("action.save")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  )
}
