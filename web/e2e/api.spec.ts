import { expect, test } from "@playwright/test"

import { createApi, E2E_API, mintToken } from "./fixtures"

test.describe("接口契约（真实后端，不 mock）", () => {
  test("公开站点配置带首页内容与能力卡片", async () => {
    const response = await fetch(`${E2E_API}/v1/site/config?locale=zh`)
    expect(response.status).toBe(200)
    const body = await response.json()
    expect(body.site).toBeTruthy()
    expect(body.home).toBeTruthy()
    expect(Array.isArray(body.features)).toBe(true)
  })

  test("平台级接口不需要租户头", async () => {
    const response = await fetch(`${E2E_API}/v1/system/site/home?locale=zh`, {
      headers: { Authorization: `Bearer ${mintToken()}` },
    })
    expect(response.status).toBe(200)
    const body = await response.json()
    expect(body.home).toBeTruthy()
  })

  test("页面乐观锁可往返，过期令牌返回 409", async () => {
    const api = createApi()
    const created = await api.post("/v1/system/site/pages", {
      slug: `e2e-lock-${Date.now()}`,
      locale: "zh",
      title: "乐观锁校验",
      bodyMd: "正文",
    })
    expect(created.status).toBe(200)
    const uid = created.json.page.item.uid

    const detail = await api.get(`/v1/system/site/pages/${uid}`)
    const draft = detail.json.page.draft
    const token = detail.json.page.item.updatedAt
    expect(token).toContain(".")

    const saved = await api.put(`/v1/system/site/pages/${uid}`, {
      ...draft,
      updatedAt: token,
    })
    expect(saved.status).toBe(200)

    const stale = await api.put(`/v1/system/site/pages/${uid}`, {
      ...draft,
      updatedAt: token,
    })
    expect(stale.status).toBe(409)

    await api.delete(`/v1/system/site/pages/${uid}`)
    await api.dispose()
  })

  test("登录配置返回当前生效来源", async () => {
    const api = createApi()
    const email = await api.get("/v1/system/auth/email-settings")
    expect(email.status).toBe(200)
    expect(["DATABASE", "CONFIG", "NONE"]).toContain(email.json.email.source)
    const oauth = await api.get("/v1/system/auth/oauth-providers")
    expect(oauth.status).toBe(200)
    expect(oauth.json.items).toHaveLength(4)
    await api.dispose()
  })
})
