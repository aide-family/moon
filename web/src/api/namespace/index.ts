import { del, get, post, put } from "@/api/http"

import type { CreateNamespaceParams, ListNamespacesParams, ListNamespacesResult, NamespaceItem, UpdateNamespaceParams } from "./types"

/** GET /v1/namespaces */
export function listNamespaces(params: ListNamespacesParams = {}): Promise<ListNamespacesResult> {
  return get<ListNamespacesResult>("/namespaces", params)
}

/** POST /v1/namespace */
export function createNamespace(body: CreateNamespaceParams): Promise<{ namespace: NamespaceItem }> {
  return post<{ namespace: NamespaceItem }>("/namespace", body)
}

/** PUT /v1/namespace/{uid} */
export function updateNamespace(body: UpdateNamespaceParams): Promise<{ namespace: NamespaceItem }> {
  return put<{ namespace: NamespaceItem }>(`/namespace/${body.uid}`, body)
}

/** PUT /v1/namespace/{uid}/status */
export function updateNamespaceStatus(uid: string, status: string): Promise<{ namespace: NamespaceItem }> {
  return put<{ namespace: NamespaceItem }>(`/namespace/${uid}/status`, { uid, status })
}

/** DELETE /v1/namespace/{uid} */
export function deleteNamespace(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/namespace/${uid}`)
}

/** PUT /v1/namespace/{uid}/owner */
export function transferNamespaceOwner(uid: string, userUid: string): Promise<{ namespace: NamespaceItem }> {
  return put<{ namespace: NamespaceItem }>(`/namespace/${uid}/owner`, { uid, userUid })
}
