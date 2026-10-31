package migration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/buildinfo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/lock"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewRelocator)

const (
	// LockKey 让多副本之间同一时间只有一个数据库迁移或迁移包操作。
	LockKey = "alioth:lock:system:migration"
	// ArchiveFormatVersion 是迁移包格式版本，导入时严格比较。
	ArchiveFormatVersion = int32(1)
	// lockTTL 覆盖一次全量导出或导入；拿到锁后不续期，操作结束主动释放。
	lockTTL = 30 * time.Minute
)

// Relocator 负责平台升级与迁移：查看环境状态、执行数据库迁移、
// 导出与导入迁移包。读写权限在用例层判断，页面隐藏按钮不能代替这里的拒绝。
type Relocator struct {
	repo   repository.MaintenanceRepo
	access access.Checker
	locks  *lock.Client
	log    *slog.Logger
}

func NewRelocator(repo repository.MaintenanceRepo, checker access.Checker, locks *lock.Client, logger *slog.Logger) *Relocator {
	return &Relocator{repo: repo, access: checker, locks: locks, log: logger}
}

// Status 返回升级与迁移页的环境状态，只需要平台读权限。
func (relocator *Relocator) Status(ctx context.Context) (*bo.MaintenanceStatus, error) {
	if _, err := relocator.access.RequireSystem(ctx, perm.SystemMigrationRead); err != nil {
		return nil, err
	}
	database, err := relocator.repo.Status(ctx)
	if err != nil {
		return nil, err
	}
	status := &bo.MaintenanceStatus{
		AppVersion:           buildinfo.Version,
		ArchiveFormatVersion: ArchiveFormatVersion,
		PostgreSQLVersion:    database.PostgreSQLVersion,
		EncryptKeyCheck:      database.EncryptKeyCheck,
		SchemaReady:          true,
		Tables:               database.Tables,
	}
	for _, table := range database.Tables {
		if !table.Archived {
			continue
		}
		status.ArchiveTableCount++
		status.ArchiveRowCount += table.Rows
		if !table.Exists {
			status.SchemaReady = false
		}
	}
	return status, nil
}

// MigrateSchema 执行数据库迁移。迁移是幂等的，升级后补跑和失败重试都可以直接点。
func (relocator *Relocator) MigrateSchema(ctx context.Context) (bo.SchemaMigration, error) {
	account, err := relocator.access.RequireSystem(ctx, perm.SystemMigrationWrite)
	if err != nil {
		return bo.SchemaMigration{}, err
	}
	hold, err := relocator.acquire(ctx)
	if err != nil {
		return bo.SchemaMigration{}, err
	}
	defer func() { _ = hold.Release(context.Background()) }()
	result, err := relocator.repo.MigrateSchema(ctx)
	if err != nil {
		relocator.log.Warn(fmt.Sprintf("system migration failed: %v", err), "operator", account.UID)
		return bo.SchemaMigration{}, err
	}
	relocator.log.Info("system migration finished", "operator", account.UID, "tables", result.TableCount, "duration", result.Duration)
	return result, nil
}

// ExportSession 持有一把迁移锁和本次导出的文件名；Close 负责释放锁。
type ExportSession struct {
	relocator *Relocator
	hold      *lock.Hold
	operator  string
	filename  string
}

// Filename 返回下载时使用的文件名。
func (session *ExportSession) Filename() string {
	return session.filename
}

// Write 把迁移包流式写到 w。写失败时返回服务端错误，具体原因已经记录在进程日志里。
func (session *ExportSession) Write(ctx context.Context, w io.Writer) (bo.ArchiveExport, error) {
	result, err := session.relocator.repo.ExportArchive(ctx, w, buildinfo.Version)
	if err != nil {
		session.relocator.log.Warn(fmt.Sprintf("system migration export failed: %v", err), "operator", session.operator)
		return result, err
	}
	session.relocator.log.Info("system migration exported", "operator", session.operator, "tables", result.Tables, "rows", result.Rows)
	return result, nil
}

// Close 释放导出期间持有的迁移锁。
func (session *ExportSession) Close() {
	if session == nil || session.hold == nil {
		return
	}
	_ = session.hold.Release(context.Background())
	session.hold = nil
}

// Export 校验权限并拿到迁移锁，返回可以流式写出的导出会话。
func (relocator *Relocator) Export(ctx context.Context) (*ExportSession, error) {
	account, err := relocator.access.RequireSystem(ctx, perm.SystemMigrationWrite)
	if err != nil {
		return nil, err
	}
	hold, err := relocator.acquire(ctx)
	if err != nil {
		return nil, err
	}
	return &ExportSession{
		relocator: relocator,
		hold:      hold,
		operator:  account.UID,
		filename:  fmt.Sprintf("alioth-backup-%s-%s.alioth-backup", buildinfo.Version, time.Now().UTC().Format("20060102-150405")),
	}, nil
}

// Import 读取迁移包并覆盖目标库的配置数据；成功后 seed 已补齐新版本默认数据。
func (relocator *Relocator) Import(ctx context.Context, r io.Reader) (bo.ArchiveImport, error) {
	account, err := relocator.access.RequireSystem(ctx, perm.SystemMigrationWrite)
	if err != nil {
		return bo.ArchiveImport{}, err
	}
	hold, err := relocator.acquire(ctx)
	if err != nil {
		return bo.ArchiveImport{}, err
	}
	defer func() { _ = hold.Release(context.Background()) }()
	started := time.Now()
	result, err := relocator.repo.ImportArchive(ctx, r)
	if err != nil {
		relocator.log.Warn(fmt.Sprintf("system migration import failed: %v", err), "operator", account.UID)
		return result, err
	}
	relocator.log.Info("system migration imported", "operator", account.UID,
		"tables", result.Tables, "rows", result.Rows, "source", result.SourceVersion, "duration", time.Since(started))
	return result, nil
}

// acquire 拿迁移锁；另一个副本正在执行时返回 409，让人等而不是并发改库。
func (relocator *Relocator) acquire(ctx context.Context) (*lock.Hold, error) {
	if relocator.locks == nil {
		return nil, merr.Server("lock failed")
	}
	hold, err := relocator.locks.Acquire(ctx, LockKey, lockTTL)
	if err != nil {
		if errors.Is(err, lock.ErrNotAcquired) {
			return nil, merr.Conflict("migration is running")
		}
		return nil, err
	}
	return hold, nil
}
