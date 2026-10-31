package repository

import (
	"context"
	"io"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// MaintenanceRepo 提供平台升级与迁移需要的数据操作：
// 读取环境状态、执行数据库迁移、导出与导入迁移包。
// 迁移包只覆盖配置与身份表，运行数据不进包，见 migratedTables 的注释。
type MaintenanceRepo interface {
	// Status 读取数据库版本、配置表与运行表的行数。
	Status(ctx context.Context) (*bo.MaintenanceDatabase, error)
	// MigrateSchema 执行 ent 建表、seed 与周分区，可重复执行。
	MigrateSchema(ctx context.Context) (bo.SchemaMigration, error)
	// ExportArchive 以 gzip 流写出迁移包，appVersion 写进 manifest 供排查。
	ExportArchive(ctx context.Context, w io.Writer, appVersion string) (bo.ArchiveExport, error)
	// ImportArchive 读取迁移包并覆盖目标库的配置数据，成功后由调用方补跑 seed。
	ImportArchive(ctx context.Context, r io.Reader) (bo.ArchiveImport, error)
}
