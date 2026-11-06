import { expect, test } from "@playwright/test"

import { storageState } from "./fixtures"

test.use({ storageState: storageState("light") })

test.describe("站点设置能力卡片", () => {
  test("拖动卡片可以调整顺序", async ({ page }) => {
    await page.goto("/system/site")
    await page
      .getByRole("button", { name: "首页内容", exact: true })
      .first()
      .click()

    const section = page.locator("#site-settings-home")
    const titles = section.locator('input[id^="feature-title-"]')
    await titles.first().waitFor({ timeout: 30_000 })
    const firstTitle = await titles.nth(0).inputValue()
    const secondTitle = await titles.nth(1).inputValue()
    expect(firstTitle).not.toBe(secondTitle)

    const rows = section.locator("[data-sort-key]")
    const grip = rows.nth(1).locator("svg.lucide-grip-vertical")
    await grip.scrollIntoViewIfNeeded()
    await page.waitForTimeout(300)
    const handle = await grip.boundingBox()
    expect(handle).not.toBeNull()

    await page.mouse.move(handle!.x + 8, handle!.y + 8)
    await page.mouse.down()
    await page.mouse.move(handle!.x + 8, handle!.y + 22, { steps: 4 })
    await page.waitForTimeout(150)
    // 拖动开始后源卡片脱离列表流，目标位置要重新量；落点在第一张卡片上缘。
    const target = await rows.nth(0).boundingBox()
    expect(target).not.toBeNull()
    await page.mouse.move(target!.x + target!.width / 2, target!.y + 8, {
      steps: 14,
      delay: 20,
    })
    await page.waitForTimeout(150)
    await page.mouse.up()

    await expect(titles.nth(0)).toHaveValue(secondTitle)
    await expect(titles.nth(1)).toHaveValue(firstTitle)
  })
})
