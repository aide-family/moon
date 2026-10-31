import { del, get, post, put } from "@/api/http"

import type {
  CreateRoleParams,
  ListRolesParams,
  ListRolesResult,
  PermissionItem,
  PermissionMenu,
  RoleItem,
  SelectRolesResult,
  UpdateRoleParams,
  UpdateRoleRemarkParams,
} from "./types"

/** GET /v1/roles */
export function listRoles(params: ListRolesParams = {}): Promise<ListRolesResult> {
  return get<ListRolesResult>("/roles", params)
}

/** GET /v1/roles/select */
export function selectRoles(limit = 100): Promise<SelectRolesResult> {
  return get<SelectRolesResult>("/roles/select", { limit })
}

/** POST /v1/roles */
export function createRole(body: CreateRoleParams): Promise<{ role: RoleItem }> {
  return post<{ role: RoleItem }>("/roles", body)
}

/** PUT /v1/role/{uid} */
export function updateRole(body: UpdateRoleParams): Promise<{ role: RoleItem }> {
  return put<{ role: RoleItem }>(`/role/${body.uid}`, body)
}

/** PUT /v1/role/{uid}/status */
export function updateRoleStatus(uid: string, status: string): Promise<{ role: RoleItem }> {
  return put<{ role: RoleItem }>(`/role/${uid}/status`, { uid, status })
}

/** DELETE /v1/role/{uid} */
export function deleteRole(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/role/${uid}`)
}

/** PUT /v1/role/{uid}/permissions */
export function updateRolePermissions(uid: string, permissions: string[]): Promise<{ role: RoleItem }> {
  return put<{ role: RoleItem }>(`/role/${uid}/permissions`, { uid, permissions })
}

/** PUT /v1/role/{uid}/remark */
export function updateRoleRemark(body: UpdateRoleRemarkParams): Promise<{ role: RoleItem }> {
  return put<{ role: RoleItem }>(`/role/${body.uid}/remark`, body)
}

/** GET /v1/permissions */
export function listPermissions(): Promise<{
  items?: PermissionItem[]
  menus?: PermissionMenu[]
}> {
  return get<{ items?: PermissionItem[]; menus?: PermissionMenu[] }>(
    "/permissions"
  )
}
