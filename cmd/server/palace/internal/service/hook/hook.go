package hook

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	hookapi "github.com/aide-family/moon/api/admin/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

// Service is a service that implements the HookServer interface.
type Service struct {
	hookapi.UnimplementedHookServer

	alarmHookBiz *biz.AlarmHookBiz
}

// NewHookService creates a new Service.
func NewHookService(alarmHook *biz.AlarmHookBiz) *Service {
	return &Service{alarmHookBiz: alarmHook}
}

// CreateHook creates a new hook.
func (s *Service) CreateHook(ctx context.Context, req *hookapi.CreateHookRequest) (*hookapi.CreateHookReply, error) {
	params := build.
		NewBuilder().
		HookModuleBuilder().
		WithContext(ctx).
		WithAPICreateAlarmHook(req).
		ToBo()
	_, err := s.alarmHookBiz.CreateAlarmHook(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.CreateHookReply{}, nil
}

// UpdateHook updates an existing hook.
func (s *Service) UpdateHook(ctx context.Context, req *hookapi.UpdateHookRequest) (*hookapi.UpdateHookReply, error) {
	params := build.NewBuilder().
		HookModuleBuilder().
		WithAPIUpdateAlarmHookRequest(req).
		ToBo()
	if err := s.alarmHookBiz.UpdateAlarmHook(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.UpdateHookReply{}, nil
}

// DeleteHook deletes an existing hook.
func (s *Service) DeleteHook(ctx context.Context, req *hookapi.DeleteHookRequest) (*hookapi.DeleteHookReply, error) {
	err := s.alarmHookBiz.DeleteAlarmHook(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.DeleteHookReply{}, nil
}

// GetHook gets an existing hook.
func (s *Service) GetHook(ctx context.Context, req *hookapi.GetHookRequest) (*hookapi.GetHookReply, error) {
	alarmHook, err := s.alarmHookBiz.GetAlarmHook(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.GetHookReply{Hook: build.
		NewBuilder().
		HookModuleBuilder().
		WithDoAlarmHook(alarmHook).
		ToAPI()}, nil
}

// ListHook lists all hooks.
func (s *Service) ListHook(ctx context.Context, req *hookapi.ListHookRequest) (*hookapi.ListHookReply, error) {
	params := build.NewBuilder().
		HookModuleBuilder().
		WithAPIQueryAlarmHookListRequest(req).
		ToBo()
	alarmHooks, err := s.alarmHookBiz.ListPage(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.ListHookReply{
		Hooks: types.SliceTo(alarmHooks, func(hook *bizmodel.AlarmHook) *admin.AlarmHookItem {
			return build.NewBuilder().HookModuleBuilder().WithDoAlarmHook(hook).ToAPI()
		}),
		Pagination: build.NewPageBuilder(params.Page).ToAPI(),
	}, nil
}

// UpdateHookStatus updates the status of an existing hook.
func (s *Service) UpdateHookStatus(ctx context.Context, req *hookapi.UpdateHookStatusRequest) (*hookapi.UpdateHookStatusReply, error) {
	params := build.NewBuilder().
		HookModuleBuilder().
		WithAPIUpdateAlarmHookStatusRequest(req).
		ToBo()
	if err := s.alarmHookBiz.UpdateStatus(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.UpdateHookStatusReply{}, nil
}

// ListHookSelectList 获取hook下拉列表
func (s *Service) ListHookSelectList(ctx context.Context, req *hookapi.ListHookRequest) (*hookapi.ListHookSelectListReply, error) {
	params := build.NewBuilder().
		HookModuleBuilder().
		WithAPIQueryAlarmHookListRequest(req).
		ToBo()
	alarmHooks, err := s.alarmHookBiz.ListPage(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.ListHookSelectListReply{
		List: types.SliceTo(alarmHooks, func(hook *bizmodel.AlarmHook) *admin.SelectItem {
			return build.NewBuilder().HookModuleBuilder().WithDoAlarmHook(hook).ToAPISelect()
		}),
	}, nil
}
