import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { banUser, listUsers, permitUser } from "@/api/user"
import type { UserItem } from "@/api/user/types"
import { useSession } from "@/auth/session"
import { useListQuery } from "@/lib/list-query"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

export function UserPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [items, setItems] = useState<UserItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const result = await listUsers({ page: query.page, pageSize: query.pageSize, name: query.name })
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
      } catch {
        if (active) {
          setFailed(true)
          setItems([])
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [query.page, query.pageSize, query.name, attempt])

  async function onStatus(item: UserItem) {
    if (item.status === "ACTIVE") {
      await banUser(item.uid)
    } else {
      await permitUser(item.uid)
    }
    setAttempt((value) => value + 1)
  }

  return (
    <section className="space-y-4">
      <h1 className="text-xl font-semibold">{t("user.title")}</h1>
      <form className="flex gap-2" onSubmit={(event) => { event.preventDefault(); query.update({ name: draft }) }}>
        <Input value={draft} placeholder={t("user.name")} onChange={(event) => setDraft(event.target.value)} />
        <Button type="submit" variant="outline">{t("action.search")}</Button>
      </form>
      {loading ? <p>{t("state.loading")}</p> : null}
      {failed ? <p>{t("state.failed")}</p> : null}
      {!loading && !failed && items.length === 0 ? <p>{t("state.empty")}</p> : null}
      {items.length > 0 ? (
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
                <TableCell>{item.status}</TableCell>
                <TableCell className="text-right">
                  {session.can("user:write") && item.uid !== session.user?.uid ? (
                    <Button variant="outline" size="sm" onClick={() => void onStatus(item)}>
                      {item.status === "ACTIVE" ? t("action.ban") : t("action.permit")}
                    </Button>
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      ) : null}
      <div className="flex items-center gap-2">
        <Button variant="outline" disabled={query.page <= 1} onClick={() => query.update({ page: query.page - 1 })}>{t("action.prev")}</Button>
        <span className="text-sm">{query.page} / {Math.max(1, Math.ceil(total / query.pageSize))}</span>
        <Button variant="outline" disabled={query.page * query.pageSize >= total} onClick={() => query.update({ page: query.page + 1 })}>{t("action.next")}</Button>
      </div>
    </section>
  )
}
