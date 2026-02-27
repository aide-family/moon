package bo

import (
	"time"

	"github.com/aide-family/magicbox/encoding"
	"github.com/aide-family/magicbox/encoding/json"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

type CreateWebhookBo struct {
	App     enum.WebhookAPP
	Name    string
	URL     string
	Method  enum.HTTPMethod
	Headers map[string]string
	Secret  string
}

func NewCreateWebhookBo(req *apiv1.CreateWebhookRequest) *CreateWebhookBo {
	return &CreateWebhookBo{
		App:     req.App,
		Name:    req.Name,
		URL:     req.Url,
		Method:  req.Method,
		Headers: req.Headers,
		Secret:  req.Secret,
	}
}

type UpdateWebhookBo struct {
	UID     snowflake.ID
	App     enum.WebhookAPP
	Name    string
	URL     string
	Method  enum.HTTPMethod
	Headers map[string]string
	Secret  string
}

func NewUpdateWebhookBo(req *apiv1.UpdateWebhookRequest) *UpdateWebhookBo {
	return &UpdateWebhookBo{
		UID:     snowflake.ParseInt64(req.Uid),
		App:     req.App,
		Name:    req.Name,
		URL:     req.Url,
		Method:  req.Method,
		Headers: req.Headers,
		Secret:  req.Secret,
	}
}

type UpdateWebhookStatusBo struct {
	UID    snowflake.ID
	Status enum.GlobalStatus
}

func NewUpdateWebhookStatusBo(req *apiv1.UpdateWebhookStatusRequest) *UpdateWebhookStatusBo {
	return &UpdateWebhookStatusBo{
		UID:    snowflake.ParseInt64(req.Uid),
		Status: req.Status,
	}
}

type WebhookItemBo struct {
	UID       snowflake.ID      `json:"uid"`
	App       enum.WebhookAPP   `json:"app"`
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Method    enum.HTTPMethod   `json:"method"`
	Headers   map[string]string `json:"headers"`
	Secret    string            `json:"secret"`
	Status    enum.GlobalStatus `json:"status"`
	CreatedAt time.Time         `json:"-"`
	UpdatedAt time.Time         `json:"-"`
}

func (b *WebhookItemBo) ToAPIV1WebhookItem() *apiv1.WebhookItem {
	return &apiv1.WebhookItem{
		Uid:       b.UID.Int64(),
		App:       b.App,
		Name:      b.Name,
		Url:       b.URL,
		Method:    b.Method,
		Headers:   b.Headers,
		Secret:    b.Secret,
		Status:    b.Status,
		CreatedAt: b.CreatedAt.Format(time.DateTime),
		UpdatedAt: b.UpdatedAt.Format(time.DateTime),
	}
}

type ListWebhookBo struct {
	*PageRequestBo
	App     enum.WebhookAPP
	Keyword string
	Status  enum.GlobalStatus
}

func NewListWebhookBo(req *apiv1.ListWebhookRequest) *ListWebhookBo {
	return &ListWebhookBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		App:           req.App,
		Keyword:       req.Keyword,
		Status:        req.Status,
	}
}

func ToAPIV1ListWebhookReply(pageResponseBo *PageResponseBo[*WebhookItemBo]) *apiv1.ListWebhookReply {
	items := make([]*apiv1.WebhookItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1WebhookItem())
	}
	return &apiv1.ListWebhookReply{
		Items:    items,
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
	}
}

// SelectWebhookBo 选择Webhook的 BO
type SelectWebhookBo struct {
	App     enum.WebhookAPP
	Keyword string
	Limit   int32
	LastUID snowflake.ID
	Status  enum.GlobalStatus
}

// NewSelectWebhookBo 从 API 请求创建 BO
func NewSelectWebhookBo(req *apiv1.SelectWebhookRequest) *SelectWebhookBo {
	var lastUID snowflake.ID
	if req.LastUID > 0 {
		lastUID = snowflake.ParseInt64(req.LastUID)
	}
	return &SelectWebhookBo{
		App:     req.App,
		Keyword: req.Keyword,
		Limit:   req.Limit,
		LastUID: lastUID,
		Status:  req.Status,
	}
}

// WebhookItemSelectBo Webhook选择项的 BO
type WebhookItemSelectBo struct {
	UID      snowflake.ID
	Name     string
	Status   enum.GlobalStatus
	Disabled bool
	Tooltip  string
	App      enum.WebhookAPP
}

// ToAPIV1WebhookItemSelect 转换为 API 响应
func (b *WebhookItemSelectBo) ToAPIV1WebhookItemSelect() *apiv1.WebhookItemSelect {
	return &apiv1.WebhookItemSelect{
		Value:    b.UID.Int64(),
		Label:    b.Name,
		Disabled: b.Disabled,
		Tooltip:  b.Tooltip,
		App:      b.App,
	}
}

// SelectWebhookBoResult Biz层返回结果
type SelectWebhookBoResult struct {
	Items   []*WebhookItemSelectBo
	Total   int64
	LastUID snowflake.ID
}

// SelectWebhookReplyParams 转换为API响应的参数
type SelectWebhookReplyParams struct {
	Items   []*WebhookItemSelectBo
	Total   int64
	LastUID snowflake.ID
	Limit   int32
}

// ToAPIV1SelectWebhookReply 转换为 API 响应
func ToAPIV1SelectWebhookReply(params *SelectWebhookReplyParams) *apiv1.SelectWebhookReply {
	selectItems := make([]*apiv1.WebhookItemSelect, 0, len(params.Items))
	for _, item := range params.Items {
		selectItems = append(selectItems, item.ToAPIV1WebhookItemSelect())
	}
	var lastUIDInt64 int64
	if params.LastUID > 0 {
		lastUIDInt64 = params.LastUID.Int64()
	}
	// hasMore: 如果返回的记录数等于limit，说明可能还有更多记录
	// 如果返回的记录数小于limit，说明已经查询完了
	hasMore := int32(len(params.Items)) == params.Limit
	return &apiv1.SelectWebhookReply{
		Items:   selectItems,
		Total:   params.Total,
		LastUID: lastUIDInt64,
		HasMore: hasMore,
	}
}

type SendWebhookBo struct {
	UID  snowflake.ID `json:"uid"`
	Data string       `json:"data"`
}

func (b *SendWebhookBo) ToMessageLog(webhookConfig *WebhookItemBo) (*CreateMessageLogBo, error) {
	jsonCodec, ok := encoding.GetCodec(json.Name)
	if !ok {
		return nil, merr.ErrorInternalServer("%s codec not found", json.Name)
	}
	webhookConfigBytes, err := jsonCodec.Marshal(webhookConfig)
	if err != nil {
		return nil, err
	}
	return NewCreateMessageLogBo(strutil.EncryptString(b.Data), strutil.EncryptString(webhookConfigBytes), enum.MessageType(webhookConfig.App)), nil
}

func NewSendWebhookBo(req *apiv1.SendWebhookRequest) *SendWebhookBo {
	return &SendWebhookBo{
		UID:  snowflake.ParseInt64(req.Uid),
		Data: req.Data,
	}
}

type SendWebhookWithTemplateBo struct {
	UID         snowflake.ID
	TemplateUID snowflake.ID
	JSONData    []byte
}

func NewSendWebhookWithTemplateBo(req *apiv1.SendWebhookWithTemplateRequest) (*SendWebhookWithTemplateBo, error) {
	jsonCodec, ok := encoding.GetCodec(json.Name)
	if !ok {
		return nil, merr.ErrorInternalServer("%s codec not found", json.Name)
	}
	if !jsonCodec.Valid([]byte(req.JsonData)) {
		return nil, merr.ErrorParams("invalid json data")
	}
	return &SendWebhookWithTemplateBo{
		UID:         snowflake.ParseInt64(req.Uid),
		TemplateUID: snowflake.ParseInt64(req.TemplateUID),
		JSONData:    []byte(req.JsonData),
	}, nil
}

func (b *SendWebhookWithTemplateBo) ToSendWebhookBo(templateDo *TemplateItemBo) (*SendWebhookBo, error) {
	if !(templateDo.MessageType < 2000 || templateDo.MessageType >= 3000) {
		return nil, merr.ErrorParams("invalid template message type, expected webhook type, got %s", templateDo.MessageType)
	}
	if templateDo.Status != enum.GlobalStatus_ENABLED {
		return nil, merr.ErrorParams("template %s(%d) is disabled", templateDo.Name, templateDo.UID)
	}
	webhookTemplateData, err := templateDo.ToWebhookTemplateData()
	if err != nil {
		return nil, err
	}
	var jsonData map[string]any
	jsonCodec, ok := encoding.GetCodec(json.Name)
	if !ok {
		return nil, merr.ErrorInternalServer("%s codec not found", json.Name)
	}
	if err := jsonCodec.Unmarshal(b.JSONData, &jsonData); err != nil {
		return nil, merr.ErrorInternalServer("unmarshal json data failed").WithCause(err)
	}

	bodyData, err := strutil.ExecuteTextTemplate(string(webhookTemplateData), jsonData)
	if err != nil {
		return nil, merr.ErrorParams("execute text template failed").WithCause(err)
	}

	return &SendWebhookBo{
		UID:  b.UID,
		Data: bodyData,
	}, nil
}
