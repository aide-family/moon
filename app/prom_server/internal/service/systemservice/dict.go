package systemservice

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/server/system"
	"github.com/aide-family/moon/app/prom_server/internal/biz"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/helper/middler"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/go-kratos/kratos/v2/log"
)

type Service struct {
	system.UnimplementedDictServer

	log *log.Helper

	pageBiz    *biz.AlarmPageBiz
	sysDictBiz *biz.SysDictBiz
}

func NewDictService(sysDictBiz *biz.SysDictBiz, pageBiz *biz.AlarmPageBiz, logger log.Logger) *Service {
	return &Service{
		log:        log.NewHelper(log.With(logger, "module", "service.Service")),
		pageBiz:    pageBiz,
		sysDictBiz: sysDictBiz,
	}
}

func (s *Service) CreateDict(ctx context.Context, req *system.CreateDictRequest) (*system.CreateDictReply, error) {
	dictBo := &bo.CreateSysDictBo{
		Name:     req.GetName(),
		Category: vobj.Category(req.GetCategory()),
		Status:   vobj.StatusEnabled,
		Remark:   req.GetRemark(),
		Color:    req.GetColor(),
	}
	newDict, err := s.sysDictBiz.CreateDict(ctx, dictBo)
	if err != nil {
		s.log.Errorf("create dict err: %v", err)
		return nil, err
	}
	return &system.CreateDictReply{Id: newDict.GetID()}, nil
}

func (s *Service) UpdateDict(ctx context.Context, req *system.UpdateDictRequest) (*system.UpdateDictReply, error) {
	dictBo := &bo.UpdateSysDictBo{
		ID:       req.GetId(),
		Name:     req.GetName(),
		Category: vobj.Category(req.GetCategory()),
		Remark:   req.GetRemark(),
		Color:    req.GetColor(),
		Status:   vobj.Status(req.GetStatus()),
	}
	newDict, err := s.sysDictBiz.UpdateDict(ctx, dictBo)
	if err != nil {
		s.log.Errorf("update dict err: %v", err)
		return nil, err
	}

	return &system.UpdateDictReply{Id: newDict.GetID()}, nil
}

func (s *Service) BatchUpdateDictStatus(ctx context.Context, req *system.BatchUpdateDictStatusRequest) (*system.BatchUpdateDictStatusReply, error) {
	if err := s.sysDictBiz.BatchUpdateDictStatus(ctx, vobj.Status(req.GetStatus()), req.GetIds()); err != nil {
		s.log.Errorf("batch update dict status err: %v", err)
		return nil, err
	}
	return &system.BatchUpdateDictStatusReply{Ids: req.GetIds()}, nil
}

func (s *Service) DeleteDict(ctx context.Context, req *system.DeleteDictRequest) (*system.DeleteDictReply, error) {
	if err := s.sysDictBiz.DeleteDictByIds(ctx, req.GetId()); err != nil {
		s.log.Errorf("delete dict err: %v", err)
		return nil, err
	}
	return &system.DeleteDictReply{Id: req.GetId()}, nil
}

func (s *Service) BatchDeleteDict(ctx context.Context, req *system.BatchDeleteDictRequest) (*system.BatchDeleteDictReply, error) {
	if err := s.sysDictBiz.DeleteDictByIds(ctx, req.GetIds()...); err != nil {
		s.log.Errorf("batch delete dict err: %v", err)
		return nil, err
	}
	return &system.BatchDeleteDictReply{Ids: req.GetIds()}, nil
}

func (s *Service) GetDict(ctx context.Context, req *system.GetDictRequest) (*system.GetDictReply, error) {
	dictDo, err := s.sysDictBiz.GetDictById(ctx, req.GetId())
	if err != nil {
		s.log.Errorf("get dict err: %v", err)
		return nil, err
	}
	reply := &system.GetDictReply{
		PromDict: dictDoToApiV1(dictDo),
	}
	return reply, nil
}

func dictDoToApiV1(dictDo *do.SysDict) *api.DictV1 {
	if pkg.IsNil(dictDo) {
		return nil
	}
	return &api.DictV1{
		Id:        dictDo.GetID(),
		Name:      dictDo.GetName(),
		Category:  api.Category(dictDo.GetCategory()),
		Color:     dictDo.GetColor(),
		Status:    api.Status(dictDo.GetStatus()),
		Remark:    dictDo.GetRemark(),
		CreatedAt: dictDo.GetCreatedAt().Unix(),
		UpdatedAt: dictDo.GetUpdatedAt().Unix(),
		DeletedAt: int64(dictDo.GetDeletedAt()),
	}
}

func buildPageApi(page bo.Pagination) *api.PageReply {
	if pkg.IsNil(page) {
		return nil
	}
	return &api.PageReply{
		Curr:  page.GetCurr(),
		Size:  page.GetSize(),
		Total: page.GetTotal(),
	}
}

func (s *Service) ListDict(ctx context.Context, req *system.ListDictRequest) (*system.ListDictReply, error) {
	pageReq := req.GetPage()
	pgInfo := bo.NewPage(pageReq.GetCurr(), pageReq.GetSize())
	dictDoList, err := s.sysDictBiz.ListDict(ctx, &bo.ListSysDictBo{
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
	list := slices.To(dictDoList, func(dictBo *do.SysDict) *api.DictV1 { return dictDoToApiV1(dictBo) })

	return &system.ListDictReply{
		Page: buildPageApi(pgInfo),
		List: list,
	}, nil
}

func dictDoToSelectApiV1(dictDo *do.SysDict) *api.DictSelectV1 {
	if pkg.IsNil(dictDo) {
		return nil
	}
	return &api.DictSelectV1{
		Value:     dictDo.GetID(),
		Label:     dictDo.GetName(),
		Category:  api.Category(dictDo.GetCategory()),
		Color:     dictDo.GetColor(),
		Status:    api.Status(dictDo.GetStatus()),
		Remark:    dictDo.GetRemark(),
		IsDeleted: dictDo.GetDeletedAt() != 0,
	}
}

func (s *Service) SelectDict(ctx context.Context, req *system.SelectDictRequest) (*system.SelectDictReply, error) {
	pageReq := req.GetPage()
	pgInfo := bo.NewPage(pageReq.GetCurr(), pageReq.GetSize())
	dictDoList, err := s.sysDictBiz.SelectDict(ctx, &bo.SelectSysDictBo{
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
	list := slices.To(dictDoList, func(dictDo *do.SysDict) *api.DictSelectV1 {
		return dictDoToSelectApiV1(dictDo)
	})

	return &system.SelectDictReply{
		Page: buildPageApi(pgInfo),
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
