package bo

import "github.com/moon-monitor/moon/pkg/api/common"

type RemoveConfigParams struct {
	TeamID string
	Name   string
	Type   common.NoticeType
}
