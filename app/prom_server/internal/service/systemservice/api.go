package systemservice

import (
	"context"
	"sort"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/server/system"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/vo"
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
	apiBo := &bo.ApiBO{
		Name:   req.GetName(),
		Path:   req.GetPath(),
		Method: req.GetMethod(),
		Remark: req.GetRemark(),
		Module: vo.Module(req.GetModule()),
		Domain: vo.Domain(req.GetDomain()),
	}

	apiBoList, err := s.apiBiz.CreateApi(ctx, apiBo)
	if err != nil {
		return nil, err
	}
	if len(apiBoList) > 0 {
		apiBo = apiBoList[0]
	}

	return &pb.CreateApiReply{
		Id: apiBo.Id,
	}, nil
}

func (s *ApiService) UpdateApi(ctx context.Context, req *pb.UpdateApiRequest) (*pb.UpdateApiReply, error) {
	apiBo := &bo.ApiBO{
		Id:     req.GetId(),
		Name:   req.GetName(),
		Path:   req.GetPath(),
		Method: req.GetMethod(),
		Remark: req.GetRemark(),
		Status: vo.Status(req.GetStatus()),
		Module: vo.Module(req.GetModule()),
		Domain: vo.Domain(req.GetDomain()),
	}
	apiBo, err := s.apiBiz.UpdateApiById(ctx, apiBo.Id, apiBo)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateApiReply{
		Id: apiBo.Id,
	}, nil
}

func (s *ApiService) DeleteApi(ctx context.Context, req *pb.DeleteApiRequest) (*pb.DeleteApiReply, error) {
	apiBo, err := s.apiBiz.GetApiById(ctx, req.GetId())
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
	apiBo, err := s.apiBiz.GetApiById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.GetApiReply{
		Detail: apiBo.ToApiV1(),
	}, nil
}

func (s *ApiService) ListApi(ctx context.Context, req *pb.ListApiRequest) (*pb.ListApiReply, error) {
	pgReq := req.GetPage()
	apiBoList, pgInfo, err := s.apiBiz.ListApi(ctx, &bo.ApiListApiReq{
		Keyword: req.GetKeyword(),
		Status:  vo.Status(req.GetStatus()),
		Curr:    pgReq.GetCurr(),
		Size:    pgReq.GetSize(),
	})
	if err != nil {
		return nil, err
	}
	list := make([]*api.ApiV1, 0, len(apiBoList))
	for _, apiBo := range apiBoList {
		list = append(list, apiBo.ToApiV1())
	}
	return &pb.ListApiReply{
		Page: &api.PageReply{
			Curr:  pgInfo.GetCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

func (s *ApiService) SelectApi(ctx context.Context, req *pb.SelectApiRequest) (*pb.SelectApiReply, error) {
	pgReq := req.GetPage()
	apiBoList, pgInfo, err := s.apiBiz.ListApi(ctx, &bo.ApiListApiReq{
		Keyword: req.GetKeyword(),
		Status:  vo.Status(req.GetStatus()),
		Curr:    pgReq.GetCurr(),
		Size:    pgReq.GetSize(),
	})
	if err != nil {
		return nil, err
	}
	list := make([]*api.ApiSelectV1, 0, len(apiBoList))
	for _, apiBo := range apiBoList {
		list = append(list, apiBo.ToApiSelectV1())
	}
	return &pb.SelectApiReply{
		Page: &api.PageReply{
			Curr:  pgInfo.GetCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

// EditApiStatus 编辑api状态
func (s *ApiService) EditApiStatus(ctx context.Context, req *pb.EditApiStatusRequest) (*pb.EditApiStatusReply, error) {
	if err := s.apiBiz.UpdateApiStatusById(ctx, vo.Status(req.GetStatus()), req.GetIds()); err != nil {
		return nil, err
	}

	return &pb.EditApiStatusReply{
		Ids: req.GetIds(),
	}, nil
}

// GetApiTree 获取api权限树
func (s *ApiService) GetApiTree(ctx context.Context, _ *pb.GetApiTreeRequest) (*pb.GetApiTreeReply, error) {
	apiBoList, err := s.apiBiz.ListAllApi(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*api.ApiTree, 0, len(apiBoList))
	domainMap := make(map[vo.Domain]map[vo.Module][]*bo.ApiBO)
	domainMapKeys := make(vo.DomainList, 0)
	moduleMapKeys := make(map[vo.Domain]vo.ModuleList)
	for _, apiBo := range apiBoList {
		if _, ok := domainMap[apiBo.Domain]; !ok {
			domainMap[apiBo.Domain] = make(map[vo.Module][]*bo.ApiBO)
			domainMapKeys = append(domainMapKeys, apiBo.Domain)
		}
		if _, ok := domainMap[apiBo.Domain][apiBo.Module]; !ok {
			domainMap[apiBo.Domain][apiBo.Module] = make([]*bo.ApiBO, 0)
			moduleMapKeys[apiBo.Domain] = append(moduleMapKeys[apiBo.Domain], apiBo.Module)
		}
		domainMap[apiBo.Domain][apiBo.Module] = append(domainMap[apiBo.Domain][apiBo.Module], apiBo)
	}

	sort.Sort(domainMapKeys)
	for domain, moduleKeys := range moduleMapKeys {
		sort.Sort(moduleKeys)
		moduleMapKeys[domain] = moduleKeys
	}

	for _, domain := range domainMapKeys {
		moduleMap := domainMap[domain]
		domainDetail := &api.ApiTree{
			Domain:       domain.Value(),
			Module:       make([]*api.Module, 0),
			DomainName:   domain.String(),
			DomainRemark: domain.Remark(),
		}
		moduleKeys := moduleMapKeys[domain]
		for _, module := range moduleKeys {
			apiItemList := moduleMap[module]
			moduleDetail := &api.Module{
				Module: module.Value(),
				Apis:   make([]*api.ApiSelectV1, 0),
				Name:   module.String(),
				Remark: module.Remark(),
			}
			for _, apiBo := range apiItemList {
				moduleDetail.Apis = append(moduleDetail.Apis, apiBo.ToApiSelectV1())
			}
			domainDetail.Module = append(domainDetail.Module, moduleDetail)
		}
		list = append(list, domainDetail)
	}

	return &pb.GetApiTreeReply{
		Tree: list,
	}, nil
}

// AuthorizeApi 授权api
func (s *ApiService) AuthorizeApi(ctx context.Context, req *pb.AuthorizeApiRequest) (*pb.AuthorizeApiReply, error) {
	return nil, nil
}
