package runtimecache

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gen/field"
)

// ProviderSetRuntimeCache is runtime_cache providers.
var ProviderSetRuntimeCache = wire.NewSet(NewRuntimeCache)

var (
	runtimeCache          RuntimeCache
	runtimeCacheInitOnce  sync.Once
	runtimeCacheTimerOnce sync.Once
)

// runtimeCacheTimer 定时同步用户和团队基础信息缓存
func runtimeCacheTimer(d *data.Data) {
	runtimeCacheTimerOnce.Do(func() {
		syncBaseInfo(d)
		tick := time.NewTicker(1 * time.Minute)
		go func() {
			defer after.RecoverX()
			for {
				select {
				case <-d.Exit():
					tick.Stop()
					log.Debugw("stop", "runtimeCacheTimer")
					return
				case <-tick.C:
					syncBaseInfo(d)
				}
			}
		}()
	})
}

func syncBaseInfo(d *data.Data) {
	runtimeCacheEnv := runtimeCache.(*env)
	ctx := context.Background()
	// 获取所有的团队列表
	teamList, err := query.Use(d.GetMainDB(ctx)).SysTeam.Find()
	if err != nil {
		return
	}

	// 获取所有的团队管理员列表
	for _, teamItem := range teamList {
		runtimeCacheEnv.teamList[teamItem.ID] = teamItem
		db, dbErr := d.GetBizGormDB(teamItem.ID)
		if dbErr != nil {
			continue
		}
		teamMemberList, queryErr := bizquery.Use(db).SysTeamMember.WithContext(ctx).Preload(field.Associations).Find()
		if queryErr != nil {
			continue
		}
		if _, exist := runtimeCacheEnv.teamAdminList[teamItem.ID]; !exist {
			runtimeCacheEnv.teamAdminList[teamItem.ID] = make(map[uint32]*bizmodel.SysTeamMember)
		}
		for _, teamMemberItem := range teamMemberList {
			runtimeCacheEnv.teamAdminList[teamItem.ID][teamMemberItem.UserID] = teamMemberItem
			if _, ok := runtimeCacheEnv.userTeamList[teamMemberItem.UserID]; !ok {
				runtimeCacheEnv.userTeamList[teamMemberItem.UserID] = make(map[uint32]*model.SysTeam)
			}
			runtimeCacheEnv.userTeamList[teamMemberItem.UserID][teamItem.ID] = teamItem
			if _, ok := runtimeCacheEnv.teamAdminList[teamItem.ID]; !ok {
				runtimeCacheEnv.teamAdminList[teamItem.ID] = make(map[uint32]*bizmodel.SysTeamMember)
			}
			if teamMemberItem.Role.IsAdmin() {
				runtimeCacheEnv.teamAdminList[teamItem.ID][teamMemberItem.UserID] = teamMemberItem
			}
		}
	}
	// 获取所有人员
	userList, err := query.Use(d.GetMainDB(ctx)).SysUser.Find()
	if err != nil {
		return
	}
	for _, userItem := range userList {
		runtimeCacheEnv.userList[userItem.ID] = userItem
	}
}

// GetRuntimeCache 获取运行时缓存的环境变量
func GetRuntimeCache() RuntimeCache {
	return runtimeCache
}

func NewRuntimeCache(d *data.Data) RuntimeCache {
	runtimeCacheInitOnce.Do(func() {
		runtimeCache = &env{
			userTeamList:   make(map[uint32]map[uint32]*model.SysTeam),
			teamAdminList:  make(map[uint32]map[uint32]*bizmodel.SysTeamMember),
			teamList:       make(map[uint32]*model.SysTeam),
			userList:       make(map[uint32]*model.SysUser),
			teamMemberList: make(map[uint32]map[uint32]*bizmodel.SysTeamMember),
			lock:           sync.Mutex{},
		}
		runtimeCacheTimer(d)
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
		// RemoveUserTeamList 移除用户团队列表
		RemoveUserTeamList(ctx context.Context, userID uint32, teamIDs []uint32)
		// ClearUserTeamList 清空用户团队列表
		ClearUserTeamList(ctx context.Context, userID uint32)
		// GetTeamAdminList 获取团队管理员列表
		GetTeamAdminList(ctx context.Context, teamID uint32) []*bizmodel.SysTeamMember
		// AppendTeamAdminList 设置团队管理员列表
		AppendTeamAdminList(ctx context.Context, teamID uint32, admins []*bizmodel.SysTeamMember)
		// RemoveTeamAdminList 移除团队管理员列表
		RemoveTeamAdminList(ctx context.Context, teamID uint32, adminIDs []uint32)
		// ClearTeamAdminList 清空团队管理员列表
		ClearTeamAdminList(ctx context.Context, teamID uint32)

		// GetUser 获取用户信息
		GetUser(ctx context.Context, userID uint32) *model.SysUser
		// GetTeam 获取团队信息
		GetTeam(ctx context.Context, teamID uint32) *model.SysTeam

		// AppendUser 设置用户信息
		AppendUser(ctx context.Context, user *model.SysUser)
		// AppendTeam 设置团队信息
		AppendTeam(ctx context.Context, team *model.SysTeam)
	}

	env struct {
		// 用户团队列表
		userTeamList map[uint32]map[uint32]*model.SysTeam
		// 团队管理员列表
		teamAdminList map[uint32]map[uint32]*bizmodel.SysTeamMember

		// 团队列表
		teamList map[uint32]*model.SysTeam
		// 用户列表
		userList map[uint32]*model.SysUser
		// 团队人员列表
		teamMemberList map[uint32]map[uint32]*bizmodel.SysTeamMember

		lock sync.Mutex
	}
)

func (e *env) AppendUser(_ context.Context, user *model.SysUser) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if types.IsNil(e.userList) {
		e.userList = make(map[uint32]*model.SysUser)
	}
	e.userList[user.ID] = user
}

func (e *env) AppendTeam(_ context.Context, team *model.SysTeam) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if types.IsNil(e.teamList) {
		e.teamList = make(map[uint32]*model.SysTeam)
	}
	e.teamList[team.ID] = team
}

func (e *env) GetUser(_ context.Context, userID uint32) *model.SysUser {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.userList[userID]
}

func (e *env) GetTeam(_ context.Context, teamID uint32) *model.SysTeam {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.teamList[teamID]
}

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

func (e *env) RemoveUserTeamList(_ context.Context, userID uint32, teamIDs []uint32) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if _, ok := e.userTeamList[userID]; ok {
		for _, teamID := range teamIDs {
			delete(e.userTeamList[userID], teamID)
		}
	}
}

func (e *env) RemoveTeamAdminList(_ context.Context, teamID uint32, adminIDs []uint32) {
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
