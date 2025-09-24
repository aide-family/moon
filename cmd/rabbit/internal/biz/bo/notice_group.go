package bo

import (
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/validate"
)

type NoticeGroup interface {
	GetName() string
	GetSmsConfigName() string
	GetEmailConfigName() string
	GetSmsReceivers() []string
	GetEmailReceivers() []string
	GetHookReceivers() []string
	GetTemplates() map[vobj.APP]Template
	GetTemplate(noticeType vobj.APP) Template
	GetSmsTemplate() Template
	GetEmailTemplate() Template
	GetHookTemplate(app vobj.APP) string
}

type GetNoticeGroupConfigParams struct {
	TeamID uint32
	Name   string
}

type SetNoticeGroupConfigParams struct {
	TeamID  uint32
	Configs []NoticeGroup
}

func NewNoticeGroup(opts ...NoticeGroupOption) NoticeGroup {
	noticeGroup := &noticeGroup{
		templates: make(map[vobj.APP]Template, 7),
	}
	for _, opt := range opts {
		opt(noticeGroup)
	}
	return noticeGroup
}

type NoticeGroupOption func(noticeGroup *noticeGroup)

type noticeGroup struct {
	name            string
	smsConfigName   string
	emailConfigName string
	smsReceivers    []string
	emailReceivers  []string
	hookReceivers   []string
	templates       map[vobj.APP]Template
}

// GetSmsReceivers implements NoticeGroup.
func (n *noticeGroup) GetSmsReceivers() []string {
	return n.smsReceivers
}

// GetEmailReceivers implements NoticeGroup.
func (n *noticeGroup) GetEmailReceivers() []string {
	return n.emailReceivers
}

// GetHookReceivers implements NoticeGroup.
func (n *noticeGroup) GetHookReceivers() []string {
	return n.hookReceivers
}

// GetName implements NoticeGroup.
func (n *noticeGroup) GetName() string {
	return n.name
}

// GetSmsConfigName implements NoticeGroup.
func (n *noticeGroup) GetSmsConfigName() string {
	return n.smsConfigName
}

// GetEmailConfigName implements NoticeGroup.
func (n *noticeGroup) GetEmailConfigName() string {
	return n.emailConfigName
}

// GetTemplates implements NoticeGroup.
func (n *noticeGroup) GetTemplates() map[vobj.APP]Template {
	return n.templates
}

func (n *noticeGroup) GetTemplate(noticeType vobj.APP) Template {
	return n.templates[noticeType]
}

func (n *noticeGroup) GetSmsTemplate() Template {
	return n.templates[vobj.APPSms]
}

func (n *noticeGroup) GetEmailTemplate() Template {
	return n.templates[vobj.APPEmail]
}

func (n *noticeGroup) GetHookTemplate(app vobj.APP) string {
	var template Template
	var ok bool
	switch app {
	case vobj.APPHookDingTalk:
		template, ok = n.templates[vobj.APPHookDingTalk]
	case vobj.APPHookWechat:
		template, ok = n.templates[vobj.APPHookWechat]
	case vobj.APPHookFeiShu:
		template, ok = n.templates[vobj.APPHookFeiShu]
	case vobj.APPHookOther:
		template, ok = n.templates[vobj.APPHookOther]
	}
	if !ok || validate.IsNil(template) {
		return ""
	}
	return template.GetTemplate()
}

func WithNoticeGroupOptionName(name string) NoticeGroupOption {
	return func(noticeGroup *noticeGroup) {
		noticeGroup.name = name
	}
}

func WithNoticeGroupOptionSmsConfigName(smsConfigName string) NoticeGroupOption {
	return func(noticeGroup *noticeGroup) {
		noticeGroup.smsConfigName = smsConfigName
	}
}

func WithNoticeGroupOptionEmailConfigName(emailConfigName string) NoticeGroupOption {
	return func(noticeGroup *noticeGroup) {
		noticeGroup.emailConfigName = emailConfigName
	}
}

func WithNoticeGroupOptionHookReceivers(hookReceivers []string) NoticeGroupOption {
	return func(noticeGroup *noticeGroup) {
		noticeGroup.hookReceivers = hookReceivers
	}
}

func WithNoticeGroupOptionSmsReceivers(smsReceivers []string) NoticeGroupOption {
	return func(noticeGroup *noticeGroup) {
		noticeGroup.smsReceivers = smsReceivers
	}
}

func WithNoticeGroupOptionEmailReceivers(emailReceivers []string) NoticeGroupOption {
	return func(noticeGroup *noticeGroup) {
		noticeGroup.emailReceivers = emailReceivers
	}
}

type Template interface {
	GetType() vobj.APP
	GetTemplate() string
	GetTemplateParameters() string
	GetSubject() string
}

func WithNoticeGroupOptionTemplates(templates []Template) NoticeGroupOption {
	return func(noticeGroup *noticeGroup) {
		for _, template := range templates {
			noticeGroup.templates[template.GetType()] = template
		}
	}
}

func WithNoticeGroupOptionTemplate(template Template) NoticeGroupOption {
	return func(noticeGroup *noticeGroup) {
		noticeGroup.templates[template.GetType()] = template
	}
}
