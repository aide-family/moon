import { KeyRound } from "lucide-react"
import { createElement, type ComponentType } from "react"

import type { OAuth2Provider } from "@/api/auth/types"
import { DingtalkIcon } from "@/components/dingtalk-icon"
import { FeishuIcon } from "@/components/feishu-icon"
import { GiteeIcon } from "@/components/gitee-icon"
import { GitHubIcon } from "@/components/github-icon"
import { cn } from "@/lib/class-names"

type ProviderIcon = ComponentType<{ className?: string }>

// 平台图标只负责图形；品牌色由 index.css 的 .oauth-brand[data-provider] 提供。
// 后端新增平台时这里退化成通用图标，不会渲染出空位。
const providerIcons: Partial<Record<OAuth2Provider, ProviderIcon>> = {
  GITHUB: GitHubIcon,
  GITEE: GiteeIcon,
  FEISHU: FeishuIcon,
  DINGTALK: DingtalkIcon,
}

/** 平台图标组件；未知来源回落到通用图标。 */
export function oauthProviderIcon(provider: OAuth2Provider): ProviderIcon {
  return providerIcons[provider] ?? KeyRound
}

interface OAuthProviderMarkProps {
  provider: OAuth2Provider
  className?: string
}

/**
 * 36px 的平台图标块。`oauth-brand` 只提供品牌色 token，
 * 图形本身沿用登录页第三方按钮的图标素材，未知来源显示 KeyRound。
 */
export function OAuthProviderMark({
  provider,
  className,
}: OAuthProviderMarkProps) {
  return (
    <span
      data-provider={provider}
      aria-hidden="true"
      className={cn(
        "oauth-brand grid size-9 shrink-0 place-items-center rounded-lg border border-border/60 bg-muted/40",
        className
      )}
    >
      <span className="oauth-provider-icon">
        {createElement(oauthProviderIcon(provider), { className: "size-5" })}
      </span>
    </span>
  )
}
