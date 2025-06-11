package build

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/timex"
)

// ToSaveTeamNoticeHookRequest 转换保存钩子请求
func ToSaveTeamNoticeHookRequest(req *palace.SaveTeamNoticeHookRequest) *bo.SaveTeamNoticeHookRequest {
	if req == nil {
		return nil
	}
	return &bo.SaveTeamNoticeHookRequest{
		HookID:  req.GetHookId(),
		Name:    req.GetName(),
		Remark:  req.GetRemark(),
		Status:  vobj.GlobalStatus(req.GetStatus()),
		URL:     req.GetUrl(),
		Method:  vobj.HTTPMethod(req.GetMethod()),
		Secret:  req.GetSecret(),
		Headers: req.GetHeaders(),
		APP:     vobj.HookApp(req.GetApp()),
	}
}

// ToListTeamNoticeHookRequest 转换列表请求
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

// ToNoticeHookItem 转换钩子信息
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
		Headers:      hook.GetHeaders(),
		App:          common.HookAPP(hook.GetApp().GetValue()),
		Creator:      ToUserBaseItem(hook.GetCreator()),
		NoticeGroups: ToNoticeGroupItems(hook.GetNoticeGroups()),
	}
}

// ToNoticeHookItems 转换钩子信息列表
func ToNoticeHookItems(hooks []do.NoticeHook) []*common.NoticeHookItem {
	return slices.Map(hooks, ToNoticeHookItem)
}
