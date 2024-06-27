package dict

import (
	"context"
	"github.com/aide-family/moon/api/admin"
	pb "github.com/aide-family/moon/api/admin/dict"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type Service struct {
	pb.UnimplementedDictServer

	dictBiz *biz.DictBiz
}

func NewDictService(dictBiz *biz.DictBiz) *Service {
	return &Service{
		dictBiz: dictBiz,
	}
}

func (s *Service) CreateDict(ctx context.Context, req *pb.CreateDictRequest) (*pb.CreateDictReply, error) {
	createParams := bo.CreateDictParams{
		Name:         req.GetName(),
		Value:        req.GetValue(),
		DictType:     vobj.DictType(req.GetDictType()),
		ColorType:    req.GetColorType(),
		CssClass:     req.GetCssClass(),
		Icon:         req.GetIcon(),
		ImageUrl:     req.GetImageUrl(),
		Status:       vobj.Status(req.GetStatus()),
		Remark:       req.GetRemark(),
		LanguageCode: req.GetLanguageCode(),
	}

	_, err := s.dictBiz.CreateDict(ctx, &createParams)
	if err != nil {
		return nil, err
	}
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.CreateDictReply{}, nil
}

func (s *Service) UpdateDict(ctx context.Context, req *pb.UpdateDictRequest) (*pb.UpdateDictReply, error) {
	data := req.GetData()
	createParams := bo.CreateDictParams{
		Name:         data.GetName(),
		Value:        data.GetValue(),
		DictType:     vobj.DictType(data.GetDictType()),
		ColorType:    data.GetColorType(),
		CssClass:     data.GetCssClass(),
		Icon:         data.GetIcon(),
		ImageUrl:     data.GetImageUrl(),
		Status:       vobj.Status(data.GetStatus()),
		Remark:       data.GetRemark(),
		LanguageCode: data.GetLanguageCode(),
	}

	updateParams := bo.UpdateDictParams{
		ID:          req.GetId(),
		UpdateParam: createParams,
	}
	if err := s.dictBiz.UpdateDict(ctx, &updateParams); !types.IsNil(err) {
		return nil, err
	}
	return &pb.UpdateDictReply{}, nil
}

func (s *Service) ListDict(ctx context.Context, req *pb.GetDictSelectListRequest) (*pb.ListDictReply, error) {
	queryParams := &bo.QueryDictListParams{
		Keyword:  req.GetKeyword(),
		Page:     types.NewPagination(req.GetPagination()),
		Status:   vobj.Status(req.GetStatus()),
		DictType: vobj.DictType(req.GetDictType()),
	}

	dictPage, err := s.dictBiz.ListDict(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.ListDictReply{
		Pagination: build.NewPageBuilder(queryParams.Page).ToApi(),
		List: types.SliceTo(dictPage, func(dict *model.SysDict) *admin.Dict {
			return build.NewDictBuild(dict).ToApi()
		}),
	}, nil
}

func (s *Service) BatchUpdateDictStatus(ctx context.Context, params *pb.BatchUpdateDictStatusRequest) (*pb.BatchUpdateDictStatusReply, error) {
	updateParams := bo.UpdateDictStatusParams{
		IDs:    params.GetIds(),
		Status: vobj.Status(params.Status),
	}

	err := s.dictBiz.UpdateDictStatusByIds(ctx, &updateParams)
	if err != nil {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return &pb.BatchUpdateDictStatusReply{}, nil
}

func (s *Service) DeleteDict(ctx context.Context, params *pb.DeleteDictRequest) (*pb.DeleteDictReply, error) {

	err := s.dictBiz.DeleteDictById(ctx, params.GetId())
	if err != nil {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return &pb.DeleteDictReply{}, nil
}

func (s *Service) GetDict(ctx context.Context, req *pb.GetDictRequest) (*pb.GetDictReply, error) {

	dictDO, err := s.dictBiz.GetDict(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	resDict := build.NewDictBuild(dictDO).ToApi()
	return &pb.GetDictReply{
		Dict: resDict,
	}, nil
}
