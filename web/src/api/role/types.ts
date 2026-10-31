export type RoleItem = {
  uid: string
  namespaceUid: string
  name: string
  status: string
  isSystem: boolean
  // 备注：说明这个角色能做什么、该分配给谁。系统角色也有备注。
  remark?: string
  permissions?: string[]
}

export type PermissionItem = {
  uid: string
  name: string
  resource: string
  action: string
}

/** 授权菜单：一级为菜单权限，子权限挂在菜单下。 */
export type PermissionMenu = {
  code: string
  group: string
  name: string
  always: boolean
  capabilities: string[]
  defaultReads: string[]
  permissions: PermissionItem[]
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
  remark?: string
}

export type UpdateRoleParams = {
  uid: string
  name: string
}

export type UpdateRoleRemarkParams = {
  uid: string
  remark: string
}

export type SelectRolesResult = {
  items?: { uid: string; name: string }[]
  lastUid?: string
}
