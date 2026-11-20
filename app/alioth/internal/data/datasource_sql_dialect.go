package data

import (
	"context"
	"database/sql"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// sqlDialect 收口 SQL 数据源在连接串、标识符、内省语句和 DDL 语法上的差异。
// 库表浏览、只读查询、分页取数和审计由 datasourceRepo 统一调度，两种数据库共用。
type sqlDialect interface {
	// driverName 是 database/sql 的驱动名。
	driverName() string
	// dsn 由数据源连接串、用户名密码和目标库生成驱动连接串。
	dsn(item bo.Datasource, database string) (string, error)
	// schemaName 把空模式归一到可用的名字。PostgreSQL 用 public，
	// MySQL 里模式就是库，空值交给内省语句用当前库兜底。
	schemaName(schema, database string) string
	// quoteIdentifier 给标识符加引号。调用方必须先用 requireSqlName 校验。
	quoteIdentifier(name string) string
	// supportsDDLTransaction 表示表结构变更能否整体放进一个事务。
	// MySQL 的 DDL 会隐式提交，只能按语句顺序执行。
	supportsDDLTransaction() bool
	// databases、schemas 返回可访问的库与模式。
	databases(ctx context.Context, db *sql.DB) ([]string, error)
	schemas(ctx context.Context, db *sql.DB, database string) ([]string, error)
	tables(ctx context.Context, db *sql.DB, schema string) ([]bo.SqlTable, error)
	columns(ctx context.Context, db *sql.DB, schema, table string) ([]bo.SqlColumn, error)
	indexes(ctx context.Context, db *sql.DB, schema, table string) ([]bo.SqlIndex, error)
	primaryKeys(ctx context.Context, db *sql.DB, schema, table string) ([]string, error)
	// tableDataStatement 生成分页取数语句。limit 和 offset 由调用方计算。
	tableDataStatement(schema, table string, primaryKeys []string, limit, offset int) (string, error)
	// createStatements、alterStatements 生成结构变更语句。MySQL 改表需要
	// 读现有字段定义，所以带上连接。
	createStatements(ctx context.Context, db *sql.DB, in bo.CreateSqlTableInput, schema string) ([]string, error)
	alterStatements(ctx context.Context, db *sql.DB, in bo.AlterSqlTableInput, schema string) ([]string, error)
	dropStatement(in bo.DropSqlTableInput, schema string) (string, error)
	// validateType、validateDefault 校验 DDL 里的类型与默认值写法。
	validateType(raw string) (string, error)
	validateDefault(raw string) (string, error)
}

// sqlDialectFor 按驱动返回对应的 SQL 方言；不是 SQL 驱动时返回明确错误。
func sqlDialectFor(driver string) (sqlDialect, error) {
	switch driver {
	case driverSQLPostgresql:
		return postgresDialect{}, nil
	case driverSQLMysql:
		return mysqlDialect{}, nil
	default:
		return nil, merr.Invalid("unsupported datasource driver")
	}
}

// isSqlDriver 判断驱动是否由 SQL 实验室处理。
func isSqlDriver(driver string) bool {
	switch driver {
	case driverSQLPostgresql, driverSQLMysql:
		return true
	default:
		return false
	}
}
