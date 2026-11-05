import { defineConfig } from "@playwright/test"

const WEB = process.env.E2E_WEB_URL ?? "http://127.0.0.1:8080"

// 用本机已安装的 Chrome（channel: "chrome"），不额外下载浏览器。
export default defineConfig({
  testDir: "./e2e",
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: [["list"]],
  use: {
    baseURL: WEB,
    channel: "chrome",
    headless: true,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
})
