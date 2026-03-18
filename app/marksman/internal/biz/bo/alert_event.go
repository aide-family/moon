package bo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

// ListRealtimeAlertTimeRangeDefault is the default lookback when StartAt/EndAt are zero.
const ListRealtimeAlertTimeRangeDefault = 14 * 24 * time.Hour

// AlertEventBo is the business object for an alert event produced by evaluator (in-memory).
type AlertEventBo struct {
	StrategyUID           snowflake.ID
	NamespaceUID          snowflake.ID
	Level                 *LevelItemBo
	Summary               string
	Description           string
	Expr                  string
	FiredAt               time.Time
	Value                 float64
	Labels                map[string]string
	DatasourceUID         snowflake.ID
	EvaluatorType         string // e.g. "metric", for identifying which evaluator produced the event
	EvaluatorSnapshotJSON string // pre-serialized evaluator snapshot JSON (used by repo to find-or-insert snapshot, then store ID on event)
	// Fingerprint is a stable identifier for this alert event, derived from firedAt and labels.
	Fingerprint string
	// EvaluateDuration is the evaluator window; used to schedule auto-recovery at 2x this duration.
	EvaluateDuration time.Duration
}

// AlertEventItemBo is the business object for a persisted alert event (real-time alert).
type AlertEventItemBo struct {
	UID                 snowflake.ID
	StrategyUID         snowflake.ID
	NamespaceUID        snowflake.ID
	LevelUID            snowflake.ID
	LevelName           string
	Summary             string
	Description         string
	Expr                string
	FiredAt             time.Time
	Value               float64
	Labels              map[string]string
	DatasourceUID       snowflake.ID
	EvaluatorType       string
	EvaluatorSnapshotID snowflake.ID
	Status              enum.AlertEventStatus
	IntervenedAt        *time.Time
	IntervenedBy        snowflake.ID
	SuppressedUntilAt   *time.Time
	SuppressedBy        snowflake.ID
	SuppressedReason    string
	RecoveredAt         *time.Time
	RecoveredBy         snowflake.ID
	RecoveredReason     string
}

func ToAPIV1AlertEventItem(b *AlertEventItemBo) *apiv1.AlertEventItem {
	if b == nil {
		return nil
	}
	item := &apiv1.AlertEventItem{
		Uid:              b.UID.Int64(),
		StrategyUid:      b.StrategyUID.Int64(),
		NamespaceUid:     b.NamespaceUID.Int64(),
		LevelUid:         b.LevelUID.Int64(),
		LevelName:        b.LevelName,
		Summary:          b.Summary,
		Description:      b.Description,
		Expr:             b.Expr,
		FiredAt:          timex.FormatTime(&b.FiredAt),
		Value:            b.Value,
		Labels:           b.Labels,
		DatasourceUid:    b.DatasourceUID.Int64(),
		Status:           b.Status,
		IntervenedAt:     timex.FormatTime(b.IntervenedAt),
		IntervenedBy:     b.IntervenedBy.Int64(),
		SuppressedBy:     b.SuppressedBy.Int64(),
		SuppressedReason: b.SuppressedReason,
		RecoveredAt:      timex.FormatTime(b.RecoveredAt),
		RecoveredBy:      b.RecoveredBy.Int64(),
		RecoveredReason:  b.RecoveredReason,
	}
	if b.IntervenedAt != nil {
		item.IntervenedAt = timex.FormatTime(b.IntervenedAt)
		item.IntervenedBy = b.IntervenedBy.Int64()
	}
	if b.SuppressedUntilAt != nil {
		item.SuppressUntilAt = timex.FormatTime(b.SuppressedUntilAt)
		item.SuppressedBy = b.SuppressedBy.Int64()
		item.SuppressedReason = b.SuppressedReason
	}
	if b.RecoveredAt != nil {
		item.RecoveredAt = timex.FormatTime(b.RecoveredAt)
		item.RecoveredBy = b.RecoveredBy.Int64()
		item.RecoveredReason = b.RecoveredReason
	}
	return item
}

type ListRealtimeAlertBo struct {
	*PageRequestBo
	AlertPageUID snowflake.ID
	Status       enum.AlertEventStatus
	StartAt      time.Time
	EndAt        time.Time
}

func NewListRealtimeAlertBo(req *apiv1.ListRealtimeAlertRequest) *ListRealtimeAlertBo {
	var startAt, endAt time.Time
	if req.GetStartAtUnix() != 0 {
		startAt = time.Unix(req.GetStartAtUnix(), 0)
	}
	if req.GetEndAtUnix() != 0 {
		endAt = time.Unix(req.GetEndAtUnix(), 0)
	}
	return &ListRealtimeAlertBo{
		PageRequestBo: NewPageRequestBo(req.GetPage(), req.GetPageSize()),
		AlertPageUID:  snowflake.ParseInt64(req.GetAlertPageUid()),
		Status:        req.GetStatus(),
		StartAt:       startAt,
		EndAt:         endAt,
	}
}

func ToAPIV1ListRealtimeAlertReply(pageResponseBo *PageResponseBo[*AlertEventItemBo]) *apiv1.ListRealtimeAlertReply {
	items := make([]*apiv1.AlertEventItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, ToAPIV1AlertEventItem(item))
	}
	return &apiv1.ListRealtimeAlertReply{
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
		Items:    items,
	}
}

// BuildAlertFingerprint builds a deterministic fingerprint from firedAt and labels.
// It sorts label keys, joins them as "k=v" pairs, and hashes with SHA-256.
func BuildAlertFingerprint(labels map[string]string) string {
	if labels == nil {
		labels = map[string]string{}
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+labels[k])
	}
	base := strings.Join(parts, ",")
	sum := sha256.Sum256([]byte(base))
	return hex.EncodeToString(sum[:])
}

type InterveneAlertBo struct {
	UID          snowflake.ID
	IntervenedBy snowflake.ID
}

func NewInterveneAlertBo(ctx context.Context, req *apiv1.InterveneAlertRequest) *InterveneAlertBo {
	return &InterveneAlertBo{
		UID:          snowflake.ParseInt64(req.GetUid()),
		IntervenedBy: contextx.GetUserUID(ctx),
	}
}

type SuppressAlertBo struct {
	UID              snowflake.ID
	SuppressUntilAt  time.Time
	SuppressedBy     snowflake.ID
	SuppressedReason string
}

func NewSuppressAlertBo(ctx context.Context, req *apiv1.SuppressAlertRequest) *SuppressAlertBo {
	suppressUntilAt := time.Now()
	if suppressUntilUnix := req.GetSuppressUntilUnix(); suppressUntilUnix > 0 {
		suppressUntilAt = time.Unix(suppressUntilUnix, 0)
	}
	return &SuppressAlertBo{
		UID:              snowflake.ParseInt64(req.GetUid()),
		SuppressUntilAt:  suppressUntilAt,
		SuppressedBy:     contextx.GetUserUID(ctx),
		SuppressedReason: req.GetSuppressedReason(),
	}
}

type RecoverAlertBo struct {
	UID             snowflake.ID
	RecoveredBy     snowflake.ID
	RecoveredReason string
}

func NewRecoverAlertBo(ctx context.Context, req *apiv1.RecoverAlertRequest) *RecoverAlertBo {
	return &RecoverAlertBo{
		UID:             snowflake.ParseInt64(req.GetUid()),
		RecoveredBy:     contextx.GetUserUID(ctx),
		RecoveredReason: req.GetRecoveredReason(),
	}
}
