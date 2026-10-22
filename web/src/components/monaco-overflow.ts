import { useEffect, useState } from "react"

/**
 * Monaco 的补全/提示浮层默认渲染在编辑器容器里。编辑器位于带
 * transform 的 Dialog 中时，`position: fixed` 的浮层会以 Dialog 为
 * 包含块，导致提示跑到输入框底部而不是光标处。
 *
 * 这里创建一个挂到 document.body 的浮层容器，配合
 * `fixedOverflowWidgets` 让 Monaco 按视口坐标定位提示框。
 * 容器必须带 `monaco-editor` 类：补全列表的宽度、层级和主题变量
 * 都定义在 `.monaco-editor .suggest-widget` 这类选择器下。
 */
export function useMonacoOverflowNode(): HTMLElement | null {
  const [node] = useState<HTMLElement | null>(() => {
    if (typeof document === "undefined") {
      return null
    }
    const element = document.createElement("div")
    element.className = "monaco-editor monaco-overflow-widgets"
    element.style.position = "fixed"
    element.style.top = "0"
    element.style.left = "0"
    element.style.zIndex = "100"
    return element
  })

  useEffect(() => {
    if (!node) {
      return
    }
    document.body.appendChild(node)
    return () => {
      node.remove()
    }
  }, [node])

  return node
}
