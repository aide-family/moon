import { loader, type Monaco } from "@monaco-editor/react"
import * as monacoEditor from "monaco-editor"
import { createWebWorker as createWebWorkerInternal } from "monaco-editor/esm/vs/editor/editor.main.js"
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker"
import htmlWorker from "monaco-editor/esm/vs/language/html/html.worker?worker"
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker"
import { configureMonacoYaml, type MonacoYaml } from "monaco-yaml"
import yamlWorker from "monaco-yaml/yaml.worker?worker"

import {
  GENERIC_KUBERNETES_SCHEMA,
  KUBERNETES_YAML_SCHEMAS,
} from "@/components/kubernetes-yaml-schema"

// Monaco 只初始化一次：worker 工厂、loader 配置、GitHub 主题与 YAML 语言服务
// 都在这里收口。页面组件只调用 ensureMonacoRuntime()，不再各自配置。
let runtimeReady = false
let themesReady = false
let yamlReady = false
let workerApiPatched = false
let yamlController: MonacoYaml | null = null

function ensureWorkerEnvironment() {
  window.MonacoEnvironment = {
    getWorker(_workerId, label) {
      if (label === "yaml") {
        return new yamlWorker()
      }
      if (label === "json") {
        return new jsonWorker()
      }
      if (label === "html" || label === "markdown" || label === "handlebars" || label === "razor") {
        return new htmlWorker()
      }
      return new editorWorker()
    },
  }
}

/** 在模块加载时执行一次，把 monaco 交给 @monaco-editor/react 的 loader。 */
export function ensureMonacoRuntime() {
  if (runtimeReady) {
    return
  }
  runtimeReady = true
  ensureWorkerEnvironment()
  loader.config({ monaco: monacoEditor })
}

// 配色直接取自 Shiki 的 github-light / github-dark，与 CodeBlock 的
// --code-bg / --code-fg 是同一套 GitHub 主题，代码块和编辑器在明暗主题下保持一致。
const githubLightRules = [
  { token: "comment", foreground: "6a737d" },
  { token: "string", foreground: "032f62" },
  { token: "string.key", foreground: "005cc5" },
  { token: "string.value", foreground: "032f62" },
  { token: "constant", foreground: "005cc5" },
  { token: "keyword", foreground: "d73a49" },
  { token: "number", foreground: "005cc5" },
  { token: "type", foreground: "22863a" },
  { token: "tag", foreground: "22863a" },
  { token: "attribute", foreground: "6f42c1" },
  { token: "variable", foreground: "e36209" },
  { token: "operator", foreground: "24292e" },
  { token: "operators", foreground: "24292e" },
  { token: "delimiter", foreground: "24292e" },
  { token: "meta", foreground: "6a737d" },
  { token: "anchor", foreground: "005cc5" },
  { token: "invalid", foreground: "b31d28" },
]

const githubDarkRules = [
  { token: "comment", foreground: "6a737d" },
  { token: "string", foreground: "9ecbff" },
  { token: "string.key", foreground: "79b8ff" },
  { token: "string.value", foreground: "9ecbff" },
  { token: "constant", foreground: "79b8ff" },
  { token: "keyword", foreground: "f97583" },
  { token: "number", foreground: "79b8ff" },
  { token: "type", foreground: "85e89d" },
  { token: "tag", foreground: "85e89d" },
  { token: "attribute", foreground: "b392f0" },
  { token: "variable", foreground: "ffab70" },
  { token: "operator", foreground: "e1e4e8" },
  { token: "operators", foreground: "e1e4e8" },
  { token: "delimiter", foreground: "e1e4e8" },
  { token: "meta", foreground: "6a737d" },
  { token: "anchor", foreground: "79b8ff" },
  { token: "invalid", foreground: "fdaeb7" },
]

function setupGithubThemes(monaco: Monaco) {
  if (themesReady) {
    return
  }
  themesReady = true
  monaco.editor.defineTheme("github-light", {
    base: "vs",
    inherit: true,
    rules: githubLightRules,
    colors: {
      "editor.background": "#ffffff",
      "editor.foreground": "#24292e",
      "editorCursor.foreground": "#044289",
      "editor.lineHighlightBackground": "#f6f8fa",
      "editor.selectionBackground": "#0366d625",
      "editor.inactiveSelectionBackground": "#0366d611",
      "editorIndentGuide.background1": "#eff2f6",
      "editorIndentGuide.activeBackground1": "#d7dbe0",
      "editorWidget.background": "#f6f8fa",
      "editorWidget.border": "#d1d5da",
      "editorSuggestWidget.background": "#f6f8fa",
      "editorSuggestWidget.border": "#d1d5da",
      "editorSuggestWidget.foreground": "#24292e",
      "editorSuggestWidget.selectedBackground": "#e2e5e9",
      "editorHoverWidget.background": "#f6f8fa",
      "editorHoverWidget.border": "#d1d5da",
      "scrollbarSlider.background": "#959da533",
      "scrollbarSlider.hoverBackground": "#959da544",
      "scrollbarSlider.activeBackground": "#959da555",
      "editorOverviewRuler.border": "#ffffff",
    },
  })
  monaco.editor.defineTheme("github-dark", {
    base: "vs-dark",
    inherit: true,
    rules: githubDarkRules,
    colors: {
      "editor.background": "#24292e",
      "editor.foreground": "#e1e4e8",
      "editorCursor.foreground": "#c8e1ff",
      "editor.lineHighlightBackground": "#2b3036",
      "editor.selectionBackground": "#3392ff44",
      "editor.inactiveSelectionBackground": "#3392ff22",
      "editorIndentGuide.background1": "#2f363d",
      "editorIndentGuide.activeBackground1": "#444d56",
      "editorWidget.background": "#1f2428",
      "editorWidget.border": "#444d56",
      "editorSuggestWidget.background": "#1f2428",
      "editorSuggestWidget.border": "#444d56",
      "editorSuggestWidget.foreground": "#e1e4e8",
      "editorSuggestWidget.selectedBackground": "#39414a",
      "editorHoverWidget.background": "#1f2428",
      "editorHoverWidget.border": "#444d56",
      "scrollbarSlider.background": "#6a737d33",
      "scrollbarSlider.hoverBackground": "#6a737d44",
      "scrollbarSlider.activeBackground": "#6a737d55",
      "editorOverviewRuler.border": "#24292e",
    },
  })
}

function schemaSettings() {
  const settings = KUBERNETES_YAML_SCHEMAS.map((item) => ({
    // 文件名带 kind，编辑器通过 model 路径匹配到对应 schema；同一个 kind
    // 在不同编辑器中共享 schema，但 model 路径带唯一前缀，互不共用文本。
    uri: `inmemory://schema/alioth/${item.kind}.json`,
    fileMatch: [`**/${item.kind}.yaml`],
    schema: item.schema,
  }))
  settings.push({
    uri: "inmemory://schema/alioth/Unknown.json",
    fileMatch: ["**/Unknown.yaml"],
    schema: GENERIC_KUBERNETES_SCHEMA,
  })
  return settings
}

/**
 * monaco-editor 0.55 把 createWebWorker 改成接收现成的 Worker，
 * 而 monaco-worker-manager（monaco-yaml 的 worker 层）仍按旧签名传
 * label/moduleId，结果会回退到同步 worker 并报 Missing requestHandler。
 * 这里把公开入口转接到 monaco 自己的新实现，保留 worker 工厂与 createData。
 */
function patchCreateWebWorker(monaco: Monaco) {
  if (workerApiPatched) {
    return
  }
  workerApiPatched = true
  // 新版 createWebWorker 处理完 worker 后会回调 monaco 自己的 editor.createWebWorker，
  // 因此带 worker 的调用必须交给原函数，否则会无限递归。
  const createWebWorkerOriginal = monaco.editor.createWebWorker
  try {
    monaco.editor.createWebWorker = ((options: { worker?: unknown } & Parameters<typeof createWebWorkerInternal>[0]) =>
      options && "worker" in options
        ? createWebWorkerOriginal(options as never)
        : createWebWorkerInternal(options)) as unknown as typeof monaco.editor.createWebWorker
  } catch {
    // 只读环境（例如被冻结的 namespace）下保持原行为，语言服务会退化为同步 worker。
  }
}

function setupYaml(monaco: Monaco) {
  if (yamlReady) {
    return
  }
  yamlReady = true
  patchCreateWebWorker(monaco)
  yamlController = configureMonacoYaml(monaco, {
    enableSchemaRequest: false,
    // 校验 + 补全 + 悬浮说明都走内置 schema，不发任何网络请求。
    validate: true,
    completion: true,
    hover: true,
    isKubernetes: true,
    schemas: schemaSettings(),
  })
}

/** CodeEditor 的 beforeMount 钩子：GitHub 主题 + YAML 语言服务只注册一次。 */
export function setupMonaco(monaco: Monaco) {
  setupGithubThemes(monaco)
  setupYaml(monaco)
}

/** 供测试或未来的 schema 热更新使用；正常页面不需要调用。 */
export function getYamlController(): MonacoYaml | null {
  return yamlController
}
