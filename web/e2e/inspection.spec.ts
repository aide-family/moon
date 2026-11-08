import { expect, test } from "@playwright/test"

import { createApi, storageState } from "./fixtures"

test.use({ storageState: storageState("light") })

/**
 * 巡检真实后端路径：创建计划 → 手动执行 → 报告快照 → 分享链接 → 公开页。
 * 依赖租户已开通 inspection 能力；未开通时跳过，避免把环境差异报成代码失败。
 */
test.describe("巡检（真实后端，不 mock）", () => {
  test("创建计划、执行、查看报告并公开分享", async ({ page }) => {
    const api = createApi()
    const templates = await api.get("/v1/inspection/templates?page=1&pageSize=20")
    test.skip(
      templates.status === 403,
      "inspection feature is not enabled for this tenant"
    )
    expect(templates.status).toBe(200)
    const template = (
      templates.json as { items: Array<{ uid: string }> }
    ).items[0]
    expect(template).toBeTruthy()

    const name = `e2e-inspection-${Date.now()}`
    const created = await api.post("/v1/inspection/plans", {
      plan: {
        name,
        description: "e2e acceptance",
        cron: "0 9 * * *",
        timezone: "Asia/Shanghai",
        timeoutSeconds: 300,
        templateUid: template.uid,
        outputFormats: ["HTML", "MARKDOWN"],
        titleTemplate: "",
        notificationGroupUids: [],
        notifyPolicy: "ON_ABNORMAL",
        labels: "{}",
        remark: "",
        items: [
          {
            kind: "NOTE",
            title: "巡检说明",
            datasourceUid: "",
            query: "## 巡检结论\n\n**全部正常**。",
            windowSeconds: 3600,
            stepSeconds: 60,
            display: "TEXT",
            params: "{}",
            enabled: true,
          },
        ],
      },
    })
    expect(created.status).toBe(200)
    const planUid = (created.json as { plan: { uid: string } }).plan.uid

    try {
      const triggered = await api.post(`/v1/inspection/plans/${planUid}/run`, {})
      expect(triggered.status).toBe(200)
      const runUid = (triggered.json as { run: { uid: string } }).run.uid

      await expect
        .poll(
          async () => {
            const response = await api.get(`/v1/inspection/runs/${runUid}`)
            return (
              response.json as { run: { status: string } }
            ).run.status
          },
          { timeout: 20_000 }
        )
        .toBe("SUCCEEDED")

      await page.goto("/inspection-runs")
      await expect(page.getByText(name).first()).toBeVisible()
      await page.getByRole("button", { name: "查看报告" }).first().click()
      const reportFrame = page.frameLocator("iframe")
      await expect(reportFrame.locator("body")).toContainText("巡检结论")

      await page.getByRole("button", { name: "分享" }).first().click()
      const shareURL = await page.locator("input[readonly]").first().inputValue()
      expect(shareURL).toContain("/share/report/")

      const publicPage = await page.context().newPage()
      await publicPage.goto(shareURL)
      await expect(
        publicPage.frameLocator("iframe").locator("body")
      ).toContainText("巡检结论")
      await publicPage.close()
    } finally {
      await api.delete(`/v1/inspection/plans/${planUid}`)
    }
  })
})
