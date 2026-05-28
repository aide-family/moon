package biz

import (
	"context"
	"strings"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/plugin/cache"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

const alertNotifyDedupTTL = 7 * 24 * time.Hour

func (b *Alert) shouldDispatchAlertNotification(ctx context.Context, payload *bo.AlertPayloadBo, routeKey string) bool {
	if b.cache == nil || payload == nil || routeKey == "" {
		return true
	}
	fingerprint := alertDedupFingerprint(payload)
	if fingerprint == "" {
		return true
	}
	namespaceUID, ok := contextx.TryGetNamespace(ctx)
	if !ok {
		return true
	}
	if isResolvedAlertStatus(payload.Status) {
		firingKey := alertNotifyDedupKey(namespaceUID.Int64(), fingerprint, routeKey, "firing")
		if err := b.cache.Del(ctx, firingKey); err != nil {
			b.helper.WithContext(ctx).Warnw("msg", "clear firing alert dedup key failed", "error", err, "key", firingKey)
		}
		resolvedKey := alertNotifyDedupKey(namespaceUID.Int64(), fingerprint, routeKey, "resolved")
		exists, err := b.cache.Exists(ctx, resolvedKey)
		if err != nil {
			b.helper.WithContext(ctx).Warnw("msg", "check resolved alert dedup failed", "error", err, "key", resolvedKey)
			return true
		}
		return !exists
	}
	firingKey := alertNotifyDedupKey(namespaceUID.Int64(), fingerprint, routeKey, "firing")
	exists, err := b.cache.Exists(ctx, firingKey)
	if err != nil {
		b.helper.WithContext(ctx).Warnw("msg", "check firing alert dedup failed", "error", err, "key", firingKey)
		return true
	}
	if exists {
		b.helper.WithContext(ctx).Debugw(
			"msg", "skip duplicate firing alert notification",
			"fingerprint", fingerprint,
			"routeKey", routeKey,
			"status", payload.Status,
		)
	}
	return !exists
}

func (b *Alert) markAlertNotificationDispatched(ctx context.Context, payload *bo.AlertPayloadBo, routeKey string) {
	if b.cache == nil || payload == nil || routeKey == "" {
		return
	}
	fingerprint := alertDedupFingerprint(payload)
	if fingerprint == "" {
		return
	}
	namespaceUID, ok := contextx.TryGetNamespace(ctx)
	if !ok {
		return
	}
	phase := "firing"
	if isResolvedAlertStatus(payload.Status) {
		phase = "resolved"
	}
	key := alertNotifyDedupKey(namespaceUID.Int64(), fingerprint, routeKey, phase)
	if err := b.cache.Set(ctx, key, "1", alertNotifyDedupTTL); err != nil {
		b.helper.WithContext(ctx).Warnw("msg", "mark alert notification dedup failed", "error", err, "key", key)
	}
}

func alertDedupFingerprint(payload *bo.AlertPayloadBo) string {
	if payload == nil {
		return ""
	}
	if payload.Fingerprint != "" {
		return payload.Fingerprint
	}
	if v := payload.Labels["fingerprint"]; v != "" {
		return v
	}
	if v := payload.Labels["alert_event_uid"]; v != "" {
		return v
	}
	if payload.GroupKey != "" {
		return payload.GroupKey
	}
	return ""
}

func isResolvedAlertStatus(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "resolved")
}

func alertNotifyDedupKey(namespaceUID int64, fingerprint, routeKey, phase string) cache.K {
	return cache.NewKey("rabbit", "alert", "notify", phase, namespaceUID, fingerprint, routeKey)
}
