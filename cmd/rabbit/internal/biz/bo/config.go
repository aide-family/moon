package bo

import "github.com/aide-family/moon/pkg/api/common"

type RemoveConfigParams struct {
	TeamID string
	Name   string
	Type   common.NoticeType
}
