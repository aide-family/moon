// Package bo is the business logic object
package bo

import (
	"net/http"
	"time"

	"github.com/aide-family/magicbox/encoding"
	"github.com/aide-family/magicbox/encoding/json"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

type SendEmailBo struct {
	UID         snowflake.ID `json:"-"`
	Subject     string       `json:"subject"`
	Body        string       `json:"body"`
	To          []string     `json:"to"`
	Cc          []string     `json:"cc"`
	ContentType string       `json:"content_type"`
	Headers     http.Header  `json:"headers"`
}

func (b *SendEmailBo) ToMessageLog(emailConfig *EmailConfigItemBo) (*CreateMessageLogBo, error) {
	jsonCodec, ok := encoding.GetCodec(json.Name)
	if !ok {
		return nil, merr.ErrorInternalServer("%s codec not found", json.Name)
	}
	messageBytes, err := jsonCodec.Marshal(b)
	if err != nil {
		return nil, err
	}
	emailConfigBytes, err := jsonCodec.Marshal(emailConfig)
	if err != nil {
		return nil, err
	}
	return NewCreateMessageLogBo(strutil.EncryptString(messageBytes), strutil.EncryptString(emailConfigBytes), enum.MessageType_EMAIL), nil
}

func NewSendEmailBo(req *apiv1.SendEmailRequest) *SendEmailBo {
	headers := make(http.Header)
	for key, value := range req.Headers {
		headers.Add(key, value)
	}
	return &SendEmailBo{
		UID:         snowflake.ParseInt64(req.Uid),
		Subject:     req.Subject,
		Body:        req.Body,
		To:          req.To,
		Cc:          req.Cc,
		ContentType: req.ContentType,
		Headers:     headers,
	}
}

type SendEmailWithTemplateBo struct {
	UID         snowflake.ID
	TemplateUID snowflake.ID
	JSONData    []byte
	To          []string
	Cc          []string
}

func NewSendEmailWithTemplateBo(req *apiv1.SendEmailWithTemplateRequest) (*SendEmailWithTemplateBo, error) {
	jsonCodec, ok := encoding.GetCodec(json.Name)
	if !ok {
		return nil, merr.ErrorInternalServer("%s codec not found", json.Name)
	}
	if !jsonCodec.Valid([]byte(req.JsonData)) {
		return nil, merr.ErrorParams("invalid json data")
	}
	return &SendEmailWithTemplateBo{
		UID:         snowflake.ParseInt64(req.Uid),
		TemplateUID: snowflake.ParseInt64(req.TemplateUID),
		JSONData:    []byte(req.JsonData),
		To:          req.To,
		Cc:          req.Cc,
	}, nil
}

func (b *SendEmailWithTemplateBo) ToSendEmailBo(templateBo *TemplateItemBo) (*SendEmailBo, error) {
	if templateBo.MessageType != enum.MessageType_EMAIL {
		return nil, merr.ErrorParams("invalid template message type, expected email type, got %d", templateBo.MessageType)
	}
	if templateBo.Status != enum.GlobalStatus_ENABLED {
		return nil, merr.ErrorParams("template %s(%d) is disabled", templateBo.Name, templateBo.UID)
	}
	emailTemplateData, err := templateBo.ToEmailTemplateData()
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

	subjectData, err := strutil.ExecuteTextTemplate(emailTemplateData.Subject, jsonData)
	if err != nil {
		return nil, merr.ErrorParams("execute text template failed").WithCause(err)
	}
	bodyData, err := strutil.ExecuteTextTemplate(emailTemplateData.Body, jsonData)
	if err != nil {
		return nil, merr.ErrorParams("execute text template failed").WithCause(err)
	}

	return &SendEmailBo{
		UID:         b.UID,
		To:          b.To,
		Cc:          b.Cc,
		Subject:     subjectData,
		Body:        bodyData,
		ContentType: emailTemplateData.ContentType,
		Headers:     emailTemplateData.Headers,
	}, nil
}

type CreateEmailConfigBo struct {
	Name     string
	Host     string
	Port     int32
	Username string
	Password string
}

func NewCreateEmailConfigBo(req *apiv1.CreateEmailConfigRequest) *CreateEmailConfigBo {
	return &CreateEmailConfigBo{
		Name:     req.Name,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
	}
}

type UpdateEmailConfigBo struct {
	UID snowflake.ID
	CreateEmailConfigBo
}

func NewUpdateEmailConfigBo(req *apiv1.UpdateEmailConfigRequest) *UpdateEmailConfigBo {
	return &UpdateEmailConfigBo{
		UID: snowflake.ParseInt64(req.Uid),
		CreateEmailConfigBo: CreateEmailConfigBo{
			Name:     req.Name,
			Host:     req.Host,
			Port:     req.Port,
			Username: req.Username,
			Password: req.Password,
		},
	}
}

type UpdateEmailConfigStatusBo struct {
	UID    snowflake.ID
	Status enum.GlobalStatus
}

func NewUpdateEmailConfigStatusBo(req *apiv1.UpdateEmailConfigStatusRequest) *UpdateEmailConfigStatusBo {
	return &UpdateEmailConfigStatusBo{
		UID:    snowflake.ParseInt64(req.Uid),
		Status: req.Status,
	}
}

type ListEmailConfigBo struct {
	*PageRequestBo
	Keyword string
	Status  enum.GlobalStatus
}

func NewListEmailConfigBo(req *apiv1.ListEmailConfigRequest) *ListEmailConfigBo {
	return &ListEmailConfigBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		Keyword:       req.Keyword,
		Status:        req.Status,
	}
}

func ToAPIV1ListEmailConfigReply(pageResponseBo *PageResponseBo[*EmailConfigItemBo]) *apiv1.ListEmailConfigReply {
	items := make([]*apiv1.EmailConfigItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1EmailConfigItem())
	}
	return &apiv1.ListEmailConfigReply{
		Items:    items,
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
	}
}

// SelectEmailConfigBo 选择Email配置的 BO
type SelectEmailConfigBo struct {
	Keyword string
	Limit   int32
	LastUID snowflake.ID
	Status  enum.GlobalStatus
}

// NewSelectEmailConfigBo 从 API 请求创建 BO
func NewSelectEmailConfigBo(req *apiv1.SelectEmailConfigRequest) *SelectEmailConfigBo {
	var lastUID snowflake.ID
	if req.LastUID > 0 {
		lastUID = snowflake.ParseInt64(req.LastUID)
	}
	return &SelectEmailConfigBo{
		Keyword: req.Keyword,
		Limit:   req.Limit,
		LastUID: lastUID,
		Status:  req.Status,
	}
}

// EmailConfigItemSelectBo Email配置选择项的 BO
type EmailConfigItemSelectBo struct {
	UID      snowflake.ID
	Name     string
	Status   enum.GlobalStatus
	Disabled bool
	Tooltip  string
}

// ToAPIV1EmailConfigItemSelect 转换为 API 响应
func (b *EmailConfigItemSelectBo) ToAPIV1EmailConfigItemSelect() *apiv1.EmailConfigItemSelect {
	return &apiv1.EmailConfigItemSelect{
		Value:    b.UID.Int64(),
		Label:    b.Name,
		Disabled: b.Disabled,
		Tooltip:  b.Tooltip,
	}
}

// SelectEmailConfigBoResult Biz层返回结果
type SelectEmailConfigBoResult struct {
	Items   []*EmailConfigItemSelectBo
	Total   int64
	LastUID snowflake.ID
}

// SelectEmailConfigReplyParams 转换为API响应的参数
type SelectEmailConfigReplyParams struct {
	Items   []*EmailConfigItemSelectBo
	Total   int64
	LastUID snowflake.ID
	Limit   int32
}

// ToAPIV1SelectEmailConfigReply 转换为 API 响应
func ToAPIV1SelectEmailConfigReply(params *SelectEmailConfigReplyParams) *apiv1.SelectEmailConfigReply {
	selectItems := make([]*apiv1.EmailConfigItemSelect, 0, len(params.Items))
	for _, item := range params.Items {
		selectItems = append(selectItems, item.ToAPIV1EmailConfigItemSelect())
	}
	var lastUIDInt64 int64
	if params.LastUID > 0 {
		lastUIDInt64 = params.LastUID.Int64()
	}
	// hasMore: 如果返回的记录数等于limit，说明可能还有更多记录
	// 如果返回的记录数小于limit，说明已经查询完了
	hasMore := int32(len(params.Items)) == params.Limit
	return &apiv1.SelectEmailConfigReply{
		Items:   selectItems,
		Total:   params.Total,
		LastUID: lastUIDInt64,
		HasMore: hasMore,
	}
}

type EmailConfigItemBo struct {
	UID       snowflake.ID      `json:"uid"`
	Name      string            `json:"name"`
	Host      string            `json:"host"`
	Port      int32             `json:"port"`
	Username  string            `json:"username"`
	Password  string            `json:"password"`
	Status    enum.GlobalStatus `json:"status"`
	CreatedAt time.Time         `json:"-"`
	UpdatedAt time.Time         `json:"-"`
}

// GetHost implements email.Config.
func (b *EmailConfigItemBo) GetHost() string {
	return b.Host
}

// GetPassword implements email.Config.
func (b *EmailConfigItemBo) GetPassword() string {
	return b.Password
}

// GetPort implements email.Config.
func (b *EmailConfigItemBo) GetPort() int32 {
	return b.Port
}

// GetUsername implements email.Config.
func (b *EmailConfigItemBo) GetUsername() string {
	return b.Username
}

func (b *EmailConfigItemBo) ToAPIV1EmailConfigItem() *apiv1.EmailConfigItem {
	return &apiv1.EmailConfigItem{
		Uid:       b.UID.Int64(),
		Name:      b.Name,
		Host:      b.Host,
		Port:      b.Port,
		Username:  b.Username,
		Password:  b.Password,
		Status:    enum.GlobalStatus(b.Status),
		CreatedAt: b.CreatedAt.Format(time.DateTime),
		UpdatedAt: b.UpdatedAt.Format(time.DateTime),
	}
}
