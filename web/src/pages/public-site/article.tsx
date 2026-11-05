import type { ReactNode } from "react"

import { SiteBody } from "@/components/site-body"

interface SiteArticleProps {
  title: string
  /** 标题下方的时间、语言等元信息，文案与顺序由调用方决定。 */
  meta?: ReactNode
  coverImageUrl?: string
  bodyMd: string
  emptyText: string
  /** 元信息与正文之间的附加提示（例如语言回退提示）。 */
  children?: ReactNode
}

/**
 * 官网文档页的固定版式：标题、元信息、封面、附加提示与 Markdown 正文。
 *
 * `/p/:slug`（已发布快照）与 `/site-preview/:uid`（后台草稿）共用这一份版式，
 * 两边的区别只在数据来源与分析行文案。
 */
export function SiteArticle({
  title,
  meta,
  coverImageUrl,
  bodyMd,
  emptyText,
  children,
}: SiteArticleProps) {
  return (
    <>
      <h1 className="font-heading text-3xl font-semibold tracking-tight break-words sm:text-4xl">
        {title}
      </h1>
      {meta}
      {coverImageUrl ? (
        <img
          src={coverImageUrl}
          alt=""
          className="mt-6 w-full rounded-2xl border border-border/60 object-cover"
        />
      ) : null}
      {children}
      <SiteBody content={bodyMd} emptyText={emptyText} className="mt-8" />
    </>
  )
}
