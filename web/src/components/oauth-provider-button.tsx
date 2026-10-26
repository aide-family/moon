import { Loader2 } from "lucide-react"
import { createElement } from "react"

import type { OAuth2ProviderItem } from "@/api/auth/types"
import { oauthProviderIcon } from "@/components/oauth-provider-mark"
import { Button } from "@/components/ui/button"

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
  return (
    <Button
      type="button"
      variant="outline"
      data-provider={provider.provider}
      className="oauth-provider oauth-brand h-11 w-full justify-center gap-2.5 px-3 text-sm font-medium"
      disabled={disabled}
      aria-busy={pending || undefined}
      onClick={onClick}
    >
      <span aria-hidden="true" className="oauth-provider-icon">
        {pending ? (
          <Loader2 className="size-5 animate-spin" />
        ) : (
          createElement(oauthProviderIcon(provider.provider), {
            className: "size-5",
          })
        )}
      </span>
      <span className="truncate">{label}</span>
    </Button>
  )
}
