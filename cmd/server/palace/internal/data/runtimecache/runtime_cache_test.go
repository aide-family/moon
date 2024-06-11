package runtimecache_test

import (
	"context"
	"testing"

	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/pkg/helper/model"
	"github.com/aide-family/moon/pkg/types"
	"github.com/aide-family/moon/pkg/vobj"
)

func TestNewRuntimeCache(t *testing.T) {
	cache := runtimecache.NewRuntimeCache()
	ctx := context.Background()
	cache.AppendUserTeamList(ctx, 1, []*model.SysTeam{
		{
			ID:        1,
			DeletedAt: 0,
			Name:      "test1",
			Status:    vobj.StatusEnable,
			Remark:    "test1 remark",
			Logo:      "",
			LeaderID:  1,
			CreatorID: 1,
			UUID:      "",
			Leader:    nil,
			Creator:   nil,
		},
	})

	t.Log(cache.GetUserTeamList(ctx, 1))

	cache.AppendTeamAdminList(ctx, 1, []*model.SysUser{
		{
			ID:        1,
			CreatedAt: types.Time{},
			UpdatedAt: types.Time{},
			DeletedAt: 0,
			Username:  "",
			Nickname:  "",
			Password:  "",
			Email:     "",
			Phone:     "",
			Remark:    "",
			Avatar:    "",
			Salt:      "",
			Gender:    0,
			Role:      0,
			Status:    0,
		},
	})
	t.Log(cache.GetTeamAdminList(ctx, 1))
}
