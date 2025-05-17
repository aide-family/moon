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

// ToSaveTeamNoticeHookRequest 转换保存钩子请求
func ToSaveTeamNoticeHookRequest(req *palace.SaveTeamNoticeHookRequest) *bo.SaveTeamNoticeHookRequest {
	if req == nil {
		return nil
	}
	headers := make(map[string]string, len(req.GetHeaders()))
	for _, header := range req.GetHeaders() {
		headers[header.GetKey()] = header.GetValue()
	}

	return &bo.SaveTeamNoticeHookRequest{
		HookID:  req.GetHookId(),
		Name:    req.GetName(),
		Remark:  req.GetRemark(),
		URL:     req.GetUrl(),
		Method:  vobj.HTTPMethod(req.GetMethod()),
		Secret:  req.GetSecret(),
		Headers: headers,
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
	headers := make([]*common.Header, 0, len(hook.GetHeaders()))
	for k, v := range hook.GetHeaders() {
		headers = append(headers, &common.Header{Key: k, Value: v})
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
		Headers:      headers,
		App:          common.HookAPP(hook.GetApp().GetValue()),
		Creator:      ToUserBaseItem(hook.GetCreator()),
		NoticeGroups: ToNoticeGroupItems(hook.GetNoticeGroups()),
	}
}

// ToNoticeHookItems 转换钩子信息列表
func ToNoticeHookItems(hooks []do.NoticeHook) []*common.NoticeHookItem {
	return slices.Map(hooks, ToNoticeHookItem)
}
