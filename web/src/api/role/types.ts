export type RoleItem = {
  uid: string
  namespaceUid: string
  name: string
  status: string
  isSystem: boolean
  permissions?: string[]
}

export type PermissionItem = {
  uid: string
  name: string
  resource: string
  action: string
}

export type ListRolesParams = {
  page?: number
  pageSize?: number
  name?: string
}

export type ListRolesResult = {
  items?: RoleItem[]
  total?: number
  page?: number
  pageSize?: number
}

export type CreateRoleParams = {
  name: string
}

export type UpdateRoleParams = {
  uid: string
  name: string
}

export type SelectRolesResult = {
  items?: { uid: string; name: string }[]
  lastUid?: string
}
