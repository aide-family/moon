package bo

import (
	"strings"
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/event"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
)

type Alert struct {
	TeamID       uint32           `json:"teamId"`
	Status       vobj.AlertStatus `json:"status"`
	Fingerprint  string           `json:"fingerprint"`
	Labels       kv.StringMap     `json:"labels"`
	Summary      string           `json:"summary"`
	Description  string           `json:"description"`
	Value        string           `json:"value"`
	GeneratorURL string           `json:"generatorURL"`
	StartsAt     time.Time        `json:"startsAt"`
	EndsAt       time.Time        `json:"endsAt"`
}

func (a *Alert) Validate() error {
	if a.StartsAt.IsZero() {
		return merr.ErrorParamsError("startsAt is required")
	}
	if !a.Status.Exist() {
		return merr.ErrorParamsError("status is required")
	}
	if strings.TrimSpace(a.Fingerprint) == "" {
		return merr.ErrorParamsError("fingerprint is required")
	}
	if a.TeamID <= 0 {
		return merr.ErrorParamsError("teamId is required")
	}
	if a.Status.IsResolved() {
		if a.EndsAt.IsZero() {
			return merr.ErrorParamsError("endsAt is required")
		}
		if a.EndsAt.Before(a.StartsAt) {
			return merr.ErrorParamsError("endsAt must be after startsAt")
		}
	}

	return nil
}

type GetAlertParams struct {
	TeamID      uint32    `json:"teamId"`
	Fingerprint string    `json:"fingerprint"`
	StartsAt    time.Time `json:"startsAt"`
}

type ListAlertParams struct {
	*PaginationRequest
	TeamID      uint32           `json:"teamId"`
	Fingerprint string           `json:"fingerprint"`
	Keyword     string           `json:"keyword"`
	TimeRange   []time.Time      `json:"timeRange"`
	Status      vobj.AlertStatus `json:"status"`
}

func (p *ListAlertParams) ToListAlertReply(items []*event.Realtime) *ListAlertReply {
	return &ListAlertReply{
		PaginationReply: p.ToReply(),
		Items:           slices.Map(items, func(item *event.Realtime) do.Realtime { return item }),
	}
}

type ListAlertReply = ListReply[do.Realtime]
