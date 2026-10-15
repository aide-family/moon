import { get, put } from "@/api/http"

import type { ListUsersParams, ListUsersResult, SelectUsersParams, SelectUsersResult, UserItem } from "./types"

/** GET /v1/users */
export function listUsers(params: ListUsersParams = {}): Promise<ListUsersResult> {
  return get<ListUsersResult>("/users", params)
}

/** GET /v1/users/select */
export function selectUsers(params: SelectUsersParams = {}): Promise<SelectUsersResult> {
  return get<SelectUsersResult>("/users/select", params)
}

/** GET /v1/user/{uid} */
export function getUser(uid: string): Promise<{ user: UserItem }> {
  return get<{ user: UserItem }>(`/user/${uid}`)
}

/** PUT /v1/user/ban/{uid} */
export function banUser(uid: string): Promise<{ user: UserItem }> {
  return put<{ user: UserItem }>(`/user/ban/${uid}`)
}

/** PUT /v1/user/permit/{uid} */
export function permitUser(uid: string): Promise<{ user: UserItem }> {
  return put<{ user: UserItem }>(`/user/permit/${uid}`)
}
