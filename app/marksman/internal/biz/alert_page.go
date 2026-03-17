package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewAlertPage(
	alertPageRepo repository.AlertPage,
	helper *klog.Helper,
) *AlertPageBiz {
	return &AlertPageBiz{
		alertPageRepo: alertPageRepo,
		helper:        klog.NewHelper(klog.With(helper.Logger(), "biz", "alert_page")),
	}
}

type AlertPageBiz struct {
	alertPageRepo repository.AlertPage
	helper        *klog.Helper
}

func (b *AlertPageBiz) CreateAlertPage(ctx context.Context, req *bo.CreateAlertPageBo) (snowflake.ID, error) {
	uid, err := b.alertPageRepo.CreateAlertPage(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "create alert page failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("create alert page failed").WithCause(err)
	}
	return uid, nil
}

func (b *AlertPageBiz) UpdateAlertPage(ctx context.Context, req *bo.UpdateAlertPageBo) error {
	if err := b.alertPageRepo.UpdateAlertPage(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("alert page not found")
		}
		b.helper.Errorw("msg", "update alert page failed", "error", err, "uid", req.UID.Int64())
		return merr.ErrorInternalServer("update alert page failed").WithCause(err)
	}
	return nil
}

func (b *AlertPageBiz) DeleteAlertPage(ctx context.Context, uid snowflake.ID) error {
	if err := b.alertPageRepo.DeleteAlertPage(ctx, uid); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("alert page not found")
		}
		b.helper.Errorw("msg", "delete alert page failed", "error", err, "uid", uid.Int64())
		return merr.ErrorInternalServer("delete alert page failed").WithCause(err)
	}
	return nil
}

func (b *AlertPageBiz) GetAlertPage(ctx context.Context, uid snowflake.ID) (*bo.AlertPageItemBo, error) {
	item, err := b.alertPageRepo.GetAlertPage(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("alert page not found")
		}
		b.helper.Errorw("msg", "get alert page failed", "error", err, "uid", uid.Int64())
		return nil, merr.ErrorInternalServer("get alert page failed").WithCause(err)
	}
	return item, nil
}

func (b *AlertPageBiz) ListAlertPage(ctx context.Context, req *bo.ListAlertPageBo) (*bo.PageResponseBo[*bo.AlertPageItemBo], error) {
	result, err := b.alertPageRepo.ListAlertPage(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "list alert page failed", "error", err)
		return nil, merr.ErrorInternalServer("list alert page failed").WithCause(err)
	}
	return result, nil
}
