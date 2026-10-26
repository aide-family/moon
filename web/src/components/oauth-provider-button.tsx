import { KeyRound, Loader2 } from "lucide-react"
import type { ComponentType } from "react"

import type { OAuth2Provider, OAuth2ProviderItem } from "@/api/auth/types"
import { DingtalkIcon } from "@/components/dingtalk-icon"
import { FeishuIcon } from "@/components/feishu-icon"
import { GiteeIcon } from "@/components/gitee-icon"
import { GitHubIcon } from "@/components/github-icon"
import { Button } from "@/components/ui/button"

type ProviderIcon = ComponentType<{ className?: string }>

// 平台图标只负责图形；品牌色由 index.css 的 .oauth-provider[data-provider] 提供。
// 后端新增平台时这里退化成通用图标，不会渲染出空位。
const providerIcons: Partial<Record<OAuth2Provider, ProviderIcon>> = {
  GITHUB: GitHubIcon,
  GITEE: GiteeIcon,
  FEISHU: FeishuIcon,
  DINGTALK: DingtalkIcon,
}

interface OAuthProviderButtonProps {
  provider: OAuth2ProviderItem
  /** 按钮文案，已按当前语言拼好（例如「使用 GitHub 登录」）。 */
  label: string
  /** 本轮正在跳转的平台：图标块换成加载动画。 */
  pending: boolean
  disabled: boolean
  onClick: () => void
}

/**
 * 登录页的主操作按钮：整行宽度、44px 高，描边与底色带平台品牌色，
 * 图标用品牌色（飞书/钉钉的素材自带配色，GitHub/Gitee 走 currentColor）。
 * 文字保持前景色——品牌色只做点缀，浅色与深色模式下文字对比度都由 token 保证。
 */
export function OAuthProviderButton({
  provider,
  label,
  pending,
  disabled,
  onClick,
}: OAuthProviderButtonProps) {
  const Icon = providerIcons[provider.provider] ?? KeyRound

  return (
    <Button
      type="button"
      variant="outline"
      data-provider={provider.provider}
      className="oauth-provider h-11 w-full justify-center gap-2.5 px-3 text-sm font-medium"
      disabled={disabled}
      aria-busy={pending || undefined}
      onClick={onClick}
    >
      <span aria-hidden="true" className="oauth-provider-icon">
        {pending ? (
          <Loader2 className="size-5 animate-spin" />
        ) : (
          <Icon className="size-5" />
        )}
      </span>
      <span className="truncate">{label}</span>
    </Button>
  )
}
