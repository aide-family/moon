import { ArrowUpRight, Loader2 } from "lucide-react"
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
 * 登录页的主操作按钮：48px 高、整行宽度，左侧品牌图标块、居中文案、右侧外链箭头。
 *
 * 图标块承接平台品牌色（飞书/钉钉素材自带配色，GitHub/Gitee 走 currentColor），
 * 文案始终使用前景色；这样浅色与深色模式的文字对比度都由 token 保证。
 * 品牌色只做描边、浅底和图标点缀，hover / focus / disabled 由 index.css 统一处理。
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
      data-pending={pending || undefined}
      className="oauth-provider oauth-brand group/oauth relative h-12 w-full justify-center rounded-xl px-3 text-sm font-medium shadow-xs focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:opacity-100"
      disabled={disabled}
      aria-busy={pending || undefined}
      onClick={onClick}
    >
      <span aria-hidden="true" className="oauth-provider-badge absolute left-3">
        {pending ? (
          <Loader2 className="size-[1.125rem] animate-spin" />
        ) : (
          createElement(oauthProviderIcon(provider.provider), {
            className: "size-[1.125rem]",
          })
        )}
      </span>
      <span className="max-w-full min-w-0 truncate px-8">{label}</span>
      <ArrowUpRight
        aria-hidden="true"
        className="absolute right-3 size-4 text-muted-foreground/60 transition-[color,transform] duration-200 group-hover/oauth:translate-x-0.5 group-hover/oauth:-translate-y-0.5 group-hover/oauth:text-foreground group-disabled/oauth:opacity-0 motion-reduce:transition-none"
      />
    </Button>
  )
}
