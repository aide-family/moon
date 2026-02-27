package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

type Member interface {
	CreateMember(ctx context.Context, req *bo.CreateMemberBo) error
	DeleteMember(ctx context.Context, uid snowflake.ID) error
	GetMember(ctx context.Context, uid snowflake.ID) (*bo.MemberItemBo, error)
	GetMemberByNamespaceAndUser(ctx context.Context, namespaceUID, userUID snowflake.ID) (*bo.MemberItemBo, error)
	ListMember(ctx context.Context, req *bo.ListMemberBo) (*bo.PageResponseBo[*bo.MemberItemBo], error)
	SelectMember(ctx context.Context, req *bo.SelectMemberBo) (*bo.SelectMemberBoResult, error)
	UpdateMemberStatus(ctx context.Context, uid snowflake.ID, status int32) error
	GetNamespaceUIDsByUserUID(ctx context.Context, userUID snowflake.ID) ([]snowflake.ID, error)
}
