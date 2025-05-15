package build

import (
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/do"
)

func ToLabelNotice(notice bo.LabelNotices) *do.LabelNotices {
	return &do.LabelNotices{
		Key:            notice.GetKey(),
		Value:          notice.GetValue(),
		ReceiverRoutes: notice.GetReceiverRoutes(),
	}
}
