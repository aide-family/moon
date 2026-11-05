// 真实后端冒烟：不做任何业务接口 mock，用真实令牌打开真实页面。
// 覆盖 M33 这一轮踩过的关键路径；新需求按同样结构补检查项。
//
// 用法：
//   SMOKE_TOKEN=<jwt> SMOKE_NAMESPACE=<tenant-uid> node web/scripts/smoke-site.mjs
// 可选：SMOKE_WEB_URL（默认 http://127.0.0.1:8080）、SMOKE_API_URL（默认同 WEB）、
//       SMOKE_CHROME（Chrome 可执行文件路径）、SMOKE_SHOTS（截图目录）
import fs from "node:fs"
import path from "node:path"
import { spawn } from "node:child_process"

const WEB = process.env.SMOKE_WEB_URL ?? "http://127.0.0.1:8080"
const API = process.env.SMOKE_API_URL ?? WEB
const TOKEN = process.env.SMOKE_TOKEN ?? ""
const NAMESPACE = process.env.SMOKE_NAMESPACE ?? ""
const CHROME =
  process.env.SMOKE_CHROME ??
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
const SHOTS = process.env.SMOKE_SHOTS ?? path.resolve(import.meta.dirname, "../../tmp/smoke")

if (!TOKEN) {
  console.error("缺少 SMOKE_TOKEN：先用本地 dev 用户签一个 JWT（不要走验证码登录）")
  process.exit(2)
}

fs.mkdirSync(SHOTS, { recursive: true })
const results = []
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

function check(name, ok, detail = "") {
  results.push({ name, ok, detail })
  console.log(`${ok ? "PASS" : "FAIL"} ${name}${detail ? ` — ${detail}` : ""}`)
}

async function api(method, requestPath, body) {
  const response = await fetch(`${API}${requestPath}`, {
    method,
    headers: {
      Accept: "application/json",
      Authorization: `Bearer ${TOKEN}`,
      "X-Namespace": NAMESPACE,
      ...(body ? { "Content-Type": "application/json" } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  })
  const text = await response.text()
  let json = null
  try {
    json = text ? JSON.parse(text) : null
  } catch {
    // 非 JSON 响应保留文本
  }
  return { status: response.status, json, text }
}

// ---------------------------------------------------------------- 接口层
console.log("== 接口")
const publicConfig = await api("GET", "/v1/site/config?locale=zh")
check(
  "公开站点配置可读且带首页内容",
  publicConfig.status === 200 &&
    Boolean(publicConfig.json?.site) &&
    Boolean(publicConfig.json?.home) &&
    Array.isArray(publicConfig.json?.features),
  `status=${publicConfig.status}`
)

const pages = await api(
  "GET",
  "/v1/system/site/pages?page=1&pageSize=20&keyword=&state=&locale=zh"
)
const firstPage = pages.json?.items?.[0]
check("官网页面列表可读", pages.status === 200 && Boolean(firstPage), `status=${pages.status}`)

if (firstPage) {
  const detail = await api("GET", `/v1/system/site/pages/${firstPage.uid}`)
  const token = detail.json?.page?.item?.updatedAt
  const draft = detail.json?.page?.draft
  check("页面详情可读且带锁令牌", detail.status === 200 && Boolean(token), `token=${token ?? ""}`)

  if (token && draft) {
    const saved = await api("PUT", `/v1/system/site/pages/${firstPage.uid}`, {
      title: draft.title,
      summary: draft.summary,
      bodyMd: draft.bodyMd,
      coverImageUrl: draft.coverImageUrl,
      seoTitle: draft.seoTitle,
      seoDescription: draft.seoDescription,
      updatedAt: token,
    })
    check("带最新 updatedAt 保存成功（乐观锁可往返）", saved.status === 200, `status=${saved.status}`)

    const stale = await api("PUT", `/v1/system/site/pages/${firstPage.uid}`, {
      title: draft.title,
      summary: draft.summary,
      bodyMd: draft.bodyMd,
      coverImageUrl: draft.coverImageUrl,
      seoTitle: draft.seoTitle,
      seoDescription: draft.seoDescription,
      updatedAt: token,
    })
    check("过期 updatedAt 被拒绝（返回 409）", stale.status === 409, `status=${stale.status}`)
  }
}

const email = await api("GET", "/v1/system/auth/email-settings")
check(
  "登录配置能带出当前生效来源",
  email.status === 200 && ["DATABASE", "CONFIG", "NONE"].includes(email.json?.email?.source),
  `source=${email.json?.email?.source}`
)

// ---------------------------------------------------------------- 浏览器层
console.log("== 页面")
const chrome = spawn(
  CHROME,
  [
    "--headless=new",
    "--disable-gpu",
    "--no-sandbox",
    "--hide-scrollbars",
    "--remote-debugging-port=9444",
    `--user-data-dir=${path.join(SHOTS, "chrome-profile")}`,
    "--window-size=1440,900",
    "about:blank",
  ],
  { stdio: "ignore" }
)

async function waitJSON(url, attempts = 80) {
  for (let index = 0; index < attempts; index += 1) {
    try {
      const response = await fetch(url)
      if (response.ok) return await response.json()
    } catch {
      // Chrome 还没起来
    }
    await sleep(500)
  }
  throw new Error("Chrome DevTools 未就绪")
}

const list = await waitJSON("http://127.0.0.1:9444/json/list")
const target = list.find((item) => item.type === "page")
const ws = new WebSocket(target.webSocketDebuggerUrl)
await new Promise((resolve, reject) => {
  ws.onopen = resolve
  ws.onerror = reject
})

let nextId = 0
const pending = new Map()
const consoleErrors = []
ws.onmessage = (event) => {
  const message = JSON.parse(event.data)
  if (message.id && pending.has(message.id)) {
    const job = pending.get(message.id)
    pending.delete(message.id)
    if (message.error) job.reject(new Error(JSON.stringify(message.error)))
    else job.resolve(message.result)
    return
  }
  if (message.method === "Runtime.exceptionThrown") {
    consoleErrors.push(message.params.exceptionDetails?.text ?? "exception")
  }
}

function send(method, params = {}) {
  const id = ++nextId
  ws.send(JSON.stringify({ id, method, params }))
  return new Promise((resolve, reject) => pending.set(id, { resolve, reject }))
}

async function openPage(pagePath, screenshotName) {
  await send("Page.navigate", { url: `${WEB}${pagePath}` })
  await sleep(4000)
  const shot = await send("Page.captureScreenshot", { format: "png", fromSurface: true })
  fs.writeFileSync(path.join(SHOTS, `${screenshotName}.png`), Buffer.from(shot.data, "base64"))
  const probe = await send("Runtime.evaluate", {
    expression: `JSON.stringify({
      text: document.body.innerText.slice(0, 4000),
      scrollWidth: document.documentElement.scrollWidth,
      clientWidth: document.documentElement.clientWidth,
      title: document.title,
    })`,
  })
  return JSON.parse(probe.result.value)
}

await send("Page.enable")
await send("Runtime.enable")
await send("Emulation.setDeviceMetricsOverride", {
  width: 1440,
  height: 900,
  deviceScaleFactor: 1,
  mobile: false,
})

// 先回同源页写会话，再逐页验证。
await send("Page.navigate", { url: `${WEB}/` })
await sleep(1500)
await send("Runtime.evaluate", {
  expression: `localStorage.setItem("token", ${JSON.stringify(TOKEN)});
    localStorage.setItem("namespace", ${JSON.stringify(NAMESPACE)});
    localStorage.setItem("theme", "light");
    localStorage.setItem("lang", "zh"); true`,
})

const home = await openPage("/", "site-home")
check("官网首页有标题且无横向滚动", home.text.length > 40 && home.scrollWidth <= home.clientWidth + 1)

if (firstPage) {
  const doc = await openPage(`/p/${firstPage.slug}`, "site-document")
  check(
    "官网文档页可访问且无横向滚动",
    doc.text.length > 20 && doc.scrollWidth <= doc.clientWidth + 1,
    `slug=${firstPage.slug}`
  )

  const preview = await openPage(`/site-preview/${firstPage.uid}`, "site-preview")
  check("后台草稿预览页可打开并显示预览横幅", preview.text.includes("预览"))
}

const settings = await openPage("/system/site", "site-settings")
check(
  "站点设置页的首页内容加载成功",
  settings.text.includes("首页内容") && !settings.text.includes("首页内容加载失败")
)
check(
  "站点设置页有左侧区块导航",
  ["站点与备案", "登录方式", "邮件服务", "第三方登录", "首页内容"].every((label) =>
    settings.text.includes(label)
  )
)

ws.close()
chrome.kill("SIGKILL")

check("页面无未捕获异常", consoleErrors.length === 0, consoleErrors.slice(0, 3).join(" | "))

const failed = results.filter((item) => !item.ok)
console.log(`\n${results.length - failed.length}/${results.length} passed，截图在 ${SHOTS}`)
process.exit(failed.length === 0 ? 0 : 1)
