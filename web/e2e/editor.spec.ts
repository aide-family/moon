import { expect, test } from "@playwright/test"

import { createApi, SHOT_DIR, storageState } from "./fixtures"

test.use({ storageState: storageState("light") })

async function seedDraft() {
  const api = createApi()
  const created = await api.post("/v1/system/site/pages", {
    slug: `e2e-editor-${Date.now()}`,
    locale: "zh",
    title: "编辑器内联预览",
    bodyMd: "## 小标题\n\n正文一段。",
  })
  return { api, uid: created.json.page.item.uid as string }
}

test.describe("官网页面编辑器", () => {
  test("内联预览在对话框里渲染完整页面", async ({ page }) => {
    const { api, uid } = await seedDraft()
    try {
      await page.goto(`/system/site/pages?uid=${uid}`)
      await expect(page.getByRole("heading", { name: "官网页面" })).toBeVisible()
      await page.getByRole("button", { name: "内联预览", exact: true }).click()
      const dialog = page.getByRole("dialog")
      await expect(dialog).toBeVisible()
      await expect(dialog.locator("h1")).toHaveText("编辑器内联预览")
      await expect(dialog.getByRole("link", { name: "在新标签打开" })).toBeVisible()
      await page.screenshot({ path: `${SHOT_DIR}/editor-inline-preview-1440-light.png` })
      await dialog.getByRole("button", { name: "关闭预览" }).click()
    } finally {
      await api.delete(`/v1/system/site/pages/${uid}`)
      await api.dispose()
    }
  })

  test("保存冲突可以覆盖保存", async ({ page }) => {
    const { api, uid } = await seedDraft()
    try {
      await page.goto(`/system/site/pages?uid=${uid}`)
      await expect(page.getByRole("heading", { name: "官网页面" })).toBeVisible()

      const title = page.locator("#site-page-title")
      await title.fill("我的本地修改")
      await page.getByRole("button", { name: "保存草稿" }).click()
      await expect(page.getByText("草稿已保存。")).toBeVisible()

      // 模拟另一个管理员保存：直接改一次 updatedAt。
      const detail = await api.get(`/v1/system/site/pages/${uid}`)
      const draft = detail.json.page.draft
      const serverSide = await api.put(`/v1/system/site/pages/${uid}`, {
        ...draft,
        updatedAt: detail.json.page.item.updatedAt,
      })
      expect(serverSide.status).toBe(200)

      await title.fill("覆盖后的标题")
      await page.getByRole("button", { name: "保存草稿" }).click()
      const dialog = page.getByRole("dialog", { name: /官网已有其他人保存/ })
      await expect(dialog).toBeVisible()
      await page.screenshot({ path: `${SHOT_DIR}/editor-conflict-1440-light.png` })
      await dialog.getByRole("button", { name: /保留我的修改并覆盖保存/ }).click()
      await expect(page.getByText("已用你的修改覆盖保存。")).toBeVisible()

      const saved = await api.get(`/v1/system/site/pages/${uid}`)
      expect(saved.json.page.draft.title).toBe("覆盖后的标题")
    } finally {
      await api.delete(`/v1/system/site/pages/${uid}`)
      await api.dispose()
    }
  })

  test("窄屏深色编辑器无横向滚动", async ({ browser }) => {
    const context = await browser.newContext({
      storageState: storageState("dark"),
      viewport: { width: 390, height: 844 },
    })
    const page = await context.newPage()
    const { api, uid } = await seedDraft()
    try {
      await page.goto(`/system/site/pages?uid=${uid}`)
      await expect(page.getByRole("heading", { name: "官网页面" })).toBeVisible()
      const overflow = await page.evaluate(
        () => document.documentElement.scrollWidth - document.documentElement.clientWidth
      )
      expect(overflow).toBeLessThanOrEqual(1)
      await page.screenshot({ path: `${SHOT_DIR}/editor-390-dark.png`, fullPage: true })
    } finally {
      await api.delete(`/v1/system/site/pages/${uid}`)
      await api.dispose()
      await context.close()
    }
  })
})
