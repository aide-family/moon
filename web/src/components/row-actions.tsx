import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
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
}

export function RowActions({
  detailLabel,
  moreLabel,
  onDetail,
  actions,
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

  if (inline.length > 2) {
    return (
      <div className="flex flex-wrap justify-end gap-2">
        <Button variant="outline" size="sm" onClick={onDetail}>
          {detailLabel}
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="sm">
              {moreLabel}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {actions.map((action) => (
              <DropdownMenuItem
                key={action.key}
                variant={action.destructive ? "destructive" : "default"}
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
    <div className="flex flex-wrap justify-end gap-2">
      {inline.map((action) => (
        <Button
          key={action.key}
          variant={action.destructive ? "destructive" : "outline"}
          size="sm"
          onClick={action.onSelect}
        >
          {action.label}
        </Button>
      ))}
    </div>
  )
}
