import { Loader2 } from "lucide-react"
import { useTranslation } from "react-i18next"

import { FormField } from "@/components/form-field"
import { Input } from "@/components/ui/input"
import type { CaptchaController } from "@/hooks/use-captcha"
import { cn } from "@/lib/class-names"

interface CaptchaFieldProps {
  value: string
  onChange: (value: string) => void
  captcha: CaptchaController
  /** 服务端或本地校验附加到 captchaCode 的错误；加载失败由 hook 自己提供。 */
  error?: string
}

/**
 * 图形验证码：输入框在左、验证码图片在右，同一行等高。
 * 题型可能是 6 位数字码或四则运算，提示文案随题型切换；点击图片刷新，
 * 图片未就绪时用同高同宽的占位，避免布局跳动。
 */
export function CaptchaField({
  value,
  onChange,
  captcha,
  error,
}: CaptchaFieldProps) {
  const { t } = useTranslation()
  const message = error || captcha.error
  const numeric = captcha.kind === "NUMERIC"

  return (
    <FormField
      label={t("form.captcha")}
      htmlFor="captcha"
      hint={numeric ? t("form.captchaNumericHint") : t("form.captchaHint")}
      required
      error={message}
    >
      <div className="flex items-center gap-2">
        <Input
          id="captcha"
          className="min-w-0 flex-1"
          value={value}
          placeholder={
            numeric
              ? t("form.captchaNumericPlaceholder")
              : t("form.captchaPlaceholder")
          }
          inputMode="numeric"
          autoComplete="off"
          aria-invalid={message ? true : undefined}
          onChange={(event) => onChange(event.target.value)}
          required
        />
        <button
          type="button"
          className="flex shrink-0 rounded-sm outline-none focus-visible:ring-3 focus-visible:ring-ring"
          aria-label={t("form.captchaRefresh")}
          aria-busy={captcha.loading || undefined}
          disabled={captcha.loading}
          onClick={() => void captcha.refresh()}
        >
          {captcha.image ? (
            <span className="relative block">
              <img
                src={captcha.image}
                alt={t("form.captcha")}
                className="h-8 w-auto rounded border"
              />
              {captcha.loading ? (
                <span
                  aria-hidden="true"
                  className="absolute inset-0 flex items-center justify-center rounded bg-background/60"
                >
                  <Loader2 className="size-4 animate-spin text-muted-foreground" />
                </span>
              ) : null}
            </span>
          ) : (
            // 图片尺寸是 280x72，按 h-8 渲染约 124px 宽；占位保持同高同宽避免布局跳动。
            <span
              aria-hidden="true"
              className={cn(
                "flex h-8 w-[124px] items-center justify-center rounded border border-dashed border-border bg-muted/40",
                captcha.loading && "animate-pulse"
              )}
            >
              {captcha.loading ? (
                <Loader2 className="size-4 animate-spin text-muted-foreground" />
              ) : null}
            </span>
          )}
        </button>
      </div>
    </FormField>
  )
}
