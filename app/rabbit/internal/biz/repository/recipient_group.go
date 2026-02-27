package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

// RecipientGroup 收件组仓储，项目内部维护
type RecipientGroup interface {
	GetRecipientGroupByName(ctx context.Context, name string) (*bo.RecipientGroupItemBo, error)
	CreateRecipientGroup(ctx context.Context, req *bo.CreateRecipientGroupBo) (snowflake.ID, error)
	GetRecipientGroup(ctx context.Context, uid snowflake.ID) (*bo.RecipientGroupDetailBo, error)
	UpdateRecipientGroup(ctx context.Context, req *bo.UpdateRecipientGroupBo) error
	UpdateRecipientGroupStatus(ctx context.Context, req *bo.UpdateRecipientGroupStatusBo) error
	DeleteRecipientGroup(ctx context.Context, uid snowflake.ID) error
	ListRecipientGroup(ctx context.Context, req *bo.ListRecipientGroupBo) (*bo.PageResponseBo[*bo.RecipientGroupItemBo], error)
	SelectRecipientGroup(ctx context.Context, req *bo.SelectRecipientGroupBo) (*bo.SelectRecipientGroupBoResult, error)
}

type RecipientMember interface {
	CreateRecipientMember(ctx context.Context, members []*bo.RecipientMemberItemBo) error
}
