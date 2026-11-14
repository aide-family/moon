import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export interface RowAction {
  key: string
  label: string
  onSelect: () => void
  destructive?: boolean
}

interface RowActionsProps {
  detailLabel: string
  moreLabel: string
  onDetail: () => void
  actions: RowAction[]
  // 兼容旧调用：现在只要存在额外操作就统一收进“更多”菜单。
  forceMenu?: boolean
  // 为真时按钮不换行，用于列宽固定、需要保持一行两个按钮的表格。
  nowrap?: boolean
  // 为真时禁用「详情」入口（例如已停用租户不允许只读查看），但仍展示按钮。
  detailDisabled?: boolean
}

export function RowActions({
  detailLabel,
  moreLabel,
  onDetail,
  actions,
  nowrap = false,
  detailDisabled = false,
}: RowActionsProps) {
  // 统一约定：存在额外操作时固定「详情 + 更多」两个入口，其余操作全部进菜单；
  // 没有额外操作时只保留详情按钮，避免不同列表的操作区形态不一致。
  const normal = actions.filter((action) => !action.destructive)
  const dangerous = actions.filter((action) => action.destructive)
  const rowClass = nowrap
    ? "flex flex-nowrap justify-center gap-2"
    : "flex flex-wrap justify-center gap-2"

  if (actions.length === 0) {
    return (
      <div className={rowClass}>
        <Button
          variant="link"
          size="sm"
          className="h-auto px-0"
          disabled={detailDisabled}
          onClick={onDetail}
        >
          {detailLabel}
        </Button>
      </div>
    )
  }

  return (
    <div className={rowClass}>
      <Button
        variant="link"
        size="sm"
        className="h-auto px-0"
        disabled={detailDisabled}
        onClick={onDetail}
      >
        {detailLabel}
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="link" size="sm" className="h-auto px-0">
            {moreLabel}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {normal.map((action) => (
            <DropdownMenuItem key={action.key} onSelect={action.onSelect}>
              {action.label}
            </DropdownMenuItem>
          ))}
          {normal.length > 0 && dangerous.length > 0 ? (
            <DropdownMenuSeparator />
          ) : null}
          {dangerous.map((action) => (
            <DropdownMenuItem
              key={action.key}
              variant="destructive"
              onSelect={action.onSelect}
            >
              {action.label}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
