package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz"
	"github.com/aide-family/rabbit/internal/biz/bo"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func NewRecipientGroupService(recipientGroupBiz *biz.RecipientGroup) *RecipientGroupService {
	return &RecipientGroupService{
		recipientGroupBiz: recipientGroupBiz,
	}
}

type RecipientGroupService struct {
	apiv1.UnimplementedRecipientGroupServiceServer
	recipientGroupBiz *biz.RecipientGroup
}

func (s *RecipientGroupService) CreateRecipientGroup(ctx context.Context, req *apiv1.CreateRecipientGroupRequest) (*apiv1.CreateRecipientGroupReply, error) {
	createBo := bo.NewCreateRecipientGroupBo(req)
	uid, err := s.recipientGroupBiz.CreateRecipientGroup(ctx, createBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.CreateRecipientGroupReply{Uid: uid.Int64()}, nil
}

func (s *RecipientGroupService) GetRecipientGroup(ctx context.Context, req *apiv1.GetRecipientGroupRequest) (*apiv1.RecipientGroupItem, error) {
	detail, err := s.recipientGroupBiz.GetRecipientGroup(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return detail.ToAPIV1RecipientGroupItemFromDetail(), nil
}

func (s *RecipientGroupService) UpdateRecipientGroup(ctx context.Context, req *apiv1.UpdateRecipientGroupRequest) (*apiv1.UpdateRecipientGroupReply, error) {
	updateBo := bo.NewUpdateRecipientGroupBo(req)
	if err := s.recipientGroupBiz.UpdateRecipientGroup(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateRecipientGroupReply{}, nil
}

func (s *RecipientGroupService) DeleteRecipientGroup(ctx context.Context, req *apiv1.DeleteRecipientGroupRequest) (*apiv1.DeleteRecipientGroupReply, error) {
	if err := s.recipientGroupBiz.DeleteRecipientGroup(ctx, snowflake.ParseInt64(req.Uid)); err != nil {
		return nil, err
	}
	return &apiv1.DeleteRecipientGroupReply{}, nil
}

func (s *RecipientGroupService) ListRecipientGroup(ctx context.Context, req *apiv1.ListRecipientGroupRequest) (*apiv1.ListRecipientGroupReply, error) {
	listBo := bo.NewListRecipientGroupBo(req)
	page, err := s.recipientGroupBiz.ListRecipientGroup(ctx, listBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListRecipientGroupReply(page), nil
}

func (s *RecipientGroupService) SelectRecipientGroup(ctx context.Context, req *apiv1.SelectRecipientGroupRequest) (*apiv1.SelectRecipientGroupReply, error) {
	selectBo := bo.NewSelectRecipientGroupBo(req)
	result, err := s.recipientGroupBiz.SelectRecipientGroup(ctx, selectBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SelectRecipientGroupReply(result.Items, result.Total, result.LastUID, req.Limit), nil
}

func (s *RecipientGroupService) UpdateRecipientGroupStatus(ctx context.Context, req *apiv1.UpdateRecipientGroupStatusRequest) (*apiv1.UpdateRecipientGroupStatusReply, error) {
	statusBo := bo.NewUpdateRecipientGroupStatusBo(req)
	if err := s.recipientGroupBiz.UpdateRecipientGroupStatus(ctx, statusBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateRecipientGroupStatusReply{}, nil
}
