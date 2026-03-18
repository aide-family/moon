package biz

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewAlertPage(
	alertPageRepo repository.AlertPage,
	userAlertPageRepo repository.UserAlertPage,
	helper *klog.Helper,
) *AlertPageBiz {
	return &AlertPageBiz{
		alertPageRepo:     alertPageRepo,
		userAlertPageRepo: userAlertPageRepo,
		helper:            klog.NewHelper(klog.With(helper.Logger(), "biz", "alert_page")),
	}
}

type AlertPageBiz struct {
	alertPageRepo     repository.AlertPage
	userAlertPageRepo repository.UserAlertPage
	helper            *klog.Helper
}

func (b *AlertPageBiz) CreateAlertPage(ctx context.Context, req *bo.CreateAlertPageBo) (snowflake.ID, error) {
	taken, err := b.alertPageRepo.AlertPageNameTaken(ctx, req.Name, 0)
	if err != nil {
		b.helper.Errorw("msg", "check alert page name taken failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("check alert page name failed").WithCause(err)
	}
	if taken {
		return 0, merr.ErrorParams("alert page name already exists, please use another name")
	}
	uid, err := b.alertPageRepo.CreateAlertPage(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "create alert page failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("create alert page failed").WithCause(err)
	}
	return uid, nil
}

func (b *AlertPageBiz) UpdateAlertPage(ctx context.Context, req *bo.UpdateAlertPageBo) error {
	taken, err := b.alertPageRepo.AlertPageNameTaken(ctx, req.Name, req.UID)
	if err != nil {
		b.helper.Errorw("msg", "check alert page name taken failed", "error", err, "name", req.Name)
		return merr.ErrorInternalServer("check alert page name failed").WithCause(err)
	}
	if taken {
		return merr.ErrorParams("alert page name already exists, please use another name")
	}
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

func (b *AlertPageBiz) ListUserAlertPages(ctx context.Context) ([]*bo.AlertPageItemBo, error) {
	userUID := contextx.GetUserUID(ctx)
	uids, err := b.userAlertPageRepo.GetUserAlertPageUIDs(ctx, userUID)
	if err != nil {
		b.helper.Errorw("msg", "get user alert page uids failed", "error", err)
		return nil, merr.ErrorInternalServer("get user alert pages failed").WithCause(err)
	}
	if len(uids) == 0 {
		return nil, nil
	}
	pages, err := b.alertPageRepo.GetAlertPagesByUIDs(ctx, uids)
	if err != nil {
		b.helper.Errorw("msg", "get alert pages by uids failed", "error", err)
		return nil, merr.ErrorInternalServer("get user alert pages failed").WithCause(err)
	}

	return pages, nil
}

func (b *AlertPageBiz) SaveUserAlertPages(ctx context.Context, alertPageUIDs []snowflake.ID) error {
	userUID := contextx.GetUserUID(ctx)
	if userUID.Int64() == 0 {
		return merr.ErrorUnauthorized("user required")
	}
	if alertPagesTotal := len(alertPageUIDs); alertPagesTotal > 0 {
		alertPagesFound, err := b.alertPageRepo.CountAlertPagesByUIDs(ctx, alertPageUIDs)
		if err != nil {
			b.helper.Errorw("msg", "count alert pages by uids failed", "error", err)
			return merr.ErrorInternalServer("validate alert pages failed").WithCause(err)
		}
		if alertPagesFound != int64(alertPagesTotal) {
			return merr.ErrorParams("some alert pages not found or not in current, please check if the alert pages are in the current namespace")
		}
	}
	if err := b.userAlertPageRepo.SaveUserAlertPages(ctx, userUID, alertPageUIDs); err != nil {
		b.helper.Errorw("msg", "save user alert pages failed", "error", err)
		return merr.ErrorInternalServer("save user alert pages failed").WithCause(err)
	}
	return nil
}
