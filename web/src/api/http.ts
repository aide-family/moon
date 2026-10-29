import i18n from "@/i18n"
import { toast } from "sonner"

const baseURL = "/v1"

const messageKeys: Record<string, string> = {
  "captcha is invalid": "error.captchaInvalid",
  "email already exists": "error.emailExists",
  "invalid email or code": "error.invalidCode",
  "email is not registered": "error.emailNotRegistered",
  "invalid email address": "error.invalidEmail",
  "invalid username": "register.usernameInvalid",
  "send code too frequently": "error.sendTooFrequent",
  "too many verification attempts": "error.tooManyAttempts",
  "too many bind email addresses": "error.tooManyBindEmails",
  "oauth login failed": "error.oauthFailed",
  // 绑定失败走回调页的失败块，这里补一条本地化映射，避免直接露英文原文。
  "oauth bind failed": "profile.oauth.linkFailed",
  "oauth provider is not configured": "error.oauthNotConfigured",
  "oauth binding is invalid or expired": "error.oauthBindingExpired",
  // 登录链路（补邮箱）也用这句原文，必须保留无插值的 key；绑定冲突的 409
  // 由个人资料页与回调页按 status 自己提示 profile.oauth.alreadyBound。
  "oauth account is already bound": "error.oauthAccountBound",
  // 解绑的 404 按「已经解绑」处理，复用页面通用的未找到文案。
  "oauth link not found": "error.notFound",
  "cannot unlink the last login method": "profile.oauth.unbindLastCredential",
  "member already exists": "error.memberExists",
  "notification group is already bound to this on-call":
    "error.groupAlreadyBound",
  "invitation expired": "error.invitationExpired",
  "invitation is not pending": "error.invitationHandled",
  "invitation belongs to another user": "error.invitationNotYours",
  "invitation email does not match": "error.invitationNotYours",
  "invalid datasource endpoint": "form.invalidUrl",
  "datasource connection already exists": "datasource.connectionExists",
  "name is too long": "form.nameTooLong",
  // Kubernetes 集群接口的后端错误原文，避免中文界面直接露英文。
  "invalid kubeconfig": "error.invalidKubeconfig",
  "kubeconfig is empty": "error.invalidKubeconfig",
  "kubeconfig is required": "error.invalidKubeconfig",
  "kubeconfig must embed credentials; exec, auth-provider, tokenFile and certificate files are not supported":
    "error.clusterUnsupportedKubeconfig",
  "kubeconfig proxy-url is not supported": "error.clusterUnsupportedKubeconfig",
  "kubeconfig must embed the certificate authority; certificate-authority file paths are not supported":
    "error.clusterUnsupportedKubeconfig",
  "cluster is unreachable": "error.clusterUnreachable",
  "cluster credential is not allowed to perform this operation":
    "error.clusterForbidden",
  "kubeconfig credential was rejected by the cluster":
    "error.clusterCredentialRejected",
  "cluster resource conflict": "error.clusterConflict",
  "cluster is rate limited": "error.clusterRateLimited",
  "cluster is unavailable": "error.clusterUnavailable",
  "pod has no containers": "error.podNoContainers",
  "access token required": "error.accessTokenRequired",
  "refresh token required": "error.refreshTokenRequired",
  "invalid token": "error.invalidToken",
  "forbidden": "error.forbidden",
  "permission denied": "error.permissionDenied",
  "user is banned": "error.userBanned",
  "user is not active": "error.userInactive",
  "cannot change super admin": "error.cannotChangeSuperAdmin",
  "cannot change another system admin": "error.cannotChangeAnotherSystemAdmin",
  "cannot change your own status": "error.cannotChangeOwnStatus",
  "cannot change your own system role": "error.cannotChangeOwnSystemRole",
  "cannot change owner status": "error.cannotChangeOwnerStatus",
  "invalid system role": "error.invalidSystemRole",
  "owner role is assigned by transfer owner": "error.ownerRoleByTransfer",
  "namespace is required": "error.namespaceRequired",
  "namespace is disabled": "error.namespaceDisabled",
  "namespace mismatch": "error.namespaceMismatch",
  "member not found": "error.memberNotFound",
  "member is not joined": "error.memberNotJoined",
  "member must accept invitation": "error.memberMustAcceptInvitation",
  "member has no email": "error.memberNoEmail",
  "member has no phone": "error.memberNoPhone",
  "member is already merged": "error.memberAlreadyMerged",
  "invalid invitation": "error.invalidInvitation",
  "invitation not found": "error.invitationNotFound",
  "name already exists": "error.nameExists",
  "too many members": "error.tooManyMembers",
  "datasource is disabled": "error.datasourceDisabled",
  "datasource is not logs": "error.datasourceNotLogs",
  "datasource is not metrics": "error.datasourceNotMetrics",
  "datasource does not support alert charts": "error.datasourceNoAlertCharts",
  "unsupported datasource driver": "error.unsupportedDriver",
  "invalid datasource index pattern": "error.invalidIndexPattern",
  "invalid label matcher": "error.invalidLabelMatcher",
  // 探测目标组 / 探测规则的校验错误原文。
  "at least one probe target is required": "error.probeTargetRequired",
  "too many probe targets": "error.probeTargetTooMany",
  "duplicate probe target": "error.probeTargetDuplicate",
  "invalid probe target": "error.probeTargetInvalid",
  "probe target is required": "error.probeTargetInvalid",
  "probe target must be http or https": "error.probeTargetInvalid",
  "probe target host is required": "error.probeTargetInvalid",
  "probe target host is blocked": "error.probeTargetBlocked",
  "unsupported probe driver": "error.probeTargetDriver",
  "unsupported probe method": "error.probeMethodInvalid",
  "invalid probe request": "error.probeRequestInvalid",
  "probe timeout out of range": "error.probeTimeoutInvalid",
  "invalid expected status range": "error.probeStatusInvalid",
  "invalid expected status code": "error.probeStatusInvalid",
  "probe header key is required": "error.probeHeaderKey",
  "probe header is controlled by the client": "error.probeHeaderControlled",
  "probe header contains newline": "error.probeHeaderNewline",
  "probe body keyword is too long": "error.probeKeywordTooLong",
  "invalid probe interval": "error.probeIntervalInvalid",
  "invalid probe metric": "error.probeMetricInvalid",
  "remark is too long": "probeTarget.remarkTooLong",
  "invalid query expression": "error.invalidQueryExpression",
  "upstream query failed": "error.upstreamQueryFailed",
  "invalid smtp endpoint": "error.invalidSmtpEndpoint",
  "invalid smtp port": "error.invalidSmtpPort",
  "invalid webhook endpoint": "error.invalidWebhookEndpoint",
  "webhook endpoint must not contain credentials": "error.webhookEndpointCredentials",
  "level is in use": "error.levelInUse",
  "group still has rules": "error.groupHasRules",
  "invalid alert event status": "error.invalidAlertEventStatus",
  "invalid alert event time range": "error.invalidAlertEventTimeRange",
  "invalid time range": "error.invalidTimeRange",
  "invalid time zone": "error.invalidTimeZone",
  "invalid suppression filter": "error.invalidSuppressionFilter",
  "history filters require HISTORY view": "error.historyFiltersNeedHistoryView",
  "alert analysis range must be within 30 days": "error.analysisRangeTooLong",
  "alert analysis range must not end in the future": "error.analysisRangeFuture",
  "alert event time range must be within 30 days": "error.eventRangeTooLong",
  "time range must be within 30 days": "error.rangeTooLong",
  "focus window must be within the analysis range": "error.focusWindowOutsideRange",
  "duplicate variable name": "error.duplicateVariableName",
  "empty variable expr": "error.emptyVariableExpr",
  "invalid variable name": "error.invalidVariableName",
  "reserved variable name": "error.reservedVariableName",
  "too many variables": "error.tooManyVariables",
  "variable expr is too long": "error.variableExprTooLong",
  "invalid summary template": "error.invalidSummaryTemplate",
  "invalid description template": "error.invalidDescriptionTemplate",
  "channel is disabled": "error.channelDisabled",
  "channel is used by a notification group": "error.channelInUseByGroup",
  "channel type and app do not match": "error.channelTypeAppMismatch",
  "invalid channel type": "error.invalidChannelType",
  "template is disabled": "error.templateDisabled",
  "template is used by a notification group": "error.templateInUseByGroup",
  "template is required for every channel": "error.templateRequiredPerChannel",
  "channel and template counts do not match": "error.channelTemplateCountMismatch",
  "duplicate notification group": "error.duplicateGroup",
  "group is bound to a notification subscription": "error.groupBoundSubscription",
  "group is bound to a notification on-call": "error.groupBoundOncall",
  "notification group requires at least one channel": "error.groupNeedsChannel",
  "only one email channel is allowed": "error.onlyOneEmailChannel",
  "only one sms channel is allowed": "error.onlyOneSmsChannel",
  "email channel requires at least one receiver": "error.emailChannelNeedsReceiver",
  "email receivers require an email channel": "error.emailReceiversNeedChannel",
  "sms channel requires an sms template": "error.smsChannelNeedsTemplate",
  "sms channel requires at least one receiver": "error.smsChannelNeedsReceiver",
  "sms receivers require an sms channel": "error.smsReceiversNeedChannel",
  "sms delivery is not implemented": "error.smsNotImplemented",
  "sms access key is required": "error.smsAccessKeyRequired",
  "sms sign name is required": "error.smsSignNameRequired",
  "template body is required": "error.templateBodyRequired",
  "receivers are required": "error.receiversRequired",
  "repeat interval seconds is out of range": "error.repeatIntervalOutOfRange",
  "too many receivers": "error.tooManyReceivers",
  "too many channels": "error.tooManyChannels",
  "too many subscription filters": "error.tooManySubscriptionFilters",
  "invalid on-call timezone": "error.invalidOncallTimezone",
  "invalid on-call weekday": "error.invalidOncallWeekday",
  "invalid on-call windows": "error.invalidOncallWindows",
  "on-call window requires at least one weekday": "error.oncallWindowNeedsWeekday",
  "on-call window start and end must differ": "error.oncallWindowSameEdge",
  "on-call window time must be HH:mm": "error.oncallWindowTimeFormat",
  "too many on-call filters": "error.tooManyOncallFilters",
  "too many on-call groups": "error.tooManyOncallGroups",
  "too many on-call time windows": "error.tooManyOncallWindows",
  "agent is offline": "error.agentOffline",
  "agent is revoked": "error.agentRevoked",
  "select an online agent": "error.selectOnlineAgent",
  "command is not approved": "error.commandNotApproved",
  "name and command are required": "error.nameCommandRequired",
  "ssh credentials missing": "error.sshCredentialsMissing",
  "invalid ssh private key": "error.invalidSshPrivateKey",
  "execute cancelled": "error.executeCancelled",
  "execute timed out": "error.executeTimedOut",
  "node name is required": "error.nodeNameRequired",
  "invalid resource kind": "error.invalidResourceKind",
  "resource name is required": "error.resourceNameRequired",
  "resource yaml is too large to preview": "error.yamlTooLarge",
  "namespace and name are required": "error.namespaceNameRequired",
  // 功能套餐：基础套餐不可改名 / 删除，被租户使用的套餐不可删除。
  "base package cannot be renamed": "packageAdmin.baseNameLocked",
  "base package cannot be deleted": "packageAdmin.baseNotDeletable",
  "package is in use": "packageAdmin.packageInUse",
  "invalid feature": "packageAdmin.invalidFeature",
}

// 服务端错误带 reason（CLIENT / INVALID / CONFLICT 等）。message 没有具体映射时用它兜底，
// 保证中文界面永远拿得到本地化文案，而不是把后端英文原文当主文案显示。
const reasonKeys: Record<string, string> = {
  CLIENT: "error.reason.CLIENT",
  UNAUTHENTICATED: "error.reason.UNAUTHENTICATED",
  FORBIDDEN: "error.reason.FORBIDDEN",
  NOT_FOUND: "error.reason.NOT_FOUND",
  CONFLICT: "error.reason.CONFLICT",
  INVALID: "error.reason.INVALID",
  SERVER: "error.reason.SERVER",
}

const fieldByMessage: Record<string, string> = {
  "captcha is invalid": "captchaCode",
  "email already exists": "email",
  "email is not registered": "email",
  "invalid username": "name",
}

export class ApiError extends Error {
  status: number
  fields: Record<string, string>
  /** 服务端原始 message（未本地化）。需要按业务含义分支时用它，不要比对本地化后的文本。 */
  raw: string
  /** 服务端错误携带的附加信息，例如发码节流的 retryAfterSeconds。 */
  metadata: Record<string, string>
  /** 服务端错误原因（CLIENT / INVALID / CONFLICT 等），未提供时为空串。 */
  reason: string

  constructor(
    status: number,
    message: string,
    fields: Record<string, string> = {},
    raw = "",
    metadata: Record<string, string> = {},
    reason = ""
  ) {
    super(message)
    this.status = status
    this.fields = fields
    this.raw = raw
    this.metadata = metadata
    this.reason = reason
  }
}

export function localizeMessage(message: string, reason = ""): string {
  const key = messageKeys[message]
  if (key) {
    return i18n.t(key)
  }
  const fallback = reasonKeys[reason]
  return fallback ? i18n.t(fallback) : message
}

function fieldName(raw: string): string {
  const snake = raw
    .replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`)
    .replace(/^_/, "")
  return snake.replace(/_([a-z])/g, (_, letter: string) => letter.toUpperCase())
}

export function parseFieldErrors(message: string): Record<string, string> {
  const known = fieldByMessage[message]
  if (known) {
    return { [known]: localizeMessage(message) }
  }
  const fields: Record<string, string> = {}
  const pattern = /invalid\s+[\w.]+\.(\w+)\s*:\s*([^;\n]+)/gi
  for (const match of message.matchAll(pattern)) {
    const name = fieldName(match[1])
    fields[name] = localizeMessage(match[2].trim())
  }
  return fields
}

function staysOnAuthForm(path: string): boolean {
  return path.startsWith("/auth/") || path === "/captcha"
}

function notifyStatus(status: number, message: string) {
  if (status === 404) {
    toast(message || i18n.t("error.notFound"))
    return
  }
  if (status >= 500) {
    toast.error(i18n.t("error.server"), {
      description: message || undefined,
      duration: 8000,
    })
  }
}

let refreshing: Promise<boolean> | null = null

function clearSession(): void {
  localStorage.removeItem("token")
  localStorage.removeItem("refreshToken")
  if (window.location.pathname !== "/login") {
    window.location.assign("/login")
  }
}

// refreshAccessToken 用 refresh token 换新的 access token。
// 并发 401 共享同一个刷新请求，避免同时发出多次刷新。
async function refreshAccessToken(): Promise<boolean> {
  const refreshToken = localStorage.getItem("refreshToken")
  if (!refreshToken) {
    return false
  }
  if (!refreshing) {
    refreshing = (async () => {
      try {
        const headers = new Headers({ Authorization: `Bearer ${refreshToken}` })
        const namespace = localStorage.getItem("namespace")
        if (namespace) {
          headers.set("X-Namespace", namespace)
        }
        const response = await fetch(`${baseURL}/self/refresh-token`, {
          headers,
        })
        if (!response.ok) {
          return false
        }
        const data = (await response.json()) as { accessToken?: string }
        if (!data.accessToken) {
          return false
        }
        localStorage.setItem("token", data.accessToken)
        return true
      } catch {
        return false
      }
    })()
  }
  try {
    return await refreshing
  } finally {
    refreshing = null
  }
}

export async function request(
  path: string,
  init: RequestInit = {},
  retry = true
): Promise<Response> {
  const headers = new Headers(init.headers)
  const token = localStorage.getItem("token") ?? ""
  const namespace = localStorage.getItem("namespace") ?? ""
  if (token) {
    headers.set("Authorization", `Bearer ${token}`)
  }
  // 切换租户时会显式传入目标租户，不能被 localStorage 里的旧值覆盖。
  if (namespace && !headers.has("X-Namespace")) {
    headers.set("X-Namespace", namespace)
  }
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json")
  }

  let response: Response
  try {
    response = await fetch(`${baseURL}${path}`, { ...init, headers })
  } catch (err) {
    // 浏览器取消请求（页面切换、输入防抖）继续向上抛，避免误报成网关错误。
    if (err instanceof DOMException && err.name === "AbortError") {
      throw err
    }
    // 后端不可达时 fetch 只给出浏览器原文（Failed to fetch）。统一包装成 502，
    // 列表页才能真正渲染标准错误页，而不是把英文原文塞在内容区。
    throw new ApiError(
      502,
      i18n.t("error.badGatewayTitle"),
      {},
      err instanceof Error ? err.message : ""
    )
  }
  if (
    response.status === 401 &&
    retry &&
    !staysOnAuthForm(path) &&
    path !== "/self/refresh-token"
  ) {
    if (await refreshAccessToken()) {
      return request(path, init, false)
    }
    clearSession()
  }
  if (
    response.status === 403 &&
    !path.startsWith("/invitation/") &&
    window.location.pathname !== "/forbidden"
  ) {
    window.location.assign("/forbidden")
  }
  return response
}

export interface RequestOptions {
  // 后台轮询和列表加载由页面自己展示错误，不弹全局 toast。
  silent?: boolean
}

export async function send<T>(
  path: string,
  init: RequestInit = {},
  options: RequestOptions = {}
): Promise<T> {
  const response = await request(path, init)
  const text = await response.text()
  let data: {
    message?: string
    metadata?: Record<string, string>
    reason?: string
  } = {}
  if (text) {
    try {
      data = JSON.parse(text) as {
        message?: string
        metadata?: Record<string, string>
        reason?: string
      }
    } catch {
      data = { message: text }
    }
  }
  if (!response.ok) {
    const raw = data.message || `request failed: ${response.status}`
    const message = localizeMessage(raw, data.reason ?? "")
    const fields = response.status === 400 ? parseFieldErrors(raw) : {}
    if (
      !options.silent &&
      (response.status === 404 || response.status >= 500)
    ) {
      notifyStatus(response.status, message)
    }
    throw new ApiError(
      response.status,
      message,
      fields,
      raw,
      data.metadata ?? {},
      data.reason ?? ""
    )
  }
  return data as T
}

// 同一份 GET 的并发请求共享已解析结果：React StrictMode 会让 effect 执行两次，
// 这里把重复请求合并成一次网络调用。promise settle 后立即从表里移除，
// 不做结果缓存——之后的重试、刷新、轮询仍然会真正访问网络。
const inFlightGets = new Map<string, Promise<unknown>>()

function requestNamespace(init?: RequestInit): string {
  const headers = new Headers(init?.headers)
  return headers.get("X-Namespace") ?? localStorage.getItem("namespace") ?? ""
}

export function get<T>(
  path: string,
  params?: object,
  init?: RequestInit,
  options?: RequestOptions
): Promise<T> {
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(params ?? {})) {
    if (
      typeof value !== "string" &&
      typeof value !== "number" &&
      typeof value !== "boolean"
    ) {
      continue
    }
    if (value === "" || value === false) {
      continue
    }
    query.set(key, String(value))
  }
  const suffix = query.size > 0 ? `?${query.toString()}` : ""
  const pathWithQuery = `${path}${suffix}`
  // AbortSignal 有独立的取消语义（验证码刷新），不能与其它调用共享同一个请求。
  const shareable = !init?.signal
  const key = `${pathWithQuery}|${requestNamespace(init)}|${options?.silent ? "silent" : "loud"}`
  if (shareable) {
    const pending = inFlightGets.get(key)
    if (pending) {
      return pending as Promise<T>
    }
  }
  const promise = send<T>(pathWithQuery, init, options)
  if (shareable) {
    inFlightGets.set(key, promise)
    const clear = () => {
      // 只清理自己这一笔；已被新请求替换时不能误删后来者。
      if (inFlightGets.get(key) === promise) {
        inFlightGets.delete(key)
      }
    }
    void promise.then(clear, clear)
  }
  return promise
}

export function post<T>(path: string, body?: unknown): Promise<T> {
  return send<T>(path, { method: "POST", body: JSON.stringify(body ?? {}) })
}

export function put<T>(path: string, body?: unknown): Promise<T> {
  return send<T>(path, { method: "PUT", body: JSON.stringify(body ?? {}) })
}

export function del<T>(path: string, options?: RequestOptions): Promise<T> {
  return send<T>(path, { method: "DELETE" }, options)
}
