package bo

import "time"

// MaintenanceDatabase 是升级与迁移页需要的数据库事实。
// 应用版本与迁移包格式版本由用例层补齐，存储层不关心构建信息。
type MaintenanceDatabase struct {
	// PostgreSQLVersion 是数据库版本号。
	PostgreSQLVersion string
	// EncryptKeyCheck 是 ENCRYPT_KEY 的 SHA-256 前 16 位。
	EncryptKeyCheck string
	// Tables 是全部业务表及其迁移归属。
	Tables []MaintenanceTable
}

// MaintenanceTable 是一张业务表。Archived 为 false 表示运行数据，不进入迁移包。
type MaintenanceTable struct {
	Name     string
	Archived bool
	Rows     int64
	Exists   bool
}

// MaintenanceStatus 是升级与迁移页的完整环境状态。
type MaintenanceStatus struct {
	AppVersion           string
	ArchiveFormatVersion int32
	PostgreSQLVersion    string
	EncryptKeyCheck      string
	ArchiveTableCount    int32
	ArchiveRowCount      int64
	SchemaReady          bool
	Tables               []MaintenanceTable
}

// SchemaMigration 是一次数据库迁移的结果。
type SchemaMigration struct {
	TableCount int32
	Duration   time.Duration
}

// ArchiveExport 是一次迁移包导出的统计。
type ArchiveExport struct {
	Tables int32
	Rows   int64
}

// ArchiveImport 是一次迁移包导入的统计。
type ArchiveImport struct {
	Tables        int32
	Rows          int64
	SourceVersion string
}
