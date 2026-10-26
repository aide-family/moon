import { useMemo } from "react"

import { useSession } from "@/auth/session"
import { type QueryDialect, type QueryStoreScope } from "@/lib/query-store"

/** 查询缓存的作用域：用户 + 租户 + 数据源 + 方言，四者任一变化就是另一份缓存。 */
export function useQueryScope(
  datasourceUid: string,
  dialect: QueryDialect
): QueryStoreScope {
  const session = useSession()
  const userUid = session.user?.uid ?? ""
  const namespaceUid = session.namespaceUid
  return useMemo(
    () => ({ userUid, namespaceUid, datasourceUid, dialect }),
    [userUid, namespaceUid, datasourceUid, dialect]
  )
}
