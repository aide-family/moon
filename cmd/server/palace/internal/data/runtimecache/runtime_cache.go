package runtimecache

import (
	"context"
	"sort"
	"sync"

	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/model"
	"github.com/aide-family/moon/pkg/helper/model/bizmodel"
	"github.com/aide-family/moon/pkg/helper/model/query"
	"github.com/google/wire"
)

// ProviderSetRuntimeCache is runtime_cache providers.
var ProviderSetRuntimeCache = wire.NewSet(NewRuntimeCache)

var (
	runtimeCache         RuntimeCache
	runtimeCacheInitOnce sync.Once
)

// GetRuntimeCache 获取运行时缓存的环境变量
func GetRuntimeCache() RuntimeCache {
	return runtimeCache
}

func NewRuntimeCache(d *data.Data) RuntimeCache {
	runtimeCacheInitOnce.Do(func() {
		runtimeCache = &env{
			userTeamList:  make(map[uint32]map[uint32]*model.SysTeam),
			teamAdminList: make(map[uint32]map[uint32]*bizmodel.SysTeamMember),
		}
		// 获取所有的团队列表
		teamList, err := query.Use(d.GetMainDB(context.Background())).SysTeam.Find()
		if err != nil {
			return
		}
		// TODO 根据真实数据映射
		runtimeCache.AppendUserTeamList(context.Background(), 1, teamList)
	})
	return runtimeCache
}

type (
	// RuntimeCache 运行时缓存的环境变量
	RuntimeCache interface {
		// GetUserTeamList 获取用户团队列表
		GetUserTeamList(ctx context.Context, userID uint32) []*model.SysTeam
		// AppendUserTeamList 设置用户团队列表
		AppendUserTeamList(ctx context.Context, userID uint32, teams []*model.SysTeam)
		// removeUserTeamList 移除用户团队列表
		removeUserTeamList(ctx context.Context, userID uint32, teamIDs []uint32)
		// ClearUserTeamList 清空用户团队列表
		ClearUserTeamList(ctx context.Context, userID uint32)
		// GetTeamAdminList 获取团队管理员列表
		GetTeamAdminList(ctx context.Context, teamID uint32) []*bizmodel.SysTeamMember
		// AppendTeamAdminList 设置团队管理员列表
		AppendTeamAdminList(ctx context.Context, teamID uint32, admins []*bizmodel.SysTeamMember)
		// removeTeamAdminList 移除团队管理员列表
		removeTeamAdminList(ctx context.Context, teamID uint32, adminIDs []uint32)
		// ClearTeamAdminList 清空团队管理员列表
		ClearTeamAdminList(ctx context.Context, teamID uint32)
	}

	env struct {
		// 用户团队列表
		userTeamList map[uint32]map[uint32]*model.SysTeam
		// 团队管理员列表
		teamAdminList map[uint32]map[uint32]*bizmodel.SysTeamMember

		lock sync.Mutex
	}
)

func (e *env) ClearUserTeamList(_ context.Context, userID uint32) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if _, ok := e.userTeamList[userID]; ok {
		delete(e.userTeamList, userID)
	}
}

func (e *env) ClearTeamAdminList(_ context.Context, teamID uint32) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if _, ok := e.teamAdminList[teamID]; ok {
		delete(e.teamAdminList, teamID)
	}
}

func (e *env) removeUserTeamList(_ context.Context, userID uint32, teamIDs []uint32) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if _, ok := e.userTeamList[userID]; ok {
		for _, teamID := range teamIDs {
			delete(e.userTeamList[userID], teamID)
		}
	}
}

func (e *env) removeTeamAdminList(_ context.Context, teamID uint32, adminIDs []uint32) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if _, ok := e.teamAdminList[teamID]; ok {
		for _, adminID := range adminIDs {
			delete(e.teamAdminList[teamID], adminID)
		}
	}
}

func (e *env) AppendUserTeamList(_ context.Context, userID uint32, teams []*model.SysTeam) {
	e.lock.Lock()
	defer e.lock.Unlock()
	for _, team := range teams {
		if _, ok := e.userTeamList[userID]; !ok {
			e.userTeamList[userID] = make(map[uint32]*model.SysTeam)
		}
		e.userTeamList[userID][team.ID] = team
	}
}

func (e *env) AppendTeamAdminList(_ context.Context, teamID uint32, admins []*bizmodel.SysTeamMember) {
	e.lock.Lock()
	defer e.lock.Unlock()
	for _, admin := range admins {
		if _, ok := e.teamAdminList[teamID]; !ok {
			e.teamAdminList[teamID] = make(map[uint32]*bizmodel.SysTeamMember)
		}
		e.teamAdminList[teamID][admin.ID] = admin
	}
}

func (e *env) GetUserTeamList(_ context.Context, userID uint32) []*model.SysTeam {
	e.lock.Lock()
	defer e.lock.Unlock()
	list := make([]*model.SysTeam, 0, len(e.userTeamList[userID]))
	for _, team := range e.userTeamList[userID] {
		list = append(list, team)
	}
	// 排序，ID倒序
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID > list[j].ID
	})
	return list
}

func (e *env) GetTeamAdminList(_ context.Context, teamID uint32) []*bizmodel.SysTeamMember {
	e.lock.Lock()
	defer e.lock.Unlock()
	list := make([]*bizmodel.SysTeamMember, 0, len(e.teamAdminList[teamID]))
	for _, admin := range e.teamAdminList[teamID] {
		list = append(list, admin)
	}
	// 排序，ID倒序
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID > list[j].ID
	})
	return list
}
