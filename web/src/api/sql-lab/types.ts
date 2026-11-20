export interface SqlTableItem {
  name: string
  /** TABLE 或 VIEW。 */
  type: string
}

export interface SqlColumnItem {
  name: string
  dataType: string
  udtName: string
  nullable: boolean
  defaultValue: string
  length: number
  numericPrecision: number
  numericScale: number
  primaryKey: boolean
  unique: boolean
  position: number
}

export interface SqlIndexItem {
  name: string
  columns: string[]
  unique: boolean
  primary: boolean
  definition: string
}

export interface SqlQueryRow {
  cells: string[]
}

export interface SqlQueryResult {
  columns: string[]
  rows: SqlQueryRow[]
  rowCount: number
  /** int64 在 protojson 里是字符串，例如 "12"。 */
  elapsedMs: string
  truncated: boolean
}

export interface SqlColumnDefinition {
  name: string
  type: string
  nullable: boolean
  defaultValue: string
  primaryKey: boolean
  unique: boolean
  comment: string
}

export interface SqlAlterColumn {
  name: string
  /** 留空表示不改类型。 */
  type?: string
  /** 只有显式传值才会修改可空性，undefined 表示不改。 */
  nullable?: boolean
  /** 留空表示不改默认值。 */
  defaultValue?: string
  dropDefault?: boolean
}

export interface SqlAuditItem {
  uid: string
  /** CREATE_TABLE、ALTER_TABLE 或 DROP_TABLE。 */
  action: string
  database: string
  schema: string
  table: string
  statement: string
  /** SUCCESS 或 FAILED。 */
  status: string
  errorMessage: string
  operatorUid: string
  operatorName: string
  /** 执行时间，Unix 秒。int64 在 protojson 里是字符串。 */
  createdAt: string
}

export interface ListSqlDatabasesResult {
  items: string[]
}

export interface ListSqlSchemasParams {
  uid: string
  database?: string
}

export interface ListSqlSchemasResult {
  items: string[]
}

export interface ListSqlTablesParams {
  uid: string
  database?: string
  schema?: string
}

export interface ListSqlTablesResult {
  items: SqlTableItem[]
}

export interface ListSqlTableColumnsParams {
  uid: string
  database?: string
  schema?: string
  table: string
}

export interface ListSqlTableColumnsResult {
  items: SqlColumnItem[]
}

export interface ListSqlTableIndexesParams {
  uid: string
  database?: string
  schema?: string
  table: string
}

export interface ListSqlTableIndexesResult {
  items: SqlIndexItem[]
}

export interface QuerySqlParams {
  uid: string
  database?: string
  statement: string
  maxRows?: number
}

export interface CreateSqlTableParams {
  uid: string
  database?: string
  schema?: string
  table: string
  columns: SqlColumnDefinition[]
  ifNotExists?: boolean
  comment?: string
}

export interface CreateSqlTableResult {
  statement: string
}

export interface AlterSqlTableParams {
  uid: string
  database?: string
  schema?: string
  table: string
  renameTo?: string
  addColumns?: SqlColumnDefinition[]
  dropColumns?: string[]
  alterColumns?: SqlAlterColumn[]
}

export interface AlterSqlTableResult {
  statements: string[]
}

export interface DropSqlTableParams {
  uid: string
  database?: string
  schema?: string
  table: string
  ifExists?: boolean
}

export interface DropSqlTableResult {
  statement: string
}

export interface ListSqlAuditsParams {
  uid: string
  page?: number
  pageSize?: number
}

export interface ListSqlAuditsResult {
  items: SqlAuditItem[]
  total: number
  page: number
  pageSize: number
}
