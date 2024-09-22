package hook

import (
	"context"

	hookapi "github.com/aide-family/moon/api/admin/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
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
	params := builder.NewParamsBuild().HookModuleBuilder().WithCreateHookRequest(req).ToBo()
	_, err := s.alarmHookBiz.CreateAlarmHook(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.CreateHookReply{}, nil
}

// UpdateHook updates an existing hook.
func (s *Service) UpdateHook(ctx context.Context, req *hookapi.UpdateHookRequest) (*hookapi.UpdateHookReply, error) {
	params := builder.NewParamsBuild().HookModuleBuilder().WithUpdateHookRequest(req).ToBo()
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
	return &hookapi.GetHookReply{
		Detail: builder.NewParamsBuild().WithContext(ctx).HookModuleBuilder().DoHookBuilder().ToAPI(alarmHook),
	}, nil
}

// ListHook lists all hooks.
func (s *Service) ListHook(ctx context.Context, req *hookapi.ListHookRequest) (*hookapi.ListHookReply, error) {
	params := builder.NewParamsBuild().HookModuleBuilder().WithListHookRequest(req).ToBo()
	alarmHooks, err := s.alarmHookBiz.ListPage(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.ListHookReply{
		List:       builder.NewParamsBuild().WithContext(ctx).HookModuleBuilder().DoHookBuilder().ToAPIs(alarmHooks),
		Pagination: builder.NewParamsBuild().PaginationModuleBuilder().ToAPI(params.Page),
	}, nil
}

// UpdateHookStatus updates the status of an existing hook.
func (s *Service) UpdateHookStatus(ctx context.Context, req *hookapi.UpdateHookStatusRequest) (*hookapi.UpdateHookStatusReply, error) {
	params := builder.NewParamsBuild().HookModuleBuilder().WithUpdateHookStatusRequest(req).ToBo()
	if err := s.alarmHookBiz.UpdateStatus(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.UpdateHookStatusReply{}, nil
}

// ListHookSelectList 获取hook下拉列表
func (s *Service) ListHookSelectList(ctx context.Context, req *hookapi.ListHookRequest) (*hookapi.ListHookSelectListReply, error) {
	params := builder.NewParamsBuild().HookModuleBuilder().WithListHookRequest(req).ToBo()
	alarmHooks, err := s.alarmHookBiz.ListPage(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.ListHookSelectListReply{
		List: builder.NewParamsBuild().WithContext(ctx).HookModuleBuilder().DoHookBuilder().ToSelects(alarmHooks),
	}, nil
}
