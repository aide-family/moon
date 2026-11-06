import { expect, test } from "@playwright/test"

import { SHOT_DIR, storageState } from "./fixtures"

test.use({ storageState: storageState("light") })

test.describe("升级策略步骤列表", () => {
  test("添加按钮在列表末尾，追加步骤后仍在视野内", async ({ page }) => {
    await page.goto("/alert-escalation-policies")
    await page.getByRole("button", { name: "新建升级策略" }).first().click()

    const dialog = page.getByRole("dialog")
    const addStep = dialog.getByRole("button", { name: "添加步骤" })
    const stepLabels = dialog.getByText(/^第 \d+ 步$/)
    await expect(addStep).toBeVisible()
    await addStep.scrollIntoViewIfNeeded()
    await expect(stepLabels).toHaveCount(1)

    // 按钮在步骤卡片下方，而不是挤在「升级步骤」标题右侧。
    const firstStepBox = await stepLabels.first().boundingBox()
    const initialAddBox = await addStep.boundingBox()
    expect(firstStepBox).not.toBeNull()
    expect(initialAddBox).not.toBeNull()
    expect(initialAddBox!.y).toBeGreaterThan(firstStepBox!.y)

    // 连续追加两步：新步骤插在按钮上方，按钮不能被顶出视口，否则又得滚上去点。
    await addStep.click()
    await addStep.click()
    await expect(stepLabels).toHaveCount(3)

    const viewport = page.viewportSize()
    expect(viewport).not.toBeNull()
    await expect
      .poll(async () => {
        const box = await addStep.boundingBox()
        return box ? Math.round(box.y + box.height) : null
      })
      .toBeLessThanOrEqual(viewport!.height)

    const lastStepBox = await stepLabels.last().boundingBox()
    const tailBox = await addStep.boundingBox()
    expect(lastStepBox).not.toBeNull()
    expect(tailBox).not.toBeNull()
    expect(tailBox!.y).toBeGreaterThan(lastStepBox!.y)

    await page.screenshot({
      path: `${SHOT_DIR}/escalation-steps-1440-light.png`,
    })
  })
  test("拖动卡片可以调整步骤顺序", async ({ page }) => {
    await page.goto("/alert-escalation-policies")
    await page.getByRole("button", { name: "新建升级策略" }).first().click()

    const dialog = page.getByRole("dialog")
    const addStep = dialog.getByRole("button", { name: "添加步骤" })
    await addStep.scrollIntoViewIfNeeded()
    await addStep.click()
    await addStep.click()
    // 添加后会把列表尾部平滑滚动到视野内，等它停下来再量位置。
    await page.waitForTimeout(900)

    const delays = dialog.locator('input[type="number"]')
    await expect(delays).toHaveCount(3)
    await expect(delays.nth(0)).toHaveValue("5")

    const rows = dialog.locator("[data-sort-key]")
    const grip = rows.nth(1).locator("svg.lucide-grip-vertical")
    await grip.scrollIntoViewIfNeeded()
    const handle = await grip.boundingBox()
    expect(handle).not.toBeNull()
    await page.mouse.move(handle!.x + 8, handle!.y + 8)
    await page.mouse.down()
    await page.mouse.move(handle!.x + 8, handle!.y + 22, { steps: 4 })
    await page.waitForTimeout(120)
    // 拖动开始后源卡片脱离列表流，目标位置要重新量。
    const target = await rows.nth(0).boundingBox()
    expect(target).not.toBeNull()
    await page.mouse.move(target!.x + target!.width / 2, target!.y + 8, {
      steps: 12,
      delay: 20,
    })
    await page.waitForTimeout(120)
    await page.mouse.up()

    // 第 2 步（延迟 10）被拖到最前面。
    await expect(delays.nth(0)).toHaveValue("10")
    await expect(delays.nth(1)).toHaveValue("5")
  })
})
