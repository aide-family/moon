package repoimpl

import (
	"context"
	"strconv"
	"time"

	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
)

const (
	// 缓存前缀
	cachePrefix = "palace"

	cacheKeyUser = "user"

	cacheKeyTeam = "team"

	cacheKeyUserTeam = "user_team"

	cacheKeyUserDict = "user_dict"
)

func userCacheKey(userID uint32) string {
	return types.TextJoin(cachePrefix, ":", cacheKeyUser, ":", strconv.Itoa(int(userID)))
}

func teamCacheKey(teamID uint32) string {
	return types.TextJoin(cachePrefix, ":", cacheKeyTeam, ":", strconv.Itoa(int(teamID)))
}

func userTeamCacheKey(teamID uint32) string {
	return types.TextJoin(cachePrefix, ":", cacheKeyUserTeam, ":", strconv.Itoa(int(teamID)))
}

func dictCacheKey(ctx context.Context, dictID uint32, isBiz bool) string {
	teamID := strconv.Itoa(int(middleware.GetTeamID(ctx)))
	if isBiz {
		return types.TextJoin(cachePrefix, ":", cacheKeyUserDict, ":", teamID, ":", "biz", ":", strconv.Itoa(int(dictID)))
	}
	return types.TextJoin(cachePrefix, ":", cacheKeyUserTeam, ":", teamID, ":", strconv.Itoa(int(dictID)))
}

// NewCacheRepository 创建缓存操作
func NewCacheRepository(data *data.Data) repository.Cache {
	return &cacheRepositoryImpl{data: data}
}

type cacheRepositoryImpl struct {
	data *data.Data
}

func (l *cacheRepositoryImpl) GetUser(ctx context.Context, userID uint32) *model.SysUser {
	var user *model.SysUser
	if err := l.data.GetCacher().GetObject(ctx, userCacheKey(userID), user); err != nil {
		userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
		user, err = userQuery.WithContext(ctx).Where(userQuery.ID.Eq(userID)).First()
		if err != nil {
			return new(model.SysUser)
		}
	}
	defer l.AppendUser(ctx, user)
	return user
}

func (l *cacheRepositoryImpl) GetTeam(ctx context.Context, teamID uint32) *model.SysTeam {
	var team *model.SysTeam
	if err := l.data.GetCacher().GetObject(ctx, teamCacheKey(teamID), team); err != nil {
		teamQuery := query.Use(l.data.GetMainDB(ctx)).SysTeam
		team, err = teamQuery.WithContext(ctx).Where(teamQuery.ID.Eq(teamID)).First()
		if err != nil {
			return new(model.SysTeam)
		}
	}
	defer l.AppendTeam(ctx, team)
	return team
}

func (l *cacheRepositoryImpl) AppendUser(ctx context.Context, user *model.SysUser) {
	_ = l.data.GetCacher().SetObject(ctx, userCacheKey(user.ID), user, 12*time.Hour)
}

func (l *cacheRepositoryImpl) AppendTeam(ctx context.Context, team *model.SysTeam) {
	_ = l.data.GetCacher().SetObject(ctx, teamCacheKey(team.ID), team, 12*time.Hour)
}

func (l *cacheRepositoryImpl) GetUserTeamList(ctx context.Context, userID uint32) []*model.SysTeam {
	var teamIDs []uint32
	teamIDsStr, err := l.data.GetCacher().Get(ctx, userTeamCacheKey(userID))
	if err != nil {
		return nil
	}

	_ = types.Unmarshal([]byte(teamIDsStr), &teamIDs)
	teamIds := make([]uint32, 0, len(teamIDs))
	list := make([]*model.SysTeam, 0, len(teamIDs))
	for _, teamID := range teamIDs {
		var team *model.SysTeam
		if err := l.data.GetCacher().GetObject(ctx, teamCacheKey(teamID), team); err != nil {
			teamIds = append(teamIds, teamID)
			continue
		}
		list = append(list, team)
	}
	if len(teamIds) > 0 {
		teamQuery := query.Use(l.data.GetMainDB(ctx)).SysTeam
		teamList, err := teamQuery.WithContext(ctx).Where(teamQuery.ID.In(teamIds...)).Find()
		if err == nil {
			list = append(list, teamList...)
		}
	}
	return list
}

func (l *cacheRepositoryImpl) SyncUserTeamList(ctx context.Context, userID uint32) {
	// 查询所有的团队
	teamQuery := query.Use(l.data.GetMainDB(ctx)).SysTeam
	teamList, err := teamQuery.WithContext(ctx).Find()
	if err != nil {
		return
	}
	teamIDs := make([]uint32, 0, len(teamList))
	for _, teamItem := range teamList {
		// 查询该用户是否在团队中
		bizQuery, err := getTeamIDBizQuery(l.data, teamItem.ID)
		if err != nil {
			continue
		}
		userTeamQuery := bizQuery.SysTeamMember
		_, err = userTeamQuery.WithContext(ctx).
			Where(userTeamQuery.UserID.Eq(userID)).First()
		if err != nil {
			continue
		}
		// 缓存用户团队列表
		teamIDs = append(teamIDs, teamItem.ID)
	}
	if len(teamIDs) > 0 {
		teamIdsStr, _ := types.Marshal(teamIDs)
		if err := l.data.GetCacher().Set(ctx, userTeamCacheKey(userID), string(teamIdsStr), 12*time.Hour); err != nil {
			log.Warnf("cache user %d team list failed: %s", userID, err)
		}
	}
}

func (l *cacheRepositoryImpl) GetUsers(ctx context.Context, userIDs []uint32) []*model.SysUser {
	users := make([]*model.SysUser, 0, len(userIDs))
	noExistIds := make([]uint32, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID <= 0 {
			continue
		}
		var user model.SysUser
		if err := l.data.GetCacher().GetObject(ctx, userCacheKey(userID), &user); err != nil {
			noExistIds = append(noExistIds, userID)
			continue
		}
		users = append(users, &user)
	}
	if len(noExistIds) > 0 {
		userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
		sysUsers, err := userQuery.WithContext(ctx).Where(userQuery.ID.In(noExistIds...)).Find()
		if err == nil {
			users = append(users, sysUsers...)
		}
	}
	for _, user := range users {
		l.AppendUser(ctx, user)
	}
	return users
}

func (l *cacheRepositoryImpl) AppendDict(ctx context.Context, dict imodel.IDict, isBiz bool) {
	_ = l.data.GetCacher().SetObject(ctx, dictCacheKey(ctx, dict.GetID(), isBiz), dict, 12*time.Hour)
}

func (l *cacheRepositoryImpl) AppendDictList(ctx context.Context, dict []imodel.IDict, isBiz bool) {
	for _, item := range dict {
		_ = l.data.GetCacher().SetObject(ctx, dictCacheKey(ctx, item.GetID(), isBiz), item, 12*time.Hour)
	}
}

func (l *cacheRepositoryImpl) GetDict(ctx context.Context, id uint32, isBiz bool) imodel.IDict {
	if isBiz {
		var dict bizmodel.SysDict
		if err := l.data.GetCacher().GetObject(ctx, dictCacheKey(ctx, id, isBiz), &dict); err != nil {
			log.Warnf("bizmodel.SysDict is nil")
			bizQuery, err := getBizQuery(ctx, l.data)
			if !types.IsNil(err) {
				return &dict
			}
			iDict, err := bizQuery.SysDict.WithContext(ctx).Where(bizQuery.SysDict.ID.Eq(id)).First()
			if err != nil {
				return &dict
			}
			return iDict
		}
		return &dict
	}
	var dict model.SysDict
	if err := l.data.GetCacher().GetObject(ctx, dictCacheKey(ctx, id, isBiz), &dict); err != nil {
		log.Warnf("model.SysDict is nil")
		mainQuery := query.Use(l.data.GetMainDB(ctx))
		sysDict, err := mainQuery.SysDict.WithContext(ctx).Where(mainQuery.SysDict.ID.Eq(id)).First()
		if err != nil {
			return &dict
		}
		return sysDict
	}
	return &dict
}

func (l *cacheRepositoryImpl) GetDictList(ctx context.Context, ids []uint32, isBiz bool) []imodel.IDict {
	dictList := make([]imodel.IDict, 0, len(ids))
	for _, id := range ids {
		dictList = append(dictList, l.GetDict(ctx, id, isBiz))
	}
	return dictList
}

// Cacher 获取缓存实例
func (l *cacheRepositoryImpl) Cacher() cache.ICacher {
	return l.data.GetCacher()
}
