// 功能套餐与租户功能授权。
//
// 接口契约（identity.proto → HTTP JSON，后端统一输出 camelCase / json_name）：
// - GET  /v1/system/features
// - GET  /v1/system/packages?page&pageSize&name
// - POST /v1/system/package
// - PUT  /v1/system/package/{uid}
// - DELETE /v1/system/package/{uid}
// - GET  /v1/system/namespace/{uid}/features
// - PUT  /v1/system/namespace/{uid}/features

/** 功能目录项。name / description 由后端提供，前端展示优先用本地双语目录。 */
export type FeatureItem = {
  code: string
  name?: string
  description?: string
  /** 该功能覆盖的租户权限码，例如 alert 对应 alert:* / alert:intervene。 */
  permissions?: string[]
}

export type ListFeaturesResult = {
  items?: FeatureItem[]
}

/** 平台级功能套餐。features 是该套餐包含的功能 code。 */
export type PackageItem = {
  uid: string
  name: string
  remark?: string
  features?: string[]
  /** 基础套餐只有一个，名称固定、不可删除，但可改功能与备注。 */
  isBase: boolean
  status: string
  /** Unix 秒（int64），展示时 *1000。 */
  createdAt?: number
}

export type ListPackagesParams = {
  page?: number
  pageSize?: number
  name?: string
}

export type ListPackagesResult = {
  items?: PackageItem[]
  total?: number
  page?: number
  pageSize?: number
}

/** 新建 / 编辑套餐共用的请求体。 */
export type PackagePayload = {
  name: string
  remark?: string
  features: string[]
}

export type UpdatePackagePayload = PackagePayload & {
  uid: string
}

/**
 * 租户当前的功能授权。
 * - packageFeatures：套餐本身包含的功能；
 * - features：按套餐 + 租户差异换算后的有效功能。
 */
export type NamespaceFeatures = {
  packageUid?: string
  packageName?: string
  features?: string[]
  packageFeatures?: string[]
}

/** PUT 的目标有效功能，服务端据此换算出租户差异。 */
export type UpdateNamespaceFeaturesPayload = {
  packageUid: string
  features: string[]
}
