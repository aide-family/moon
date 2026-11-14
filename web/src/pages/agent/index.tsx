import { Search } from "lucide-react"
import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { formErrorMessage, readFormError } from "@/lib/form-error"

import {
  createAgent,
  deleteAgent,
  deployAgent,
  getAgentDeploy,
  listAgents,
  revokeAgent,
  rotateAgentToken,
  updateAgent,
} from "@/api/agent"
import type { AgentDeployItem, AgentDeployStep, AgentItem } from "@/api/agent/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import type { FilterChip } from "@/components/filter-chips"
import { FormField } from "@/components/form-field"
import {
  LabelEditor,
  labelErrorKey,
  labelsFromJSON,
  labelsToJSON,
  readLabelEntries,
  type LabelRow,
} from "@/components/label-editor"
import { LabelTags } from "@/components/label-tags"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Field, FieldLabel } from "@/components/ui/field"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import { Textarea } from "@/components/ui/textarea"
import { requireTrimmed, collectFieldErrors, hasFieldErrors } from "@/lib/form"
import { TABLE_MIN_WIDTH } from "@/lib/table-layout"
import { useFilterDraft } from "@/hooks/use-filter-draft"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"

import { DeployConsole } from "./deploy-console"

// 安装失败的下一步提示：按失败步骤给出可照做的原因，终端里保留原始报错。
const deployHintKeys: Record<string, string> = {
  connect: "agent.deployHintConnect",
  inspect: "agent.deployHintInspect",
  package: "agent.deployHintPackage",
  upload: "agent.deployHintUpload",
  install: "agent.deployHintInstall",
  configure: "agent.deployHintConfigure",
  service: "agent.deployHintService",
  online: "agent.deployHintOnline",
}

/** 步骤状态的颜色：文字与圆点同色，同时靠文案本身表达状态。 */
function deployStepTone(status: string): string {
  if (status === "SUCCEEDED") {
    return "text-success"
  }
  if (status === "RUNNING") {
    return "text-info"
  }
  if (status === "FAILED") {
    return "text-danger"
  }
  return "text-muted-foreground"
}

function AgentLabelChips({ labels }: { labels: string }) {
  const entries = readLabelEntries(labels)
  if (!entries || entries.length === 0) {
    return <span className="text-xs text-muted-foreground">—</span>
  }
  // 列表里最多展示 3 个，其余折叠成 +N；完整标签在详情抽屉里。
  const shown = entries.slice(0, 3)
  const rest = entries.length - shown.length
  return (
    <div className="flex min-w-0 flex-wrap gap-1">
      {shown.map((entry) => (
        <Badge
          key={entry.key}
          variant="secondary"
          className="max-w-full font-mono text-[11px] font-normal whitespace-normal"
        >
          <span className="min-w-0 break-all">
            <span className="text-muted-foreground">{entry.key}</span>=
            {entry.value}
          </span>
        </Badge>
      ))}
      {rest > 0 ? (
        <Badge variant="outline" className="font-mono text-[11px] font-normal">
          +{rest}
        </Badge>
      ) : null}
    </div>
  )
}

export function AgentPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const session = useSession()
  const query = useListQuery()
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty } = useFilterDraft({ name: query.name })
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<AgentItem | null>(null)
  const [name, setName] = useState("")
  const [sshHost, setSshHost] = useState("")
  const [sshPort, setSshPort] = useState("22")
  const [sshUser, setSshUser] = useState("")
  const [sshPassword, setSshPassword] = useState("")
  const [sshPrivateKey, setSshPrivateKey] = useState("")
  const [remark, setRemark] = useState("")
  const [labelRows, setLabelRows] = useState<LabelRow[]>([])
  const [token, setToken] = useState("")
  const [tokenOpen, setTokenOpen] = useState(false)
  const [autoDeploy, setAutoDeploy] = useState(true)
  const [deployOpen, setDeployOpen] = useState(false)
  const [deploy, setDeploy] = useState<AgentDeployItem | null>(null)
  // 保留整条机器记录：安装失败时可以直接打开编辑对话框改凭据，不用回列表再找一遍。
  const [deployItem, setDeployItem] = useState<AgentItem | null>(null)
  const [deployError, setDeployError] = useState("")
  const [deployPending, setDeployPending] = useState(false)
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<AgentItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "revoke" | "delete"; item: AgentItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("agent:write")
  const canExec = session.can("ssh:exec")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<AgentItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}`,
    fetch: (page, pageSize) => listAgents({ page, pageSize, name: query.name }),
  })

  function openCreate() {
    setEditing(null)
    setName("")
    setSshHost("")
    setSshPort("22")
    setSshUser("")
    setSshPassword("")
    setSshPrivateKey("")
    setRemark("")
    setLabelRows([])
    setAutoDeploy(true)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function openEdit(item: AgentItem) {
    setEditing(item)
    setName(item.name)
    setSshHost(item.sshHost)
    setSshPort(String(item.sshPort || 22))
    setSshUser(item.sshUser)
    setSshPassword("")
    setSshPrivateKey("")
    setRemark(item.remark)
    setLabelRows(labelsFromJSON(item.labels ?? ""))
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const labelResult = labelsToJSON(labelRows)
    const entries: Array<[string, string | null]> = [
      ["name", requireTrimmed(name, t("form.required"))],
      ["labels", labelResult.error ? t(labelErrorKey(labelResult.error)) : null],
    ]
    const shouldDeploy = !editing && autoDeploy
    if (shouldDeploy) {
      entries.push(
        ["sshHost", requireTrimmed(sshHost, t("form.required"))],
        ["sshUser", requireTrimmed(sshUser, t("form.required"))],
        ["sshPassword", requireTrimmed(sshPassword, t("form.required"))]
      )
    }
    const errors = collectFieldErrors(entries)
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const body = {
        name: name.trim(),
        sshHost,
        sshPort: Number(sshPort) || 22,
        sshUser,
        sshPassword,
        sshPrivateKey,
        remark,
        labels: labelResult.json,
      }
      if (editing) {
        await updateAgent({ ...body, uid: editing.uid })
        setOpen(false)
      } else {
        const result = await createAgent(body)
        setOpen(false)
        if (shouldDeploy) {
          void startDeploy(result.agent)
        } else {
          setToken(result.token)
          setTokenOpen(true)
        }
      }
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function startDeploy(item: AgentItem) {
    setDeployItem(item)
    setDeploy(null)
    setDeployError("")
    setDeployPending(true)
    setDeployOpen(true)
    try {
      const result = await deployAgent(item.uid)
      setDeploy(result.deploy)
    } catch (err) {
      setDeployError(formErrorMessage(err, t("state.failed")))
    } finally {
      setDeployPending(false)
    }
  }

  // 安装进度存在服务端内存里，页面每秒拉一次；安装结束后停表并刷新列表。
  useEffect(() => {
    if (!deployOpen || !deployItem || deploy?.status !== "RUNNING") {
      return
    }
    let cancelled = false
    const timer = setInterval(() => {
      getAgentDeploy(deployItem.uid)
        .then((result) => {
          if (cancelled || !result.deploy) {
            return
          }
          setDeploy(result.deploy)
          if (result.deploy.status !== "RUNNING") {
            reload()
          }
        })
        .catch((err) => {
          if (!cancelled) {
            setDeployError(formErrorMessage(err, t("state.failed")))
          }
        })
    }, 1000)
    return () => {
      cancelled = true
      clearInterval(timer)
    }
  }, [deployOpen, deployItem, deploy?.status, reload, t])

  async function onRotate(item: AgentItem) {
    setPending(true)
    try {
      const result = await rotateAgentToken(item.uid)
      setToken(result.token)
      setTokenOpen(true)
      reload()
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
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
        await deleteAgent(confirm.item.uid)
      } else {
        await revokeAgent(confirm.item.uid)
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: AgentItem): RowAction[] {
    const list: RowAction[] = []
    if (canExec && item.online) {
      list.push({ key: "login", label: t("agent.login"), onSelect: () => navigate(`/agents/${item.uid}/shell`) })
    }
    if (!canWrite) {
      return list
    }
    list.push(
      { key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) },
      { key: "deploy", label: t("agent.deploy"), onSelect: () => void startDeploy(item) },
      { key: "rotate", label: t("agent.rotate"), onSelect: () => void onRotate(item) },
      { key: "revoke", label: t("agent.revoke"), destructive: true, onSelect: () => setConfirm({ action: "revoke", item }) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    )
    return list
  }

  const deployStepLabels: Record<string, string> = {
    connect: t("agent.deployStepConnect"),
    inspect: t("agent.deployStepInspect"),
    package: t("agent.deployStepPackage"),
    upload: t("agent.deployStepUpload"),
    install: t("agent.deployStepInstall"),
    configure: t("agent.deployStepConfigure"),
    service: t("agent.deployStepService"),
    online: t("agent.deployStepOnline"),
  }

  function deployStepTitle(step: AgentDeployStep): string {
    return deployStepLabels[step.key] ?? step.title
  }

  const deployTarget = deployItem
    ? `${deployItem.sshUser}@${deployItem.sshHost}:${deployItem.sshPort || 22}`
    : ""
  const failedStep = deploy?.steps.find((step) => step.status === "FAILED")?.key ?? ""
  const deployHint =
    deploy?.status === "FAILED"
      ? t(deployHintKeys[failedStep] ?? "agent.deployHintUnknown")
      : ""

  function deployStatusBadge(status: string): { value: string; label: string } {
    if (status === "SUCCEEDED") {
      return { value: "DONE", label: t("agent.deploySucceeded") }
    }
    if (status === "RUNNING") {
      return { value: "WORKING", label: t("agent.deployRunning") }
    }
    if (status === "FAILED") {
      return { value: "FAILED", label: t("agent.deployFailed") }
    }
    return { value: "PENDING", label: t("agent.deployPending") }
  }

  const filterChips: FilterChip[] = query.name
    ? [
        {
          key: "name",
          label: `${t("agent.name")}: ${query.name}`,
          onRemove: () => query.update({ name: "" }),
        },
      ]
    : []

  function applyFilters() {
    query.update({ name: draft.name.trim() })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ name: "" })
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("agent.title")} description={t("agent.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <ListFilter
        id="agent-filters"
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        chips={filterChips}
        onClearAll={clearFilters}
        onClear={clearFilters}
        appliedCount={filterChips.length}
        dirty={dirty}
        onApply={applyFilters}
      >
        <Field className="min-w-0 w-full max-w-[500px] flex-1 basis-56">
          <FieldLabel
            htmlFor="agent-name-filter"
            className="text-xs text-muted-foreground"
          >
            {t("agent.name")}
          </FieldLabel>
          <div className="relative">
            <Search
              aria-hidden="true"
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="agent-name-filter"
              className="pl-9"
              value={draft.name}
              placeholder={t("agent.namePlaceholder")}
              onChange={(event) => update("name", event.target.value)}
            />
          </div>
        </Field>
      </ListFilter>
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
        {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("agent.empty")}</p> : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("agent.name")}</TableHead>
                <TableHead>{t("agent.labels")}</TableHead>
                <TableHead>{t("agent.hostname")}</TableHead>
                <TableHead>{t("agent.localIp")}</TableHead>
                <TableHead>{t("agent.os")}</TableHead>
                <TableHead>{t("agent.cpu")}</TableHead>
                <TableHead>{t("agent.memory")}</TableHead>
                <TableHead>{t("agent.disk")}</TableHead>
                <TableHead>{t("agent.online")}</TableHead>
                <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell className="max-w-64 whitespace-normal">
                    <AgentLabelChips labels={item.labels ?? ""} />
                  </TableCell>
                  <TableCell>{item.hostname || "-"}</TableCell>
                  <TableCell>{item.localIp || "-"}</TableCell>
                  <TableCell>{item.osVersion || item.os || "-"}</TableCell>
                  <TableCell className="max-w-48 truncate" title={item.cpu}>{item.cpu || "-"}</TableCell>
                  <TableCell>{item.memory || "-"}</TableCell>
                  <TableCell>{item.disk || "-"}</TableCell>
                  <TableCell><StatusBadge value={item.online ? "ONLINE" : "OFFLINE"} label={item.online ? t("agent.onlineYes") : t("agent.onlineNo")} /></TableCell>
                  <TableCell><StatusBadge value={item.status} /></TableCell>
                  <TableCell>
                    <RowActions nowrap detailLabel={t("action.detail")} moreLabel={t("action.more")} onDetail={() => setDetail(item)} actions={actions(item)} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onSave}>
            <DialogBody className="grid gap-4 sm:grid-cols-3">
              <FormField
                label={t("agent.name")}
                htmlFor="agent-name"
                className="sm:col-span-full"
                required
                error={fieldErrors.name}
              >
                <Input id="agent-name" value={name} placeholder={t("agent.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
              </FormField>
              <FormField label={t("agent.sshHost")} htmlFor="agent-host" error={fieldErrors.sshHost}>
                <Input id="agent-host" value={sshHost} placeholder={t("agent.sshHostPlaceholder")} onChange={(event) => setSshHost(event.target.value)} />
              </FormField>
              <FormField label={t("agent.sshPort")} htmlFor="agent-port">
                <Input id="agent-port" value={sshPort} placeholder={t("agent.sshPortPlaceholder")} onChange={(event) => setSshPort(event.target.value)} />
              </FormField>
              <FormField label={t("agent.sshUser")} htmlFor="agent-user" error={fieldErrors.sshUser}>
                <Input id="agent-user" value={sshUser} placeholder={t("agent.sshUserPlaceholder")} onChange={(event) => setSshUser(event.target.value)} />
              </FormField>
              <FormField label={t("agent.sshPassword")} htmlFor="agent-password" className="sm:col-span-full" error={fieldErrors.sshPassword}>
                <Input id="agent-password" type="password" value={sshPassword} onChange={(event) => setSshPassword(event.target.value)} placeholder={t("agent.keepSecret")} />
              </FormField>
              <FormField label={t("agent.sshPrivateKey")} htmlFor="agent-key" className="sm:col-span-full">
                <Textarea id="agent-key" value={sshPrivateKey} onChange={(event) => setSshPrivateKey(event.target.value)} placeholder={editing ? t("agent.keepSecret") : t("agent.sshPrivateKeyPlaceholder")} />
              </FormField>
              <FormField label={t("agent.remark")} htmlFor="agent-remark" className="sm:col-span-full">
                <Textarea id="agent-remark" value={remark} placeholder={t("agent.remarkPlaceholder")} onChange={(event) => setRemark(event.target.value)} />
              </FormField>
              <FormField
                label={t("agent.labels")}
                htmlFor="agent-label-key-0"
                hint={t("agent.labelsHint")}
                className="sm:col-span-full"
                error={fieldErrors.labels}
              >
                <LabelEditor
                  labels={labelRows}
                  onChange={setLabelRows}
                  idPrefix="agent-label"
                />
              </FormField>
              {!editing ? (
                <label className="flex items-start gap-2.5 text-sm sm:col-span-full">
                  <Checkbox
                    checked={autoDeploy}
                    onCheckedChange={(value) => setAutoDeploy(value === true)}
                  />
                  <span>
                    {t("agent.deployAfterCreate")}
                    <span className="mt-1 block text-xs text-muted-foreground">
                      {t("agent.deployAfterCreateHint")}
                    </span>
                  </span>
                </label>
              ) : null}
            </DialogBody>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                {t("action.cancel")}
              </Button>
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={tokenOpen} onOpenChange={setTokenOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("agent.tokenTitle")}</DialogTitle>
          </DialogHeader>
          <DialogBody className="grid gap-4">
            <p className="text-sm text-muted-foreground">{t("agent.tokenHint")}</p>
            <Input readOnly value={token} />
          </DialogBody>
          <DialogFooter>
            <Button onClick={() => setTokenOpen(false)}>{t("action.confirm")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog open={deployOpen} onOpenChange={setDeployOpen}>
        <DialogContent className="sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>{t("agent.deployTitle")}</DialogTitle>
          </DialogHeader>
          <DialogBody className="flex flex-col gap-3">
            <div className="flex flex-wrap items-center gap-2">
              <p className="min-w-0 flex-1 break-words text-sm text-muted-foreground">
                {deployItem
                  ? `${deployItem.name}${deployTarget ? ` · ${deployTarget}` : ""}`
                  : t("agent.deployDescription")}
              </p>
              <StatusBadge {...deployStatusBadge(deploy?.status ?? "PENDING")} />
            </div>
            {deployHint ? (
              <div className="flex flex-wrap items-center gap-2 rounded-md border border-danger-border bg-danger-soft px-3 py-2">
                <p className="min-w-0 flex-1 break-words text-sm text-danger">
                  {deployHint}
                </p>
                {failedStep === "connect" && canWrite && deployItem ? (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setDeployOpen(false)
                      openEdit(deployItem)
                    }}
                  >
                    {t("agent.deployFixCredentials")}
                  </Button>
                ) : null}
              </div>
            ) : null}
            {deploy ? (
              <div className="flex flex-wrap items-baseline gap-x-3 gap-y-1.5">
                <span className="text-xs text-muted-foreground">
                  {t("agent.deploySteps")}
                </span>
                {deploy.steps.map((step) => (
                  <span
                    key={step.key}
                    className={`inline-flex items-center gap-1.5 text-xs ${deployStepTone(step.status)}`}
                  >
                    <span aria-hidden className="size-1.5 rounded-full bg-current" />
                    {deployStepTitle(step)}
                  </span>
                ))}
              </div>
            ) : null}
            <DeployConsole
              title={deployTarget ? `${t("agent.deployLog")} · ${deployTarget}` : t("agent.deployLog")}
              log={deploy?.log ?? ""}
              running={deployPending || deploy?.status === "RUNNING"}
              truncated={deploy?.logTruncated ?? false}
              empty={deployPending ? t("agent.deployWaiting") : t("agent.deployNoRecord")}
              truncatedNotice={t("agent.deployLogTruncated")}
            />
            {deployError ? <p className="text-sm text-destructive">{deployError}</p> : null}
          </DialogBody>
          <DialogFooter>
            {deploy && deploy.status !== "RUNNING" && deployItem ? (
              <Button
                variant="outline"
                disabled={deployPending}
                onClick={() => void startDeploy(deployItem)}
              >
                {t("agent.deployRetry")}
              </Button>
            ) : null}
            <Button onClick={() => setDeployOpen(false)}>{t("agent.deployClose")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <DetailSheet
        wide
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
                { label: t("agent.hostname"), value: detail.hostname || "-" },
                { label: t("agent.machineId"), value: detail.machineId || "-" },
                { label: t("agent.localIp"), value: detail.localIp || "-" },
                { label: t("agent.os"), value: detail.os || "-" },
                { label: t("agent.arch"), value: detail.arch || "-" },
                { label: t("agent.osVersion"), value: detail.osVersion || "-" },
                { label: t("agent.cpu"), value: detail.cpu || "-" },
                { label: t("agent.disk"), value: detail.disk || "-" },
                { label: t("agent.memory"), value: detail.memory || "-" },
                { label: t("agent.bootTime"), value: detail.bootTime || "-" },
                { label: t("agent.timeServer"), value: detail.timeServer || "-" },
                { label: t("agent.version"), value: detail.version || "-" },
                { label: t("agent.online"), value: <StatusBadge value={detail.online ? "ONLINE" : "OFFLINE"} label={detail.online ? t("agent.onlineYes") : t("agent.onlineNo")} /> },
                { label: t("agent.sshHost"), value: detail.sshHost || "-" },
                { label: t("agent.sshPort"), value: String(detail.sshPort || 22) },
                { label: t("agent.sshUser"), value: detail.sshUser || "-" },
                { label: t("agent.remark"), value: detail.remark || "-" },
                {
                  label: t("agent.labels"),
                  value: <LabelTags labels={detail.labels ?? ""} />,
                  span: 2,
                },
              ]
            : []
        }
      />
      <ConfirmAction
        open={confirm !== null}
        title={confirm?.action === "delete" ? t("action.delete") : t("agent.revoke")}
        description={t("agent.confirmBody")}
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
