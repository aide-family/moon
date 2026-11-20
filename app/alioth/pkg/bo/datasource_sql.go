package bo

import "time"

// SqlTable 是 PostgreSQL 数据源里的一张表或视图。
type SqlTable struct {
	Name string
	// Type 取值为 TABLE 或 VIEW。
	Type string
}

// SqlColumn 是表的一个字段。长度、精度和小数位只对相应类型有效，其它类型为 0。
type SqlColumn struct {
	Name             string
	DataType         string
	UDTName          string
	Nullable         bool
	DefaultValue     string
	Length           int
	NumericPrecision int
	NumericScale     int
	PrimaryKey       bool
	Unique           bool
	Position         int
}

// SqlIndex 是表上的一个索引。
type SqlIndex struct {
	Name       string
	Columns    []string
	Unique     bool
	Primary    bool
	Definition string
}

// SqlQueryResult 是一次只读查询的结果。单元格统一转成字符串，NULL 为空串。
type SqlQueryResult struct {
	Columns   []string
	Rows      [][]string
	ElapsedMS int64
	Truncated bool
}

// SqlTableRows 是表数据的一页。HasMore 表示还有下一页。
type SqlTableRows struct {
	Columns   []string
	Rows      [][]string
	HasMore   bool
	ElapsedMS int64
}

// SqlColumnDefinition 是建表或加字段时的一列。
type SqlColumnDefinition struct {
	Name string
	// Type 是 PostgreSQL 类型名，可带长度、精度和数组后缀。
	Type         string
	Nullable     bool
	DefaultValue string
	PrimaryKey   bool
	Unique       bool
	Comment      string
}

// SqlAlterColumn 是对一个已有字段的修改。字段留空表示这一项不改。
type SqlAlterColumn struct {
	Name string
	// Type 为空表示不改类型。
	Type string
	// Nullable 为 nil 表示不改可空性。
	Nullable *bool
	// DefaultValue 为空表示不改默认值。
	DefaultValue string
	// DropDefault 为真时清除默认值，忽略 DefaultValue。
	DropDefault bool
}

// SqlTableRef 定位一张表。Database 和 Schema 留空时分别使用连接串里的库和 public。
type SqlTableRef struct {
	Database string
	Schema   string
	Table    string
}

// CreateSqlTableInput 是建表请求。
type CreateSqlTableInput struct {
	SqlTableRef
	Columns     []SqlColumnDefinition
	IfNotExists bool
	Comment     string
}

// AlterSqlTableInput 是改表请求。RenameTo、AddColumns、DropColumns 和
// AlterColumns 至少有一项非空。
type AlterSqlTableInput struct {
	SqlTableRef
	RenameTo     string
	AddColumns   []SqlColumnDefinition
	DropColumns  []string
	AlterColumns []SqlAlterColumn
}

// DropSqlTableInput 是删表请求。
type DropSqlTableInput struct {
	SqlTableRef
	IfExists bool
}

// SqlAudit 是一条 DDL 审计。语句在用户数据库里执行，审计保留结果和操作人。
type SqlAudit struct {
	UID          string
	Action       string
	Database     string
	Schema       string
	Table        string
	Statement    string
	Status       string
	ErrorMessage string
	OperatorUID  string
	OperatorName string
	CreatedAt    time.Time
}
