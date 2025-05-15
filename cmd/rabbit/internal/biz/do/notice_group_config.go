package do

import (
	"encoding/json"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/pkg/api/common"
	apicommon "github.com/aide-family/moon/pkg/api/rabbit/common"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/validate"
)

var _ cache.Object = (*NoticeGroupConfig)(nil)

type NoticeGroupConfig struct {
	Name            string                          `json:"name"`
	SMSConfigName   string                          `json:"smsConfigName"`
	EmailConfigName string                          `json:"emailConfigName"`
	HookConfigNames []string                        `json:"hookConfigNames"`
	SMSUserNames    []string                        `json:"smsUserNames"`
	EmailUserNames  []string                        `json:"emailUserNames"`
	Templates       map[common.NoticeType]*Template `json:"templates"`
}

// GetEmailConfigName implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetEmailConfigName() string {
	if n == nil {
		return ""
	}
	return n.EmailConfigName
}

// GetEmailTemplate implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetEmailTemplate() bo.Template {
	if n == nil {
		return nil
	}
	return n.Templates[common.NoticeType_NOTICE_TYPE_EMAIL]
}

// GetEmailUserNames implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetEmailUserNames() []string {
	if n == nil {
		return nil
	}
	return n.EmailUserNames
}

// GetHookConfigNames implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetHookConfigNames() []string {
	if n == nil {
		return nil
	}
	return n.HookConfigNames
}

// GetHookTemplate implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetHookTemplate(app apicommon.HookAPP) string {
	if n == nil {
		return ""
	}
	t, ok := n.Templates[common.NoticeType_NOTICE_TYPE_HOOK_DINGTALK]
	if !ok || validate.IsNil(t) {
		return ""
	}
	return t.Template
}

// GetName implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetName() string {
	if n == nil {
		return ""
	}
	return n.Name
}

// GetSmsConfigName implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetSmsConfigName() string {
	if n == nil {
		return ""
	}
	return n.SMSConfigName
}

// GetSmsTemplate implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetSmsTemplate() bo.Template {
	if n == nil {
		return nil
	}
	return n.Templates[common.NoticeType_NOTICE_TYPE_SMS]
}

// GetSmsUserNames implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetSmsUserNames() []string {
	if n == nil {
		return nil
	}
	return n.SMSUserNames
}

// GetTemplate implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetTemplate(noticeType common.NoticeType) bo.Template {
	if n == nil {
		return nil
	}
	return n.Templates[noticeType]
}

// GetTemplates implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetTemplates() map[common.NoticeType]bo.Template {
	if n == nil {
		return nil
	}
	templates := make(map[common.NoticeType]bo.Template, len(n.Templates))
	for k, v := range n.Templates {
		templates[k] = v
	}
	return templates
}

type Template struct {
	Type           common.NoticeType `json:"type"`
	Template       string            `json:"template"`
	TemplateParams string            `json:"templateParams"`
	Subject        string            `json:"subject"`
}

// GetTemplate implements bo.Template.
func (t *Template) GetTemplate() string {
	if t == nil {
		return ""
	}
	return t.Template
}

// GetTemplateParameters implements bo.Template.
func (t *Template) GetTemplateParameters() string {
	if t == nil {
		return ""
	}
	return t.TemplateParams
}

// GetSubject implements bo.Template.
func (t *Template) GetSubject() string {
	if t == nil {
		return ""
	}
	return t.Subject
}

// GetType implements bo.Template.
func (t *Template) GetType() common.NoticeType {
	if t == nil {
		return common.NoticeType_NOTICE_TYPE_UNKNOWN
	}
	return t.Type
}

// MarshalBinary implements cache.Object.
func (n *NoticeGroupConfig) MarshalBinary() (data []byte, err error) {
	return json.Marshal(n)
}

// UnmarshalBinary implements cache.Object.
func (n *NoticeGroupConfig) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, n)
}

func (n *NoticeGroupConfig) UniqueKey() string {
	return n.Name
}
