import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { formMessage, readFormError } from "@/lib/form-error"

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
import type { NotificationSubscriptionItem } from "@/api/notification-subscription/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
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
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"
import { parseLabelPairs, serializeLabelPairs, type LabelPair } from "@/lib/labels"
import { run } from "@/lib/action"

export function NotificationSubscriptionPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [groups, setGroups] = useState<GroupOption[]>([])
  const [levels, setLevels] = useState<AlertDictionaryItem[]>([])
  const [datasources, setDatasources] = useState<DatasourceOption[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<NotificationSubscriptionItem | null>(null)
  const [name, setName] = useState("")
  const [labels, setLabels] = useState<LabelPair[]>([{ key: "", value: "" }])
  const [levelUids, setLevelUids] = useState<string[]>([])
  const [datasourceUids, setDatasourceUids] = useState<string[]>([])
  const [groupUids, setGroupUids] = useState<string[]>([])
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

  useEffect(() => {
    let active = true
    void Promise.all([
      selectNotificationGroups({ limit: 100 }),
      listAlertDictionaries({ page: 1, pageSize: 100, type: "ALERT" }),
      selectDatasources({ limit: 100 }),
    ])
      .then(([groupResult, levelResult, datasourceResult]) => {
        if (active) {
          setGroups(groupResult.items ?? [])
          setLevels(levelResult.items ?? [])
          setDatasources(datasourceResult.items ?? [])
        }
      })
      .catch(() => {
        if (active) {
          setGroups([])
          setLevels([])
          setDatasources([])
        }
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  function openCreate() {
    setEditing(null)
    setName("")
    setLabels([{ key: "", value: "" }])
    setLevelUids([])
    setDatasourceUids([])
    setGroupUids([])
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
    setGroupUids(item.groupUids ?? [])
    setRemark(item.remark)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function toggle(values: string[], uid: string): string[] {
    return values.includes(uid) ? values.filter((value) => value !== uid) : [...values, uid]
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
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const payload = {
        name: name.trim(),
        labels: serializeLabelPairs(labels),
        levelUids,
        datasourceUids,
        groupUids,
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
      setConfirmError(formMessage(err, t("state.failed")))
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
        : { key: "enable", label: t("action.enable"), onSelect: () => run(updateNotificationSubscriptionStatus(item.uid, "ACTIVE").then(() => reload()), setError) },
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
                <TableHead>{t("namespace.status")}</TableHead>
                <TableHead className="text-center">{t("action.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell className="max-w-md text-sm text-muted-foreground">{conditionSummary(item)}</TableCell>
                  <TableCell>
                    {(item.groupUids ?? []).map((uid) => groups.find((group) => group.uid === uid)?.name ?? uid).join(", ")}
                  </TableCell>
                  <TableCell><StatusBadge value={item.status} /></TableCell>
                  <TableCell>
                    <RowActions
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
        <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="grid gap-4" onSubmit={onSave}>
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
              hint={t("subscription.groupsHint")}
              required
              error={fieldErrors.groups}
            >
              <div
                className={`flex flex-wrap gap-2 rounded-md ${fieldErrors.groups ? "border border-destructive p-2" : ""}`}
              >
                {groups.map((item) => (
                  <label
                    key={item.uid}
                    className="inline-flex cursor-pointer items-center gap-2 rounded-md border bg-background px-3 py-1.5 text-sm"
                  >
                    <input
                      type="checkbox"
                      checked={groupUids.includes(item.uid)}
                      onChange={() => setGroupUids(toggle(groupUids, item.uid))}
                    />
                    {item.name}
                  </label>
                ))}
                {groups.length === 0 ? <p className="text-sm text-muted-foreground">{t("subscription.noGroups")}</p> : null}
              </div>
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
            <FormField label={t("subscription.levels")} hint={t("subscription.levelsHint")}>
              <div className="flex flex-wrap gap-2 rounded-md">
                {levels.map((item) => (
                  <label
                    key={item.uid}
                    className="inline-flex cursor-pointer items-center gap-2 rounded-md border bg-background px-3 py-1.5 text-sm"
                  >
                    <input type="checkbox" checked={levelUids.includes(item.uid)} onChange={() => setLevelUids(toggle(levelUids, item.uid))} />
                    {item.name}
                  </label>
                ))}
                {levels.length === 0 ? <p className="text-sm text-muted-foreground">{t("subscription.noLevels")}</p> : null}
              </div>
            </FormField>
            <FormField label={t("subscription.datasources")} hint={t("subscription.datasourcesHint")}>
              <div className="flex flex-wrap gap-2 rounded-md">
                {datasources.map((item) => (
                  <label
                    key={item.uid}
                    className="inline-flex cursor-pointer items-center gap-2 rounded-md border bg-background px-3 py-1.5 text-sm"
                  >
                    <input
                      type="checkbox"
                      checked={datasourceUids.includes(item.uid)}
                      onChange={() => setDatasourceUids(toggle(datasourceUids, item.uid))}
                    />
                    {item.name}
                  </label>
                ))}
                {datasources.length === 0 ? <p className="text-sm text-muted-foreground">{t("subscription.noDatasources")}</p> : null}
              </div>
            </FormField>
            <FormField label={t("subscription.remark")} htmlFor="subscription-remark">
              <Input id="subscription-remark" value={remark} placeholder={t("subscription.remarkPlaceholder")} onChange={(event) => setRemark(event.target.value)} />
            </FormField>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
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
