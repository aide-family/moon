import { Star } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { useRepositoryStars } from "@/hooks/use-repository-stars"
import { formatStarCount, repositorySlug } from "@/lib/github-stars"

/**
 * 右上角的 GitHub Star 入口：Star 图标 + 数量摘要，点击打开仓库的 star 页面。
 *
 * - 数量来自 GitHub 公共 API（带缓存，见 lib/github-stars），拿不到时只显示
 *   “Star” 文案，不占位、不弹错，也不阻塞登录表单。
 * - 两种状态下文本都是四个字符左右，数字回来时按钮宽度几乎不变，
 *   右对齐的行里其他入口不会被推着走。
 */
export function RepositoryStarLink({
  repositoryURL,
}: {
  repositoryURL: string
}) {
  const { t, i18n } = useTranslation()
  const count = useRepositoryStars(repositorySlug(repositoryURL))
  const summary = count === null ? null : formatStarCount(count)
  // 读屏与悬浮提示带上准确数字；可见摘要用 k 缩写，两者不冲突。
  const label =
    count === null
      ? t("shell.githubStar")
      : t("shell.githubStarWithCount", {
          count: new Intl.NumberFormat(i18n.language).format(count),
        })

  return (
    <Button
      variant="outline"
      className="shrink-0 gap-1.5 rounded-full px-3 text-xs"
      asChild
    >
      <a
        href={`${repositoryURL.replace(/\/+$/, "")}/stargazers`}
        target="_blank"
        rel="noreferrer"
        aria-label={label}
        title={label}
        data-slot="repository-star-link"
      >
        <Star
          aria-hidden="true"
          className="size-3.5 transition-transform duration-200 group-hover/button:-rotate-12 motion-reduce:transition-none"
        />
        <span className="tabular-nums">{summary ?? t("shell.star")}</span>
      </a>
    </Button>
  )
}
