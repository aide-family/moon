package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

type NoticeHook interface {
	GetID() uint32
	GetName() string
	GetRemark() string
	GetURL() string
	GetMethod() vobj.HTTPMethod
	GetSecret() string
	GetHeaders() []*kv.KV
	GetApp() vobj.HookApp
}

// SaveTeamNoticeHookRequest represents a request to save a team notification hook
type SaveTeamNoticeHookRequest struct {
	hookDo  do.NoticeHook
	HookID  uint32          `json:"hookId"`
	Name    string          `json:"name"`
	Remark  string          `json:"remark"`
	URL     string          `json:"url"`
	Method  vobj.HTTPMethod `json:"method"`
	Secret  string          `json:"secret"`
	Headers []*kv.KV        `json:"headers"`
	APP     vobj.HookApp    `json:"app"`
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

func (r *SaveTeamNoticeHookRequest) GetHeaders() []*kv.KV {
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

// ListTeamNoticeHookRequest represents a request to list team notification hooks
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

// ListTeamNoticeHookReply represents a response containing a list of team notification hooks
type ListTeamNoticeHookReply = ListReply[do.NoticeHook]

// UpdateTeamNoticeHookStatusRequest represents a request to update the status of a team notification hook
type UpdateTeamNoticeHookStatusRequest struct {
	HookID uint32            `json:"hookId"`
	Status vobj.GlobalStatus `json:"status"`
}

// TeamNoticeHookSelectRequest represents a request to select team notification hooks
type TeamNoticeHookSelectRequest struct {
	*PaginationRequest
	Status  vobj.GlobalStatus `json:"status"`
	Keyword string            `json:"keyword"`
	Apps    []vobj.HookApp    `json:"apps"`
	URL     string            `json:"url"`
}

func (r *TeamNoticeHookSelectRequest) ToSelectReply(hooks []do.NoticeHook) *TeamNoticeHookSelectReply {
	return &TeamNoticeHookSelectReply{
		PaginationReply: r.ToReply(),
		Items: slices.Map(hooks, func(hook do.NoticeHook) SelectItem {
			return &selectItem{
				Value:    hook.GetID(),
				Label:    hook.GetName(),
				Disabled: !hook.GetStatus().IsEnable() || hook.GetDeletedAt() != 0,
				Extra: &selectItemExtra{
					Remark: hook.GetRemark(),
					Color:  hook.GetMethod().String(),
					Icon:   hook.GetApp().String(),
				},
			}
		}),
	}
}

// TeamNoticeHookSelectReply represents a response containing selected team notification hooks
type TeamNoticeHookSelectReply = ListReply[SelectItem]
