import { del, get, post, put } from "@/api/http"

import type {
  CopyRegistryTagParams,
  CreateImageSyncsResult,
  DeleteRegistryTagParams,
  GetRegistryTagParams,
  ImageSyncInput,
  ImageSyncItem,
  ListImageSyncsParams,
  ListImageSyncsResult,
  ListRegistriesParams,
  ListRegistriesResult,
  ListRegistryRepositoriesParams,
  ListRegistryRepositoriesResult,
  ListRegistryTagsParams,
  ListRegistryTagsResult,
  PreviewImageSyncParams,
  PreviewImageSyncResult,
  ProbeRegistryParams,
  RegistryItem,
  RegistryProbe,
  RegistryTagDetail,
  RegistryTagItem,
  SaveRegistryParams,
  UpdateRegistryParams,
} from "./types"

function segment(value: string): string {
  return encodeURIComponent(value)
}

/** GET /v1/registries */
export function listRegistries(
  params: ListRegistriesParams = {}
): Promise<ListRegistriesResult> {
  return get<ListRegistriesResult>("/registries", params)
}

/** GET /v1/registry/{uid} */
export function getRegistry(uid: string): Promise<{ registry: RegistryItem }> {
  return get<{ registry: RegistryItem }>(`/registry/${segment(uid)}`)
}

/** POST /v1/registry */
export function createRegistry(
  body: SaveRegistryParams
): Promise<{ registry: RegistryItem }> {
  return post<{ registry: RegistryItem }>("/registry", body)
}

/** PUT /v1/registry/{uid} */
export function updateRegistry(
  body: UpdateRegistryParams
): Promise<{ registry: RegistryItem }> {
  return put<{ registry: RegistryItem }>(`/registry/${segment(body.uid)}`, body)
}

/** DELETE /v1/registry/{uid} */
export function deleteRegistry(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/registry/${segment(uid)}`)
}

/** POST /v1/registry/{uid}/test */
export function testRegistry(
  uid: string
): Promise<{ probe: RegistryProbe; registry: RegistryItem }> {
  return post<{ probe: RegistryProbe; registry: RegistryItem }>(
    `/registry/${segment(uid)}/test`
  )
}

/** POST /v1/registry/probe */
export function probeRegistry(
  body: ProbeRegistryParams
): Promise<{ probe: RegistryProbe }> {
  return post<{ probe: RegistryProbe }>("/registry/probe", body)
}

/** GET /v1/registry/{uid}/projects */
export function listRegistryProjects(
  uid: string,
  name = ""
): Promise<{ items: import("./types").RegistryProjectItem[] }> {
  return get<{ items: import("./types").RegistryProjectItem[] }>(
    `/registry/${segment(uid)}/projects`,
    { name }
  )
}

/** GET /v1/registry/{uid}/repositories */
export function listRegistryRepositories(
  uid: string,
  params: ListRegistryRepositoriesParams = {}
): Promise<ListRegistryRepositoriesResult> {
  return get<ListRegistryRepositoriesResult>(
    `/registry/${segment(uid)}/repositories`,
    params
  )
}

/** GET /v1/registry/{uid}/tags */
export function listRegistryTags(
  uid: string,
  params: ListRegistryTagsParams
): Promise<ListRegistryTagsResult> {
  return get<ListRegistryTagsResult>(
    `/registry/${segment(uid)}/tags`,
    params
  )
}

/** GET /v1/registry/{uid}/tag */
export function getRegistryTag(
  uid: string,
  params: GetRegistryTagParams
): Promise<{ tag: RegistryTagDetail }> {
  return get<{ tag: RegistryTagDetail }>(`/registry/${segment(uid)}/tag`, params)
}

/** POST /v1/registry/{uid}/tag/copy */
export function copyRegistryTag(
  uid: string,
  body: CopyRegistryTagParams
): Promise<{ tag: RegistryTagItem }> {
  return post<{ tag: RegistryTagItem }>(
    `/registry/${segment(uid)}/tag/copy`,
    body
  )
}

/** POST /v1/registry/{uid}/tag/delete */
export function deleteRegistryTag(
  uid: string,
  body: DeleteRegistryTagParams
): Promise<Record<string, never>> {
  return post<Record<string, never>>(
    `/registry/${segment(uid)}/tag/delete`,
    body
  )
}

/** GET /v1/image-syncs */
export function listImageSyncs(
  params: ListImageSyncsParams = {}
): Promise<ListImageSyncsResult> {
  return get<ListImageSyncsResult>("/image-syncs", params)
}

/** POST /v1/image-syncs/preview */
export function previewImageSync(
  body: PreviewImageSyncParams
): Promise<PreviewImageSyncResult> {
  return post<PreviewImageSyncResult>("/image-syncs/preview", body)
}

/** POST /v1/image-syncs */
export function createImageSyncs(
  items: ImageSyncInput[]
): Promise<CreateImageSyncsResult> {
  return post<CreateImageSyncsResult>("/image-syncs", { items })
}

/** GET /v1/image-syncs/{uid} */
export function getImageSync(uid: string): Promise<{ item: ImageSyncItem }> {
  return get<{ item: ImageSyncItem }>(`/image-syncs/${segment(uid)}`)
}

/** POST /v1/image-syncs/{uid}/retry */
export function retryImageSync(
  uid: string
): Promise<{ item: ImageSyncItem }> {
  return post<{ item: ImageSyncItem }>(`/image-syncs/${segment(uid)}/retry`)
}

/** POST /v1/image-syncs/{uid}/cancel */
export function cancelImageSync(
  uid: string
): Promise<{ item: ImageSyncItem }> {
  return post<{ item: ImageSyncItem }>(`/image-syncs/${segment(uid)}/cancel`)
}

export type * from "./types"
