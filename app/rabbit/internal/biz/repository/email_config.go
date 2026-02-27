package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

type EmailConfig interface {
	CreateEmailConfig(ctx context.Context, req *bo.CreateEmailConfigBo) (snowflake.ID, error)
	UpdateEmailConfig(ctx context.Context, req *bo.UpdateEmailConfigBo) error
	UpdateEmailConfigStatus(ctx context.Context, req *bo.UpdateEmailConfigStatusBo) error
	DeleteEmailConfig(ctx context.Context, uid snowflake.ID) error
	GetEmailConfig(ctx context.Context, uid snowflake.ID) (*bo.EmailConfigItemBo, error)
	GetEmailConfigByName(ctx context.Context, name string) (*bo.EmailConfigItemBo, error)
	ListEmailConfig(ctx context.Context, req *bo.ListEmailConfigBo) (*bo.PageResponseBo[*bo.EmailConfigItemBo], error)
	SelectEmailConfig(ctx context.Context, req *bo.SelectEmailConfigBo) (*bo.SelectEmailConfigBoResult, error)
}
