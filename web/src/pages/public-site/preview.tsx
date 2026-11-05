import { ArrowLeft, Eye } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link, useParams } from "react-router-dom"

import { SiteShell } from "@/components/site-shell"
import { Button } from "@/components/ui/button"

import { SitePreviewView } from "./site-preview-view"

function PreviewBanner({ uid }: { uid: string }) {
  const { t } = useTranslation()
  return (
    <div role="note" className="border-b border-warning-border bg-warning-soft">
      <div className="mx-auto flex w-full max-w-6xl flex-wrap items-center gap-x-3 gap-y-2 px-4 py-2.5 sm:px-6">
        <span className="flex min-w-0 items-center gap-2 text-sm text-warning">
          <Eye aria-hidden="true" className="size-4 shrink-0" />
          <span className="font-medium break-words">
            {t("sitePages.preview.banner")}
          </span>
        </span>
        <Button asChild variant="outline" size="sm" className="ml-auto">
          <Link to={`/system/site/pages?uid=${encodeURIComponent(uid)}`}>
            <ArrowLeft aria-hidden="true" data-icon="inline-start" />
            {t("sitePages.preview.backToEditor")}
          </Link>
        </Button>
      </div>
    </div>
  )
}

/**
 * `/site-preview/:uid`：完整官网外壳 + 预览横幅 + 草稿内容。
 * 内容渲染与后台内联预览共用 `SitePreviewView`。
 */
export function SitePreviewPage() {
  const { uid = "" } = useParams()
  return (
    <SiteShell>
      <PreviewBanner uid={uid} />
      <SitePreviewView uid={uid} />
    </SiteShell>
  )
}
