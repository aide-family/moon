/** 官网（公开站）的站点设置，字段与后端 SiteSettingsItem 一致。 */
export interface SiteSettings {
  /** 站点名称，留空时前端回落为 Alioth。 */
  siteName: string
  /** logo 地址。 */
  logoUrl: string
  /** favicon 地址。 */
  faviconUrl: string
  /** 官方联系邮箱。 */
  contactEmail: string
  /** 代码仓库地址。 */
  repositoryUrl: string
  /** 页脚版权文案，留空时前端按当前年份回落。 */
  copyrightText: string
  /** ICP 备案号，留空表示不展示。 */
  filingIcpNumber: string
  /** ICP 备案查询地址。 */
  filingIcpUrl: string
  /** 公安备案号，留空表示不展示。 */
  filingPoliceNumber: string
  /** 公安备案查询地址。 */
  filingPoliceUrl: string
  /** 备案补充说明。 */
  filingExtraText: string
  /** 最后修改人。 */
  updatedBy: string
  /** 最后修改时间，RFC3339。 */
  updatedAt: string
}

/** 登录方式开关；两种方式至少要保留一种。 */
export interface LoginMethods {
  emailEnabled: boolean
  oauth2Enabled: boolean
  updatedBy: string
  updatedAt: string
}

/** 保存站点设置的载荷：全量覆盖，不含只读的修改人与修改时间。 */
export type SaveSiteSettingsParams = Omit<
  SiteSettings,
  "updatedBy" | "updatedAt"
>

/** 保存登录方式开关的载荷。 */
export type SaveLoginMethodsParams = Pick<
  LoginMethods,
  "emailEnabled" | "oauth2Enabled"
>

/** 公开接口返回的站点设置与登录方式。 */
export interface SiteConfig {
  site: SiteSettings
  loginMethods: LoginMethods
}

/** 公开页面：只含已发布快照字段。 */
export interface PublicSitePage {
  slug: string
  locale: string
  title: string
  summary: string
  bodyMd: string
  coverImageUrl: string
  seoTitle: string
  seoDescription: string
  /** RFC3339；从未发布时为空。 */
  publishedAt: string
  /** 请求语言没有可用页面、回落到另一种语言时为 true。 */
  localeFallback: boolean
}

/** 公开导航项：已发布且标记展示在导航里的页面。 */
export interface SiteNavItem {
  slug: string
  title: string
  sortOrder: number
}

/** 页面状态枚举。 */
export type SitePageState = "DRAFT" | "PUBLISHED" | "ARCHIVED"

/** 官网页面列表项（草稿视角）。 */
export interface SitePageItem {
  uid: string
  slug: string
  locale: string
  state: string
  showInNav: boolean
  sortOrder: number
  title: string
  summary: string
  hasUnpublishedChanges: boolean
  updatedAt: string
  publishedAt: string
}

/** 草稿与发布快照共用的内容字段。 */
export interface SitePageContent {
  title: string
  summary: string
  bodyMd: string
  coverImageUrl: string
  seoTitle: string
  seoDescription: string
}

/** 后台页面详情：草稿与已发布快照各一份。 */
export interface SitePageDetail {
  item: SitePageItem
  draft: SitePageContent
  published: SitePageContent
  publishedBy: string
}

/** 列表查询参数；空串表示不限。 */
export interface ListSitePagesParams {
  page?: number
  pageSize?: number
  keyword?: string
  state?: string
  locale?: string
}

export interface SitePageList {
  items: SitePageItem[]
  total: number
  page: number
  pageSize: number
}

/** 新建页面：slug 与语言仅创建时提交，创建后不可修改。 */
export interface CreateSitePageParams {
  slug: string
  locale: string
  title: string
  summary: string
  bodyMd: string
  coverImageUrl: string
  seoTitle: string
  seoDescription: string
}

/** 保存草稿：带 updatedAt 做乐观锁。 */
export interface UpdateSitePageParams {
  uid: string
  title: string
  summary: string
  bodyMd: string
  coverImageUrl: string
  seoTitle: string
  seoDescription: string
  updatedAt: string
}

/** 修改导航展示与顺序。 */
export interface UpdateSitePageNavParams {
  uid: string
  showInNav: boolean
  sortOrder: number
}
