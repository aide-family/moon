import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { getMyAlertColumns, saveMyAlertColumns } from "@/api/alert-page"
import type { AlertColumnItem } from "@/api/alert-page/types"
import { useSession } from "@/auth/session"
import { formErrorMessage } from "@/lib/form-error"

/**
 * 读取并保存某个告警视图的自定义列。
 *
 * 列配置按「命名空间 + 用户 + scope」存在后端，scope 是实时告警的 mine、
 * 告警页面 UID，或历史告警明细的 history。两个告警页共用这套读写逻辑，
 * 页面只负责渲染表头和列设置弹窗。
 */
export function useAlertColumns(
  scope: string,
  defaults: AlertColumnItem[],
  reloadKey = 0
) {
  const { t } = useTranslation()
  const session = useSession()
  const [columns, setColumns] = useState<AlertColumnItem[]>(defaults)
  const [open, setOpen] = useState(false)
  const [pending, setPending] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => {
    let alive = true
    async function load() {
      try {
        const result = await getMyAlertColumns(scope)
        if (!alive) {
          return
        }
        setColumns(result.columns?.length ? result.columns : defaults)
      } catch {
        // 读取失败按未配置处理，回退默认列，不阻塞告警列表本身。
        if (alive) {
          setColumns(defaults)
        }
      }
    }
    void load()
    return () => {
      alive = false
    }
  }, [scope, reloadKey, session.namespaceUid, defaults])

  async function save(next: AlertColumnItem[]) {
    setPending(true)
    setError("")
    try {
      const result = await saveMyAlertColumns(scope, next)
      setColumns(result.columns?.length ? result.columns : defaults)
      setOpen(false)
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  return { columns, open, setOpen, pending, error, setError, save }
}
