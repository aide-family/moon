package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/crypto"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.NoticeHook = (*NoticeHook)(nil)

const tableNameNoticeHook = "team_notice_hooks"

type NoticeHook struct {
	do.TeamModel
	Name         string                   `gorm:"column:name;type:varchar(64);not null;comment:名称" json:"name"`
	Remark       string                   `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status       vobj.GlobalStatus        `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	URL          string                   `gorm:"column:url;type:varchar(255);not null;comment:URL" json:"url"`
	Method       vobj.HTTPMethod          `gorm:"column:method;type:tinyint(2);not null;comment:请求方法" json:"method"`
	Secret       crypto.String            `gorm:"column:secret;type:varchar(255);not null;comment:密钥" json:"secret"`
	Headers      *crypto.Object[[]*kv.KV] `gorm:"column:headers;type:text;not null;comment:请求头" json:"headers"`
	NoticeGroups []*NoticeGroup           `gorm:"many2many:team_notice_group_hooks" json:"noticeGroups"`
	APP          vobj.HookApp             `gorm:"column:app;type:tinyint(2);not null;comment:应用" json:"app"`
}

func (n *NoticeHook) GetName() string {
	if n == nil {
		return ""
	}
	return n.Name
}

func (n *NoticeHook) GetRemark() string {
	if n == nil {
		return ""
	}
	return n.Remark
}

func (n *NoticeHook) GetStatus() vobj.GlobalStatus {
	if n == nil {
		return vobj.GlobalStatusUnknown
	}
	return n.Status
}

func (n *NoticeHook) GetURL() string {
	if n == nil {
		return ""
	}
	return n.URL
}

func (n *NoticeHook) GetMethod() vobj.HTTPMethod {
	if n == nil {
		return vobj.HTTPMethodUnknown
	}
	return n.Method
}

func (n *NoticeHook) GetSecret() string {
	if n == nil {
		return ""
	}
	return string(n.Secret)
}

func (n *NoticeHook) GetHeaders() []*kv.KV {
	if n == nil {
		return nil
	}
	return n.Headers.Get()
}

func (n *NoticeHook) GetApp() vobj.HookApp {
	if n == nil {
		return vobj.HookAppOther
	}
	return n.APP
}

func (n *NoticeHook) GetNoticeGroups() []do.NoticeGroup {
	if n == nil {
		return nil
	}
	return slices.Map(n.NoticeGroups, func(v *NoticeGroup) do.NoticeGroup { return v })
}

func (n *NoticeHook) TableName() string {
	return tableNameNoticeHook
}
