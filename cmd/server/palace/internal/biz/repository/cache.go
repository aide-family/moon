package repository

import (
	"context"

	"github.com/aide-family/moon/pkg/palace/imodel"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/plugin/cache"
)

// Cache 缓存接口
type Cache interface {
	// Cacher 获取缓存实例
	Cacher() cache.ICacher

	// GetUser 获取用户信息
	GetUser(ctx context.Context, userID uint32) *model.SysUser
	// GetTeam 获取团队信息
	GetTeam(ctx context.Context, teamID uint32) *model.SysTeam

	// AppendUser 设置用户信息
	AppendUser(ctx context.Context, user *model.SysUser)
	// AppendTeam 设置团队信息
	AppendTeam(ctx context.Context, team *model.SysTeam)

	// GetUserTeamList 获取用户团队列表
	GetUserTeamList(ctx context.Context, userID uint32) []*model.SysTeam
	// SyncUserTeamList 同步用户团队列表
	SyncUserTeamList(ctx context.Context, userID uint32)

	GetUsers(ctx context.Context, userIDs []uint32) []*model.SysUser

	// AppendDict 设置字典信息
	AppendDict(ctx context.Context, dict imodel.IDict, isBiz bool)
	// AppendDictList 设置字典信息列表
	AppendDictList(ctx context.Context, dict []imodel.IDict, isBiz bool)
	// GetDict 获取字典信息
	GetDict(ctx context.Context, id uint32, isBiz bool) imodel.IDict
	// GetDictList 获取字典信息列表
	GetDictList(ctx context.Context, ids []uint32, isBiz bool) []imodel.IDict
}
