package repository

import (
	"context"
)

// System 系统管理模块
type System interface {
	// ResetTeam 重置团队
	ResetTeam(ctx context.Context, teamID uint32) error

	// RestoreData 还原数据
	RestoreData(ctx context.Context, teamID uint32) error

	// DeleteBackup 删除备份
	DeleteBackup(ctx context.Context, teamID uint32)
}
