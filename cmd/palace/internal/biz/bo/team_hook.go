package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
)

type NoticeHook interface {
	GetID() uint32
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetURL() string
	GetMethod() vobj.HTTPMethod
	GetSecret() string
	GetHeaders() kv.StringMap
	GetApp() vobj.HookApp
}

// SaveTeamNoticeHookRequest 保存团队通知钩子请求
type SaveTeamNoticeHookRequest struct {
	hookDo  do.NoticeHook
	HookID  uint32            `json:"hookId"`
	Name    string            `json:"name"`
	Remark  string            `json:"remark"`
	Status  vobj.GlobalStatus `json:"status"`
	URL     string            `json:"url"`
	Method  vobj.HTTPMethod   `json:"method"`
	Secret  string            `json:"secret"`
	Headers kv.StringMap      `json:"headers"`
	APP     vobj.HookApp      `json:"app"`
}

func (r *SaveTeamNoticeHookRequest) GetID() uint32 {
	if r == nil || r.hookDo == nil {
		return 0
	}
	return r.hookDo.GetID()
}

func (r *SaveTeamNoticeHookRequest) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *SaveTeamNoticeHookRequest) GetRemark() string {
	if r == nil {
		return ""
	}
	return r.Remark
}

func (r *SaveTeamNoticeHookRequest) GetStatus() vobj.GlobalStatus {
	if r == nil {
		return vobj.GlobalStatusUnknown
	}
	return r.Status
}

func (r *SaveTeamNoticeHookRequest) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

func (r *SaveTeamNoticeHookRequest) GetMethod() vobj.HTTPMethod {
	if r == nil {
		return vobj.HTTPMethodPost
	}
	return r.Method
}

func (r *SaveTeamNoticeHookRequest) GetSecret() string {
	if r == nil {
		return ""
	}
	return r.Secret
}

func (r *SaveTeamNoticeHookRequest) GetHeaders() kv.StringMap {
	if r == nil {
		return nil
	}
	return r.Headers
}

func (r *SaveTeamNoticeHookRequest) GetApp() vobj.HookApp {
	if r == nil {
		return vobj.HookAppOther
	}
	return r.APP
}

func (r *SaveTeamNoticeHookRequest) WithUpdateHookRequest(hook do.NoticeHook) NoticeHook {
	r.hookDo = hook
	return r
}

// ListTeamNoticeHookRequest 列表请求
type ListTeamNoticeHookRequest struct {
	*PaginationRequest
	Status  vobj.GlobalStatus `json:"status"`
	Keyword string            `json:"keyword"`
	Apps    []vobj.HookApp    `json:"apps"`
}

func (r *ListTeamNoticeHookRequest) ToListTeamNoticeHookReply(hooks []*team.NoticeHook) *ListTeamNoticeHookReply {
	return &ListTeamNoticeHookReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(hooks, func(hook *team.NoticeHook) do.NoticeHook { return hook }),
	}
}

// ListTeamNoticeHookReply 列表响应
type ListTeamNoticeHookReply = ListReply[do.NoticeHook]

// UpdateTeamNoticeHookStatusRequest 更新状态请求
type UpdateTeamNoticeHookStatusRequest struct {
	HookID uint32            `json:"hookId"`
	Status vobj.GlobalStatus `json:"status"`
}
