package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
)

// ToSaveTeamNoticeHookRequest converts save hook request
func ToSaveTeamNoticeHookRequest(req *palace.SaveTeamNoticeHookRequest) *bo.SaveTeamNoticeHookRequest {
	if req == nil {
		return nil
	}

	return &bo.SaveTeamNoticeHookRequest{
		HookID:  req.GetHookId(),
		Name:    req.GetName(),
		Remark:  req.GetRemark(),
		URL:     req.GetUrl(),
		Method:  vobj.HTTPMethod(req.GetMethod()),
		Secret:  req.GetSecret(),
		Headers: ToKVs(req.GetHeaders()),
		APP:     vobj.HookApp(req.GetApp()),
	}
}

// ToListTeamNoticeHookRequest converts list request
func ToListTeamNoticeHookRequest(req *palace.ListTeamNoticeHookRequest) *bo.ListTeamNoticeHookRequest {
	if req == nil {
		return nil
	}
	return &bo.ListTeamNoticeHookRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Status:            vobj.GlobalStatus(req.GetStatus()),
		Keyword:           req.GetKeyword(),
		Apps:              slices.Map(req.GetApps(), func(app common.HookAPP) vobj.HookApp { return vobj.HookApp(app) }),
	}
}

// ToTeamNoticeHookSelectRequest converts select request
func ToTeamNoticeHookSelectRequest(req *palace.TeamNoticeHookSelectRequest) *bo.TeamNoticeHookSelectRequest {
	if req == nil {
		return nil
	}
	return &bo.TeamNoticeHookSelectRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Status:            vobj.GlobalStatus(req.GetStatus()),
		Apps:              slices.Map(req.GetApps(), func(app common.HookAPP) vobj.HookApp { return vobj.HookApp(app) }),
		Keyword:           req.GetKeyword(),
		URL:               req.GetUrl(),
	}
}

// ToNoticeHookItem converts hook information
func ToNoticeHookItem(hook do.NoticeHook) *common.NoticeHookItem {
	if hook == nil {
		return nil
	}
	return &common.NoticeHookItem{
		NoticeHookId: hook.GetID(),
		CreatedAt:    timex.Format(hook.GetCreatedAt()),
		UpdatedAt:    timex.Format(hook.GetUpdatedAt()),
		Name:         hook.GetName(),
		Remark:       hook.GetRemark(),
		Status:       common.GlobalStatus(hook.GetStatus().GetValue()),
		Url:          hook.GetURL(),
		Method:       common.HTTPMethod(hook.GetMethod().GetValue()),
		Secret:       hook.GetSecret(),
		Headers:      ToKVsItems(hook.GetHeaders()),
		App:          common.HookAPP(hook.GetApp().GetValue()),
		Creator:      ToUserBaseItem(hook.GetCreator()),
		NoticeGroups: ToNoticeGroupItems(hook.GetNoticeGroups()),
	}
}

// ToNoticeHookItems converts hook information list
func ToNoticeHookItems(hooks []do.NoticeHook) []*common.NoticeHookItem {
	return slices.Map(hooks, ToNoticeHookItem)
}
