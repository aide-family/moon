import { ExternalLink, RefreshCw, TriangleAlert } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { cn } from "@/lib/class-names"

/**
 * 面板加载排查浮层。跨域 iframe 的具体错误浏览器不会暴露给父页面，
 * 因此这里只提供固定的排查步骤；连接失败由调用方的 no-cors 探测标记。
 */
interface BoardLoadHintProps {
  /** 连通性探测失败时高亮警告图标并在浮层顶部给出结论。 */
  failed?: boolean
  /** 面板原始地址，用于「打开原图」新标签排查。 */
  openOriginalURL?: string
  /** 重新探测并重建 iframe。 */
  onReload?: () => void
  /** 页面级入口显示文字；面板右上角只显示图标。 */
  showLabel?: boolean
  align?: "start" | "center" | "end"
  className?: string
}

export function BoardLoadHint({
  failed = false,
  openOriginalURL,
  onReload,
  showLabel = false,
  align = "end",
  className,
}: BoardLoadHintProps) {
  const { t } = useTranslation()
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size={showLabel ? "sm" : "icon-sm"}
          className={cn(
            "shrink-0",
            failed
              ? "text-destructive hover:text-destructive"
              : "text-muted-foreground",
            className
          )}
          aria-label={t("board.view.loadHintButton")}
          title={t("board.view.loadHintButton")}
        >
          <TriangleAlert aria-hidden="true" />
          {showLabel ? <span>{t("board.view.loadHintButton")}</span> : null}
        </Button>
      </PopoverTrigger>
      <PopoverContent
        align={align}
        className="w-[28rem] max-w-[calc(100vw-2rem)] p-3"
      >
        <p className="text-sm font-medium">{t("board.view.loadHintTitle")}</p>
        <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
          {t("board.view.loadHintIntro")}
        </p>
        {failed ? (
          <p
            role="status"
            className="mt-2 rounded-md border border-warning-border bg-warning-soft px-2 py-1.5 text-xs leading-relaxed text-warning"
          >
            {t("board.view.loadHintProbeFailed")}
          </p>
        ) : null}
        <ul className="mt-2 grid list-disc gap-1.5 pl-4 text-xs leading-relaxed text-muted-foreground">
          <li>{t("board.view.loadHintIntranet")}</li>
          <li>{t("board.view.loadHintCertificate")}</li>
          <li>{t("board.view.loadHintEmbedding")}</li>
          <li>{t("board.view.loadHintMixedContent")}</li>
          <li>{t("board.view.loadHintLogin")}</li>
        </ul>
        <p className="mt-2 text-xs leading-relaxed text-muted-foreground">
          {t("board.view.loadHintCrossOrigin")}
        </p>
        {onReload || openOriginalURL ? (
          <div className="mt-3 flex flex-wrap justify-end gap-2">
            {onReload ? (
              <Button type="button" variant="outline" size="sm" onClick={onReload}>
                <RefreshCw aria-hidden="true" />
                {t("board.view.loadHintRetry")}
              </Button>
            ) : null}
            {openOriginalURL ? (
              <Button variant="outline" size="sm" asChild>
                <a
                  href={openOriginalURL}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <ExternalLink aria-hidden="true" />
                  {t("board.view.openOriginal")}
                </a>
              </Button>
            ) : null}
          </div>
        ) : null}
      </PopoverContent>
    </Popover>
  )
}
