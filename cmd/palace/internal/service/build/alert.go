package build

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	apicommon "github.com/moon-monitor/moon/pkg/api/common"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/util/kv/label"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/timex"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func ToAlertParams(req *apicommon.AlertItem) *bo.Alert {
	annotations := label.NewAnnotationFromMap(req.GetAnnotations())
	labels := label.NewLabel(req.GetLabels())
	return &bo.Alert{
		Status:       vobj.AlertStatus(req.Status),
		Labels:       labels.ToMap(),
		Summary:      annotations.GetSummary(),
		Description:  annotations.GetDescription(),
		Value:        req.GetValue(),
		GeneratorURL: req.GetGeneratorURL(),
		TeamID:       labels.GetTeamId(),
		Fingerprint:  req.GetFingerprint(),
		StartsAt:     timex.ParseX(req.GetStartsAt()),
		EndsAt:       timex.ParseX(req.GetEndsAt()),
	}
}

func ToListAlertParams(req *palace.ListAlertParams) *bo.ListAlertParams {
	return &bo.ListAlertParams{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		TimeRange:         ToTimeRange(req.GetTimeRange()),
		Fingerprint:       req.GetFingerprint(),
		Keyword:           req.GetKeyword(),
		TeamID:            0,
		Status:            vobj.AlertStatus(req.GetStatus()),
	}
}

func ToRealtimeAlertItems(items []do.Realtime) []*palace.RealtimeAlertItem {
	return slices.Map(items, ToRealtimeAlertItem)
}

func ToRealtimeAlertItem(item do.Realtime) *palace.RealtimeAlertItem {
	if validate.IsNil(item) {
		return nil
	}
	return &palace.RealtimeAlertItem{
		AlertId:      item.GetID(),
		Status:       apicommon.AlertStatus(item.GetStatus().GetValue()),
		Fingerprint:  item.GetFingerprint(),
		Labels:       item.GetLabels().ToMap(),
		Summary:      item.GetSummary(),
		Description:  item.GetDescription(),
		Value:        item.GetValue(),
		GeneratorURL: item.GetGeneratorURL(),
		StartsAt:     timex.Format(item.GetStartsAt()),
		EndsAt:       timex.Format(item.GetEndsAt()),
	}
}
