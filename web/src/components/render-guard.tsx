import { Component, type ReactNode } from "react"

import { Button } from "@/components/ui/button"

/**
 * 局部渲染护栏：预览这类「把另一套版式嵌进当前页面」的场景，一旦子树抛错，
 * 不能让整个后台白屏。兜底只影响被包裹的区域，并给出重试。
 */
interface RenderGuardProps {
  children: ReactNode
  /** 兜底文案，由调用方本地化。 */
  message: string
  retryLabel: string
}

interface RenderGuardState {
  failed: boolean
}

export class RenderGuard extends Component<RenderGuardProps, RenderGuardState> {
  state: RenderGuardState = { failed: false }

  static getDerivedStateFromError(): RenderGuardState {
    return { failed: true }
  }

  render() {
    if (this.state.failed) {
      return (
        <div
          role="alert"
          className="grid gap-3 rounded-lg border border-danger-border bg-danger-soft p-4"
        >
          <p className="text-sm text-danger">{this.props.message}</p>
          <div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => this.setState({ failed: false })}
            >
              {this.props.retryLabel}
            </Button>
          </div>
        </div>
      )
    }
    return this.props.children
  }
}
