package bo

import "github.com/aide-family/moon/pkg/api/common"

type RemoveConfigParams struct {
	TeamID uint32
	Name   string
	Type   common.NoticeType
}
