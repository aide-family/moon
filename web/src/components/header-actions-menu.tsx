import { EllipsisVertical } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export interface HeaderAction {
  key: string
  label: string
  onSelect: () => void
  destructive?: boolean
}

interface HeaderActionsMenuProps {
  label: string
  actions: HeaderAction[]
}

/** 页头「更多」：收纳次级/破坏性操作，避免页头平铺 4–5 个按钮。 */
export function HeaderActionsMenu({ label, actions }: HeaderActionsMenuProps) {
  const normal = actions.filter((action) => !action.destructive)
  const dangerous = actions.filter((action) => action.destructive)
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">
          <EllipsisVertical aria-hidden="true" />
          {label}
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
  )
}
