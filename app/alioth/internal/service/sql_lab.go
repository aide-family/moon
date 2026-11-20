package service

import (
	"context"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/datasource"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// SqlLabService 提供 PostgreSQL 数据源的库表浏览、只读查询和结构变更。
type SqlLabService struct {
	v1.UnimplementedSqlLabServiceServer
	gateway *datasource.Gateway
}

func NewSqlLabService(gateway *datasource.Gateway) *SqlLabService {
	return &SqlLabService{gateway: gateway}
}

func (s *SqlLabService) ListSqlDatabases(ctx context.Context, req *v1.ListSqlDatabasesRequest) (*v1.ListSqlDatabasesResponse, error) {
	items, err := s.gateway.SqlDatabases(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.ListSqlDatabasesResponse{Items: items}, nil
}

func (s *SqlLabService) ListSqlSchemas(ctx context.Context, req *v1.ListSqlSchemasRequest) (*v1.ListSqlSchemasResponse, error) {
	items, err := s.gateway.SqlSchemas(ctx, req.GetUid(), req.GetDatabase())
	if err != nil {
		return nil, err
	}
	return &v1.ListSqlSchemasResponse{Items: items}, nil
}

func (s *SqlLabService) ListSqlTables(ctx context.Context, req *v1.ListSqlTablesRequest) (*v1.ListSqlTablesResponse, error) {
	tables, err := s.gateway.SqlTables(ctx, req.GetUid(), req.GetDatabase(), req.GetSchema())
	if err != nil {
		return nil, err
	}
	items := make([]*v1.SqlTableItem, 0, len(tables))
	for _, table := range tables {
		items = append(items, &v1.SqlTableItem{Name: table.Name, Type: table.Type})
	}
	return &v1.ListSqlTablesResponse{Items: items}, nil
}

func (s *SqlLabService) ListSqlTableColumns(ctx context.Context, req *v1.ListSqlTableColumnsRequest) (*v1.ListSqlTableColumnsResponse, error) {
	columns, err := s.gateway.SqlTableColumns(ctx, req.GetUid(), req.GetDatabase(), req.GetSchema(), req.GetTable())
	if err != nil {
		return nil, err
	}
	items := make([]*v1.SqlColumnItem, 0, len(columns))
	for _, column := range columns {
		items = append(items, &v1.SqlColumnItem{
			Name: column.Name, DataType: column.DataType, UdtName: column.UDTName,
			Nullable: column.Nullable, DefaultValue: column.DefaultValue,
			Length: int32(column.Length), NumericPrecision: int32(column.NumericPrecision),
			NumericScale: int32(column.NumericScale), PrimaryKey: column.PrimaryKey,
			Unique: column.Unique, Position: int32(column.Position),
		})
	}
	return &v1.ListSqlTableColumnsResponse{Items: items}, nil
}

func (s *SqlLabService) ListSqlTableIndexes(ctx context.Context, req *v1.ListSqlTableIndexesRequest) (*v1.ListSqlTableIndexesResponse, error) {
	indexes, err := s.gateway.SqlTableIndexes(ctx, req.GetUid(), req.GetDatabase(), req.GetSchema(), req.GetTable())
	if err != nil {
		return nil, err
	}
	items := make([]*v1.SqlIndexItem, 0, len(indexes))
	for _, index := range indexes {
		items = append(items, &v1.SqlIndexItem{
			Name: index.Name, Columns: index.Columns, Unique: index.Unique,
			Primary: index.Primary, Definition: index.Definition,
		})
	}
	return &v1.ListSqlTableIndexesResponse{Items: items}, nil
}

func (s *SqlLabService) QuerySql(ctx context.Context, req *v1.QuerySqlRequest) (*v1.QuerySqlResponse, error) {
	result, err := s.gateway.QuerySql(ctx, req.GetUid(), req.GetDatabase(), req.GetStatement(), req.GetMaxRows())
	if err != nil {
		return nil, err
	}
	rows := make([]*v1.SqlRow, 0, len(result.Rows))
	for _, row := range result.Rows {
		rows = append(rows, &v1.SqlRow{Cells: row})
	}
	return &v1.QuerySqlResponse{
		Columns: result.Columns, Rows: rows, RowCount: int32(len(result.Rows)),
		ElapsedMs: result.ElapsedMS, Truncated: result.Truncated,
	}, nil
}

func (s *SqlLabService) CreateSqlTable(ctx context.Context, req *v1.CreateSqlTableRequest) (*v1.CreateSqlTableResponse, error) {
	statement, err := s.gateway.CreateSqlTable(ctx, req.GetUid(), createSqlTableInput(req))
	if err != nil {
		return nil, err
	}
	return &v1.CreateSqlTableResponse{Statement: statement}, nil
}

func (s *SqlLabService) AlterSqlTable(ctx context.Context, req *v1.AlterSqlTableRequest) (*v1.AlterSqlTableResponse, error) {
	statements, err := s.gateway.AlterSqlTable(ctx, req.GetUid(), alterSqlTableInput(req))
	if err != nil {
		return nil, err
	}
	return &v1.AlterSqlTableResponse{Statements: statements}, nil
}

func (s *SqlLabService) DropSqlTable(ctx context.Context, req *v1.DropSqlTableRequest) (*v1.DropSqlTableResponse, error) {
	statement, err := s.gateway.DropSqlTable(ctx, req.GetUid(), bo.DropSqlTableInput{
		SqlTableRef: bo.SqlTableRef{Database: req.GetDatabase(), Schema: req.GetSchema(), Table: req.GetTable()},
		IfExists:    req.GetIfExists(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.DropSqlTableResponse{Statement: statement}, nil
}

func (s *SqlLabService) ListSqlAudits(ctx context.Context, req *v1.ListSqlAuditsRequest) (*v1.ListSqlAuditsResponse, error) {
	items, total, page, pageSize, err := s.gateway.SqlAudits(ctx, req.GetUid(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	audits := make([]*v1.SqlAuditItem, 0, len(items))
	for _, item := range items {
		audits = append(audits, &v1.SqlAuditItem{
			Uid: item.UID, Action: item.Action, Database: item.Database, Schema: item.Schema,
			Table: item.Table, Statement: item.Statement, Status: item.Status,
			ErrorMessage: item.ErrorMessage, OperatorUid: item.OperatorUID,
			OperatorName: item.OperatorName, CreatedAt: item.CreatedAt.Unix(),
		})
	}
	return &v1.ListSqlAuditsResponse{
		Items: audits, Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func createSqlTableInput(req *v1.CreateSqlTableRequest) bo.CreateSqlTableInput {
	return bo.CreateSqlTableInput{
		SqlTableRef: bo.SqlTableRef{Database: req.GetDatabase(), Schema: req.GetSchema(), Table: req.GetTable()},
		Columns:     sqlColumnDefinitions(req.GetColumns()),
		IfNotExists: req.GetIfNotExists(),
		Comment:     req.GetComment(),
	}
}

func alterSqlTableInput(req *v1.AlterSqlTableRequest) bo.AlterSqlTableInput {
	alterColumns := make([]bo.SqlAlterColumn, 0, len(req.GetAlterColumns()))
	for _, column := range req.GetAlterColumns() {
		var nullable *bool
		if column.Nullable != nil {
			value := column.GetNullable()
			nullable = &value
		}
		alterColumns = append(alterColumns, bo.SqlAlterColumn{
			Name: column.GetName(), Type: column.GetType(), Nullable: nullable,
			DefaultValue: column.GetDefaultValue(), DropDefault: column.GetDropDefault(),
		})
	}
	return bo.AlterSqlTableInput{
		SqlTableRef:  bo.SqlTableRef{Database: req.GetDatabase(), Schema: req.GetSchema(), Table: req.GetTable()},
		RenameTo:     req.GetRenameTo(),
		AddColumns:   sqlColumnDefinitions(req.GetAddColumns()),
		DropColumns:  req.GetDropColumns(),
		AlterColumns: alterColumns,
	}
}

func sqlColumnDefinitions(columns []*v1.SqlColumnDefinition) []bo.SqlColumnDefinition {
	out := make([]bo.SqlColumnDefinition, 0, len(columns))
	for _, column := range columns {
		out = append(out, bo.SqlColumnDefinition{
			Name: column.GetName(), Type: column.GetType(), Nullable: column.GetNullable(),
			DefaultValue: column.GetDefaultValue(), PrimaryKey: column.GetPrimaryKey(),
			Unique: column.GetUnique(), Comment: column.GetComment(),
		})
	}
	return out
}
