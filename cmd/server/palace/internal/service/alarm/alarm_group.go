package alarm

import (
	"context"
	"fmt"
	"strings"

	alarmyapi "github.com/aide-family/moon/api/admin/alarm"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/util/types"
)

// GroupService 告警管理服务
type GroupService struct {
	alarmyapi.UnimplementedAlarmServer
	alarmGroupBiz *biz.AlarmGroupBiz
}

// NewAlarmService 创建告警管理服务
func NewAlarmService(alarmGroupBiz *biz.AlarmGroupBiz) *GroupService {
	return &GroupService{
		alarmGroupBiz: alarmGroupBiz,
	}
}

// CreateAlarmGroup 创建告警组
func (s *GroupService) CreateAlarmGroup(ctx context.Context, req *alarmyapi.CreateAlarmGroupRequest) (*alarmyapi.CreateAlarmGroupReply, error) {
	// 校验通知人是否重复
	if has := types.SlicesHasDuplicates(req.GetNoticeUser(), func(request *alarmyapi.CreateNoticeUserRequest) string {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%d-", request.GetUserId()))
		return sb.String()
	}); has {
		return nil, merr.ErrorI18nAlarmNoticeRepeatErr(ctx)
	}
	param := build.NewBuilder().WithContext(ctx).
		AlarmGroupModule().
		WithAPICreateAlarmGroupRequest(req).
		ToBo()
	if _, err := s.alarmGroupBiz.CreateAlarmGroup(ctx, param); !types.IsNil(err) {
		return nil, err
	}
	return &alarmyapi.CreateAlarmGroupReply{}, nil
}

// DeleteAlarmGroup 删除告警组
func (s *GroupService) DeleteAlarmGroup(ctx context.Context, req *alarmyapi.DeleteAlarmGroupRequest) (*alarmyapi.DeleteAlarmGroupReply, error) {
	if err := s.alarmGroupBiz.DeleteAlarmGroup(ctx, req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &alarmyapi.DeleteAlarmGroupReply{}, nil
}

// ListAlarmGroup 获取告警组列表
func (s *GroupService) ListAlarmGroup(ctx context.Context, req *alarmyapi.ListAlarmGroupRequest) (*alarmyapi.ListAlarmGroupReply, error) {
	param := build.NewBuilder().WithContext(ctx).AlarmGroupModule().WithAPIQueryAlarmGroupListRequest(req).ToBo()
	alarmGroups, err := s.alarmGroupBiz.ListPage(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &alarmyapi.ListAlarmGroupReply{
		Pagination: build.NewPageBuilder(param.Page).ToAPI(),
		List: build.NewBuilder().
			WithContext(ctx).
			AlarmGroupModule().
			WithDosAlarmGroup(alarmGroups).
			ToAPIs(),
	}, nil
}

// GetAlarmGroup 获取告警组详细信息
func (s *GroupService) GetAlarmGroup(ctx context.Context, req *alarmyapi.GetAlarmGroupRequest) (*alarmyapi.GetAlarmGroupReply, error) {
	detail, err := s.alarmGroupBiz.GetAlarmGroupDetail(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &alarmyapi.GetAlarmGroupReply{Detail: build.NewBuilder().
		WithContext(ctx).
		AlarmGroupModule().
		WithDoAlarmGroup(detail).
		ToAPI(),
	}, nil
}

// UpdateAlarmGroup 更新告警组信息
func (s *GroupService) UpdateAlarmGroup(ctx context.Context, req *alarmyapi.UpdateAlarmGroupRequest) (*alarmyapi.UpdateAlarmGroupReply, error) {
	// 校验通知人是否重复
	if has := types.SlicesHasDuplicates(req.GetUpdate().GetNoticeUser(), func(request *alarmyapi.CreateNoticeUserRequest) string {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%d-", request.GetUserId()))
		return sb.String()
	}); has {
		return nil, merr.ErrorI18nAlarmNoticeRepeatErr(ctx)
	}
	param := build.NewBuilder().WithContext(ctx).
		AlarmGroupModule().
		WithAPIUpdateAlarmGroupRequest(req).
		ToBo()
	err := s.alarmGroupBiz.UpdateAlarmGroup(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &alarmyapi.UpdateAlarmGroupReply{}, nil
}

// UpdateAlarmGroupStatus 更新告警组状态
func (s *GroupService) UpdateAlarmGroupStatus(ctx context.Context, req *alarmyapi.UpdateAlarmGroupStatusRequest) (*alarmyapi.UpdateAlarmGroupStatusReply, error) {
	param := build.NewBuilder().WithContext(ctx).
		AlarmGroupModule().
		WithAPIUpdateStatusAlarmGroupRequest(req).
		ToBo()
	err := s.alarmGroupBiz.UpdateStatus(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &alarmyapi.UpdateAlarmGroupStatusReply{}, nil
}
