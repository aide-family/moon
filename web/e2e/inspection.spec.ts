import { expect, test } from "@playwright/test"

import { createApi, storageState } from "./fixtures"

test.use({ storageState: storageState("light") })

/**
 * 巡检真实后端路径：创建计划 → 手动执行 → 报告快照 → 分享链接 → 公开页。
 * 依赖租户已开通 inspection 能力；未开通时跳过，避免把环境差异报成代码失败。
 */
test.describe("巡检（真实后端，不 mock）", () => {
  test("检查项测试接口执行说明项并拒绝非法参数", async () => {
    const api = createApi()
    const ok = await api.post("/v1/inspection/test-item", {
      item: {
        kind: "NOTE",
        title: "测试说明",
        datasourceUid: "",
        query: "## 结论\n正常",
        windowSeconds: 3600,
        stepSeconds: 60,
        display: "TEXT",
        params: "{}",
        enabled: true,
      },
    })
    expect(ok.status).toBe(200)
    expect((ok.json as { status: string }).status).toBe("OK")

    const bad = await api.post("/v1/inspection/test-item", {
      item: {
        kind: "METRICS_QUERY",
        title: "非法参数",
        datasourceUid: "missing",
        query: "up",
        windowSeconds: 3600,
        stepSeconds: 60,
        display: "METRIC",
        params: '{"unknown":true}',
        enabled: true,
      },
    })
    expect(bad.status).toBe(400)
  })

  test("新建向导按类型切换检查项编辑器", async ({ page }) => {
    page.on("dialog", (dialog) => void dialog.accept())
    await page.goto("/inspection-plans")
    await page.getByRole("button", { name: "新建巡检" }).first().click()
    const dialog = page.getByRole("dialog")
    await dialog.locator("input").first().fill("e2e-editor")
    await page.getByRole("button", { name: "下一步" }).click()

    async function switchKind(name: string) {
      await dialog.getByRole("button", { name }).click()
      const confirm = page.getByRole("alertdialog")
      if (await confirm.isVisible().catch(() => false)) {
        await confirm.getByRole("button", { name: "放弃修改" }).click()
      }
    }
    await switchKind("日志查询")
    await expect(dialog.getByText("级别过滤")).toBeVisible()
    await switchKind("探测目标")
    await expect(dialog.getByText("探测目标组")).toBeVisible()
    await switchKind("告警事件")
    await expect(dialog.getByText("告警级别").first()).toBeVisible()
    await switchKind("静态说明")
    await expect(dialog.getByText("说明内容（Markdown）")).toBeVisible()
    await dialog.locator("textarea:not(.ime-text-area)").first().fill("## 结论\n\n全部正常。")
    await dialog.getByRole("button", { name: "预览结果" }).click()
    const preview = page.getByRole("dialog").last()
    await expect(preview.getByText("全部正常")).toBeVisible()
    await preview.getByRole("button", { name: "取消" }).click()
  })

  test("创建计划、执行、查看报告并公开分享", async ({ page }) => {
    const api = createApi()
    const templates = await api.get("/v1/inspection/templates?page=1&pageSize=20")
    test.skip(
      templates.status === 403,
      "inspection feature is not enabled for this tenant"
    )
    expect(templates.status).toBe(200)
    const templateItems = (
      templates.json as { items: Array<{ uid: string; format?: string }> }
    ).items
    const template =
      templateItems.find((item) => item.format === "HTML") ?? templateItems[0]
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
      const runRow = page.locator("table tbody tr").filter({ hasText: name }).first()
      await expect(runRow).toBeVisible()
      await runRow.getByRole("button", { name: "查看报告" }).click()
      const reportFrame = page.frameLocator("iframe")
      await expect(reportFrame.locator("body")).toContainText("巡检结论")

      await page.getByRole("button", { name: "分享" }).first().click()
      await page.getByRole("button", { name: "生成链接" }).click()
      const shareURL = await page
        .locator("input[readonly]")
        .first()
        .inputValue({ timeout: 10_000 })
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
