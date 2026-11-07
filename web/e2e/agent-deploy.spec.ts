import { expect } from "@playwright/test"

import { storageState, test } from "./fixtures"

// 新增机器时勾选自动安装：服务端用录入的 SSH 凭据尝试安装摇光，
// 页面必须实时显示每一步和命令日志。目标机不可达时也应看到失败原因，
// 因此这里只断言「安装结束」和过程可见，不依赖本机是否有可安装的 Linux 机器。
test.use({ storageState: storageState("light") })

test("新增机器自动安装并显示安装步骤与日志", async ({ page, api }) => {
  const name = `e2e-deploy-${Date.now()}`
  let uid = ""
  try {
    await page.goto("/agents")
    await page.getByRole("button", { name: "新建", exact: true }).click()

    const dialog = page.getByRole("dialog")
    await dialog.locator("#agent-name").fill(name)
    await dialog.locator("#agent-host").fill("127.0.0.1")
    await dialog.locator("#agent-port").fill("2222")
    await dialog.locator("#agent-user").fill("root")
    await dialog.locator("#agent-password").fill("alioth")

    const checkbox = dialog.getByRole("checkbox")
    await expect(checkbox).toBeChecked()
    await dialog.getByRole("button", { name: "保存", exact: true }).click()

    // 保存后不再弹令牌，而是直接打开安装过程。
    const deployDialog = page.getByRole("dialog").filter({ hasText: "安装摇光 Agent" })
    await expect(deployDialog).toBeVisible()
    await expect(deployDialog).toContainText("安装步骤")
    await expect(deployDialog).toContainText("建立 SSH 连接")
    await expect(deployDialog).toContainText("安装日志")
    await expect(deployDialog.getByText(/安装(失败|完成)/).first()).toBeVisible({
      timeout: 50_000,
    })
    const log = await deployDialog.locator("pre").innerText()
    expect(log.length).toBeGreaterThan(0)
    expect(log).toContain("==>")

    // 从列表确认机器已创建（安装失败也会保留记录，方便重试）。
    const created = await api.get("/v1/agents?page=1&pageSize=100&name=" + encodeURIComponent(name))
    const body = created.json as { items?: Array<{ uid: string }> }
    uid = body.items?.[0]?.uid ?? ""
    expect(uid).not.toBe("")
  } finally {
    if (uid) {
      await api.delete(`/v1/agent/${uid}`)
    }
  }
})
