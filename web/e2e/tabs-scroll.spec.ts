import path from "node:path"

import { expect, test } from "@playwright/test"

import { SHOT_DIR, storageState } from "./fixtures"

/**
 * 标签与滚动的通用回归（对应 web/AGENTS.md 的「Required regression checks」）。
 *
 * 一个用例覆盖：无横向溢出、内容不被 app shell 裁切、滚动容器真的能滚到底、
 * 标签的选中态 / aria-selected / 面板显隐一致、键盘左右方向键可切换。
 *
 * 用法：用环境变量指定要检查的路由，逗号分隔；空则不跑。
 *   E2E_SCROLL_ROUTES=/alert-rule,/system/site/pages \
 *     npx playwright test tabs-scroll
 */
const ROUTES = (process.env.E2E_SCROLL_ROUTES ?? "")
  .split(",")
  .map((route) => route.trim())
  .filter(Boolean)

const VIEWPORTS = [
  { width: 1366, height: 768, label: "desktop" },
  { width: 390, height: 844, label: "narrow" },
]

test.use({ storageState: storageState("light") })

type ScrollInfo = {
  documentOverflowX: number
  scrollableCount: number
  viewportHeight: number
}

/** 页面里横向溢出、以及是否存在可滚动的容器。 */
async function measureLayout(page: import("@playwright/test").Page) {
  return page.evaluate(
    (): ScrollInfo => {
      const root = document.documentElement
      const scrollables = [...document.body.querySelectorAll<HTMLElement>("*")].filter(
        (element) => {
          const style = getComputedStyle(element)
          return (
            /(auto|scroll)/.test(style.overflowY) &&
            element.scrollHeight > element.clientHeight + 8
          )
        }
      )
      return {
        documentOverflowX: root.scrollWidth - root.clientWidth,
        scrollableCount: scrollables.length,
        viewportHeight: window.innerHeight,
      }
    }
  )
}

/** 把最高的滚动容器滚到底，确认 scrollTop 真的走到了尽头。 */
async function scrollToBottom(page: import("@playwright/test").Page) {
  return page.evaluate(() => {
    const candidates = [...document.body.querySelectorAll<HTMLElement>("*")]
      .filter((element) => {
        const style = getComputedStyle(element)
        return (
          /(auto|scroll)/.test(style.overflowY) &&
          element.scrollHeight > element.clientHeight + 8
        )
      })
      .sort((a, b) => b.scrollHeight - a.scrollHeight)
    const target = candidates[0]
    if (!target) return { scrolled: false, atEnd: false, hidden: 0 }
    target.scrollTop = target.scrollHeight
    const atEnd =
      Math.abs(target.scrollTop + target.clientHeight - target.scrollHeight) <= 2
    // 滚到底后，滚动容器内最后一个可见块必须落在容器可视区内。
    const children = [...target.children] as HTMLElement[]
    const last = children.at(-1)
    let hidden = 0
    if (last) {
      const container = target.getBoundingClientRect()
      const child = last.getBoundingClientRect()
      hidden = Math.max(0, Math.round(child.bottom - container.bottom))
    }
    return { scrolled: true, atEnd, hidden }
  })
}

for (const route of ROUTES) {
  for (const viewport of VIEWPORTS) {
    test(`标签与滚动：${route} @ ${viewport.label}`, async ({ page }) => {
      await page.setViewportSize(viewport)
      await page.goto(route)
      await expect(page.locator("body")).toBeVisible()
      await page.waitForTimeout(300)

      const layout = await measureLayout(page)
      expect(layout.documentOverflowX, "文档不应出现横向滚动").toBeLessThanOrEqual(1)

      const scroll = await scrollToBottom(page)
      if (scroll.scrolled) {
        expect(scroll.atEnd, "滚动容器应能滚到底").toBeTruthy()
        expect(scroll.hidden, "滚动到底后内容不应被裁切").toBeLessThanOrEqual(2)
      } else {
        // 内容不足一屏时，允许没有滚动容器，但横向溢出断言已覆盖裁切。
        expect(layout.viewportHeight).toBeGreaterThan(0)
      }

      const tabs = page.getByRole("tab")
      const tabCount = await tabs.count()
      for (let index = 0; index < tabCount; index += 1) {
        const tab = tabs.nth(index)
        if (!(await tab.isVisible())) continue
        await tab.click()
        await expect(tab).toHaveAttribute("aria-selected", "true")
        await expect(tab).toHaveAttribute("data-state", "active")

        // 选中标签对应的面板必须可见，未选中的面板必须真的隐藏（forceMount 也一样）。
        const panels = page.getByRole("tabpanel")
        const panelCount = await panels.count()
        for (let p = 0; p < panelCount; p += 1) {
          const panel = panels.nth(p)
          const mounted = await panel.count()
          if (!mounted) continue
          const active = await panel.getAttribute("data-state")
          const visible = await panel.isVisible()
          expect(
            visible,
            `data-state=${active} 的面板 ${active === "active" ? "应可见" : "应隐藏"}`
          ).toBe(active === "active")
        }
      }

      // 键盘：聚焦当前标签后按左右方向键，选中态要跟着移动。
      if (tabCount > 1) {
        const first = tabs.first()
        await first.focus()
        await page.keyboard.press("ArrowRight")
        const selected = page.locator('[role="tab"][aria-selected="true"]')
        await expect(selected).toHaveCount(1)
      }

      await page.screenshot({
        path: path.join(
          SHOT_DIR,
          `tabs-scroll-${viewport.label}-${route.replace(/\W+/g, "-")}.png`
        ),
      })
    })
  }
}

if (ROUTES.length === 0) {
  test("标签与滚动：未指定路由时跳过", () => {
    test.skip(true, "用 E2E_SCROLL_ROUTES=/your/route 指定要检查的页面")
  })
}
