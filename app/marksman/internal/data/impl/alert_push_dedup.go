package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/plugin/cache"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
)

const alertPushDedupTTL = time.Hour

func NewAlertPushDedupRepository(d *data.Data) repository.AlertPushDedup {
	return &alertPushDedupRepository{Data: d}
}

type alertPushDedupRepository struct {
	*data.Data
}

func (r *alertPushDedupRepository) ExistsFiringPush(ctx context.Context, namespaceUID snowflake.ID, fingerprint string) (bool, error) {
	if fingerprint == "" || namespaceUID.Int64() == 0 {
		return false, nil
	}
	return r.Cache().Exists(ctx, alertPushDedupKey(namespaceUID.Int64(), fingerprint))
}

func (r *alertPushDedupRepository) MarkFiringPush(ctx context.Context, namespaceUID snowflake.ID, fingerprint string) error {
	if fingerprint == "" || namespaceUID.Int64() == 0 {
		return nil
	}
	return r.Cache().Set(ctx, alertPushDedupKey(namespaceUID.Int64(), fingerprint), "1", alertPushDedupTTL)
}

func (r *alertPushDedupRepository) ClearFiringPush(ctx context.Context, namespaceUID snowflake.ID, fingerprint string) error {
	if fingerprint == "" || namespaceUID.Int64() == 0 {
		return nil
	}
	return r.Cache().Del(ctx, alertPushDedupKey(namespaceUID.Int64(), fingerprint))
}

func alertPushDedupKey(namespaceUID int64, fingerprint string) cache.K {
	return cache.NewKey("marksman", "alert", "push", "firing", namespaceUID, fingerprint)
}
