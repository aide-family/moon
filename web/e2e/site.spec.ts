import { expect, test } from "@playwright/test"

import { SHOT_DIR, createApi, storageState } from "./fixtures"

test.use({ storageState: storageState("light") })

test.describe("官网公开面", () => {
  test("首页可读且无横向滚动（桌面与窄屏）", async ({ page }) => {
    await page.goto("/")
    await expect(page.locator("h1")).toBeVisible()
    await expect(page.locator("h1")).not.toHaveText("")
    for (const viewport of [
      { width: 1440, height: 900 },
      { width: 390, height: 844 },
    ]) {
      await page.setViewportSize(viewport)
      await page.waitForTimeout(300)
      const overflow = await page.evaluate(
        () =>
          document.documentElement.scrollWidth -
          document.documentElement.clientWidth
      )
      expect(overflow).toBeLessThanOrEqual(1)
    }
    await page.screenshot({ path: `${SHOT_DIR}/site-home-390-light.png` })
  })

  test("文档页渲染 Markdown 代码高亮与右侧目录", async ({ page }) => {
    const slug = `e2e-md-${Date.now()}`
    const api = createApi()
    const created = await api.post("/v1/system/site/pages", {
      slug,
      locale: "zh",
      title: "Markdown 验收",
      bodyMd:
        "# Markdown 验收\n\n## 配置示例\n\n```ts\nconst answer: number = 42\n```\n\n## 下一步\n\n正文。",
    })
    const uid = created.json.page.item.uid
    await api.post(`/v1/system/site/pages/${uid}/publish`, {})

    try {
      await page.setViewportSize({ width: 1440, height: 900 })
      await page.goto(`/p/${slug}`)
      await expect(page.locator("h1")).toBeVisible()

      // 真语法高亮：代码里出现 hljs 关键字 span，且颜色不是正文色。
      const keyword = page
        .locator("pre code .hljs-keyword, pre code .hljs-built_in")
        .first()
      await expect(keyword).toBeVisible()
      const color = await keyword.evaluate(
        (node) => getComputedStyle(node).color
      )
      expect(color).not.toBe("rgb(31, 35, 40)")

      const toc = page.getByRole("navigation", { name: /本页目录/ })
      await expect(toc).toBeVisible()
      await expect(toc.getByRole("link", { name: "配置示例" })).toBeVisible()

      await page.screenshot({
        path: `${SHOT_DIR}/site-doc-1440-light.png`,
        fullPage: true,
      })

      // 窄屏：目录隐藏、无横向滚动。
      await page.setViewportSize({ width: 390, height: 844 })
      await expect(toc).toBeHidden()
      const overflow = await page.evaluate(
        () =>
          document.documentElement.scrollWidth -
          document.documentElement.clientWidth
      )
      expect(overflow).toBeLessThanOrEqual(1)
    } finally {
      await api.delete(`/v1/system/site/pages/${uid}`)
      await api.dispose()
    }
  })

  test("深色主题下代码块是深底浅字", async ({ browser }) => {
    const context = await browser.newContext({
      storageState: storageState("dark"),
      viewport: { width: 1440, height: 900 },
    })
    const page = await context.newPage()
    const slug = `e2e-md-dark-${Date.now()}`
    const api = createApi()
    const created = await api.post("/v1/system/site/pages", {
      slug,
      locale: "zh",
      title: "深色 Markdown",
      bodyMd: "# 深色\n\n```bash\naliothctl version\n```",
    })
    const uid = created.json.page.item.uid
    await api.post(`/v1/system/site/pages/${uid}/publish`, {})
    try {
      await page.goto(`/p/${slug}`)
      const block = page.locator('[data-slot="code-block"]').first()
      await expect(block).toBeVisible()
      const background = await block.evaluate(
        (node) => getComputedStyle(node).backgroundColor
      )
      expect(background).toBe("rgb(22, 27, 34)")
      await page.screenshot({ path: `${SHOT_DIR}/site-doc-dark-390.png` })
    } finally {
      await api.delete(`/v1/system/site/pages/${uid}`)
      await api.dispose()
      await context.close()
    }
  })

  test("预览页显示草稿与预览横幅", async ({ page }) => {
    const api = createApi()
    const created = await api.post("/v1/system/site/pages", {
      slug: `e2e-preview-${Date.now()}`,
      locale: "zh",
      title: "草稿预览页",
      bodyMd: "草稿正文",
    })
    const uid = created.json.page.item.uid
    try {
      await page.goto(`/site-preview/${uid}`)
      await expect(page.getByRole("note")).toContainText("预览")
      await expect(page.locator("h1")).toHaveText("草稿预览页")
    } finally {
      await api.delete(`/v1/system/site/pages/${uid}`)
      await api.dispose()
    }
  })
})

test.describe("后台设置页", () => {
  test.beforeEach(async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 })
  })

  test("左侧快捷导航跳转后标题仍然可见", async ({ page }) => {
    await page.goto("/system/site")
    await expect(page.getByRole("heading", { name: "站点设置" })).toBeVisible()
    await page
      .getByRole("button", { name: "首页内容", exact: true })
      .first()
      .click()
    await page.waitForTimeout(600)
    const headingBox = await page
      .getByRole("heading", { name: "站点设置" })
      .boundingBox()
    expect(headingBox?.y ?? -1).toBeGreaterThanOrEqual(0)
    await expect(page.getByLabel("主标题")).toBeVisible()
  })

  test("首页内容用共享原语渲染，且有内联预览", async ({ page }) => {
    await page.goto("/system/site")
    await page
      .getByRole("button", { name: "首页内容", exact: true })
      .first()
      .click()
    await page.waitForTimeout(400)

    // 用的是共享原语（Radix Select 的 trigger 与 Textarea 的 data-slot），
    // 而不是页面里手写的原生控件。
    await expect(
      page.locator('[data-slot="select-trigger"]').first()
    ).toBeVisible()
    await expect(page.locator('[data-slot="textarea"]').first()).toBeVisible()

    await page.getByLabel("主标题").fill("E2E 主标题")
    await expect(page.getByText("有未保存的改动")).toBeVisible()

    await page.getByRole("button", { name: "预览首页" }).click()
    const dialog = page.getByRole("dialog")
    await expect(dialog).toBeVisible()
    await expect(dialog.locator("h1")).toHaveText("E2E 主标题")
    await page.screenshot({
      path: `${SHOT_DIR}/settings-home-preview-1440-light.png`,
    })
    await dialog.getByRole("button", { name: "关闭预览" }).click()
  })

  test("切换语言会提示丢弃改动", async ({ page }) => {
    await page.goto("/system/site")
    await page
      .getByRole("button", { name: "首页内容", exact: true })
      .first()
      .click()
    await page.waitForTimeout(400)
    await page.getByLabel("主标题").fill("未保存的标题")
    await page.getByRole("button", { name: "English" }).click()
    const dialog = page.getByRole("dialog")
    await expect(dialog).toBeVisible()
    await expect(dialog).toContainText("切换语言")
    await dialog.getByRole("button", { name: "放弃改动并切换" }).click()
    await expect(page.getByRole("dialog")).toBeHidden()
  })

  test("窄屏深色下设置页无横向滚动", async ({ browser }) => {
    const context = await browser.newContext({
      storageState: storageState("dark"),
      viewport: { width: 390, height: 844 },
    })
    const page = await context.newPage()
    await page.goto("/system/site")
    await page
      .getByRole("button", { name: "首页内容", exact: true })
      .first()
      .click()
    await page.waitForTimeout(400)
    const overflow = await page.evaluate(
      () =>
        document.documentElement.scrollWidth -
        document.documentElement.clientWidth
    )
    expect(overflow).toBeLessThanOrEqual(1)
    await page.screenshot({
      path: `${SHOT_DIR}/settings-390-dark.png`,
      fullPage: true,
    })
    await context.close()
  })
})
