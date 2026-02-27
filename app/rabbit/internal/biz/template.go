package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewTemplate(
	templateRepo repository.Template,
	helper *klog.Helper,
) *Template {
	return &Template{
		templateRepo: templateRepo,
		helper:       klog.NewHelper(klog.With(helper.Logger(), "biz", "template")),
	}
}

type Template struct {
	helper       *klog.Helper
	templateRepo repository.Template
}

func (t *Template) CreateTemplate(ctx context.Context, req *bo.CreateTemplateBo) (snowflake.ID, error) {
	if template, err := t.templateRepo.GetTemplateByName(ctx, req.Name); err == nil {
		return 0, merr.ErrorParams("template %s already exists, uid: %s", req.Name, template.UID)
	} else if !merr.IsNotFound(err) {
		t.helper.Errorw("msg", "check template exists failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("create template %s failed", req.Name).WithCause(err)
	}
	uid, err := t.templateRepo.CreateTemplate(ctx, req)
	if err != nil {
		t.helper.Errorw("msg", "create template failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("create template %s failed", req.Name).WithCause(err)
	}
	return uid, nil
}

func (t *Template) UpdateTemplate(ctx context.Context, req *bo.UpdateTemplateBo) error {
	existTemplate, err := t.templateRepo.GetTemplateByName(ctx, req.Name)
	if err != nil && !merr.IsNotFound(err) {
		t.helper.Errorw("msg", "check template exists failed", "error", err, "name", req.Name)
		return merr.ErrorInternalServer("update template %s failed", req.Name).WithCause(err)
	} else if existTemplate != nil && existTemplate.UID != req.UID {
		return merr.ErrorParams("template %s already exists", req.Name)
	}
	if err := t.templateRepo.UpdateTemplate(ctx, req); err != nil {
		t.helper.Errorw("msg", "update template failed", "error", err, "uid", req.UID)
		return merr.ErrorInternalServer("update template %s failed", req.UID).WithCause(err)
	}
	return nil
}

func (t *Template) UpdateTemplateStatus(ctx context.Context, req *bo.UpdateTemplateStatusBo) error {
	if err := t.templateRepo.UpdateTemplateStatus(ctx, req); err != nil {
		t.helper.Errorw("msg", "update template status failed", "error", err, "uid", req.UID)
		return merr.ErrorInternalServer("update template status %s failed", req.UID).WithCause(err)
	}
	return nil
}

func (t *Template) DeleteTemplate(ctx context.Context, uid snowflake.ID) error {
	if err := t.templateRepo.DeleteTemplate(ctx, uid); err != nil {
		t.helper.Errorw("msg", "delete template failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete template %s failed", uid).WithCause(err)
	}
	return nil
}

func (t *Template) GetTemplate(ctx context.Context, uid snowflake.ID) (*bo.TemplateItemBo, error) {
	templateBo, err := t.templateRepo.GetTemplate(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("template %s not found", uid)
		}
		t.helper.Errorw("msg", "get template failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get template %s failed", uid).WithCause(err)
	}
	return templateBo, nil
}

func (t *Template) ListTemplate(ctx context.Context, req *bo.ListTemplateBo) (*bo.PageResponseBo[*bo.TemplateItemBo], error) {
	pageResponseBo, err := t.templateRepo.ListTemplate(ctx, req)
	if err != nil {
		t.helper.Errorw("msg", "list template failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list template failed").WithCause(err)
	}
	return pageResponseBo, nil
}

func (t *Template) SelectTemplate(ctx context.Context, req *bo.SelectTemplateBo) (*bo.SelectTemplateBoResult, error) {
	result, err := t.templateRepo.SelectTemplate(ctx, req)
	if err != nil {
		t.helper.Errorw("msg", "select template failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("select template failed").WithCause(err)
	}
	return result, nil
}
