// 平台升级与迁移。
//
// 接口契约（maintenance.proto → HTTP JSON，后端统一输出 camelCase）：
// - GET  /v1/system/maintenance           环境状态（system:migration:read）
// - POST /v1/system/maintenance/migrate   执行数据库迁移（system:migration:write）
// - GET  /v1/system/maintenance/archive   下载迁移包（裸 HTTP，Bearer JWT）
// - POST /v1/system/maintenance/archive   上传迁移包（裸 HTTP，请求体即文件内容）

/** 一张业务表及其迁移归属。 */
export type MaintenanceTable = {
  name: string
  /** true 表示进入迁移包；false 是运行数据，不迁移。 */
  archived: boolean
  /** 行数。int64 在 JSON 里是字符串，展示前用 Number() 归一。 */
  rows: number | string
  /** 表是否已存在于当前库。 */
  exists: boolean
}

export type MaintenanceStatus = {
  appVersion: string
  archiveFormatVersion: number
  postgresVersion: string
  archiveTableCount: number
  /** 配置数据总行数。int64 在 JSON 里是字符串。 */
  archiveRowCount: number | string
  /** ENCRYPT_KEY 的 SHA-256 前 16 位。 */
  encryptKeyCheck: string
  /** 配置表是否都已建好；false 表示还没执行过数据库迁移。 */
  schemaReady: boolean
  tables: MaintenanceTable[]
}

export type SchemaMigrationResult = {
  tableCount: number
  /** 耗时毫秒。int64 在 JSON 里是字符串。 */
  durationMs: number | string
  appVersion: string
}

export type ArchiveImportResult = {
  tables: number
  rows: number
  sourceVersion: string
  /** 导入会替换当前用户，前端需要清会话并回到登录页。 */
  reloginRequired: boolean
}
