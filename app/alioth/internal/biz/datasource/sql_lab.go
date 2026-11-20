package datasource

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// sqlDriver 是 PostgreSQL 数据库数据源的驱动编码。SQL 实验室只服务这一种驱动，
// 其它 SQL 驱动接入时再按驱动拆分浏览和执行的实现。
const sqlDriver = "SQL_POSTGRESQL"

// SqlDatabases 列出数据源可以连接的库。
func (gateway *Gateway) SqlDatabases(ctx context.Context, uid string) ([]string, error) {
	item, err := gateway.sqlSource(ctx, uid, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	databases, err := gateway.datasources.SqlDatabases(ctx, *item)
	return databases, gateway.access.Wrap(err)
}

// SqlSchemas 列出库里的模式。
func (gateway *Gateway) SqlSchemas(ctx context.Context, uid, database string) ([]string, error) {
	item, err := gateway.sqlSource(ctx, uid, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	schemas, err := gateway.datasources.SqlSchemas(ctx, *item, database)
	return schemas, gateway.access.Wrap(err)
}

// SqlTables 列出模式里的表和视图。
func (gateway *Gateway) SqlTables(ctx context.Context, uid, database, schema string) ([]bo.SqlTable, error) {
	item, err := gateway.sqlSource(ctx, uid, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	tables, err := gateway.datasources.SqlTables(ctx, *item, database, schema)
	return tables, gateway.access.Wrap(err)
}

// SqlTableColumns 列出表的字段。
func (gateway *Gateway) SqlTableColumns(ctx context.Context, uid, database, schema, table string) ([]bo.SqlColumn, error) {
	item, err := gateway.sqlSource(ctx, uid, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	columns, err := gateway.datasources.SqlTableColumns(ctx, *item, database, schema, table)
	return columns, gateway.access.Wrap(err)
}

// SqlTableIndexes 列出表的索引。
func (gateway *Gateway) SqlTableIndexes(ctx context.Context, uid, database, schema, table string) ([]bo.SqlIndex, error) {
	item, err := gateway.sqlSource(ctx, uid, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	indexes, err := gateway.datasources.SqlTableIndexes(ctx, *item, database, schema, table)
	return indexes, gateway.access.Wrap(err)
}

// QuerySql 执行只读查询。查询能读到数据，因此只要求数据源读权限。
func (gateway *Gateway) QuerySql(ctx context.Context, uid, database, statement string, maxRows int32) (*bo.SqlQueryResult, error) {
	item, err := gateway.sqlSource(ctx, uid, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	if statement == "" {
		return nil, merr.Invalid("invalid request")
	}
	result, err := gateway.datasources.QuerySql(ctx, *item, database, statement, int(maxRows))
	return result, gateway.access.Wrap(err)
}

// SqlTableData 分页读取表数据。和只读查询一样只要求数据源读权限。
func (gateway *Gateway) SqlTableData(ctx context.Context, uid, database, schema, table string, pageNo, pageSize int32) (*bo.SqlTableRows, error) {
	item, err := gateway.sqlSource(ctx, uid, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	if table == "" {
		return nil, merr.Invalid("invalid request")
	}
	rows, err := gateway.datasources.SqlTableData(ctx, *item, database, schema, table, int(pageNo), int(pageSize))
	return rows, gateway.access.Wrap(err)
}

// CreateSqlTable 建表。结构变更要求数据源写权限，成功后由数据层写审计。
func (gateway *Gateway) CreateSqlTable(ctx context.Context, uid string, in bo.CreateSqlTableInput) (string, error) {
	item, err := gateway.sqlSource(ctx, uid, perm.DatasourceWrite)
	if err != nil {
		return "", err
	}
	if in.Table == "" || len(in.Columns) == 0 {
		return "", merr.Invalid("invalid request")
	}
	operatorUID, _ := actor.UserID(ctx)
	statement, err := gateway.datasources.CreateSqlTable(ctx, *item, in, operatorUID)
	return statement, gateway.access.Wrap(err)
}

// AlterSqlTable 改表。改名、加字段、删字段和改字段一次提交。
func (gateway *Gateway) AlterSqlTable(ctx context.Context, uid string, in bo.AlterSqlTableInput) ([]string, error) {
	item, err := gateway.sqlSource(ctx, uid, perm.DatasourceWrite)
	if err != nil {
		return nil, err
	}
	if in.Table == "" {
		return nil, merr.Invalid("invalid request")
	}
	if in.RenameTo == "" && len(in.AddColumns) == 0 && len(in.DropColumns) == 0 && len(in.AlterColumns) == 0 {
		return nil, merr.Invalid("nothing to alter")
	}
	operatorUID, _ := actor.UserID(ctx)
	statements, err := gateway.datasources.AlterSqlTable(ctx, *item, in, operatorUID)
	return statements, gateway.access.Wrap(err)
}

// DropSqlTable 删表。删表是不可逆的结构变更，同样要求写权限并记审计。
func (gateway *Gateway) DropSqlTable(ctx context.Context, uid string, in bo.DropSqlTableInput) (string, error) {
	item, err := gateway.sqlSource(ctx, uid, perm.DatasourceWrite)
	if err != nil {
		return "", err
	}
	if in.Table == "" {
		return "", merr.Invalid("invalid request")
	}
	operatorUID, _ := actor.UserID(ctx)
	statement, err := gateway.datasources.DropSqlTable(ctx, *item, in, operatorUID)
	return statement, gateway.access.Wrap(err)
}

// SqlAudits 返回某个数据源的 DDL 审计，最新的在前。
func (gateway *Gateway) SqlAudits(ctx context.Context, uid string, pageNo, pageSize int32) ([]bo.SqlAudit, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	item, err := gateway.datasources.GetDatasource(ctx, ns, uid)
	if err != nil {
		return nil, 0, 0, 0, gateway.access.Wrap(err)
	}
	// 套餐降级后不再开通的驱动，连审计也不返回。
	if err := gateway.requireDriverCapability(ctx, item.Driver); err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := gateway.datasources.ListSqlAudits(ctx, ns, item.UID, p, s)
	return items, total, p, s, gateway.access.Wrap(err)
}

// sqlSource 取 SQL 数据源：权限、驱动能力、启用状态和类型都要匹配。
// 读查询和结构变更的差别只在 permission 上。
func (gateway *Gateway) sqlSource(ctx context.Context, uid, permission string) (*bo.Datasource, error) {
	ns, err := gateway.access.Allow(ctx, permission)
	if err != nil {
		return nil, err
	}
	item, err := gateway.datasources.GetDatasource(ctx, ns, uid)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if err := gateway.requireDriverCapability(ctx, item.Driver); err != nil {
		return nil, err
	}
	if item.Status != "ACTIVE" {
		return nil, merr.Invalid("datasource is disabled")
	}
	if item.Type != "SQL" || item.Driver != sqlDriver {
		return nil, merr.Invalid("datasource is not sql")
	}
	return item, nil
}
