package bo

import (
	stdjson "encoding/json"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

type PrometheusAlertItemBo struct {
	Status       string
	Labels       map[string]string
	Annotations  map[string]string
	StartsAt     time.Time
	EndsAt       *time.Time
	GeneratorURL string
	Fingerprint  string
}

type ReceivePrometheusWebhookBo struct {
	Version           string
	GroupKey          string
	Status            string
	Receiver          string
	ExternalURL       string
	GroupLabels       map[string]string
	CommonLabels      map[string]string
	CommonAnnotations map[string]string
	Alerts            []*PrometheusAlertItemBo
	TruncatedAlerts   int32
	Source            string
}

func NewReceivePrometheusWebhookBo(req *apiv1.ReceivePrometheusWebhookRequest) *ReceivePrometheusWebhookBo {
	alerts := make([]*PrometheusAlertItemBo, 0, len(req.GetAlerts()))
	for _, item := range req.GetAlerts() {
		if item == nil {
			continue
		}
		alerts = append(alerts, &PrometheusAlertItemBo{
			Status:       item.GetStatus(),
			Labels:       item.GetLabels(),
			Annotations:  item.GetAnnotations(),
			StartsAt:     parseTime(item.GetStartsAt()),
			EndsAt:       parseTimePtr(item.GetEndsAt()),
			GeneratorURL: item.GetGeneratorURL(),
			Fingerprint:  item.GetFingerprint(),
		})
	}
	return &ReceivePrometheusWebhookBo{
		Version:           req.GetVersion(),
		GroupKey:          req.GetGroupKey(),
		Status:            req.GetStatus(),
		Receiver:          req.GetReceiver(),
		ExternalURL:       req.GetExternalURL(),
		GroupLabels:       req.GetGroupLabels(),
		CommonLabels:      req.GetCommonLabels(),
		CommonAnnotations: req.GetCommonAnnotations(),
		Alerts:            alerts,
		TruncatedAlerts:   req.GetTruncatedAlerts(),
		Source:            req.GetSource(),
	}
}

func (b *ReceivePrometheusWebhookBo) NamespaceUID() (snowflake.ID, error) {
	candidates := []map[string]string{b.CommonLabels}
	for _, alert := range b.Alerts {
		if alert != nil {
			candidates = append(candidates, alert.Labels)
		}
	}
	for _, labels := range candidates {
		if labels == nil {
			continue
		}
		if value, ok := labels["namespace_uid"]; ok && value != "" {
			uid, err := strconv.ParseInt(value, 10, 64)
			if err != nil || uid <= 0 {
				return 0, merr.ErrorParams("invalid namespace_uid label: %s", value)
			}
			return snowflake.ID(uid), nil
		}
	}
	return 0, merr.ErrorParams("namespace_uid label is required")
}

type AlertPayloadBo struct {
	Source       string
	Receiver     string
	Status       string
	Fingerprint  string
	GroupKey     string
	StartsAt     time.Time
	EndsAt       *time.Time
	GeneratorURL string
	Labels       map[string]string
	Annotations  map[string]string
}

func NewAlertPayloadBo(req *ReceivePrometheusWebhookBo, alert *PrometheusAlertItemBo) *AlertPayloadBo {
	if req == nil || alert == nil {
		return nil
	}
	return &AlertPayloadBo{
		Source:       req.Source,
		Receiver:     req.Receiver,
		Status:       alert.Status,
		Fingerprint:  alert.Fingerprint,
		GroupKey:     req.GroupKey,
		StartsAt:     alert.StartsAt,
		EndsAt:       alert.EndsAt,
		GeneratorURL: alert.GeneratorURL,
		Labels:       mergeStringMap(req.CommonLabels, alert.Labels),
		Annotations:  mergeStringMap(req.CommonAnnotations, alert.Annotations),
	}
}

func mergeStringMap(base map[string]string, extra map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(extra))
	for key, value := range base {
		result[key] = value
	}
	for key, value := range extra {
		result[key] = value
	}
	return result
}

type AlertSubscriptionMemberBo struct {
	MemberUID    int64
	IsEmail      bool
	IsSMS        bool
	IsPhone      bool
	MemberName   string
	MemberAvatar string
	MemberEmail  string
	MemberPhone  string
}

func NewAlertSubscriptionMembersBo(reqs []*apiv1.AlertSubscriptionMemberRequest) []*AlertSubscriptionMemberBo {
	items := make([]*AlertSubscriptionMemberBo, 0, len(reqs))
	for _, item := range reqs {
		if item == nil {
			continue
		}
		items = append(items, &AlertSubscriptionMemberBo{
			MemberUID: item.GetMemberUid(),
			IsEmail:   item.GetIsEmail(),
			IsSMS:     item.GetIsSms(),
			IsPhone:   item.GetIsPhone(),
		})
	}
	return items
}

func (b *AlertSubscriptionMemberBo) ToAPIV1() *apiv1.AlertSubscriptionMemberItem {
	return &apiv1.AlertSubscriptionMemberItem{
		MemberUid:    b.MemberUID,
		IsEmail:      b.IsEmail,
		IsSms:        b.IsSMS,
		IsPhone:      b.IsPhone,
		MemberName:   b.MemberName,
		MemberAvatar: b.MemberAvatar,
		MemberEmail:  b.MemberEmail,
		MemberPhone:  b.MemberPhone,
	}
}

type AlertSubscriptionItemBo struct {
	UID                    snowflake.ID
	Name                   string
	Remark                 string
	Labels                 map[string]string
	ExcludeLabels          map[string]string
	RecipientGroupUIDs     []int64
	Members                []*AlertSubscriptionMemberBo
	DirectMemberEmailConfigUID snowflake.ID
	DirectMemberTemplateUID    snowflake.ID
	Status                 enum.GlobalStatus
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func (b *AlertSubscriptionItemBo) ToAPIV1() *apiv1.AlertSubscriptionItem {
	members := make([]*apiv1.AlertSubscriptionMemberItem, 0, len(b.Members))
	for _, item := range b.Members {
		members = append(members, item.ToAPIV1())
	}
	return &apiv1.AlertSubscriptionItem{
		Uid:                      b.UID.Int64(),
		Name:                     b.Name,
		Remark:                   b.Remark,
		Labels:                   b.Labels,
		ExcludeLabels:            b.ExcludeLabels,
		RecipientGroupUids:       b.RecipientGroupUIDs,
		Members:                  members,
		DirectMemberEmailConfigUid: b.DirectMemberEmailConfigUID.Int64(),
		DirectMemberTemplateUid:    b.DirectMemberTemplateUID.Int64(),
		Status:                   b.Status,
		CreatedAt:                timex.FormatTime(&b.CreatedAt),
		UpdatedAt:                timex.FormatTime(&b.UpdatedAt),
	}
}

type CreateAlertSubscriptionBo struct {
	Name                     string
	Remark                   string
	Labels                   map[string]string
	ExcludeLabels            map[string]string
	RecipientGroupUIDs       []int64
	Members                  []*AlertSubscriptionMemberBo
	DirectMemberEmailConfigUID snowflake.ID
	DirectMemberTemplateUID    snowflake.ID
}

func NewCreateAlertSubscriptionBo(req *apiv1.CreateAlertSubscriptionRequest) *CreateAlertSubscriptionBo {
	return &CreateAlertSubscriptionBo{
		Name:                     req.GetName(),
		Remark:                   req.GetRemark(),
		Labels:                   req.GetLabels(),
		ExcludeLabels:            req.GetExcludeLabels(),
		RecipientGroupUIDs:       req.GetRecipientGroupUids(),
		Members:                  NewAlertSubscriptionMembersBo(req.GetMembers()),
		DirectMemberEmailConfigUID: snowflake.ParseInt64(req.GetDirectMemberEmailConfigUid()),
		DirectMemberTemplateUID:    snowflake.ParseInt64(req.GetDirectMemberTemplateUid()),
	}
}

type UpdateAlertSubscriptionBo struct {
	UID                      snowflake.ID
	Name                     string
	Remark                   string
	Labels                   map[string]string
	ExcludeLabels            map[string]string
	RecipientGroupUIDs       []int64
	Members                  []*AlertSubscriptionMemberBo
	DirectMemberEmailConfigUID snowflake.ID
	DirectMemberTemplateUID    snowflake.ID
}

func NewUpdateAlertSubscriptionBo(req *apiv1.UpdateAlertSubscriptionRequest) *UpdateAlertSubscriptionBo {
	return &UpdateAlertSubscriptionBo{
		UID:                      snowflake.ParseInt64(req.GetUid()),
		Name:                     req.GetName(),
		Remark:                   req.GetRemark(),
		Labels:                   req.GetLabels(),
		ExcludeLabels:            req.GetExcludeLabels(),
		RecipientGroupUIDs:       req.GetRecipientGroupUids(),
		Members:                  NewAlertSubscriptionMembersBo(req.GetMembers()),
		DirectMemberEmailConfigUID: snowflake.ParseInt64(req.GetDirectMemberEmailConfigUid()),
		DirectMemberTemplateUID:    snowflake.ParseInt64(req.GetDirectMemberTemplateUid()),
	}
}

type UpdateAlertSubscriptionStatusBo struct {
	UID    snowflake.ID
	Status enum.GlobalStatus
}

func NewUpdateAlertSubscriptionStatusBo(req *apiv1.UpdateAlertSubscriptionStatusRequest) *UpdateAlertSubscriptionStatusBo {
	return &UpdateAlertSubscriptionStatusBo{
		UID:    snowflake.ParseInt64(req.GetUid()),
		Status: req.GetStatus(),
	}
}

type ListAlertSubscriptionBo struct {
	*PageRequestBo
	Keyword string
	Status  enum.GlobalStatus
}

func NewListAlertSubscriptionBo(req *apiv1.ListAlertSubscriptionsRequest) *ListAlertSubscriptionBo {
	return &ListAlertSubscriptionBo{
		PageRequestBo: NewPageRequestBo(req.GetPage(), req.GetPageSize()),
		Keyword:       req.GetKeyword(),
		Status:        req.GetStatus(),
	}
}

func ToAPIV1ListAlertSubscriptionsReply(page *PageResponseBo[*AlertSubscriptionItemBo]) *apiv1.ListAlertSubscriptionsReply {
	items := make([]*apiv1.AlertSubscriptionItem, 0, len(page.GetItems()))
	for _, item := range page.GetItems() {
		items = append(items, item.ToAPIV1())
	}
	return &apiv1.ListAlertSubscriptionsReply{
		Items:    items,
		Total:    page.GetTotal(),
		Page:     page.GetPage(),
		PageSize: page.GetPageSize(),
	}
}

func (b *AlertSubscriptionItemBo) MatchesLabels(labels map[string]string) bool {
	if b == nil {
		return false
	}
	if len(b.ExcludeLabels) > 0 {
		excluded := true
		for key, value := range b.ExcludeLabels {
			if labels[key] != value {
				excluded = false
				break
			}
		}
		if excluded {
			return false
		}
	}
	if len(b.Labels) == 0 {
		return false
	}
	for key, value := range b.Labels {
		if labels[key] != value {
			return false
		}
	}
	return true
}

func (b *AlertSubscriptionItemBo) DirectEmailEnabled() bool {
	if b.DirectMemberEmailConfigUID <= 0 {
		return false
	}
	return slices.ContainsFunc(b.Members, func(item *AlertSubscriptionMemberBo) bool {
		return item != nil && item.IsEmail
	})
}

func BuildAlertTemplateData(payload *AlertPayloadBo) string {
	payloadJSON, _ := stdjson.Marshal(map[string]any{
		"source":       payload.Source,
		"receiver":     payload.Receiver,
		"status":       payload.Status,
		"fingerprint":  payload.Fingerprint,
		"groupKey":     payload.GroupKey,
		"startsAt":     timex.FormatTime(&payload.StartsAt),
		"endsAt":       timex.FormatTime(payload.EndsAt),
		"generatorURL": payload.GeneratorURL,
		"labels":       payload.Labels,
		"annotations":  payload.Annotations,
	})
	return string(payloadJSON)
}

func BuildDefaultAlertSubject(payload *AlertPayloadBo) string {
	alertName := payload.Labels["alertname"]
	if alertName == "" {
		alertName = payload.Fingerprint
	}
	return strings.ToUpper(payload.Status) + " " + alertName
}

func BuildDefaultAlertBody(payload *AlertPayloadBo) string {
	builder := strings.Builder{}
	builder.WriteString("Status: " + payload.Status + "\n")
	builder.WriteString("Fingerprint: " + payload.Fingerprint + "\n")
	if summary := payload.Annotations["summary"]; summary != "" {
		builder.WriteString("Summary: " + summary + "\n")
	}
	if description := payload.Annotations["description"]; description != "" {
		builder.WriteString("Description: " + description + "\n")
	}
	builder.WriteString("Labels:\n")
	keys := make([]string, 0, len(payload.Labels))
	for key := range payload.Labels {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		builder.WriteString("- " + key + "=" + payload.Labels[key] + "\n")
	}
	return builder.String()
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseTimePtr(value string) *time.Time {
	if value == "" {
		return nil
	}
	t := parseTime(value)
	if t.IsZero() {
		return nil
	}
	return &t
}
