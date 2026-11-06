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
})
