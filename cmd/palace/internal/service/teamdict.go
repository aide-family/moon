package service

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/service/build"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

type TeamDictService struct {
	palace.UnimplementedTeamDictServer

	dictBiz *biz.Dict
}

func NewTeamDictService(dictBiz *biz.Dict) *TeamDictService {
	return &TeamDictService{
		dictBiz: dictBiz,
	}
}

func (s *TeamDictService) SaveTeamDict(ctx context.Context, req *palace.SaveTeamDictRequest) (*common.EmptyReply, error) {
	var params = &bo.SaveDictReq{
		DictID: req.GetDictId(),
		Key:    req.GetKey(),
		Value:  req.GetValue(),
		Status: vobj.GlobalStatusEnable,
		Type:   vobj.DictType(req.GetDictType()),
		Color:  req.GetColor(),
		Lang:   req.GetLang(),
	}
	if err := s.dictBiz.SaveDict(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "保存团队字典成功"}, nil
}

func (s *TeamDictService) UpdateTeamDictStatus(ctx context.Context, req *palace.UpdateTeamDictStatusRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateDictStatusReq{
		DictIds: req.GetDictIds(),
		Status:  vobj.GlobalStatus(req.GetStatus()),
	}
	if err := s.dictBiz.UpdateDictStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "更新团队字典状态成功"}, nil
}

func (s *TeamDictService) DeleteTeamDict(ctx context.Context, req *palace.DeleteTeamDictRequest) (*common.EmptyReply, error) {
	params := &bo.OperateOneDictReq{DictID: req.GetDictId()}
	if err := s.dictBiz.DeleteDict(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "删除团队字典成功"}, nil
}

func (s *TeamDictService) GetTeamDict(ctx context.Context, req *palace.GetTeamDictRequest) (*common.TeamDictItem, error) {
	params := &bo.OperateOneDictReq{DictID: req.GetDictId()}
	dict, err := s.dictBiz.GetDict(ctx, params)
	if err != nil {
		return nil, err
	}
	return build.ToDictItem(dict), nil
}

func (s *TeamDictService) ListTeamDict(ctx context.Context, req *palace.ListTeamDictRequest) (*palace.ListTeamDictReply, error) {
	params := &bo.ListDictReq{
		PaginationRequest: build.ToPaginationRequest(req.GetPagination()),
		DictTypes:         slices.Map(req.GetDictTypes(), func(dictType common.DictType) vobj.DictType { return vobj.DictType(dictType) }),
		Status:            vobj.GlobalStatus(req.GetStatus()),
		Keyword:           req.GetKeyword(),
		Langs:             req.GetLangs(),
	}
	listDictReply, err := s.dictBiz.ListDict(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.ListTeamDictReply{
		Pagination: build.ToPaginationReply(listDictReply.PaginationReply),
		Items:      build.ToDictItems(listDictReply.Items),
	}, nil
}
