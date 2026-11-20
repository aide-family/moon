import { del, get, post, put } from "@/api/http"

import type {
  AlterSqlTableParams,
  AlterSqlTableResult,
  CreateSqlTableParams,
  CreateSqlTableResult,
  DropSqlTableParams,
  DropSqlTableResult,
  ListSqlAuditsParams,
  ListSqlAuditsResult,
  ListSqlDatabasesResult,
  ListSqlSchemasParams,
  ListSqlSchemasResult,
  ListSqlTableColumnsParams,
  ListSqlTableColumnsResult,
  ListSqlTableIndexesParams,
  ListSqlTableIndexesResult,
  ListSqlTablesParams,
  ListSqlTablesResult,
  QuerySqlParams,
  QuerySqlTableDataParams,
  SqlQueryResult,
  SqlTableDataResult,
} from "./types"

/** 拼出 DELETE 查询串；空字符串和 false 表示使用默认值，不发送。 */
function optionalQuery(
  params: Record<string, string | number | boolean | undefined>
): string {
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === "" || value === false) {
      continue
    }
    query.set(key, String(value))
  }
  const suffix = query.toString()
  return suffix ? `?${suffix}` : ""
}

/** 列出可以连接的库 GET /v1/datasource/{uid}/sql/databases */
export function listSqlDatabases(uid: string): Promise<ListSqlDatabasesResult> {
  return get<ListSqlDatabasesResult>(`/datasource/${uid}/sql/databases`)
}

/** 列出库里的模式 GET /v1/datasource/{uid}/sql/schemas */
export function listSqlSchemas(
  params: ListSqlSchemasParams
): Promise<ListSqlSchemasResult> {
  return get<ListSqlSchemasResult>(`/datasource/${params.uid}/sql/schemas`, {
    database: params.database,
  })
}

/** 列出模式里的表和视图 GET /v1/datasource/{uid}/sql/tables */
export function listSqlTables(
  params: ListSqlTablesParams
): Promise<ListSqlTablesResult> {
  return get<ListSqlTablesResult>(`/datasource/${params.uid}/sql/tables`, {
    database: params.database,
    schema: params.schema,
  })
}

/** 列出一张表的字段 GET /v1/datasource/{uid}/sql/columns */
export function listSqlTableColumns(
  params: ListSqlTableColumnsParams
): Promise<ListSqlTableColumnsResult> {
  return get<ListSqlTableColumnsResult>(
    `/datasource/${params.uid}/sql/columns`,
    {
      database: params.database,
      schema: params.schema,
      table: params.table,
    }
  )
}

/** 列出一张表的索引 GET /v1/datasource/{uid}/sql/indexes */
export function listSqlTableIndexes(
  params: ListSqlTableIndexesParams
): Promise<ListSqlTableIndexesResult> {
  return get<ListSqlTableIndexesResult>(
    `/datasource/${params.uid}/sql/indexes`,
    {
      database: params.database,
      schema: params.schema,
      table: params.table,
    }
  )
}

/** 分页读取表数据 POST /v1/datasource/{uid}/sql/table-data */
export function querySqlTableData(
  params: QuerySqlTableDataParams
): Promise<SqlTableDataResult> {
  const { uid, ...body } = params
  return post<SqlTableDataResult>(`/datasource/${uid}/sql/table-data`, body)
}

/** 执行只读查询 POST /v1/datasource/{uid}/sql/query */
export function querySql(params: QuerySqlParams): Promise<SqlQueryResult> {
  const { uid, ...body } = params
  return post<SqlQueryResult>(`/datasource/${uid}/sql/query`, body)
}

/** 建表 POST /v1/datasource/{uid}/sql/tables */
export function createSqlTable(
  params: CreateSqlTableParams
): Promise<CreateSqlTableResult> {
  const { uid, ...body } = params
  return post<CreateSqlTableResult>(`/datasource/${uid}/sql/tables`, body)
}

/** 改表 PUT /v1/datasource/{uid}/sql/tables */
export function alterSqlTable(
  params: AlterSqlTableParams
): Promise<AlterSqlTableResult> {
  const { uid, ...body } = params
  return put<AlterSqlTableResult>(`/datasource/${uid}/sql/tables`, body)
}

/** 删表 DELETE /v1/datasource/{uid}/sql/tables */
export function dropSqlTable(
  params: DropSqlTableParams
): Promise<DropSqlTableResult> {
  const query = optionalQuery({
    database: params.database,
    schema: params.schema,
    table: params.table,
    ifExists: params.ifExists,
  })
  return del<DropSqlTableResult>(
    `/datasource/${params.uid}/sql/tables${query}`
  )
}

/** 列出 DDL 审计 GET /v1/datasource/{uid}/sql/audits */
export function listSqlAudits(
  params: ListSqlAuditsParams
): Promise<ListSqlAuditsResult> {
  return get<ListSqlAuditsResult>(`/datasource/${params.uid}/sql/audits`, {
    page: params.page,
    pageSize: params.pageSize,
  })
}

export type {
  SqlAlterColumn,
  AlterSqlTableParams,
  AlterSqlTableResult,
  CreateSqlTableParams,
  CreateSqlTableResult,
  DropSqlTableParams,
  DropSqlTableResult,
  ListSqlAuditsParams,
  ListSqlAuditsResult,
  ListSqlDatabasesResult,
  ListSqlSchemasParams,
  ListSqlSchemasResult,
  ListSqlTableColumnsParams,
  ListSqlTableColumnsResult,
  ListSqlTableIndexesParams,
  ListSqlTableIndexesResult,
  ListSqlTablesParams,
  ListSqlTablesResult,
  QuerySqlParams,
  SqlAuditItem,
  SqlColumnDefinition,
  SqlColumnItem,
  SqlIndexItem,
  SqlQueryResult,
  SqlQueryRow,
  SqlTableItem,
} from "./types"
