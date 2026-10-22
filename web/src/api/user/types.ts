export type UserItem = {
  uid: string
  name: string
  nickname: string
  email: string
  phone: string
  avatar: string
  remark: string
  status: string
  systemRole: string
  authorizedBy: string
}

export type ListUsersParams = {
  page?: number
  pageSize?: number
  name?: string
}

export type ListUsersResult = {
  items?: UserItem[]
  total?: number
  page?: number
  pageSize?: number
}

export type SelectUsersParams = {
  limit?: number
  lastUid?: string
}

export type SelectItem = {
  uid: string
  name: string
}

export type SelectUsersResult = {
  items?: SelectItem[]
  lastUid?: string
}
