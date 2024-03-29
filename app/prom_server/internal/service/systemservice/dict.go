package systemservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	"prometheus-manager/api/server/system"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
	"prometheus-manager/pkg/helper/middler"
	"prometheus-manager/pkg/util/slices"
)

type Service struct {
	system.UnimplementedDictServer

	log *log.Helper

	dictBiz *biz.DictBiz
	pageBiz *biz.AlarmPageBiz
}

func NewDictService(dictBiz *biz.DictBiz, pageBiz *biz.AlarmPageBiz, logger log.Logger) *Service {
	return &Service{
		log:     log.NewHelper(log.With(logger, "module", "service.Service")),
		dictBiz: dictBiz,
		pageBiz: pageBiz,
	}
}

func (s *Service) CreateDict(ctx context.Context, req *system.CreateDictRequest) (*system.CreateDictReply, error) {
	dictBo := &bo.DictBO{
		Name:     req.GetName(),
		Category: vobj.Category(req.GetCategory()),
		Remark:   req.GetRemark(),
		Color:    req.GetColor(),
	}
	newDict, err := s.dictBiz.CreateDict(ctx, dictBo)
	if err != nil {
		s.log.Errorf("create dict err: %v", err)
		return nil, err
	}
	return &system.CreateDictReply{Id: newDict.Id}, nil
}

func (s *Service) UpdateDict(ctx context.Context, req *system.UpdateDictRequest) (*system.UpdateDictReply, error) {
	dictBo := &bo.DictBO{
		Id:       req.GetId(),
		Name:     req.GetName(),
		Category: vobj.Category(req.GetCategory()),
		Remark:   req.GetRemark(),
		Color:    req.GetColor(),
		Status:   vobj.Status(req.GetStatus()),
	}
	newDict, err := s.dictBiz.UpdateDict(ctx, dictBo)
	if err != nil {
		s.log.Errorf("update dict err: %v", err)
		return nil, err
	}

	return &system.UpdateDictReply{Id: newDict.Id}, nil
}

func (s *Service) BatchUpdateDictStatus(ctx context.Context, req *system.BatchUpdateDictStatusRequest) (*system.BatchUpdateDictStatusReply, error) {
	if err := s.dictBiz.BatchUpdateDictStatus(ctx, vobj.Status(req.GetStatus()), req.GetIds()); err != nil {
		s.log.Errorf("batch update dict status err: %v", err)
		return nil, err
	}
	return &system.BatchUpdateDictStatusReply{Ids: req.GetIds()}, nil
}

func (s *Service) DeleteDict(ctx context.Context, req *system.DeleteDictRequest) (*system.DeleteDictReply, error) {
	if err := s.dictBiz.DeleteDictByIds(ctx, req.GetId()); err != nil {
		s.log.Errorf("delete dict err: %v", err)
		return nil, err
	}
	return &system.DeleteDictReply{Id: req.GetId()}, nil
}

func (s *Service) BatchDeleteDict(ctx context.Context, req *system.BatchDeleteDictRequest) (*system.BatchDeleteDictReply, error) {
	if err := s.dictBiz.DeleteDictByIds(ctx, req.GetIds()...); err != nil {
		s.log.Errorf("batch delete dict err: %v", err)
		return nil, err
	}
	return &system.BatchDeleteDictReply{Ids: req.GetIds()}, nil
}

func (s *Service) GetDict(ctx context.Context, req *system.GetDictRequest) (*system.GetDictReply, error) {
	dictBo, err := s.dictBiz.GetDictById(ctx, req.GetId())
	if err != nil {
		s.log.Errorf("get dict err: %v", err)
		return nil, err
	}
	reply := &system.GetDictReply{
		PromDict: dictBo.ToApiV1(),
	}
	return reply, nil
}

func (s *Service) ListDict(ctx context.Context, req *system.ListDictRequest) (*system.ListDictReply, error) {
	pageReq := req.GetPage()
	pgInfo := bo.NewPage(pageReq.GetCurr(), pageReq.GetSize())
	dictBoList, err := s.dictBiz.ListDict(ctx, &bo.ListDictRequest{
		Page:      pgInfo,
		Keyword:   req.GetKeyword(),
		Category:  vobj.Category(req.GetCategory()),
		Status:    vobj.Status(req.GetStatus()),
		IsDeleted: req.GetIsDeleted(),
	})
	if err != nil {
		s.log.Errorf("list dict err: %v", err)
		return nil, err
	}
	list := make([]*api.DictV1, 0, len(dictBoList))
	for _, dictBo := range dictBoList {
		list = append(list, dictBo.ToApiV1())
	}

	pg := req.GetPage()
	return &system.ListDictReply{
		Page: &api.PageReply{
			Curr:  pg.GetCurr(),
			Size:  pg.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

func (s *Service) SelectDict(ctx context.Context, req *system.SelectDictRequest) (*system.SelectDictReply, error) {
	pageReq := req.GetPage()
	pgInfo := bo.NewPage(pageReq.GetCurr(), pageReq.GetSize())
	dictBoList, err := s.dictBiz.ListDict(ctx, &bo.ListDictRequest{
		Page:      pgInfo,
		Keyword:   req.GetKeyword(),
		Category:  vobj.Category(req.GetCategory()),
		Status:    vobj.Status(req.GetStatus()),
		IsDeleted: req.GetIsDeleted(),
	})
	if err != nil {
		s.log.Errorf("select dict err: %v", err)
		return nil, err
	}
	list := make([]*api.DictSelectV1, 0, len(dictBoList))
	for _, dictBo := range dictBoList {
		list = append(list, dictBo.ToApiSelectV1())
	}
	pg := req.GetPage()
	return &system.SelectDictReply{
		Page: &api.PageReply{
			Curr:  pg.GetCurr(),
			Size:  pg.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

// CountAlarmPage 统计各告警页面的告警数量
func (s *Service) CountAlarmPage(ctx context.Context, req *system.CountAlarmPageRequest) (*system.CountAlarmPageReply, error) {
	count, err := s.pageBiz.CountAlarmPageByIds(ctx, req.GetIds()...)
	if err != nil {
		return nil, err
	}
	return &system.CountAlarmPageReply{
		AlarmCount: count,
	}, nil
}

// ListMyAlarmPage 获取我的告警页面列表
func (s *Service) ListMyAlarmPage(ctx context.Context, _ *system.ListMyAlarmPageRequest) (*system.ListMyAlarmPageReply, error) {
	userId := middler.GetUserId(ctx)
	userAlarmPages, err := s.pageBiz.GetUserAlarmPages(ctx, userId)
	if err != nil {
		return nil, err
	}
	return &system.ListMyAlarmPageReply{
		List: slices.To(userAlarmPages, func(alarmPageBO *bo.DictBO) *api.DictV1 {
			return alarmPageBO.ToApiV1()
		}),
	}, nil
}

// MyAlarmPagesConfig 我的告警页面列表配置
func (s *Service) MyAlarmPagesConfig(ctx context.Context, req *system.MyAlarmPagesConfigRequest) (*system.MyAlarmPagesConfigReply, error) {
	userId := middler.GetUserId(ctx)
	if err := s.pageBiz.BindUserPages(ctx, userId, req.GetAlarmIds()); err != nil {
		return nil, err
	}
	return &system.MyAlarmPagesConfigReply{}, nil
}
