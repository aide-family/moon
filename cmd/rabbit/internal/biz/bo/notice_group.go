package bo

import (
	"github.com/aide-family/moon/pkg/api/common"
	apicommon "github.com/aide-family/moon/pkg/api/rabbit/common"
	"github.com/aide-family/moon/pkg/util/validate"
)

type NoticeGroup interface {
	GetName() string
	GetSmsConfigName() string
	GetEmailConfigName() string
	GetSmsReceivers() []string
	GetEmailReceivers() []string
	GetHookReceivers() []string
	GetTemplates() map[common.NoticeType]Template
	GetTemplate(noticeType common.NoticeType) Template
	GetSmsTemplate() Template
	GetEmailTemplate() Template
	GetHookTemplate(app apicommon.HookAPP) string
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
		templates: make(map[common.NoticeType]Template, 7),
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
	templates       map[common.NoticeType]Template
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
func (n *noticeGroup) GetTemplates() map[common.NoticeType]Template {
	return n.templates
}

func (n *noticeGroup) GetTemplate(noticeType common.NoticeType) Template {
	return n.templates[noticeType]
}

func (n *noticeGroup) GetSmsTemplate() Template {
	return n.templates[common.NoticeType_NOTICE_TYPE_SMS]
}

func (n *noticeGroup) GetEmailTemplate() Template {
	return n.templates[common.NoticeType_NOTICE_TYPE_EMAIL]
}

func (n *noticeGroup) GetHookTemplate(app apicommon.HookAPP) string {
	var template Template
	var ok bool
	switch app {
	case apicommon.HookAPP_DINGTALK:
		template, ok = n.templates[common.NoticeType_NOTICE_TYPE_HOOK_DINGTALK]
	case apicommon.HookAPP_WECHAT:
		template, ok = n.templates[common.NoticeType_NOTICE_TYPE_HOOK_WECHAT]
	case apicommon.HookAPP_FEISHU:
		template, ok = n.templates[common.NoticeType_NOTICE_TYPE_HOOK_FEISHU]
	case apicommon.HookAPP_OTHER:
		template, ok = n.templates[common.NoticeType_NOTICE_TYPE_HOOK_WEBHOOK]
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
	GetType() common.NoticeType
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
