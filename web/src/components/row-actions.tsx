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
  // 为真时所有操作都收进“更多”菜单，适合把只读操作和写操作放在一起的列表。
  forceMenu?: boolean
  // 为真时按钮不换行，用于列宽固定、需要保持一行两个按钮的表格。
  nowrap?: boolean
}

export function RowActions({
  detailLabel,
  moreLabel,
  onDetail,
  actions,
  forceMenu = false,
  nowrap = false,
}: RowActionsProps) {
  const inline = [
    {
      key: "detail",
      label: detailLabel,
      onSelect: onDetail,
      destructive: false,
    },
    ...actions,
  ]
  // 危险操作统一排在普通操作之后，并用分割线隔开，避免和编辑、详情混在一起。
  const normal = actions.filter((action) => !action.destructive)
  const dangerous = actions.filter((action) => action.destructive)
  const rowClass = nowrap
    ? "flex flex-nowrap justify-center gap-2"
    : "flex flex-wrap justify-center gap-2"

  if (inline.length > 2 || (forceMenu && actions.length > 0)) {
    return (
      <div className={rowClass}>
        <Button variant="link" size="sm" className="h-auto px-0" onClick={onDetail}>
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

  return (
    <div className={rowClass}>
      {[
        ...inline.filter((action) => !action.destructive),
        ...inline.filter((action) => action.destructive),
      ].map((action) => (
        <Button
          key={action.key}
          variant={action.destructive ? "destructive" : "link"}
          size="sm"
          className={action.destructive ? undefined : "h-auto px-0"}
          onClick={action.onSelect}
        >
          {action.label}
        </Button>
      ))}
    </div>
  )
}
