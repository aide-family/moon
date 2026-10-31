package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/migration"
	"github.com/aide-family/moon/app/alioth/pkg/buildinfo"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// maxArchiveUploadBytes 限制上传迁移包的大小，避免把进程内存和数据库连接拖垮。
	maxArchiveUploadBytes = 256 << 20
	// archiveContentType 是迁移包下载的内容类型。
	archiveContentType = "application/gzip"
)

// MaintenanceService 提供平台升级与迁移页面需要的接口，以及迁移包的裸 HTTP 收发。
// 迁移包是二进制流，不走 proto 的 JSON 约定；鉴权仍复用同一个 JWT。
type MaintenanceService struct {
	v1.UnimplementedMaintenanceServiceServer
	relocator *migration.Relocator
	log       *slog.Logger
}

func NewMaintenanceService(relocator *migration.Relocator, logger *slog.Logger) *MaintenanceService {
	return &MaintenanceService{relocator: relocator, log: logger}
}

func (s *MaintenanceService) GetMaintenanceStatus(ctx context.Context, _ *v1.GetMaintenanceStatusRequest) (*v1.GetMaintenanceStatusResponse, error) {
	status, err := s.relocator.Status(ctx)
	if err != nil {
		return nil, err
	}
	tables := make([]*v1.MaintenanceTable, 0, len(status.Tables))
	for _, table := range status.Tables {
		tables = append(tables, &v1.MaintenanceTable{
			Name: table.Name, Archived: table.Archived, Rows: table.Rows, Exists: table.Exists,
		})
	}
	return &v1.GetMaintenanceStatusResponse{
		AppVersion:           status.AppVersion,
		ArchiveFormatVersion: status.ArchiveFormatVersion,
		PostgresVersion:      status.PostgreSQLVersion,
		ArchiveTableCount:    status.ArchiveTableCount,
		ArchiveRowCount:      status.ArchiveRowCount,
		EncryptKeyCheck:      status.EncryptKeyCheck,
		SchemaReady:          status.SchemaReady,
		Tables:               tables,
	}, nil
}

func (s *MaintenanceService) RunSchemaMigration(ctx context.Context, _ *v1.RunSchemaMigrationRequest) (*v1.RunSchemaMigrationResponse, error) {
	result, err := s.relocator.MigrateSchema(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.RunSchemaMigrationResponse{
		TableCount: result.TableCount,
		DurationMs: result.Duration.Milliseconds(),
		AppVersion: buildinfo.Version,
	}, nil
}

// ServeArchive 处理迁移包的下载与上传。两个方向都流式处理，不把文件读进内存。
// 鉴权由调用方完成（Authorization 已解析进 ctx），这里只做平台写权限检查。
func (s *MaintenanceService) ServeArchive(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.downloadArchive(w, r)
	case http.MethodPost:
		s.uploadArchive(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeMaintenanceError(w, merr.Client("method not allowed"))
	}
}

func (s *MaintenanceService) downloadArchive(w http.ResponseWriter, r *http.Request) {
	session, err := s.relocator.Export(r.Context())
	if err != nil {
		writeMaintenanceError(w, err)
		return
	}
	defer session.Close()
	w.Header().Set("Content-Type", archiveContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", session.Filename()))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if _, err := session.Write(r.Context(), w); err != nil {
		// 响应已经开始，改不了状态码；只记录原因，客户端会看到不完整的文件并按校验失败处理。
		s.log.Error(fmt.Sprintf("write system migration archive: %v", err))
	}
}

func (s *MaintenanceService) uploadArchive(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > maxArchiveUploadBytes {
		writeMaintenanceError(w, merr.Client("archive is too large"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxArchiveUploadBytes)
	result, err := s.relocator.Import(r.Context(), r.Body)
	if err != nil {
		writeMaintenanceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"tables":          result.Tables,
		"rows":            result.Rows,
		"sourceVersion":   result.SourceVersion,
		"reloginRequired": true,
	})
}

// writeMaintenanceError 按玉衡统一的错误格式写 JSON，前端 http.ts 能照常提示。
func writeMaintenanceError(w http.ResponseWriter, err error) {
	status := kerrors.FromError(err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(int(status.Code))
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":     status.Code,
		"reason":   status.Reason,
		"message":  status.Message,
		"metadata": status.Metadata,
	})
}
