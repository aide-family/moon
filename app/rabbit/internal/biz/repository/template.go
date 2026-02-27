package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

type Template interface {
	CreateTemplate(ctx context.Context, req *bo.CreateTemplateBo) (snowflake.ID, error)
	UpdateTemplate(ctx context.Context, req *bo.UpdateTemplateBo) error
	UpdateTemplateStatus(ctx context.Context, req *bo.UpdateTemplateStatusBo) error
	DeleteTemplate(ctx context.Context, uid snowflake.ID) error
	GetTemplate(ctx context.Context, uid snowflake.ID) (*bo.TemplateItemBo, error)
	GetTemplateByName(ctx context.Context, name string) (*bo.TemplateItemBo, error)
	ListTemplate(ctx context.Context, req *bo.ListTemplateBo) (*bo.PageResponseBo[*bo.TemplateItemBo], error)
	SelectTemplate(ctx context.Context, req *bo.SelectTemplateBo) (*bo.SelectTemplateBoResult, error)
}
