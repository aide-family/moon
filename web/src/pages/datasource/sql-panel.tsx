import type { TFunction } from "i18next"
import { Plus, RefreshCw, Trash2 } from "lucide-react"
import type { ReactNode } from "react"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import type { DatasourceItem } from "@/api/datasource/types"
import {
  alterSqlTable,
  createSqlTable,
  dropSqlTable,
  listSqlAudits,
  listSqlDatabases,
  listSqlSchemas,
  listSqlTableColumns,
  listSqlTableIndexes,
  listSqlTables,
  querySql,
  querySqlTableData,
} from "@/api/sql-lab"
import type {
  SqlAlterColumn,
  SqlAuditItem,
  SqlColumnDefinition,
  SqlColumnItem,
  SqlIndexItem,
  SqlQueryResult,
  SqlTableDataResult,
  SqlTableItem,
} from "@/api/sql-lab/types"
import { CodeBlock } from "@/components/code-block"
import { FormField } from "@/components/form-field"
import { StatusBadge } from "@/components/status-badge"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import { cn } from "@/lib/class-names"
import { checkField, firstError } from "@/lib/form"
import { failureMessage } from "@/lib/form-error"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

type SqlTab = "query" | "browse" | "audits"
type BrowseTab = "data" | "columns" | "indexes"
type NullableChoice = "keep" | "true" | "false"

const AUDIT_PAGE_SIZE = 20
// 表数据默认每页 100 行，翻页时只换 offset。
const TABLE_DATA_PAGE_SIZE = 100
const CONNECTION_DATABASE = "__connection__"
const MAX_ROWS_OPTIONS = [100, 500, 1000]
const SQL_IDENTIFIER = /^[A-Za-z_][A-Za-z0-9_]{0,62}$/
const VIEW_TYPE = "VIEW"

interface ColumnDraft {
  name: string
  dataType: string
  nullable: boolean
  defaultValue: string
  primaryKey: boolean
  unique: boolean
  comment: string
}

interface AlterRowDraft {
  name: string
  dataType: string
  nullable: NullableChoice
  defaultValue: string
  dropDefault: boolean
}

function blankColumn(): ColumnDraft {
  return {
    name: "",
    dataType: "",
    nullable: true,
    defaultValue: "",
    primaryKey: false,
    unique: false,
    comment: "",
  }
}

function blankAlterRow(): AlterRowDraft {
  return {
    name: "",
    dataType: "",
    nullable: "keep",
    defaultValue: "",
    dropDefault: false,
  }
}

function toColumnDefinition(column: ColumnDraft): SqlColumnDefinition {
  return {
    name: column.name.trim(),
    type: column.dataType.trim(),
    nullable: column.nullable,
    defaultValue: column.defaultValue.trim(),
    primaryKey: column.primaryKey,
    unique: column.unique,
    comment: column.comment.trim(),
  }
}

/** 校验字段行；返回第一条错误，错误文案自带字段名与行号。 */
function validateColumns(
  columns: ColumnDraft[],
  t: TFunction,
  allowEmpty = false
): string | null {
  if (columns.length === 0) {
    return allowEmpty ? null : t("sqlLab.columnsRequired")
  }
  for (const [index, column] of columns.entries()) {
    const rowLabel = `${t("sqlLab.column")} ${index + 1}`
    const error = firstError(
      checkField({
        label: `${rowLabel} ${t("sqlLab.columnName")}`,
        value: column.name,
        required: true,
        max: 63,
        pattern: SQL_IDENTIFIER,
        hint: t("sqlLab.identifierHint"),
      }),
      checkField({
        label: `${rowLabel} ${t("sqlLab.columnDataType")}`,
        value: column.dataType,
        required: true,
        max: 64,
      }),
      checkField({
        label: `${rowLabel} ${t("sqlLab.columnComment")}`,
        value: column.comment,
        max: 200,
      })
    )
    if (error) {
      return error
    }
  }
  return null
}

function formatAuditTarget(audit: SqlAuditItem): string {
  return [audit.database, audit.schema, audit.table].filter(Boolean).join(".")
}

function formatAuditTime(seconds: string): string {
  const value = Number(seconds)
  if (!Number.isFinite(value) || value <= 0) {
    return "—"
  }
  const date = new Date(value * 1000)
  const pad = (value: number) => String(value).padStart(2, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

/** 带小标题的表单控件，用于字段行里密集排列的输入框。 */
function ColumnField({
  label,
  children,
}: {
  label: string
  children: ReactNode
}) {
  return (
    <label className="grid min-w-0 gap-1">
      <span className="text-xs leading-4 text-muted-foreground">{label}</span>
      {children}
    </label>
  )
}

/** 建表/加字段共用的字段行编辑器。 */
function ColumnRowsEditor({
  columns,
  onChange,
}: {
  columns: ColumnDraft[]
  onChange: (columns: ColumnDraft[]) => void
}) {
  const { t } = useTranslation()

  function patch(index: number, changes: Partial<ColumnDraft>) {
    onChange(
      columns.map((column, current) =>
        current === index ? { ...column, ...changes } : column
      )
    )
  }

  return (
    <div className="grid min-w-0 gap-2">
      {columns.map((column, index) => (
        <div
          key={index}
          className="relative grid min-w-0 gap-2 rounded-md border p-2 pr-10 sm:grid-cols-2 lg:grid-cols-4"
        >
          <ColumnField label={t("sqlLab.columnName")}>
            <Input
              value={column.name}
              aria-label={t("sqlLab.columnName")}
              placeholder={t("sqlLab.columnNamePlaceholder")}
              onChange={(event) => patch(index, { name: event.target.value })}
            />
          </ColumnField>
          <ColumnField label={t("sqlLab.columnDataType")}>
            <Input
              value={column.dataType}
              aria-label={t("sqlLab.columnDataType")}
              placeholder={t("sqlLab.columnTypePlaceholder")}
              onChange={(event) =>
                patch(index, { dataType: event.target.value })
              }
            />
          </ColumnField>
          <ColumnField label={t("sqlLab.columnDefault")}>
            <Input
              value={column.defaultValue}
              aria-label={t("sqlLab.columnDefault")}
              placeholder={t("sqlLab.columnDefaultPlaceholder")}
              onChange={(event) =>
                patch(index, { defaultValue: event.target.value })
              }
            />
          </ColumnField>
          <ColumnField label={t("sqlLab.columnComment")}>
            <Input
              value={column.comment}
              aria-label={t("sqlLab.columnComment")}
              placeholder={t("sqlLab.columnCommentPlaceholder")}
              onChange={(event) =>
                patch(index, { comment: event.target.value })
              }
            />
          </ColumnField>
          <div className="flex flex-wrap items-center gap-4 sm:col-span-2 lg:col-span-4">
            <label className="flex items-center gap-2 text-sm">
              <Checkbox
                checked={column.nullable}
                aria-label={t("sqlLab.columnNullable")}
                onCheckedChange={(value) =>
                  patch(index, { nullable: value === true })
                }
              />
              <span>{t("sqlLab.columnNullable")}</span>
            </label>
            <label className="flex items-center gap-2 text-sm">
              <Checkbox
                checked={column.primaryKey}
                aria-label={t("sqlLab.columnPrimaryKey")}
                onCheckedChange={(value) =>
                  patch(index, { primaryKey: value === true })
                }
              />
              <span>{t("sqlLab.columnPrimaryKey")}</span>
            </label>
            <label className="flex items-center gap-2 text-sm">
              <Checkbox
                checked={column.unique}
                aria-label={t("sqlLab.columnUnique")}
                onCheckedChange={(value) =>
                  patch(index, { unique: value === true })
                }
              />
              <span>{t("sqlLab.columnUnique")}</span>
            </label>
          </div>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className="absolute top-1 right-1"
            aria-label={t("sqlLab.removeColumn")}
            title={t("sqlLab.removeColumn")}
            onClick={() =>
              onChange(columns.filter((_, current) => current !== index))
            }
          >
            <Trash2 aria-hidden="true" className="size-3.5" />
          </Button>
        </div>
      ))}
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="justify-self-start"
        onClick={() => onChange([...columns, blankColumn()])}
      >
        <Plus aria-hidden="true" className="size-3.5" />
        {t("sqlLab.addColumn")}
      </Button>
    </div>
  )
}

/** 改表弹窗里的字段修改行：选择已有字段，再按需改类型/可空/默认值。 */
function AlterRowsEditor({
  rows,
  columns,
  onChange,
}: {
  rows: AlterRowDraft[]
  columns: SqlColumnItem[]
  onChange: (rows: AlterRowDraft[]) => void
}) {
  const { t } = useTranslation()

  function patch(index: number, changes: Partial<AlterRowDraft>) {
    onChange(
      rows.map((row, current) =>
        current === index ? { ...row, ...changes } : row
      )
    )
  }

  return (
    <div className="grid min-w-0 gap-2">
      {rows.map((row, index) => (
        <div
          key={index}
          className="relative grid min-w-0 gap-2 rounded-md border p-2 pr-10 sm:grid-cols-2 lg:grid-cols-4"
        >
          <ColumnField label={t("sqlLab.alterColumnField")}>
            <Select
              value={row.name || undefined}
              onValueChange={(value) => patch(index, { name: value })}
            >
              <SelectTrigger aria-label={t("sqlLab.alterColumnField")}>
                <SelectValue
                  placeholder={t("sqlLab.alterColumnFieldPlaceholder")}
                />
              </SelectTrigger>
              <SelectContent>
                {columns.map((column) => (
                  <SelectItem key={column.name} value={column.name}>
                    {column.name} · {column.dataType}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </ColumnField>
          <ColumnField label={t("sqlLab.columnDataType")}>
            <Input
              value={row.dataType}
              aria-label={t("sqlLab.columnDataType")}
              placeholder={t("sqlLab.alterColumnTypePlaceholder")}
              onChange={(event) =>
                patch(index, { dataType: event.target.value })
              }
            />
          </ColumnField>
          <ColumnField label={t("sqlLab.columnNullable")}>
            <Select
              value={row.nullable}
              onValueChange={(value) =>
                patch(index, { nullable: value as NullableChoice })
              }
            >
              <SelectTrigger aria-label={t("sqlLab.columnNullable")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="keep">
                  {t("sqlLab.nullableKeep")}
                </SelectItem>
                <SelectItem value="true">
                  {t("sqlLab.nullableTrue")}
                </SelectItem>
                <SelectItem value="false">
                  {t("sqlLab.nullableFalse")}
                </SelectItem>
              </SelectContent>
            </Select>
          </ColumnField>
          <ColumnField label={t("sqlLab.columnDefault")}>
            <Input
              value={row.defaultValue}
              aria-label={t("sqlLab.columnDefault")}
              placeholder={t("sqlLab.alterColumnDefaultPlaceholder")}
              onChange={(event) =>
                patch(index, { defaultValue: event.target.value })
              }
            />
          </ColumnField>
          <div className="flex flex-wrap items-center gap-4 sm:col-span-2 lg:col-span-4">
            <label className="flex items-center gap-2 text-sm">
              <Checkbox
                checked={row.dropDefault}
                aria-label={t("sqlLab.dropDefault")}
                onCheckedChange={(value) =>
                  patch(index, { dropDefault: value === true })
                }
              />
              <span>{t("sqlLab.dropDefault")}</span>
            </label>
          </div>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className="absolute top-1 right-1"
            aria-label={t("sqlLab.removeColumn")}
            title={t("sqlLab.removeColumn")}
            onClick={() =>
              onChange(rows.filter((_, current) => current !== index))
            }
          >
            <Trash2 aria-hidden="true" className="size-3.5" />
          </Button>
        </div>
      ))}
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="justify-self-start"
        disabled={columns.length === 0}
        onClick={() => onChange([...rows, blankAlterRow()])}
      >
        <Plus aria-hidden="true" className="size-3.5" />
        {t("sqlLab.addAlterColumn")}
      </Button>
    </div>
  )
}

function QueryResultTable({ result }: { result: SqlQueryResult }) {
  const { t } = useTranslation()

  if (result.columns.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">{t("sqlLab.queryEmpty")}</p>
    )
  }

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-lg border">
      <Table className="w-max min-w-full">
        <TableHeader>
          <TableRow>
            {result.columns.map((column, index) => (
              <TableHead key={`${column}-${index}`} className="whitespace-nowrap">
                {column}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {result.rows.map((row, rowIndex) => (
            <TableRow key={rowIndex}>
              {result.columns.map((_, columnIndex) => (
                <TableCell
                  key={`${rowIndex}-${columnIndex}`}
                  className="font-mono text-xs whitespace-nowrap"
                >
                  {row.cells[columnIndex] ?? ""}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

/** 表数据预览：按页展示，NULL 单元格显示为占位符。 */
function TableDataTable({ result }: { result: SqlTableDataResult }) {
  const { t } = useTranslation()
  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-lg border">
      <Table className="w-max min-w-full">
        <TableHeader>
          <TableRow>
            {result.columns.map((column) => (
              <TableHead
                key={column}
                className="sticky top-0 z-10 whitespace-nowrap bg-background font-mono text-xs"
              >
                {column}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {result.rows.length === 0 ? (
            <TableRow>
              <TableCell
                colSpan={Math.max(result.columns.length, 1)}
                className="text-muted-foreground"
              >
                {t("sqlLab.dataEmpty")}
              </TableCell>
            </TableRow>
          ) : (
            result.rows.map((row, rowIndex) => (
              <TableRow key={rowIndex}>
                {row.cells.map((cell, cellIndex) => (
                  <TableCell
                    key={cellIndex}
                    className="max-w-[24rem] truncate font-mono text-xs"
                    title={cell || undefined}
                  >
                    {cell || "—"}
                  </TableCell>
                ))}
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  )
}

export function SqlPanel({ datasource }: { datasource: DatasourceItem }) {
  const { t } = useTranslation()
  const [tab, setTab] = useState<SqlTab>("query")
  // MySQL 里模式和库是同一层，不需要单独选模式，查询统一用连接库。
  const isMysql = datasource.driver === "SQL_MYSQL"
  const [database, setDatabase] = useState("")
  const [schema, setSchema] = useState(isMysql ? "" : "public")
  const [databases, setDatabases] = useState<string[]>([])
  const [schemas, setSchemas] = useState<string[]>([])
  const [catalogLoading, setCatalogLoading] = useState(false)
  const [catalogError, setCatalogError] = useState("")
  const [catalogVersion, setCatalogVersion] = useState(0)

  // 查询页
  const [statement, setStatement] = useState("")
  const [maxRows, setMaxRows] = useState(String(MAX_ROWS_OPTIONS[0]))
  const [queryResult, setQueryResult] = useState<SqlQueryResult | null>(null)
  const [queryPending, setQueryPending] = useState(false)
  const [queryError, setQueryError] = useState("")

  // 库表页
  const [browseTab, setBrowseTab] = useState<BrowseTab>("data")
  const [tables, setTables] = useState<SqlTableItem[]>([])
  const [tablesLoading, setTablesLoading] = useState(false)
  const [tablesError, setTablesError] = useState("")
  const [tableFilter, setTableFilter] = useState("")
  const [selectedTable, setSelectedTable] = useState("")
  const [columns, setColumns] = useState<SqlColumnItem[]>([])
  const [indexes, setIndexes] = useState<SqlIndexItem[]>([])
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState("")
  const [actionStatement, setActionStatement] = useState("")
  const [browseVersion, setBrowseVersion] = useState(0)
  const [tableData, setTableData] = useState<SqlTableDataResult | null>(null)
  const [dataPage, setDataPage] = useState(1)
  const [dataLoading, setDataLoading] = useState(false)
  const [dataError, setDataError] = useState("")

  // 新建表
  const [createOpen, setCreateOpen] = useState(false)
  const [createTable, setCreateTable] = useState("")
  const [createComment, setCreateComment] = useState("")
  const [createIfNotExists, setCreateIfNotExists] = useState(false)
  const [createColumns, setCreateColumns] = useState<ColumnDraft[]>([
    blankColumn(),
  ])
  const [createFieldErrors, setCreateFieldErrors] = useState<{
    table?: string
    comment?: string
    columns?: string
  }>({})
  const [createError, setCreateError] = useState("")
  const [createPending, setCreatePending] = useState(false)

  // 改表
  const [alterOpen, setAlterOpen] = useState(false)
  const [alterName, setAlterName] = useState("")
  const [addColumns, setAddColumns] = useState<ColumnDraft[]>([])
  const [dropColumns, setDropColumns] = useState<string[]>([])
  const [alterRows, setAlterRows] = useState<AlterRowDraft[]>([])
  const [alterFieldErrors, setAlterFieldErrors] = useState<{
    rename?: string
    add?: string
    alter?: string
  }>({})
  const [alterError, setAlterError] = useState("")
  const [alterPending, setAlterPending] = useState(false)

  // 删表
  const [dropOpen, setDropOpen] = useState(false)
  const [dropConfirm, setDropConfirm] = useState("")
  const [dropIfExists, setDropIfExists] = useState(false)
  const [dropError, setDropError] = useState("")
  const [dropPending, setDropPending] = useState(false)

  // 审计
  const [audits, setAudits] = useState<SqlAuditItem[]>([])
  const [auditTotal, setAuditTotal] = useState(0)
  const [auditPage, setAuditPage] = useState(1)
  const [auditLoading, setAuditLoading] = useState(false)
  const [auditError, setAuditError] = useState("")
  const [auditVersion, setAuditVersion] = useState(0)
  const [auditDetail, setAuditDetail] = useState<SqlAuditItem | null>(null)

  const selectedTableItem = useMemo(
    () => tables.find((item) => item.name === selectedTable) ?? null,
    [selectedTable, tables]
  )
  const filteredTables = useMemo(() => {
    const keyword = tableFilter.trim().toLowerCase()
    if (!keyword) {
      return tables
    }
    return tables.filter((item) => item.name.toLowerCase().includes(keyword))
  }, [tableFilter, tables])
  const auditPages = Math.max(
    1,
    Math.ceil(auditTotal / Math.max(AUDIT_PAGE_SIZE, 1))
  )

  // 库和模式是面板的全局上下文：数据库变化时重新读取模式，并尽量保留 public。
  useEffect(() => {
    let active = true
    async function load() {
      setCatalogLoading(true)
      setCatalogError("")
      try {
        const [databaseResult, schemaResult] = await Promise.all([
          listSqlDatabases(datasource.uid),
          isMysql
            ? Promise.resolve({ items: [] as string[] })
            : listSqlSchemas({
                uid: datasource.uid,
                database: database || undefined,
              }),
        ])
        if (!active) {
          return
        }
        setDatabases(databaseResult.items ?? [])
        const nextSchemas = schemaResult.items ?? []
        setSchemas(nextSchemas)
        setSchema((current) => {
          if (nextSchemas.length === 0 || nextSchemas.includes(current)) {
            return current
          }
          return nextSchemas.includes("public") ? "public" : nextSchemas[0]
        })
      } catch (err) {
        if (active) {
          setCatalogError(failureMessage(err, t("state.failed")))
        }
      } finally {
        if (active) {
          setCatalogLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [catalogVersion, database, datasource.uid, t])

  // 库表列表只在库表页需要，切页后再加载。
  useEffect(() => {
    if (tab !== "browse") {
      return
    }
    let active = true
    async function load() {
      setTablesLoading(true)
      setTablesError("")
      try {
        const result = await listSqlTables({
          uid: datasource.uid,
          database: database || undefined,
          schema: schema || undefined,
        })
        if (!active) {
          return
        }
        const items = result.items ?? []
        setTables(items)
        setSelectedTable((current) =>
          current && items.some((item) => item.name === current) ? current : ""
        )
      } catch (err) {
        if (active) {
          setTablesError(failureMessage(err, t("state.failed")))
        }
      } finally {
        if (active) {
          setTablesLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [browseVersion, database, datasource.uid, schema, t, tab])

  // 选中表后才读取字段和索引；DDL 成功后用同一个版本号重载。
  useEffect(() => {
    if (!selectedTable) {
      return
    }
    let active = true
    async function load() {
      setDetailLoading(true)
      setDetailError("")
      try {
        const [columnResult, indexResult] = await Promise.all([
          listSqlTableColumns({
            uid: datasource.uid,
            database: database || undefined,
            schema: schema || undefined,
            table: selectedTable,
          }),
          listSqlTableIndexes({
            uid: datasource.uid,
            database: database || undefined,
            schema: schema || undefined,
            table: selectedTable,
          }),
        ])
        if (!active) {
          return
        }
        setColumns(columnResult.items ?? [])
        setIndexes(indexResult.items ?? [])
      } catch (err) {
        if (active) {
          setDetailError(failureMessage(err, t("state.failed")))
        }
      } finally {
        if (active) {
          setDetailLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [browseVersion, database, datasource.uid, schema, selectedTable, t])

  // 表数据按页读取；切表、切库、切模式和 DDL 之后由调用方把页码重置到第一页。
  useEffect(() => {
    if (!selectedTable || tab !== "browse") {
      return
    }
    let active = true
    async function load() {
      setDataLoading(true)
      setDataError("")
      try {
        const result = await querySqlTableData({
          uid: datasource.uid,
          database: database || undefined,
          schema: schema || undefined,
          table: selectedTable,
          page: dataPage,
          pageSize: TABLE_DATA_PAGE_SIZE,
        })
        if (active) {
          setTableData(result)
        }
      } catch (err) {
        if (active) {
          setTableData(null)
          setDataError(failureMessage(err, t("state.failed")))
        }
      } finally {
        if (active) {
          setDataLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [
    browseVersion,
    dataPage,
    database,
    datasource.uid,
    schema,
    selectedTable,
    t,
    tab,
  ])

  // 审计页分页加载。
  useEffect(() => {
    if (tab !== "audits") {
      return
    }
    let active = true
    async function load() {
      setAuditLoading(true)
      setAuditError("")
      try {
        const result = await listSqlAudits({
          uid: datasource.uid,
          page: auditPage,
          pageSize: AUDIT_PAGE_SIZE,
        })
        if (!active) {
          return
        }
        setAudits(result.items ?? [])
        setAuditTotal(result.total ?? 0)
        setAuditPage(result.page || 1)
      } catch (err) {
        if (active) {
          setAuditError(failureMessage(err, t("state.failed")))
        }
      } finally {
        if (active) {
          setAuditLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [auditPage, auditVersion, datasource.uid, t, tab])

  function changeDatabase(value: string) {
    setDatabase(value === CONNECTION_DATABASE ? "" : value)
    setSchema(isMysql ? "" : "public")
    setTables([])
    setSelectedTable("")
    setColumns([])
    setIndexes([])
    setTableData(null)
    setDataPage(1)
    setActionStatement("")
    setQueryResult(null)
    setQueryError("")
    setAuditPage(1)
  }

  function changeSchema(value: string) {
    setSchema(value)
    setSelectedTable("")
    setColumns([])
    setIndexes([])
    setTableData(null)
    setDataPage(1)
    setActionStatement("")
  }

  function selectTable(item: SqlTableItem) {
    setSelectedTable(item.name)
    setActionStatement("")
    setTableData(null)
    setDataPage(1)
  }

  function refreshPanel() {
    setCatalogVersion((value) => value + 1)
    setBrowseVersion((value) => value + 1)
    setAuditVersion((value) => value + 1)
    setDataPage(1)
  }

  async function runQuery() {
    if (queryPending) {
      return
    }
    const trimmed = statement.trim()
    if (!trimmed) {
      setQueryError(t("sqlLab.queryRequired"))
      return
    }
    setQueryPending(true)
    setQueryError("")
    try {
      const result = await querySql({
        uid: datasource.uid,
        database: database || undefined,
        statement: trimmed,
        maxRows: Number(maxRows),
      })
      setQueryResult(result)
    } catch (err) {
      setQueryResult(null)
      setQueryError(failureMessage(err, t("state.failed")))
    } finally {
      setQueryPending(false)
    }
  }

  function openCreate() {
    setCreateTable("")
    setCreateComment("")
    setCreateIfNotExists(false)
    setCreateColumns([blankColumn()])
    setCreateFieldErrors({})
    setCreateError("")
    setCreateOpen(true)
  }

  function resetCreate() {
    setCreateTable("")
    setCreateComment("")
    setCreateIfNotExists(false)
    setCreateColumns([blankColumn()])
    setCreateFieldErrors({})
    setCreateError("")
  }

  async function submitCreate() {
    if (createPending) {
      return
    }
    const tableError = checkField({
      label: t("sqlLab.tableName"),
      value: createTable,
      required: true,
      max: 63,
      pattern: SQL_IDENTIFIER,
      hint: t("sqlLab.identifierHint"),
    })
    const commentError = checkField({
      label: t("sqlLab.tableComment"),
      value: createComment,
      max: 200,
    })
    const columnsError = validateColumns(createColumns, t)
    if (tableError || commentError || columnsError) {
      setCreateError("")
      setCreateFieldErrors({
        table: tableError ?? undefined,
        comment: commentError ?? undefined,
        columns: columnsError ?? undefined,
      })
      return
    }
    setCreateFieldErrors({})
    setCreatePending(true)
    setCreateError("")
    try {
      const result = await createSqlTable({
        uid: datasource.uid,
        database: database || undefined,
        schema: schema || undefined,
        table: createTable.trim(),
        columns: createColumns.map(toColumnDefinition),
        ifNotExists: createIfNotExists,
        comment: createComment.trim() || undefined,
      })
      setActionStatement(result.statement)
      setBrowseVersion((value) => value + 1)
      setDataPage(1)
      setAuditVersion((value) => value + 1)
      setCreateOpen(false)
      setTab("browse")
    } catch (err) {
      setCreateError(failureMessage(err, t("state.failed")))
    } finally {
      setCreatePending(false)
    }
  }

  function openAlter() {
    if (!selectedTable) {
      return
    }
    setAlterName("")
    setAddColumns([])
    setDropColumns([])
    setAlterRows([])
    setAlterFieldErrors({})
    setAlterError("")
    setAlterOpen(true)
  }

  function resetAlter() {
    setAlterName("")
    setAddColumns([])
    setDropColumns([])
    setAlterRows([])
    setAlterFieldErrors({})
    setAlterError("")
  }

  async function submitAlter() {
    if (!selectedTable || alterPending) {
      return
    }
    const renameValue = alterName.trim()
    const renameChanged = renameValue !== "" && renameValue !== selectedTable
    const renameError = renameChanged
      ? checkField({
          label: t("sqlLab.renameTo"),
          value: renameValue,
          max: 63,
          pattern: SQL_IDENTIFIER,
          hint: t("sqlLab.identifierHint"),
        })
      : null
    const addError = validateColumns(addColumns, t, true)

    let alterColumnError: string | null = null
    const alterPayload: SqlAlterColumn[] = []
    for (const [index, row] of alterRows.entries()) {
      const rowLabel = `${t("sqlLab.alterColumnField")} ${index + 1}`
      const hasDraft =
        row.dataType.trim() !== "" ||
        row.nullable !== "keep" ||
        row.defaultValue.trim() !== "" ||
        row.dropDefault
      if (!row.name) {
        if (hasDraft) {
          alterColumnError = t("field.required", { label: rowLabel })
          break
        }
        continue
      }
      if (row.dropDefault && row.defaultValue.trim() !== "") {
        alterColumnError = t("sqlLab.defaultConflict")
        break
      }
      if (!hasDraft) {
        continue
      }
      const change: SqlAlterColumn = { name: row.name }
      if (row.dataType.trim()) {
        change.type = row.dataType.trim()
      }
      if (row.nullable !== "keep") {
        change.nullable = row.nullable === "true"
      }
      if (row.defaultValue.trim()) {
        change.defaultValue = row.defaultValue.trim()
      }
      if (row.dropDefault) {
        change.dropDefault = true
      }
      alterPayload.push(change)
    }

    if (renameError || addError || alterColumnError) {
      setAlterError("")
      setAlterFieldErrors({
        rename: renameError ?? undefined,
        add: addError ?? undefined,
        alter: alterColumnError ?? undefined,
      })
      return
    }
    const changed =
      renameChanged ||
      addColumns.length > 0 ||
      dropColumns.length > 0 ||
      alterPayload.length > 0
    if (!changed) {
      setAlterFieldErrors({})
      setAlterError(t("sqlLab.noChanges"))
      return
    }
    setAlterFieldErrors({})
    setAlterPending(true)
    setAlterError("")
    try {
      const result = await alterSqlTable({
        uid: datasource.uid,
        database: database || undefined,
        schema: schema || undefined,
        table: selectedTable,
        renameTo: renameChanged ? renameValue : undefined,
        addColumns:
          addColumns.length > 0
            ? addColumns.map(toColumnDefinition)
            : undefined,
        dropColumns: dropColumns.length > 0 ? dropColumns : undefined,
        alterColumns: alterPayload.length > 0 ? alterPayload : undefined,
      })
      setActionStatement((result.statements ?? []).join("\n"))
      setSelectedTable(renameChanged ? renameValue : selectedTable)
      setBrowseVersion((value) => value + 1)
      setAuditVersion((value) => value + 1)
      setDataPage(1)
      setAlterOpen(false)
    } catch (err) {
      setAlterError(failureMessage(err, t("state.failed")))
    } finally {
      setAlterPending(false)
    }
  }

  function openDrop() {
    if (!selectedTable) {
      return
    }
    setDropConfirm("")
    setDropIfExists(false)
    setDropError("")
    setDropOpen(true)
  }

  async function submitDrop() {
    if (!selectedTable || dropConfirm !== selectedTable || dropPending) {
      return
    }
    setDropPending(true)
    setDropError("")
    try {
      const result = await dropSqlTable({
        uid: datasource.uid,
        database: database || undefined,
        schema: schema || undefined,
        table: selectedTable,
        ifExists: dropIfExists,
      })
      setActionStatement(result.statement)
      setSelectedTable("")
      setColumns([])
      setIndexes([])
      setTableData(null)
      setDataPage(1)
      setBrowseVersion((value) => value + 1)
      setAuditVersion((value) => value + 1)
      setDropOpen(false)
    } catch (err) {
      setDropError(failureMessage(err, t("state.failed")))
    } finally {
      setDropPending(false)
    }
  }

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-3">
      <div className="flex shrink-0 flex-wrap items-center gap-2">
        <Select
          value={database || CONNECTION_DATABASE}
          onValueChange={changeDatabase}
        >
          <SelectTrigger className="w-56" aria-label={t("sqlLab.database")}>
            <SelectValue placeholder={t("sqlLab.database")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={CONNECTION_DATABASE}>
              {t("sqlLab.databaseConnection")}
            </SelectItem>
            {databases.map((item) => (
              <SelectItem key={item} value={item}>
                {item}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {isMysql ? null : (
          <Select value={schema} onValueChange={changeSchema}>
            <SelectTrigger className="w-44" aria-label={t("sqlLab.schema")}>
              <SelectValue placeholder={t("sqlLab.schema")} />
            </SelectTrigger>
            <SelectContent>
              {(schema && !schemas.includes(schema)
                ? [schema, ...schemas]
                : schemas
              ).map((item) => (
                <SelectItem key={item} value={item}>
                  {item}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
        <Button
          type="button"
          variant="outline"
          disabled={catalogLoading}
          onClick={refreshPanel}
        >
          <RefreshCw
            aria-hidden="true"
            className={cn(
              "size-4",
              catalogLoading && "animate-spin motion-reduce:animate-none"
            )}
          />
          {t("action.refresh")}
        </Button>
        {catalogError ? (
          <span className="text-sm text-destructive">{catalogError}</span>
        ) : null}
      </div>
      <Tabs
        value={tab}
        onValueChange={(value) => setTab(value as SqlTab)}
        className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-hidden"
      >
        <TabsList className="shrink-0 self-start">
          <TabsTrigger value="query">{t("sqlLab.tabQuery")}</TabsTrigger>
          <TabsTrigger value="browse">{t("sqlLab.tabBrowse")}</TabsTrigger>
          <TabsTrigger value="audits">{t("sqlLab.tabAudits")}</TabsTrigger>
        </TabsList>

        <TabsContent
          value="query"
          className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-hidden"
        >
          <div className="shrink-0 p-1.5 -m-1.5">
            <FormField
              label={t("sqlLab.statement")}
              htmlFor="sql-statement"
              required
            >
              <Textarea
                id="sql-statement"
                className="min-h-28 resize-y font-mono text-xs"
                value={statement}
                placeholder={t("sqlLab.statementPlaceholder")}
                spellCheck={false}
                required
                onChange={(event) => setStatement(event.target.value)}
                onKeyDown={(event) => {
                  if (
                    (event.metaKey || event.ctrlKey) &&
                    event.key === "Enter"
                  ) {
                    event.preventDefault()
                    void runQuery()
                  }
                }}
              />
            </FormField>
            <p className="mt-1 text-xs text-muted-foreground">
              {t("sqlLab.readOnlyHint")}
            </p>
            <div className="mt-2 flex flex-wrap items-center gap-2">
              <Select value={maxRows} onValueChange={setMaxRows}>
                <SelectTrigger
                  className="w-28"
                  aria-label={t("sqlLab.maxRows")}
                >
                  <SelectValue placeholder={t("sqlLab.maxRows")} />
                </SelectTrigger>
                <SelectContent>
                  {MAX_ROWS_OPTIONS.map((item) => (
                    <SelectItem key={item} value={String(item)}>
                      {item}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Button
                type="button"
                disabled={queryPending}
                onClick={() => void runQuery()}
              >
                {queryPending ? t("sqlLab.running") : t("sqlLab.runQuery")}
              </Button>
              <span className="text-xs text-muted-foreground">
                {t("sqlLab.shortcut")}
              </span>
            </div>
          </div>
          {queryError ? (
            <p className="shrink-0 text-sm text-destructive">{queryError}</p>
          ) : null}
          <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-2">
            {queryResult ? (
              <div className="flex shrink-0 flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
                <span className="font-mono">
                  {t("sqlLab.resultSummary", {
                    rows: queryResult.rowCount,
                    ms: Number(queryResult.elapsedMs),
                  })}
                </span>
                {queryResult.truncated ? (
                  <span className="rounded border border-warning-border bg-warning-soft px-1.5 py-0.5 text-warning">
                    {t("sqlLab.truncated")}
                  </span>
                ) : null}
              </div>
            ) : null}
            {queryResult ? (
              <QueryResultTable result={queryResult} />
            ) : (
              <div className="flex min-h-0 flex-1 items-center justify-center rounded-lg border bg-muted/10 p-6 text-center text-sm text-muted-foreground">
                {t("sqlLab.queryEmpty")}
              </div>
            )}
          </div>
        </TabsContent>

        <TabsContent
          value="browse"
          className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-hidden"
        >
          <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-hidden lg:flex-row">
            <aside className="flex h-44 min-h-0 shrink-0 flex-col rounded-lg border lg:h-auto lg:w-56">
              <div className="shrink-0 border-b p-2">
                <Input
                  value={tableFilter}
                  aria-label={t("sqlLab.filterTables")}
                  placeholder={t("sqlLab.filterTables")}
                  onChange={(event) => setTableFilter(event.target.value)}
                />
              </div>
              <div className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto p-1">
                {tablesLoading ? (
                  <p className="p-2 text-sm text-muted-foreground">
                    {t("state.loading")}
                  </p>
                ) : tablesError ? (
                  <p className="p-2 text-sm text-destructive">{tablesError}</p>
                ) : filteredTables.length === 0 ? (
                  <p className="p-2 text-sm text-muted-foreground">
                    {tables.length === 0
                      ? t("sqlLab.tablesEmpty")
                      : t("sqlLab.tableFilterEmpty")}
                  </p>
                ) : (
                  filteredTables.map((item) => (
                    <Button
                      key={item.name}
                      type="button"
                      variant={selectedTable === item.name ? "secondary" : "ghost"}
                      size="sm"
                      className="w-full justify-between gap-2"
                      aria-pressed={selectedTable === item.name}
                      onClick={() => selectTable(item)}
                    >
                      <span className="min-w-0 truncate">{item.name}</span>
                      <span className="shrink-0 rounded border bg-muted/60 px-1 py-0.5 text-[10px] leading-none font-medium text-muted-foreground">
                        {item.type === VIEW_TYPE
                          ? t("sqlLab.view")
                          : t("sqlLab.table")}
                      </span>
                    </Button>
                  ))
                )}
              </div>
            </aside>
            <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-hidden">
              <div className="flex shrink-0 flex-wrap items-center justify-between gap-2">
                <div className="flex min-w-0 items-center gap-2">
                  <span className="min-w-0 truncate text-sm font-medium">
                    {selectedTableItem
                      ? selectedTableItem.name
                      : t("sqlLab.pickTable")}
                  </span>
                  {selectedTableItem ? (
                    <span className="shrink-0 rounded border bg-muted/60 px-1.5 py-0.5 text-[10px] leading-none font-medium text-muted-foreground">
                      {selectedTableItem.type === VIEW_TYPE
                        ? t("sqlLab.view")
                        : t("sqlLab.table")}
                    </span>
                  ) : null}
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    onClick={openCreate}
                  >
                    <Plus aria-hidden="true" className="size-3.5" />
                    {t("sqlLab.createTable")}
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={!selectedTable}
                    onClick={openAlter}
                  >
                    {t("sqlLab.alterTable")}
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="destructive"
                    disabled={!selectedTable}
                    onClick={openDrop}
                  >
                    {t("sqlLab.dropTable")}
                  </Button>
                </div>
              </div>
              {actionStatement ? (
                <CodeBlock
                  className="shrink-0"
                  code={actionStatement}
                  language="text"
                  label={t("sqlLab.statementResult")}
                  copyable
                  wrap
                  maxHeightClass="max-h-32"
                />
              ) : null}
              {selectedTable ? (
                <Tabs
                  value={browseTab}
                  onValueChange={(value) => setBrowseTab(value as BrowseTab)}
                  className="flex min-h-0 min-w-0 flex-1 flex-col gap-2 overflow-hidden"
                >
                  <TabsList className="shrink-0 self-start">
                    <TabsTrigger value="data">{t("sqlLab.data")}</TabsTrigger>
                    <TabsTrigger value="columns">
                      {t("sqlLab.columns")}
                    </TabsTrigger>
                    <TabsTrigger value="indexes">
                      {t("sqlLab.indexes")}
                    </TabsTrigger>
                  </TabsList>
                  <TabsContent
                    value="data"
                    className="flex min-h-0 min-w-0 flex-1 flex-col gap-2 overflow-hidden"
                  >
                    <div className="flex shrink-0 flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
                      <span className="font-mono">
                        {tableData
                          ? t("sqlLab.dataSummary", {
                              rows: tableData.rowCount,
                              ms: Number(tableData.elapsedMs),
                            })
                          : t("sqlLab.pageLabel", { page: dataPage })}
                      </span>
                      <div className="flex items-center gap-2">
                        <Button
                          type="button"
                          size="sm"
                          variant="outline"
                          disabled={dataLoading || dataPage <= 1}
                          onClick={() =>
                            setDataPage((value) => Math.max(1, value - 1))
                          }
                        >
                          {t("action.prev")}
                        </Button>
                        <span className="font-mono">
                          {t("sqlLab.pageLabel", { page: dataPage })}
                        </span>
                        <Button
                          type="button"
                          size="sm"
                          variant="outline"
                          disabled={dataLoading || !tableData?.hasMore}
                          onClick={() => setDataPage((value) => value + 1)}
                        >
                          {t("action.next")}
                        </Button>
                      </div>
                    </div>
                    {dataLoading ? (
                      <p className="p-3 text-sm text-muted-foreground">
                        {t("state.loading")}
                      </p>
                    ) : dataError ? (
                      <p className="p-3 text-sm text-destructive">
                        {dataError}
                      </p>
                    ) : tableData ? (
                      <TableDataTable result={tableData} />
                    ) : (
                      <div className="flex min-h-0 flex-1 items-center justify-center rounded-lg border bg-muted/10 p-6 text-center text-sm text-muted-foreground">
                        {t("sqlLab.dataEmpty")}
                      </div>
                    )}
                  </TabsContent>
                  <TabsContent
                    value="columns"
                    className="flex min-h-0 min-w-0 flex-1 flex-col gap-2 overflow-hidden"
                  >
                  <section className="flex min-h-0 min-w-0 flex-1 flex-col gap-2">
                    <div className="flex shrink-0 items-center justify-between gap-2">
                      <h3 className="text-sm font-medium">
                        {t("sqlLab.columns")}
                      </h3>
                      <span className="font-mono text-xs text-muted-foreground">
                        {columns.length}
                      </span>
                    </div>
                    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-lg border">
                      {detailLoading ? (
                        <p className="p-3 text-sm text-muted-foreground">
                          {t("state.loading")}
                        </p>
                      ) : detailError ? (
                        <p className="p-3 text-sm text-destructive">
                          {detailError}
                        </p>
                      ) : (
                        <Table className="w-max min-w-full">
                          <TableHeader>
                            <TableRow>
                              <TableHead>{t("sqlLab.columnName")}</TableHead>
                              <TableHead>{t("sqlLab.columnDataType")}</TableHead>
                              <TableHead>{t("sqlLab.columnNullable")}</TableHead>
                              <TableHead>{t("sqlLab.columnDefault")}</TableHead>
                              <TableHead>
                                {t("sqlLab.columnPrimaryKey")}
                              </TableHead>
                              <TableHead>{t("sqlLab.columnUnique")}</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {columns.length === 0 ? (
                              <TableRow>
                                <TableCell
                                  colSpan={6}
                                  className="text-muted-foreground"
                                >
                                  {t("sqlLab.columnsEmpty")}
                                </TableCell>
                              </TableRow>
                            ) : (
                              columns.map((column) => (
                                <TableRow key={column.name}>
                                  <TableCell className="font-medium">
                                    {column.name}
                                  </TableCell>
                                  <TableCell className="font-mono text-xs">
                                    {column.dataType || column.udtName || "—"}
                                  </TableCell>
                                  <TableCell>
                                    {column.nullable
                                      ? t("sqlLab.yes")
                                      : t("sqlLab.no")}
                                  </TableCell>
                                  <TableCell
                                    className="max-w-[18rem] truncate font-mono text-xs"
                                    title={column.defaultValue || undefined}
                                  >
                                    {column.defaultValue || "—"}
                                  </TableCell>
                                  <TableCell>
                                    {column.primaryKey
                                      ? t("sqlLab.yes")
                                      : t("sqlLab.no")}
                                  </TableCell>
                                  <TableCell>
                                    {column.unique
                                      ? t("sqlLab.yes")
                                      : t("sqlLab.no")}
                                  </TableCell>
                                </TableRow>
                              ))
                            )}
                          </TableBody>
                        </Table>
                      )}
                    </div>
                  </section>
                  </TabsContent>
                  <TabsContent
                    value="indexes"
                    className="flex min-h-0 min-w-0 flex-1 flex-col gap-2 overflow-hidden"
                  >
                  <section className="flex min-h-0 min-w-0 flex-1 flex-col gap-2">
                    <div className="flex shrink-0 items-center justify-between gap-2">
                      <h3 className="text-sm font-medium">
                        {t("sqlLab.indexes")}
                      </h3>
                      <span className="font-mono text-xs text-muted-foreground">
                        {indexes.length}
                      </span>
                    </div>
                    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-lg border">
                      {detailLoading ? (
                        <p className="p-3 text-sm text-muted-foreground">
                          {t("state.loading")}
                        </p>
                      ) : detailError ? (
                        <p className="p-3 text-sm text-destructive">
                          {detailError}
                        </p>
                      ) : (
                        <Table className="w-max min-w-full">
                          <TableHeader>
                            <TableRow>
                              <TableHead>{t("sqlLab.indexName")}</TableHead>
                              <TableHead>{t("sqlLab.indexUnique")}</TableHead>
                              <TableHead>{t("sqlLab.indexPrimary")}</TableHead>
                              <TableHead>{t("sqlLab.indexColumns")}</TableHead>
                              <TableHead>
                                {t("sqlLab.indexDefinition")}
                              </TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {indexes.length === 0 ? (
                              <TableRow>
                                <TableCell
                                  colSpan={5}
                                  className="text-muted-foreground"
                                >
                                  {t("sqlLab.indexesEmpty")}
                                </TableCell>
                              </TableRow>
                            ) : (
                              indexes.map((index) => (
                                <TableRow key={index.name}>
                                  <TableCell className="font-medium">
                                    {index.name}
                                  </TableCell>
                                  <TableCell>
                                    {index.unique
                                      ? t("sqlLab.yes")
                                      : t("sqlLab.no")}
                                  </TableCell>
                                  <TableCell>
                                    {index.primary
                                      ? t("sqlLab.yes")
                                      : t("sqlLab.no")}
                                  </TableCell>
                                  <TableCell className="font-mono text-xs">
                                    {index.columns.join(", ") || "—"}
                                  </TableCell>
                                  <TableCell
                                    className="max-w-[28rem] truncate font-mono text-xs"
                                    title={index.definition || undefined}
                                  >
                                    {index.definition || "—"}
                                  </TableCell>
                                </TableRow>
                              ))
                            )}
                          </TableBody>
                        </Table>
                      )}
                    </div>
                  </section>
                  </TabsContent>
                </Tabs>
              ) : (
                <div className="flex min-h-0 flex-1 items-center justify-center rounded-lg border bg-muted/10 p-6 text-center text-sm text-muted-foreground">
                  {t("sqlLab.pickTable")}
                </div>
              )}
            </div>
          </div>
        </TabsContent>

        <TabsContent
          value="audits"
          className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-hidden"
        >
          <div className="flex shrink-0 flex-wrap items-center justify-between gap-2">
            <span className="font-mono text-sm text-muted-foreground">
              {t("list.total", { count: auditTotal })}
            </span>
            {auditTotal > 0 ? (
              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={auditLoading || auditPage <= 1}
                  onClick={() =>
                    setAuditPage((value) => Math.max(1, value - 1))
                  }
                >
                  {t("action.prev")}
                </Button>
                <span className="font-mono text-sm text-muted-foreground">
                  {auditPage} / {auditPages}
                </span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={
                    auditLoading || auditPage * AUDIT_PAGE_SIZE >= auditTotal
                  }
                  onClick={() => setAuditPage((value) => value + 1)}
                >
                  {t("action.next")}
                </Button>
              </div>
            ) : null}
          </div>
          {auditLoading ? (
            <p className="text-sm text-muted-foreground">
              {t("state.loading")}
            </p>
          ) : auditError ? (
            <p className="text-sm text-destructive">{auditError}</p>
          ) : audits.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {t("sqlLab.auditsEmpty")}
            </p>
          ) : (
            <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-lg border">
              <Table className={TABLE_MIN_WIDTH.default}>
                <TableHeader>
                  <TableRow>
                    <TableHead
                      className={cn(TABLE_COLUMN_WIDTH.time, TABLE_NOWRAP)}
                    >
                      {t("sqlLab.auditTime")}
                    </TableHead>
                    <TableHead
                      className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                    >
                      {t("sqlLab.auditAction")}
                    </TableHead>
                    <TableHead>{t("sqlLab.auditTarget")}</TableHead>
                    <TableHead
                      className={cn(TABLE_COLUMN_WIDTH.status, TABLE_NOWRAP)}
                    >
                      {t("sqlLab.auditStatus")}
                    </TableHead>
                    <TableHead
                      className={cn(TABLE_COLUMN_WIDTH.identity, TABLE_NOWRAP)}
                    >
                      {t("sqlLab.auditOperator")}
                    </TableHead>
                    <TableHead>{t("sqlLab.auditStatement")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {audits.map((audit) => (
                    <TableRow
                      key={audit.uid}
                      className="cursor-pointer"
                      onClick={() => setAuditDetail(audit)}
                    >
                      <TableCell
                        className={cn(
                          TABLE_COLUMN_WIDTH.time,
                          TABLE_NOWRAP,
                          "font-mono text-xs tabular-nums text-muted-foreground"
                        )}
                      >
                        {formatAuditTime(audit.createdAt)}
                      </TableCell>
                      <TableCell
                        className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                      >
                        <StatusBadge value={audit.action} />
                      </TableCell>
                      <TableCell className="max-w-[16rem] break-all whitespace-normal">
                        {formatAuditTarget(audit) || "—"}
                      </TableCell>
                      <TableCell
                        className={cn(TABLE_COLUMN_WIDTH.status, TABLE_NOWRAP)}
                      >
                        <StatusBadge value={audit.status} />
                      </TableCell>
                      <TableCell
                        className={cn(
                          TABLE_COLUMN_WIDTH.identity,
                          TABLE_NOWRAP
                        )}
                      >
                        {audit.operatorName || audit.operatorUid || "—"}
                      </TableCell>
                      <TableCell>
                        <span
                          className="block max-w-[24rem] truncate font-mono text-xs"
                          title={audit.statement}
                        >
                          {audit.statement || "—"}
                        </span>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </TabsContent>
      </Tabs>

      <Dialog
        open={createOpen}
        onOpenChange={(open) => {
          setCreateOpen(open)
          if (!open) {
            resetCreate()
          }
        }}
      >
        <DialogContent className="sm:max-w-4xl">
          <DialogHeader>
            <DialogTitle>{t("sqlLab.createTitle")}</DialogTitle>
          </DialogHeader>
          <form
            className="flex min-h-0 flex-1 flex-col gap-4"
            onSubmit={(event) => {
              event.preventDefault()
              void submitCreate()
            }}
          >
            <DialogBody className="grid gap-4 sm:grid-cols-2">
              <FormField
                label={t("sqlLab.tableName")}
                htmlFor="sql-create-table"
                required
                error={createFieldErrors.table}
              >
                <Input
                  id="sql-create-table"
                  value={createTable}
                  placeholder={t("sqlLab.tableNamePlaceholder")}
                  onChange={(event) => {
                    setCreateTable(event.target.value)
                    if (createFieldErrors.table) {
                      setCreateFieldErrors((current) => ({
                        ...current,
                        table: undefined,
                      }))
                    }
                  }}
                />
              </FormField>
              <FormField
                label={t("sqlLab.tableComment")}
                htmlFor="sql-create-comment"
                error={createFieldErrors.comment}
              >
                <Input
                  id="sql-create-comment"
                  value={createComment}
                  placeholder={t("sqlLab.tableCommentPlaceholder")}
                  onChange={(event) => {
                    setCreateComment(event.target.value)
                    if (createFieldErrors.comment) {
                      setCreateFieldErrors((current) => ({
                        ...current,
                        comment: undefined,
                      }))
                    }
                  }}
                />
              </FormField>
              <label className="flex items-center gap-2 text-sm sm:col-span-full">
                <Checkbox
                  checked={createIfNotExists}
                  onCheckedChange={(value) =>
                    setCreateIfNotExists(value === true)
                  }
                />
                <span>{t("sqlLab.ifNotExists")}</span>
              </label>
              <FormField
                label={t("sqlLab.columns")}
                required
                className="sm:col-span-full"
                error={createFieldErrors.columns}
              >
                <ColumnRowsEditor
                  columns={createColumns}
                  onChange={(next) => {
                    setCreateColumns(next)
                    if (createFieldErrors.columns) {
                      setCreateFieldErrors((current) => ({
                        ...current,
                        columns: undefined,
                      }))
                    }
                  }}
                />
              </FormField>
            </DialogBody>
            {createError ? (
              <p className="text-sm text-destructive">{createError}</p>
            ) : null}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setCreateOpen(false)}
              >
                {t("action.cancel")}
              </Button>
              <Button type="submit" disabled={createPending}>
                {createPending ? t("action.saving") : t("action.create")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog
        open={alterOpen}
        onOpenChange={(open) => {
          setAlterOpen(open)
          if (!open) {
            resetAlter()
          }
        }}
      >
        <DialogContent className="sm:max-w-4xl">
          <DialogHeader>
            <DialogTitle>{t("sqlLab.alterTitle")}</DialogTitle>
          </DialogHeader>
          <form
            className="flex min-h-0 flex-1 flex-col gap-4"
            onSubmit={(event) => {
              event.preventDefault()
              void submitAlter()
            }}
          >
            <DialogBody className="grid gap-4 sm:grid-cols-2">
              <FormField
                label={t("sqlLab.renameTo")}
                htmlFor="sql-alter-rename"
                error={alterFieldErrors.rename}
              >
                <Input
                  id="sql-alter-rename"
                  value={alterName}
                  placeholder={t("sqlLab.renameToPlaceholder")}
                  onChange={(event) => {
                    setAlterName(event.target.value)
                    if (alterFieldErrors.rename) {
                      setAlterFieldErrors((current) => ({
                        ...current,
                        rename: undefined,
                      }))
                    }
                  }}
                />
              </FormField>
              <div className="grid content-start gap-2">
                <span className="text-sm font-medium">
                  {t("sqlLab.dropColumns")}
                </span>
                {columns.length === 0 ? (
                  <p className="text-xs text-muted-foreground">
                    {t("sqlLab.dropColumnsEmpty")}
                  </p>
                ) : (
                  <div className="flex max-h-32 flex-wrap gap-x-4 gap-y-2 overflow-y-auto rounded-md border p-2">
                    {columns.map((column) => (
                      <label
                        key={column.name}
                        className="flex items-center gap-2 text-sm"
                      >
                        <Checkbox
                          checked={dropColumns.includes(column.name)}
                          aria-label={column.name}
                          onCheckedChange={(value) =>
                            setDropColumns((current) =>
                              value === true
                                ? [...current, column.name]
                                : current.filter(
                                    (item) => item !== column.name
                                  )
                            )
                          }
                        />
                        <span className="font-mono text-xs">{column.name}</span>
                      </label>
                    ))}
                  </div>
                )}
              </div>
              <FormField
                label={t("sqlLab.addColumns")}
                className="sm:col-span-full"
                error={alterFieldErrors.add}
              >
                <ColumnRowsEditor
                  columns={addColumns}
                  onChange={(next) => {
                    setAddColumns(next)
                    if (alterFieldErrors.add) {
                      setAlterFieldErrors((current) => ({
                        ...current,
                        add: undefined,
                      }))
                    }
                  }}
                />
              </FormField>
              <FormField
                label={t("sqlLab.alterColumns")}
                className="sm:col-span-full"
                error={alterFieldErrors.alter}
              >
                <AlterRowsEditor
                  rows={alterRows}
                  columns={columns}
                  onChange={(next) => {
                    setAlterRows(next)
                    if (alterFieldErrors.alter) {
                      setAlterFieldErrors((current) => ({
                        ...current,
                        alter: undefined,
                      }))
                    }
                  }}
                />
              </FormField>
            </DialogBody>
            {alterError ? (
              <p className="text-sm text-destructive">{alterError}</p>
            ) : null}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setAlterOpen(false)}
              >
                {t("action.cancel")}
              </Button>
              <Button type="submit" disabled={alterPending}>
                {alterPending ? t("action.saving") : t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={dropOpen}
        onOpenChange={(open) => {
          setDropOpen(open)
          if (!open) {
            setDropConfirm("")
            setDropIfExists(false)
            setDropError("")
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("sqlLab.dropTitle")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("sqlLab.dropDescription", { table: selectedTable })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <FormField
            label={t("sqlLab.dropConfirmLabel", { table: selectedTable })}
            htmlFor="sql-drop-confirm"
            error={
              dropConfirm && dropConfirm !== selectedTable
                ? t("sqlLab.dropMismatch")
                : undefined
            }
          >
            <Input
              id="sql-drop-confirm"
              value={dropConfirm}
              placeholder={t("sqlLab.dropConfirmPlaceholder")}
              onChange={(event) => {
                setDropConfirm(event.target.value)
                if (dropError) {
                  setDropError("")
                }
              }}
            />
          </FormField>
          <label className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={dropIfExists}
              onCheckedChange={(value) => setDropIfExists(value === true)}
            />
            <span>{t("sqlLab.dropIfExists")}</span>
          </label>
          {dropError ? (
            <p className="text-sm text-destructive">{dropError}</p>
          ) : null}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={dropPending}>
              {t("action.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={dropPending || dropConfirm !== selectedTable}
              onClick={(event) => {
                event.preventDefault()
                void submitDrop()
              }}
            >
              {dropPending ? t("action.saving") : t("action.delete")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Dialog
        open={auditDetail !== null}
        onOpenChange={(open) => {
          if (!open) {
            setAuditDetail(null)
          }
        }}
      >
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{t("sqlLab.auditDetailTitle")}</DialogTitle>
          </DialogHeader>
          {auditDetail ? (
            <DialogBody className="grid gap-4">
              <div className="grid gap-2">
                <span className="text-xs font-medium text-muted-foreground">
                  {t("sqlLab.auditFullStatement")}
                </span>
                <CodeBlock
                  code={auditDetail.statement}
                  language="text"
                  copyable
                  wrap
                  maxHeightClass="max-h-60"
                />
              </div>
              {auditDetail.errorMessage ? (
                <div className="grid gap-1">
                  <span className="text-xs font-medium text-muted-foreground">
                    {t("sqlLab.auditError")}
                  </span>
                  <p className="text-sm break-words whitespace-pre-wrap text-destructive">
                    {auditDetail.errorMessage}
                  </p>
                </div>
              ) : null}
            </DialogBody>
          ) : null}
        </DialogContent>
      </Dialog>
    </div>
  )
}
