package dictservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/dict"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"
)

type Service struct {
	pb.UnimplementedDictServer

	log *log.Helper

	dictBiz *biz.DictBiz
}

func NewDictService(dictBiz *biz.DictBiz, logger log.Logger) *Service {
	return &Service{
		log:     log.NewHelper(log.With(logger, "module", "service.Service")),
		dictBiz: dictBiz,
	}
}

func (s *Service) CreateDict(ctx context.Context, req *pb.CreateDictRequest) (*pb.CreateDictReply, error) {
	dictBo := &bo.DictBO{
		Name:     req.GetName(),
		Category: valueobj.Category(req.GetCategory()),
		Remark:   req.GetRemark(),
		Color:    req.GetColor(),
	}
	newDict, err := s.dictBiz.CreateDict(ctx, dictBo)
	if err != nil {
		s.log.Errorf("create dict err: %v", err)
		return nil, err
	}
	return &pb.CreateDictReply{Id: newDict.Id}, nil
}

func (s *Service) UpdateDict(ctx context.Context, req *pb.UpdateDictRequest) (*pb.UpdateDictReply, error) {
	dictBo := &bo.DictBO{
		Id:       req.GetId(),
		Name:     req.GetName(),
		Category: valueobj.Category(req.GetCategory()),
		Remark:   req.GetRemark(),
		Color:    req.GetColor(),
		Status:   valueobj.Status(req.GetStatus()),
	}
	newDict, err := s.dictBiz.UpdateDict(ctx, dictBo)
	if err != nil {
		s.log.Errorf("update dict err: %v", err)
		return nil, err
	}

	return &pb.UpdateDictReply{Id: newDict.Id}, nil
}

func (s *Service) BatchUpdateDictStatus(ctx context.Context, req *pb.BatchUpdateDictStatusRequest) (*pb.BatchUpdateDictStatusReply, error) {
	list := slices.To[uint32, uint](req.GetIds(), func(u uint32) uint {
		return uint(u)
	})
	if err := s.dictBiz.BatchUpdateDictStatus(ctx, req.GetStatus(), list); err != nil {
		s.log.Errorf("batch update dict status err: %v", err)
		return nil, err
	}
	return &pb.BatchUpdateDictStatusReply{Ids: req.GetIds()}, nil
}

func (s *Service) DeleteDict(ctx context.Context, req *pb.DeleteDictRequest) (*pb.DeleteDictReply, error) {
	if err := s.dictBiz.DeleteDictByIds(ctx, uint(req.GetId())); err != nil {
		s.log.Errorf("delete dict err: %v", err)
		return nil, err
	}
	return &pb.DeleteDictReply{Id: req.GetId()}, nil
}

func (s *Service) BatchDeleteDict(ctx context.Context, req *pb.BatchDeleteDictRequest) (*pb.BatchDeleteDictReply, error) {
	ids := slices.To[uint32, uint](req.GetIds(), func(u uint32) uint {
		return uint(u)
	})
	if err := s.dictBiz.DeleteDictByIds(ctx, ids...); err != nil {
		s.log.Errorf("batch delete dict err: %v", err)
		return nil, err
	}
	return &pb.BatchDeleteDictReply{Ids: req.GetIds()}, nil
}

func (s *Service) GetDict(ctx context.Context, req *pb.GetDictRequest) (*pb.GetDictReply, error) {
	dictBo, err := s.dictBiz.GetDictById(ctx, uint(req.GetId()))
	if err != nil {
		s.log.Errorf("get dict err: %v", err)
		return nil, err
	}
	reply := &pb.GetDictReply{
		PromDict: dictBo.ToApiV1(),
	}
	return reply, nil
}

func (s *Service) ListDict(ctx context.Context, req *pb.ListDictRequest) (*pb.ListDictReply, error) {
	dictBoList, pgInfo, err := s.dictBiz.ListDict(ctx, req)
	if err != nil {
		s.log.Errorf("list dict err: %v", err)
		return nil, err
	}
	list := make([]*api.DictV1, 0, len(dictBoList))
	for _, dictBo := range dictBoList {
		list = append(list, dictBo.ToApiV1())
	}

	pg := req.GetPage()
	return &pb.ListDictReply{
		Page: &api.PageReply{
			Curr:  pg.GetCurr(),
			Size:  pg.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

func (s *Service) SelectDict(ctx context.Context, req *pb.SelectDictRequest) (*pb.SelectDictReply, error) {
	dictBoList, pgInfo, err := s.dictBiz.SelectDict(ctx, req)
	if err != nil {
		s.log.Errorf("select dict err: %v", err)
		return nil, err
	}
	list := make([]*api.DictSelectV1, 0, len(dictBoList))
	for _, dictBo := range dictBoList {
		list = append(list, dictBo.ToApiSelectV1())
	}
	pg := req.GetPage()
	return &pb.SelectDictReply{
		Page: &api.PageReply{
			Curr:  pg.GetCurr(),
			Size:  pg.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}
