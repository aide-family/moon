package systemservice

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/system"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/valueobj"
)

type ApiService struct {
	pb.UnimplementedApiServer

	log    *log.Helper
	apiBiz *biz.ApiBiz
}

func NewApiService(apiBiz *biz.ApiBiz, logger log.Logger) *ApiService {
	return &ApiService{
		log:    log.NewHelper(log.With(logger, "module", "service.system.api")),
		apiBiz: apiBiz,
	}
}

func (s *ApiService) CreateApi(ctx context.Context, req *pb.CreateApiRequest) (*pb.CreateApiReply, error) {
	apiBo := &dobo.ApiBO{
		Name:   req.GetName(),
		Path:   req.GetPath(),
		Method: req.GetMethod(),
		Remark: req.GetRemark(),
	}

	apiBoList, err := s.apiBiz.CreateApi(ctx, apiBo)
	if err != nil {
		return nil, err
	}
	if len(apiBoList) > 0 {
		apiBo = apiBoList[0]
	}

	return &pb.CreateApiReply{
		Id: uint32(apiBo.Id),
	}, nil
}

func (s *ApiService) UpdateApi(ctx context.Context, req *pb.UpdateApiRequest) (*pb.UpdateApiReply, error) {
	apiBo := &dobo.ApiBO{
		Id:     uint(req.GetId()),
		Name:   req.GetName(),
		Path:   req.GetPath(),
		Method: req.GetMethod(),
		Remark: req.GetRemark(),
		Status: valueobj.Status(req.GetStatus()),
	}
	apiBo, err := s.apiBiz.UpdateApiById(ctx, apiBo.Id, apiBo)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateApiReply{
		Id: uint32(apiBo.Id),
	}, nil
}

func (s *ApiService) DeleteApi(ctx context.Context, req *pb.DeleteApiRequest) (*pb.DeleteApiReply, error) {
	apiBo, err := s.apiBiz.GetApiById(ctx, uint(req.GetId()))
	if err != nil {
		return nil, err
	}

	if err = s.apiBiz.DeleteApiById(ctx, apiBo.Id); err != nil {
		return nil, err
	}

	return &pb.DeleteApiReply{
		Id: req.GetId(),
	}, nil
}

func (s *ApiService) GetApi(ctx context.Context, req *pb.GetApiRequest) (*pb.GetApiReply, error) {
	apiBo, err := s.apiBiz.GetApiById(ctx, uint(req.GetId()))
	if err != nil {
		return nil, err
	}

	return &pb.GetApiReply{
		Detail: apiBo.ToV1(),
	}, nil
}

func (s *ApiService) ListApi(ctx context.Context, req *pb.ListApiRequest) (*pb.ListApiReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	apiBoList, err := s.apiBiz.ListApi(ctx, pgInfo)
	if err != nil {
		return nil, err
	}
	list := make([]*api.ApiV1, 0, len(apiBoList))
	for _, apiBo := range apiBoList {
		list = append(list, apiBo.ToV1())
	}
	return &pb.ListApiReply{
		Page: &api.PageReply{
			Curr:  pgReq.GetCurr(),
			Size:  pgReq.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

func (s *ApiService) SelectApi(ctx context.Context, req *pb.SelectApiRequest) (*pb.SelectApiReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	apiBoList, err := s.apiBiz.ListApi(ctx, pgInfo)
	if err != nil {
		return nil, err
	}
	list := make([]*api.ApiSelectV1, 0, len(apiBoList))
	for _, apiBo := range apiBoList {
		list = append(list, apiBo.ToSelectV1())
	}
	return &pb.SelectApiReply{
		Page: &api.PageReply{
			Curr:  pgReq.GetCurr(),
			Size:  pgReq.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}
