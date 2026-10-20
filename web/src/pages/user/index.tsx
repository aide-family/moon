import { useState } from "react"
import { useTranslation } from "react-i18next"
import { formMessage } from "@/lib/form-error"

import { banUser, listUsers, permitUser } from "@/api/user"
import type { UserItem } from "@/api/user/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"
import { ListFrame } from "@/components/list-frame"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

export function UserPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [detail, setDetail] = useState<UserItem | null>(null)
  const [banning, setBanning] = useState<UserItem | null>(null)
  const [pending, setPending] = useState(false)
  const [confirmError, setConfirmError] = useState("")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<UserItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${query.name}|${session.namespaceUid}`,
    fetch: (page, pageSize) => listUsers({ page, pageSize, name: query.name }),
  })

  async function onPermit(item: UserItem) {
    await permitUser(item.uid)
    reload()
  }

  async function onBan() {
    if (!banning) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      await banUser(banning.uid)
      setBanning(null)
      reload()
    } catch (err) {
      setConfirmError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function userActions(item: UserItem): RowAction[] {
    if (!session.can("user:write") || item.uid === session.user?.uid) {
      return []
    }
    if (item.status === "ACTIVE") {
      return [
        {
          key: "ban",
          label: t("action.ban"),
          destructive: true,
          onSelect: () => {
            setConfirmError("")
            setBanning(item)
          },
        },
      ]
    }
    return [
      {
        key: "permit",
        label: t("action.permit"),
        onSelect: () => void onPermit(item),
      },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("user.title")} description={t("user.description")} />
      <form
        className="flex gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input
          className="max-w-sm"
          value={draft}
          placeholder={t("user.name")}
          onChange={(event) => setDraft(event.target.value)}
        />
        <Button type="submit" variant="outline">
          {t("action.search")}
        </Button>
      </form>
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
          {loading ? (
            <p className="text-sm text-muted-foreground">
              {t("state.loading")}
            </p>
          ) : null}
          {!loading && failed ? (
            <div className="flex items-center gap-2">
              <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
              <Button
                variant="outline"
                onClick={() => reload()}
              >
                {t("action.retry")}
              </Button>
            </div>
          ) : null}
          {!loading && !failed && items.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t("user.empty")}</p>
          ) : null}
          {!loading && !failed && items.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("user.name")}</TableHead>
                  <TableHead>{t("login.emailLabel")}</TableHead>
                  <TableHead>{t("namespace.status")}</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.uid}>
                    <TableCell>{item.name}</TableCell>
                    <TableCell>{item.email}</TableCell>
                    <TableCell>{enumLabel(t, item.status)}</TableCell>
                    <TableCell>
                      <RowActions
                        detailLabel={t("action.detail")}
                        moreLabel={t("action.more")}
                        onDetail={() => setDetail(item)}
                        actions={userActions(item)}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : null}
      </ListFrame>
      <DetailSheet
        open={detail !== null}
        title={detail?.name || detail?.email || t("action.detail")}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
          }
        }}
        fields={[
          { label: t("user.name"), value: detail?.name ?? "" },
          { label: t("user.nickname"), value: detail?.nickname ?? "" },
          { label: t("login.emailLabel"), value: detail?.email ?? "" },
          { label: t("namespace.status"), value: enumLabel(t, detail?.status ?? "") },
          { label: t("profile.remark"), value: detail?.remark ?? "" },
          { label: t("detail.uid"), value: detail?.uid ?? "" },
        ]}
      />
      <ConfirmAction
        open={banning !== null}
        title={t("confirm.banTitle")}
        description={t("confirm.banDescription")}
        confirmLabel={t("action.confirm")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void onBan()}
        onOpenChange={(value) => {
          if (!value) {
            setBanning(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
