package do

import (
	"encoding/json"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/validate"
)

var _ cache.Object = (*NoticeGroupConfig)(nil)

type NoticeGroupConfig struct {
	Name            string                 `json:"name"`
	SMSConfigName   string                 `json:"smsConfigName"`
	EmailConfigName string                 `json:"emailConfigName"`
	HookReceivers   []string               `json:"hookReceivers"`
	SMSReceivers    []string               `json:"smsReceivers"`
	EmailReceivers  []string               `json:"emailReceivers"`
	Templates       map[vobj.APP]*Template `json:"templates"`
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
	return n.Templates[vobj.APPEmail]
}

// GetEmailReceivers implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetEmailReceivers() []string {
	if n == nil {
		return nil
	}
	return n.EmailReceivers
}

func (n *NoticeGroupConfig) GetHookReceivers() []string {
	if n == nil {
		return nil
	}
	return n.HookReceivers
}

// GetHookTemplate implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetHookTemplate(app vobj.APP) string {
	if n == nil {
		return ""
	}
	t, ok := n.Templates[vobj.APPHookDingTalk]
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
	return n.Templates[vobj.APPSms]
}

// GetSmsReceivers implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetSmsReceivers() []string {
	if n == nil {
		return nil
	}
	return n.SMSReceivers
}

// GetTemplate implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetTemplate(noticeType vobj.APP) bo.Template {
	if n == nil {
		return nil
	}
	return n.Templates[noticeType]
}

// GetTemplates implements bo.NoticeGroup.
func (n *NoticeGroupConfig) GetTemplates() map[vobj.APP]bo.Template {
	if n == nil {
		return nil
	}
	templates := make(map[vobj.APP]bo.Template, len(n.Templates))
	for k, v := range n.Templates {
		templates[k] = v
	}
	return templates
}

type Template struct {
	Type           vobj.APP `json:"type"`
	Template       string   `json:"template"`
	TemplateParams string   `json:"templateParams"`
	Subject        string   `json:"subject"`
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
func (t *Template) GetType() vobj.APP {
	if t == nil {
		return vobj.APPUnknown
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
