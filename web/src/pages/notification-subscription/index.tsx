import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { formErrorMessage, readFormError } from "@/lib/form-error"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import type { AlertDictionaryItem } from "@/api/alert-dictionary/types"
import { selectDatasources } from "@/api/datasource"
import type { SelectItem as DatasourceOption } from "@/api/datasource/types"
import { selectNotificationGroups } from "@/api/notification-group"
import type { SelectItem as GroupOption } from "@/api/notification-group/types"
import {
  createNotificationSubscription,
  deleteNotificationSubscription,
  listNotificationSubscriptions,
  updateNotificationSubscription,
  updateNotificationSubscriptionStatus,
} from "@/api/notification-subscription"
import type {
  CreateNotificationSubscriptionParams,
  NotificationSubscriptionItem,
} from "@/api/notification-subscription/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { ListFrame } from "@/components/list-frame"
import { MultiSelect } from "@/components/multi-select"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"
import { parseLabelPairs, serializeLabelPairs, type LabelPair } from "@/lib/labels"
import { loadAllPagedItems, loadAllSelectItems } from "@/lib/select-pages"
import { runAction } from "@/lib/action"

/** 重复发送间隔按分钟编辑，后端按秒保存；0 表示不限制。 */
const MAX_REPEAT_INTERVAL_MINUTES = 720

function toOptions(items: Array<{ uid: string; name: string }>) {
  return items.map((item) => ({ value: item.uid, label: item.name }))
}

/** 把后端返回的秒数换算成分钟展示，非整分钟按四舍五入。 */
function toMinutes(seconds: number | undefined): number {
  return Math.round((seconds ?? 0) / 60)
}

export function NotificationSubscriptionPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [groups, setGroups] = useState<GroupOption[]>([])
  const [levels, setLevels] = useState<AlertDictionaryItem[]>([])
  const [datasourceLevels, setDatasourceLevels] = useState<AlertDictionaryItem[]>([])
  const [datasources, setDatasources] = useState<DatasourceOption[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<NotificationSubscriptionItem | null>(null)
  const [name, setName] = useState("")
  const [labels, setLabels] = useState<LabelPair[]>([{ key: "", value: "" }])
  const [levelUids, setLevelUids] = useState<string[]>([])
  const [datasourceUids, setDatasourceUids] = useState<string[]>([])
  const [datasourceLevelUids, setDatasourceLevelUids] = useState<string[]>([])
  const [groupUids, setGroupUids] = useState<string[]>([])
  const [repeatIntervalMinutes, setRepeatIntervalMinutes] = useState("0")
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<NotificationSubscriptionItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; item: NotificationSubscriptionItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("notification:write")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<NotificationSubscriptionItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}`,
    fetch: (page, pageSize) => listNotificationSubscriptions({ page, pageSize, name: query.name }),
  })

  // 订阅弹窗的附加选择器按模块权限加载，缺权限时跳过，避免 403 全局跳转。
  const canReadLevels = session.can("alert_rule:read")
  const canReadSources = session.can("datasource:read")

  useEffect(() => {
    let active = true
    void Promise.all([
      canReadLevels
        ? loadAllSelectItems(
            (lastUid) => selectNotificationGroups({ limit: 100, lastUid }),
            100
          )
        : Promise.resolve([]),
      canReadLevels
        ? loadAllPagedItems((page, pageSize) =>
            listAlertDictionaries({ page, pageSize, type: "ALERT" })
          )
        : Promise.resolve([]),
      canReadLevels
        ? loadAllPagedItems((page, pageSize) =>
            listAlertDictionaries({ page, pageSize, type: "DATASOURCE" })
          )
        : Promise.resolve([]),
      canReadSources
        ? loadAllSelectItems(
            (lastUid) => selectDatasources({ limit: 100, lastUid }),
            100
          )
        : Promise.resolve([]),
    ])
      .then(([groupResult, levelResult, datasourceLevelResult, datasourceResult]) => {
        if (active) {
          setGroups(groupResult)
          setLevels(levelResult)
          setDatasourceLevels(datasourceLevelResult)
          setDatasources(datasourceResult)
        }
      })
      .catch(() => {
        if (active) {
          setGroups([])
          setLevels([])
          setDatasourceLevels([])
          setDatasources([])
        }
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid, canReadLevels, canReadSources])

  function selectionSummary(selected: number, total: number): string {
    return selected >= total && total > 0
      ? t("multiSelect.allSelected", { count: selected })
      : t("multiSelect.summary", { count: selected, total })
  }

  function intervalError(): string | null {
    const trimmed = repeatIntervalMinutes.trim()
    if (!/^\d+$/.test(trimmed)) {
      return t("subscription.repeatIntervalInvalid")
    }
    const minutes = Number(trimmed)
    if (minutes > MAX_REPEAT_INTERVAL_MINUTES) {
      return t("subscription.repeatIntervalInvalid")
    }
    return null
  }

  /** 列表和详情把秒数展示为分钟；0 表示该告警不限制重复通知。 */
  function intervalSummary(seconds: number | undefined): string {
    const minutes = toMinutes(seconds)
    return minutes > 0
      ? t("subscription.repeatIntervalValue", { minutes })
      : t("subscription.repeatIntervalOff")
  }

  function openCreate() {
    setEditing(null)
    setName("")
    setLabels([{ key: "", value: "" }])
    setLevelUids([])
    setDatasourceUids([])
    setDatasourceLevelUids([])
    setGroupUids([])
    setRepeatIntervalMinutes("0")
    setRemark("")
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function openEdit(item: NotificationSubscriptionItem) {
    setEditing(item)
    setName(item.name)
    setLabels(parseLabelPairs(item.labels))
    setLevelUids(item.levelUids ?? [])
    setDatasourceUids(item.datasourceUids ?? [])
    setDatasourceLevelUids(item.datasourceLevelUids ?? [])
    setGroupUids(item.groupUids ?? [])
    setRepeatIntervalMinutes(String(toMinutes(item.repeatIntervalSeconds)))
    setRemark(item.remark)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
      ["groups", groupUids.length === 0 ? t("form.required") : null],
      [
        "labels",
        labels.some(
          (pair) => pair.key.trim() === "" && pair.value.trim() !== "",
        )
          ? t("form.required")
          : null,
      ],
      ["repeatIntervalSeconds", intervalError()],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const payload: CreateNotificationSubscriptionParams = {
        name: name.trim(),
        labels: serializeLabelPairs(labels),
        levelUids,
        datasourceUids,
        datasourceLevelUids,
        groupUids,
        repeatIntervalSeconds: Number(repeatIntervalMinutes.trim()) * 60,
        remark,
      }
      if (editing) {
        await updateNotificationSubscription({ ...payload, uid: editing.uid })
      } else {
        await createNotificationSubscription(payload)
      }
      setOpen(false)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function runConfirm() {
    if (!confirm) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      if (confirm.action === "delete") {
        await deleteNotificationSubscription(confirm.item.uid)
      } else {
        await updateNotificationSubscriptionStatus(confirm.item.uid, "DISABLED")
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: NotificationSubscriptionItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      { key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) },
      item.status === "ACTIVE"
        ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => setConfirm({ action: "disable", item }) }
        : { key: "enable", label: t("action.enable"), onSelect: () => runAction(updateNotificationSubscriptionStatus(item.uid, "ACTIVE").then(() => reload()), setError) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    ]
  }

  function conditionSummary(item: NotificationSubscriptionItem): string {
    const parts: string[] = []
    const pairs = parseLabelPairs(item.labels).filter((pair) => pair.key.trim() !== "")
    if (pairs.length > 0) {
      parts.push(pairs.map((pair) => `${pair.key}=${pair.value}`).join(", "))
    }
    const levelNames = (item.levelUids ?? []).map((uid) => levels.find((level) => level.uid === uid)?.name ?? uid)
    if (levelNames.length > 0) {
      parts.push(`${t("subscription.levels")}: ${levelNames.join(", ")}`)
    }
    const datasourceNames = (item.datasourceUids ?? []).map((uid) => datasources.find((datasource) => datasource.uid === uid)?.name ?? uid)
    if (datasourceNames.length > 0) {
      parts.push(`${t("subscription.datasources")}: ${datasourceNames.join(", ")}`)
    }
    const datasourceLevelNames = (item.datasourceLevelUids ?? []).map((uid) => datasourceLevels.find((level) => level.uid === uid)?.name ?? uid)
    if (datasourceLevelNames.length > 0) {
      parts.push(`${t("subscription.datasourceLevels")}: ${datasourceLevelNames.join(", ")}`)
    }
    return parts.length > 0 ? parts.join(" · ") : t("subscription.anyEvent")
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("subscription.title")} description={t("subscription.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input className="max-w-sm" value={draft} placeholder={t("subscription.namePlaceholder")} onChange={(event) => setDraft(event.target.value)} />
        <Button type="submit" variant="outline">
          {t("action.search")}
        </Button>
      </form>
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
            <Button variant="outline" onClick={() => reload()}>
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("subscription.empty")}</p> : null}
        {!loading && !failed && items.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("subscription.name")}</TableHead>
                <TableHead>{t("subscription.conditions")}</TableHead>
                <TableHead>{t("subscription.groups")}</TableHead>
                <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell className="max-w-md whitespace-normal text-sm break-words text-muted-foreground">{conditionSummary(item)}</TableCell>
                  <TableCell className="whitespace-normal break-words">
                    {(item.groupUids ?? []).map((uid) => groups.find((group) => group.uid === uid)?.name ?? uid).join(", ")}
                  </TableCell>
                  <TableCell><StatusBadge value={item.status} /></TableCell>
                  <TableCell>
                    <RowActions nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(item)}
                      actions={actions(item)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="gap-0 p-0 sm:max-w-3xl">
          <DialogHeader className="shrink-0 border-b px-4 py-3 pr-12">
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="flex min-h-0 flex-1 flex-col" onSubmit={onSave}>
            <DialogBody className="grid gap-4 px-4 py-4">
              <FormField
                label={t("subscription.name")}
                htmlFor="subscription-name"
                required
                error={fieldErrors.name}
              >
                <Input id="subscription-name" value={name} placeholder={t("subscription.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
              </FormField>
              <FormField
                label={t("subscription.groups")}
                htmlFor="subscription-groups"
                hint={t("subscription.groupsHint")}
                required
                error={fieldErrors.groups}
              >
                <MultiSelect
                  id="subscription-groups"
                  invalid={Boolean(fieldErrors.groups)}
                  values={groupUids}
                  options={toOptions(groups)}
                  onChange={setGroupUids}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("subscription.noGroups")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                />
              </FormField>
              <FormField label={t("subscription.labels")} hint={t("subscription.labelsHint")} error={fieldErrors.labels}>
                <div className="grid gap-2">
                  {labels.map((pair, index) => (
                    <div key={index} className="flex items-center gap-2">
                      <Input
                        value={pair.key}
                        placeholder={t("subscription.labelKey")}
                        onChange={(event) => {
                          const next = [...labels]
                          next[index] = { ...pair, key: event.target.value }
                          setLabels(next)
                        }}
                      />
                      <Input
                        value={pair.value}
                        placeholder={t("subscription.labelValue")}
                        onChange={(event) => {
                          const next = [...labels]
                          next[index] = { ...pair, value: event.target.value }
                          setLabels(next)
                        }}
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() => setLabels(labels.filter((_, current) => current !== index))}
                      >
                        {t("action.delete")}
                      </Button>
                    </div>
                  ))}
                  <Button type="button" variant="outline" onClick={() => setLabels([...labels, { key: "", value: "" }])}>
                    {t("subscription.addLabel")}
                  </Button>
                </div>
              </FormField>
              <FormField label={t("subscription.levels")} htmlFor="subscription-levels" hint={t("subscription.levelsHint")}>
                <MultiSelect
                  id="subscription-levels"
                  values={levelUids}
                  options={toOptions(levels)}
                  onChange={setLevelUids}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("subscription.noLevels")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                />
              </FormField>
              <FormField label={t("subscription.datasources")} htmlFor="subscription-datasources" hint={t("subscription.datasourcesHint")}>
                <MultiSelect
                  id="subscription-datasources"
                  values={datasourceUids}
                  options={toOptions(datasources)}
                  onChange={setDatasourceUids}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("subscription.noDatasources")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                />
              </FormField>
              <FormField
                label={t("subscription.datasourceLevels")}
                htmlFor="subscription-datasource-levels"
                hint={t("subscription.datasourceLevelsHint")}
              >
                <MultiSelect
                  id="subscription-datasource-levels"
                  values={datasourceLevelUids}
                  options={toOptions(datasourceLevels)}
                  onChange={setDatasourceLevelUids}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("subscription.noDatasourceLevels")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                />
              </FormField>
              <FormField
                label={t("subscription.repeatInterval")}
                htmlFor="subscription-repeat-interval"
                hint={t("subscription.repeatIntervalHint")}
                error={fieldErrors.repeatIntervalSeconds}
              >
                <Input
                  id="subscription-repeat-interval"
                  type="number"
                  min={0}
                  max={MAX_REPEAT_INTERVAL_MINUTES}
                  step={1}
                  inputMode="numeric"
                  className="max-w-40"
                  value={repeatIntervalMinutes}
                  onChange={(event) => setRepeatIntervalMinutes(event.target.value)}
                />
              </FormField>
              <FormField label={t("subscription.remark")} htmlFor="subscription-remark">
                <Input id="subscription-remark" value={remark} placeholder={t("subscription.remarkPlaceholder")} onChange={(event) => setRemark(event.target.value)} />
              </FormField>
            </DialogBody>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter className="mx-0 mb-0 shrink-0 border-t px-4 py-3">
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <DetailSheet
        open={detail !== null}
        title={detail?.name ?? ""}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
          }
        }}
        fields={
          detail
            ? [
                { label: t("detail.uid"), value: detail.uid },
                { label: t("subscription.conditions"), value: conditionSummary(detail) },
                { label: t("subscription.repeatInterval"), value: intervalSummary(detail.repeatIntervalSeconds) },
                {
                  label: t("subscription.groups"),
                  value: (detail.groupUids ?? []).map((uid) => groups.find((group) => group.uid === uid)?.name ?? uid).join(", "),
                },
                { label: t("namespace.status"), value: <StatusBadge value={detail.status} /> },
                { label: t("subscription.remark"), value: detail.remark },
              ]
            : []
        }
      />
      <ConfirmAction
        open={confirm !== null}
        title={confirm?.action === "delete" ? t("confirm.deleteTitle") : t("confirm.disableTitle")}
        description={confirm?.action === "delete" ? t("confirm.deleteDescription") : t("confirm.disableDescription")}
        confirmLabel={t("action.confirm")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void runConfirm()}
        onOpenChange={(value) => {
          if (!value) {
            setConfirm(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
