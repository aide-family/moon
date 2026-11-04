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
