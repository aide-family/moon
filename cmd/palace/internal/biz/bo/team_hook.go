package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/validate"
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
	if validate.IsNil(r) {
		return 0
	}
	if validate.IsNil(r.hookDo) {
		return r.HookID
	}
	return r.hookDo.GetID()
}

func (r *SaveTeamNoticeHookRequest) GetName() string {
	if validate.IsNil(r) {
		return ""
	}
	return r.Name
}

func (r *SaveTeamNoticeHookRequest) GetRemark() string {
	if validate.IsNil(r) {
		return ""
	}
	return r.Remark
}

func (r *SaveTeamNoticeHookRequest) GetStatus() vobj.GlobalStatus {
	if validate.IsNil(r) {
		return vobj.GlobalStatusUnknown
	}
	return r.Status
}

func (r *SaveTeamNoticeHookRequest) GetURL() string {
	if validate.IsNil(r) {
		return ""
	}
	return r.URL
}

func (r *SaveTeamNoticeHookRequest) GetMethod() vobj.HTTPMethod {
	if validate.IsNil(r) {
		return vobj.HTTPMethodPost
	}
	return r.Method
}

func (r *SaveTeamNoticeHookRequest) GetSecret() string {
	if validate.IsNil(r) {
		return ""
	}
	return r.Secret
}

func (r *SaveTeamNoticeHookRequest) GetHeaders() kv.StringMap {
	if validate.IsNil(r) {
		return nil
	}
	return r.Headers
}

func (r *SaveTeamNoticeHookRequest) GetApp() vobj.HookApp {
	if validate.IsNil(r) {
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

func (r *ListTeamNoticeHookRequest) ToListReply(hooks []do.NoticeHook) *ListTeamNoticeHookReply {
	return &ListTeamNoticeHookReply{
		PaginationReply: r.ToReply(),
		Items:           hooks,
	}
}

// ListTeamNoticeHookReply 列表响应
type ListTeamNoticeHookReply = ListReply[do.NoticeHook]

// UpdateTeamNoticeHookStatusRequest 更新状态请求
type UpdateTeamNoticeHookStatusRequest struct {
	HookID uint32            `json:"hookId"`
	Status vobj.GlobalStatus `json:"status"`
}
