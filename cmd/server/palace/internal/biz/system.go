package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/log"
)

func NewSystemBiz(systemRepository repository.System, teamRepository repository.Team) *SystemBiz {
	return &SystemBiz{
		systemRepository: systemRepository,
		teamRepository:   teamRepository,
	}
}

// SystemBiz .
type SystemBiz struct {
	systemRepository repository.System
	teamRepository   repository.Team
}

// ResetTeam 重置团队数据信息
func (s *SystemBiz) ResetTeam(ctx context.Context, teamID uint32) (err error) {
	// 还原数据(如果有)
	if err = s.systemRepository.RestoreData(ctx, teamID); !types.IsNil(err) {
		return err
	}
	defer func() {
		if !types.IsNil(err) {
			log.Error(err)
			err = merr.ErrorAlert("重置团队信息失败").
				WithCause(err).
				WithCause(s.systemRepository.RestoreData(ctx, teamID))
			return
		}
		// 删除备份数据
		s.systemRepository.DeleteBackup(ctx, teamID)
	}()
	// 获取团队基础信息
	team, err := s.teamRepository.GetTeamDetail(ctx, teamID)
	if !types.IsNil(err) {
		return err
	}
	// 获取团队成员列表
	members, err := s.teamRepository.MemberList(ctx, teamID)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}

	if err = s.systemRepository.ResetTeam(ctx, teamID); !types.IsNil(err) {
		return err
	}

	return s.teamRepository.SyncTeamBaseData(ctx, team, members)
}
