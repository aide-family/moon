import { resolve } from "node:path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": resolve(import.meta.dirname, "./src"),
    },
  },
  optimizeDeps: {
    // monaco-yaml 的 worker 依赖 CJS 包 path-browserify。worker 的依赖默认不预构建，
    // 开发服务器会把它当 ESM 直接加载并报 “module is not defined”，YAML 语言服务
    // 随之退回主线程。显式预构建后 dev 与生产构建行为一致。
    include: ["path-browserify"],
  },
  server: {
    host: "127.0.0.1",
    port: 8080,
    strictPort: true,
    proxy: {
      "/v1": {
        target: process.env.ALIOTH_DEV_PROXY ?? "http://127.0.0.1:8000",
        changeOrigin: true,
        ws: true,
      },
    },
  },
})
