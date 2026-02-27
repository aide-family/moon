package bo

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

// CreateTemplateBo 创建模板的 BO
type CreateTemplateBo struct {
	Name        string
	MessageType enum.MessageType
	JSONData    string
}

// NewCreateTemplateBo 从 API 请求创建 BO
func NewCreateTemplateBo(req *apiv1.CreateTemplateRequest) (*CreateTemplateBo, error) {
	if !json.Valid([]byte(req.JsonData)) {
		return nil, merr.ErrorParams("invalid json data")
	}
	return &CreateTemplateBo{
		Name:        req.Name,
		MessageType: req.MessageType,
		JSONData:    req.JsonData,
	}, nil
}

// UpdateTemplateBo 更新模板的 BO
type UpdateTemplateBo struct {
	UID         snowflake.ID
	Name        string
	MessageType enum.MessageType
	JSONData    string
}

// NewUpdateTemplateBo 从 API 请求创建 BO
func NewUpdateTemplateBo(req *apiv1.UpdateTemplateRequest) (*UpdateTemplateBo, error) {
	if !json.Valid([]byte(req.JsonData)) {
		return nil, merr.ErrorParams("invalid json data")
	}
	return &UpdateTemplateBo{
		UID:         snowflake.ParseInt64(req.Uid),
		Name:        req.Name,
		MessageType: req.MessageType,
		JSONData:    req.JsonData,
	}, nil
}

// UpdateTemplateStatusBo 更新模板状态的 BO
type UpdateTemplateStatusBo struct {
	UID    snowflake.ID
	Status enum.GlobalStatus
}

// NewUpdateTemplateStatusBo 从 API 请求创建 BO
func NewUpdateTemplateStatusBo(req *apiv1.UpdateTemplateStatusRequest) *UpdateTemplateStatusBo {
	return &UpdateTemplateStatusBo{
		UID:    snowflake.ParseInt64(req.Uid),
		Status: req.Status,
	}
}

// EmailTemplateData Email 模板的数据结构
type EmailTemplateData struct {
	Subject     string      `json:"subject"`
	Body        string      `json:"body"`
	ContentType string      `json:"content_type"`
	Headers     http.Header `json:"headers,omitempty"`
}

// WebhookTemplateData Webhook 模板的数据结构
type WebhookTemplateData string

// SMSTemplateData SMS 模板的数据结构
type SMSTemplateData struct {
	Content string            `json:"content"`
	Params  map[string]string `json:"params,omitempty"`
}

// TemplateItemBo 模板项的 BO
type TemplateItemBo struct {
	UID         snowflake.ID
	Name        string
	MessageType enum.MessageType
	JSONData    string
	Status      enum.GlobalStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ToEmailTemplateData 将 JSONData 转换为 EmailTemplateData
func (t *TemplateItemBo) ToEmailTemplateData() (*EmailTemplateData, error) {
	var data EmailTemplateData
	if err := json.Unmarshal([]byte(t.JSONData), &data); err != nil {
		return nil, err
	}
	if data.ContentType == "" {
		data.ContentType = "text/html"
	}
	return &data, nil
}

// ToWebhookTemplateData 将 JSONData 转换为 WebhookTemplateData
func (t *TemplateItemBo) ToWebhookTemplateData() (WebhookTemplateData, error) {
	return WebhookTemplateData(t.JSONData), nil
}

// ToSMSTemplateData 将 JSONData 转换为 SMSTemplateData
func (t *TemplateItemBo) ToSMSTemplateData() (*SMSTemplateData, error) {
	var data SMSTemplateData
	if err := json.Unmarshal([]byte(t.JSONData), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// ToAPIV1TemplateItem 转换为 API 响应
func (t *TemplateItemBo) ToAPIV1TemplateItem() *apiv1.TemplateItem {
	return &apiv1.TemplateItem{
		Uid:         t.UID.Int64(),
		Name:        t.Name,
		MessageType: t.MessageType,
		JsonData:    t.JSONData,
		Status:      enum.GlobalStatus(t.Status),
		CreatedAt:   t.CreatedAt.Format(time.DateTime),
		UpdatedAt:   t.UpdatedAt.Format(time.DateTime),
	}
}

// ListTemplateBo 列表查询的 BO
type ListTemplateBo struct {
	*PageRequestBo
	Keyword     string
	Status      enum.GlobalStatus
	MessageType enum.MessageType
}

// NewListTemplateBo 从 API 请求创建 BO
func NewListTemplateBo(req *apiv1.ListTemplateRequest) *ListTemplateBo {
	return &ListTemplateBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		Keyword:       req.Keyword,
		Status:        req.Status,
		MessageType:   req.MessageType,
	}
}

// ToAPIV1ListTemplateReply 转换为 API 响应
func ToAPIV1ListTemplateReply(pageResponseBo *PageResponseBo[*TemplateItemBo]) *apiv1.ListTemplateReply {
	items := make([]*apiv1.TemplateItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1TemplateItem())
	}
	return &apiv1.ListTemplateReply{
		Items:    items,
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
	}
}

// SelectTemplateBo 选择模板的 BO
type SelectTemplateBo struct {
	MessageType enum.MessageType
	Keyword     string
	Limit       int32
	LastUID     snowflake.ID
	Status      enum.GlobalStatus
}

// NewSelectTemplateBo 从 API 请求创建 BO
func NewSelectTemplateBo(req *apiv1.SelectTemplateRequest) *SelectTemplateBo {
	var lastUID snowflake.ID
	if req.LastUID > 0 {
		lastUID = snowflake.ParseInt64(req.LastUID)
	}
	return &SelectTemplateBo{
		MessageType: req.MessageType,
		Keyword:     req.Keyword,
		Limit:       req.Limit,
		LastUID:     lastUID,
		Status:      req.Status,
	}
}

// TemplateItemSelectBo 模板选择项的 BO
type TemplateItemSelectBo struct {
	UID      snowflake.ID
	Name     string
	Status   enum.GlobalStatus
	Disabled bool
	Tooltip  string
}

// ToAPIV1TemplateItemSelect 转换为 API 响应
func (b *TemplateItemSelectBo) ToAPIV1TemplateItemSelect() *apiv1.TemplateItemSelect {
	return &apiv1.TemplateItemSelect{
		Value:    b.UID.Int64(),
		Label:    b.Name,
		Disabled: b.Disabled,
		Tooltip:  b.Tooltip,
	}
}

// SelectTemplateBoResult Biz层返回结果
type SelectTemplateBoResult struct {
	Items   []*TemplateItemSelectBo
	Total   int64
	LastUID snowflake.ID
}

// SelectTemplateReplyParams 转换为API响应的参数
type SelectTemplateReplyParams struct {
	Items   []*TemplateItemSelectBo
	Total   int64
	LastUID snowflake.ID
	Limit   int32
}

// ToAPIV1SelectTemplateReply 转换为 API 响应
func ToAPIV1SelectTemplateReply(params *SelectTemplateReplyParams) *apiv1.SelectTemplateReply {
	selectItems := make([]*apiv1.TemplateItemSelect, 0, len(params.Items))
	for _, item := range params.Items {
		selectItems = append(selectItems, item.ToAPIV1TemplateItemSelect())
	}
	var lastUIDInt64 int64
	if params.LastUID > 0 {
		lastUIDInt64 = params.LastUID.Int64()
	}
	// hasMore: 如果返回的记录数等于limit，说明可能还有更多记录
	// 如果返回的记录数小于limit，说明已经查询完了
	hasMore := int32(len(params.Items)) == params.Limit
	return &apiv1.SelectTemplateReply{
		Items:   selectItems,
		Total:   params.Total,
		LastUID: lastUIDInt64,
		HasMore: hasMore,
	}
}
