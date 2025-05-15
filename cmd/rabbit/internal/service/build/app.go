package build

import (
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/rabbit/common"
)

func ToSendHookAPP(app common.HookAPP) vobj.APP {
	switch app {
	case common.HookAPP_OTHER:
		return vobj.APPHookOther
	case common.HookAPP_DINGTALK:
		return vobj.APPHookDingTalk
	case common.HookAPP_FEISHU:
		return vobj.APPHookFeiShu
	case common.HookAPP_WECHAT:
		return vobj.APPHookWechat
	default:
		return vobj.APPUnknown
	}
}
