import type { NamespaceItem } from "@/api/namespace/types"
import type { UserItem } from "@/api/user/types"

export type SelfInfo = {
  user: UserItem
  systemRole?: string
  systemPermissions?: string[]
}

export type SelfNamespaces = {
  items?: NamespaceItem[]
}

export type SelfPermissions = {
  permissions?: string[]
}
