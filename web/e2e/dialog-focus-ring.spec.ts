import { expect, test } from "@playwright/test"

import { SHOT_DIR, storageState } from "./fixtures"

/**
 * 对话框内焦点环的通用回归（对应 web/AGENTS.md 的「Dialogs and modals」）。
 *
 * DialogBody 是滚动容器，裁剪边界压在内容边缘上；表单控件选中时的焦点环是外扩的
 * `ring-3`（3px box-shadow），如果正文没有给焦点环留出净空，左右两侧就会被裁掉，
 * 只剩 1px 的边框色。用例直接量控件的四边净空：任何一边小于 3px 都会重新引入这个缺陷。
 * 同时确认留净空用的负外边距没有移动字段，字段左边界仍与对话框标题一致。
 */
type Clearance = { left: number; right: number; top: number; bottom: number }

type DialogMetrics = {
  titleLeft: number | null
  bodyPaddingLeft: number
  bodyPaddingTop: number
  bodyMarginLeft: number
  bodyMarginTop: number
  overflowX: number
  controls: { slot: string; clearance: Clearance; left: number }[]
}

const CONTROLS = '[data-slot="input"], textarea, [data-slot="select-trigger"]'

test.use({ storageState: storageState("light") })

async function measureDialog(page: import("@playwright/test").Page) {
  return page.evaluate((selector): DialogMetrics => {
    const dialog = document.querySelector<HTMLElement>('[data-slot="dialog-content"]')
    const body = dialog?.querySelector<HTMLElement>('[data-slot="dialog-body"]')
    const title = dialog?.querySelector<HTMLElement>('[data-slot="dialog-title"]')
    if (!dialog || !body) {
      throw new Error("对话框或 DialogBody 不存在")
    }
    const bodyStyle = getComputedStyle(body)
    const bodyRect = body.getBoundingClientRect()
    // 滚动容器按内边距盒裁剪，没有边框时内边距盒就是元素本身的矩形。
    const clip = {
      left: bodyRect.left + parseFloat(bodyStyle.borderLeftWidth),
      right: bodyRect.right - parseFloat(bodyStyle.borderRightWidth),
      top: bodyRect.top + parseFloat(bodyStyle.borderTopWidth),
      bottom: bodyRect.bottom - parseFloat(bodyStyle.borderBottomWidth),
    }
    const controls = [...body.querySelectorAll<HTMLElement>(selector)].map((control) => {
      const rect = control.getBoundingClientRect()
      return {
        slot: control.dataset.slot ?? control.tagName.toLowerCase(),
        clearance: {
          left: rect.left - clip.left,
          right: clip.right - rect.right,
          top: rect.top - clip.top,
          bottom: clip.bottom - rect.bottom,
        },
        left: rect.left,
      }
    })
    return {
      titleLeft: title ? title.getBoundingClientRect().left : null,
      bodyPaddingLeft: parseFloat(bodyStyle.paddingLeft),
      bodyPaddingTop: parseFloat(bodyStyle.paddingTop),
      bodyMarginLeft: parseFloat(bodyStyle.marginLeft),
      bodyMarginTop: parseFloat(bodyStyle.marginTop),
      overflowX: body.scrollWidth - body.clientWidth,
      controls,
    }
  }, CONTROLS)
}

async function openDialog(page: import("@playwright/test").Page, path: string, action: string) {
  await page.goto(path, { waitUntil: "domcontentloaded" })
  const button = page.getByRole("button", { name: action, exact: true }).first()
  await expect(button).toBeVisible()
  await button.click()
  await expect(page.locator('[data-slot="dialog-content"]')).toBeVisible()
  // 等 zoom-in 动画结束，避免按缩放中的矩形量净空。
  await page.waitForTimeout(250)
}

test.describe("对话框焦点环不被裁切", () => {
  for (const viewport of [
    { width: 1440, height: 900, label: "desktop" },
    { width: 390, height: 844, label: "narrow" },
  ]) {
    test(`标准三段式对话框的字段焦点环完整（${viewport.label}）`, async ({ page }) => {
      await page.setViewportSize(viewport)
      await openDialog(page, "/roles", "新建")
      const metrics = await measureDialog(page)

      if (!metrics.controls.length) {
        throw new Error("对话框正文里没有找到表单控件")
      }
      for (const control of metrics.controls) {
        // ring-3 外扩 3px；净空小于 3px 就会被滚动容器裁掉。
        expect(control.clearance.left, `${control.slot} 左侧焦点环被裁`).toBeGreaterThanOrEqual(3)
        expect(control.clearance.right, `${control.slot} 右侧焦点环被裁`).toBeGreaterThanOrEqual(3)
      }
      // 留净空用的是等量负外边距，字段不能因此移动，否则标题与字段左边界会错位。
      expect(Math.abs((metrics.titleLeft ?? 0) - metrics.controls[0].left)).toBeLessThanOrEqual(1)
      expect(metrics.overflowX).toBeLessThanOrEqual(1)
    })
  }

  test("自带内边距的对话框字段仍与标题对齐", async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 })
    await openDialog(page, "/alert-maintenance-windows", "新建维护窗口")
    const metrics = await measureDialog(page)

    // 这类对话框用 DialogContent p-0 + DialogBody px-4 py-4：自己的内边距就是净空，
    // 用 m-0 抵消基线的负外边距，字段左边界必须和标题一致。
    expect(metrics.bodyMarginLeft).toBe(0)
    expect(metrics.bodyMarginTop).toBe(0)
    expect(metrics.bodyPaddingLeft).toBeGreaterThanOrEqual(3)
    expect(metrics.bodyPaddingTop).toBeGreaterThanOrEqual(3)
    expect(Math.abs((metrics.titleLeft ?? 0) - metrics.controls[0].left)).toBeLessThanOrEqual(1)
    for (const control of metrics.controls) {
      expect(control.clearance.left).toBeGreaterThanOrEqual(3)
      expect(control.clearance.right).toBeGreaterThanOrEqual(3)
    }
    await page.screenshot({ path: `${SHOT_DIR}/dialog-focus-ring-padded-1440-light.png` })
  })

  test("短视口下正文滚到底，页脚与焦点环仍然完整", async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await openDialog(page, "/alert-maintenance-windows", "新建维护窗口")
    const scrolled = await page.evaluate(() => {
      const body = document.querySelector<HTMLElement>('[data-slot="dialog-body"]')
      if (!body) return null
      body.scrollTop = body.scrollHeight
      return {
        scrollTop: body.scrollTop,
        atEnd: Math.abs(body.scrollTop + body.clientHeight - body.scrollHeight) <= 2,
        scrollable: body.scrollHeight > body.clientHeight,
      }
    })
    if (!scrolled) {
      throw new Error("DialogBody 不存在")
    }
    // 内容够长时正文必须真的能滚到底；不够长时至少不能出现「滚不动」的裁切。
    if (scrolled.scrollable) {
      expect(scrolled.atEnd).toBe(true)
    }
    await expect(page.locator('[data-slot="dialog-footer"]')).toBeVisible()
    await expect(page.locator('[data-slot="dialog-header"]')).toBeVisible()
    const metrics = await measureDialog(page)
    for (const control of metrics.controls) {
      expect(control.clearance.left).toBeGreaterThanOrEqual(3)
      expect(control.clearance.right).toBeGreaterThanOrEqual(3)
    }
    await page.screenshot({ path: `${SHOT_DIR}/dialog-focus-ring-narrow-390-light.png` })
  })
})
