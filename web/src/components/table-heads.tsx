import type { ReactNode } from "react"

import { TableCell, TableHead } from "@/components/ui/table"
import { cn } from "@/lib/class-names"
import { TABLE_COLUMN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

interface TableColumnHeadProps {
  children: ReactNode
  className?: string
}

/** 操作列表头：固定 7rem，容纳两个内联操作。 */
export function TableActionsHead({
  children,
  className,
}: TableColumnHeadProps) {
  return (
    <TableHead
      className={cn(
        TABLE_COLUMN_WIDTH.actions,
        TABLE_NOWRAP,
        "text-center",
        className
      )}
    >
      {children}
    </TableHead>
  )
}

/** 操作列单元格：与 TableActionsHead 同宽、同轴居中。 */
export function TableActionsCell({
  children,
  className,
}: {
  children: ReactNode
  className?: string
}) {
  return (
    <TableCell
      className={cn(TABLE_COLUMN_WIDTH.actions, "text-center", className)}
    >
      {children}
    </TableCell>
  )
}

/** 状态列：固定宽度 + 不换行，各页面宽度一致。 */
export function TableStatusHead({ children, className }: TableColumnHeadProps) {
  return (
    <TableHead
      className={cn(TABLE_COLUMN_WIDTH.status, TABLE_NOWRAP, className)}
    >
      {children}
    </TableHead>
  )
}

/** 类型/渠道/模板类型列：固定宽度 + 不换行。 */
export function TableTypeHead({ children, className }: TableColumnHeadProps) {
  return (
    <TableHead className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP, className)}>
      {children}
    </TableHead>
  )
}
