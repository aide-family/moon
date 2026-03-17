package evaluator

import (
	"context"
	"fmt"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/server/cron"
	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data/impl/do"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"
)

func NewAlerting(
	alertEventUID snowflake.ID,
	info *bo.AlertEventBo,
	alertEventRepo repository.AlertEvent,
	alertingRepo repository.Alerting,
) cron.CronJob {
	return &alerting{
		alertEventUID:  alertEventUID,
		info:           info,
		alertEventRepo: alertEventRepo,
		alertingRepo:   alertingRepo,
	}
}

type alerting struct {
	alertEventUID  snowflake.ID
	info           *bo.AlertEventBo
	alertEventRepo repository.AlertEvent
	alertingRepo   repository.Alerting
}

// Index implements [cron.CronJob].
func (a *alerting) Index() string {
	return fmt.Sprintf("alerting--%d--%d--%s", a.info.NamespaceUID.Int64(), a.alertEventUID.Int64(), a.info.Fingerprint)
}

// IsImmediate implements [cron.CronJob].
func (a *alerting) IsImmediate() bool {
	return false
}

// Run implements [cron.CronJob].
func (a *alerting) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()
	ctx = contextx.WithNamespace(ctx, a.info.NamespaceUID)

	alertEvent, err := a.alertEventRepo.GetAlertEventByFingerprint(ctx, a.alertEventUID, a.info.Fingerprint)
	if err != nil {
		klog.Errorw("msg", "get alert event failed", "error", err, "alertEventUID", a.alertEventUID.Int64(), "fingerprint", a.info.Fingerprint)
		return
	}
	if alertEvent.Status != do.AlertEventStatusFiring {
		return
	}
	if err := a.alertEventRepo.AutoRecoverAlert(ctx, a.alertEventUID); err != nil {
		klog.Errorw("msg", "auto recover alert failed", "error", err, "alertEventUID", a.alertEventUID.Int64())
		return
	}
	a.alertingRepo.Remove(a.Index())
}

// Spec implements [cron.CronJob].
func (a *alerting) Spec() cron.CronSpec {
	return cron.CronSpecEvery(a.info.EvaluateDuration * 2)
}
