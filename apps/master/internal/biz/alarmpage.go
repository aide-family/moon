package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/api"
	"prometheus-manager/api/perrors"
	"prometheus-manager/api/prom"
	pb "prometheus-manager/api/prom/v1"

	"prometheus-manager/dal/model"

	"prometheus-manager/apps/master/internal/service"
)

type (
	IAlarmPageV1Repo interface {
		V1Repo
		CreateAlarmPage(ctx context.Context, m *model.PromAlarmPage) error
		UpdateAlarmPageById(ctx context.Context, id int32, m *model.PromAlarmPage) error
		UpdateAlarmPagesStatusByIds(ctx context.Context, ids []int32, status prom.Status) error
		DeleteAlarmPageById(ctx context.Context, id int32) error
		GetAlarmPageById(ctx context.Context, req *pb.GetAlarmPageRequest) (*model.PromAlarmPage, error)
		ListAlarmPage(ctx context.Context, req *pb.ListAlarmPageRequest) ([]*model.PromAlarmPage, int64, error)
	}

	AlarmPageLogic struct {
		logger *log.Helper
		repo   IAlarmPageV1Repo
	}
)

var _ service.IAlarmPageV1Logic = (*AlarmPageLogic)(nil)

func NewAlarmPageLogic(repo IAlarmPageV1Repo, logger log.Logger) *AlarmPageLogic {
	return &AlarmPageLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", alarmPageModuleName))}
}

// CreateAlarmPage 创建告警页面
//
//	ctx 上下文
//	req 请求参数
func (s *AlarmPageLogic) CreateAlarmPage(ctx context.Context, req *pb.CreateAlarmPageRequest) (*pb.CreateAlarmPageReply, error) {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageLogic.CreateAlarmPage")
	defer span.End()

	if err := s.repo.CreateAlarmPage(ctx, alarmPageToModel(req.GetAlarmPage())); err != nil {
		s.logger.WithContext(ctx).Errorf("CreateAlarmPage error: %v", err)
		return nil, perrors.ErrorLogicCreateAlertPageFailed("创建告警页面失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	return &pb.CreateAlarmPageReply{Response: &api.Response{Message: "创建告警页面成功"}}, nil
}

// UpdateAlarmPage 更新告警页面
//
//	ctx 上下文
//	req 请求参数
func (s *AlarmPageLogic) UpdateAlarmPage(ctx context.Context, req *pb.UpdateAlarmPageRequest) (*pb.UpdateAlarmPageReply, error) {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageLogic.UpdateAlarmPage")
	defer span.End()

	if err := s.repo.UpdateAlarmPageById(ctx, req.GetId(), alarmPageToModel(req.GetAlarmPage())); err != nil {
		s.logger.WithContext(ctx).Errorf("UpdateAlarmPage error: %v", err)
		return nil, perrors.ErrorLogicEditAlertPageFailed("更新告警页面失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	return &pb.UpdateAlarmPageReply{Response: &api.Response{Message: "更新告警页面成功"}}, nil
}

// UpdateAlarmPagesStatus 更新告警页面状态
//
//	ctx 上下文
//	req 请求参数
func (s *AlarmPageLogic) UpdateAlarmPagesStatus(ctx context.Context, req *pb.UpdateAlarmPagesStatusRequest) (*pb.UpdateAlarmPagesStatusReply, error) {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageLogic.UpdateAlarmPagesStatus")
	defer span.End()

	if err := s.repo.UpdateAlarmPagesStatusByIds(ctx, req.GetIds(), req.GetStatus()); err != nil {
		s.logger.WithContext(ctx).Errorf("UpdateAlarmPagesStatus error: %v", err)
		return nil, perrors.ErrorLogicEditAlertPageFailed("更新告警页面状态失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	return &pb.UpdateAlarmPagesStatusReply{Response: &api.Response{Message: "更新告警页面状态成功"}}, nil
}

// DeleteAlarmPage 删除告警页面
//
//	ctx 上下文
//	req 请求参数
func (s *AlarmPageLogic) DeleteAlarmPage(ctx context.Context, req *pb.DeleteAlarmPageRequest) (*pb.DeleteAlarmPageReply, error) {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageLogic.DeleteAlarmPage")
	defer span.End()

	if err := s.repo.DeleteAlarmPageById(ctx, req.GetId()); err != nil {
		s.logger.WithContext(ctx).Errorf("DeleteAlarmPage error: %v", err)
		return nil, perrors.ErrorLogicDeleteAlertPageFailed("删除告警页面失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	return &pb.DeleteAlarmPageReply{Response: &api.Response{Message: "删除告警页面成功"}}, nil
}

// GetAlarmPage 获取告警页面详情
//
//	ctx 上下文
//	req 请求参数
func (s *AlarmPageLogic) GetAlarmPage(ctx context.Context, req *pb.GetAlarmPageRequest) (*pb.GetAlarmPageReply, error) {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageLogic.GetAlarmPage")
	defer span.End()

	detail, err := s.repo.GetAlarmPageById(ctx, req)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("GetAlarmPage error: %v", err)
		if perrors.IsLogicDataNotFound(err) {
			return nil, perrors.ErrorClientNotFound("告警页面不存在").WithCause(err).WithMetadata(map[string]string{
				"req": req.String(),
			})
		}

		return nil, perrors.ErrorServerDatabaseError("获取告警页面失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	return &pb.GetAlarmPageReply{Response: &api.Response{Message: "获取告警页面成功"}, AlarmPage: alarmPageToProm(detail)}, nil
}

// ListAlarmPage 获取告警页面列表
//
//	ctx 上下文
//	req 请求参数
func (s *AlarmPageLogic) ListAlarmPage(ctx context.Context, req *pb.ListAlarmPageRequest) (*pb.ListAlarmPageReply, error) {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageLogic.ListAlarmPage")
	defer span.End()

	alarmPages, total, err := s.repo.ListAlarmPage(ctx, req)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("ListAlarmPage error: %v", err)
		return nil, perrors.ErrorServerDatabaseError("获取告警页面列表失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	list := make([]*prom.AlarmPageItem, 0, len(alarmPages))
	for _, alarmPage := range alarmPages {
		list = append(list, alarmPageToProm(alarmPage))
	}

	return &pb.ListAlarmPageReply{
		Response: &api.Response{Message: "获取告警页面列表成功"},
		Result: &api.ListQueryResult{
			Page: &api.PageReply{
				Current: req.GetQuery().GetPage().GetCurrent(),
				Size:    req.GetQuery().GetPage().GetSize(),
				Total:   total,
			},
			// TODO 暂时不返回fields
			Fields: nil,
		},
		AlarmPages: list,
	}, nil
}
