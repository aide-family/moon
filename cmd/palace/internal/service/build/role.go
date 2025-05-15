package build

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/timex"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

type ListRoleRequest interface {
	GetPagination() *common.PaginationRequest
	GetStatus() common.GlobalStatus
	GetKeyword() string
}

func ToListRoleRequest(req ListRoleRequest) *bo.ListRoleReq {
	return &bo.ListRoleReq{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Status:            vobj.GlobalStatus(req.GetStatus()),
		Keyword:           req.GetKeyword(),
	}
}

func ToTeamRoleItem(role do.TeamRole) *common.TeamRoleItem {
	if validate.IsNil(role) {
		return nil
	}
	return &common.TeamRoleItem{
		TeamRoleId: role.GetID(),
		Name:       role.GetName(),
		Remark:     role.GetRemark(),
		Status:     common.GlobalStatus(role.GetStatus().GetValue()),
		Resources:  nil,
		Members:    nil,
		CreatedAt:  timex.Format(role.GetCreatedAt()),
		UpdatedAt:  timex.Format(role.GetUpdatedAt()),
		Creator:    ToUserBaseItem(role.GetCreator()),
	}
}

func ToTeamRoleItems(roles []do.TeamRole) []*common.TeamRoleItem {
	return slices.Map(roles, ToTeamRoleItem)
}

func ToSystemRoleItem(role do.Role) *common.SystemRoleItem {
	if validate.IsNil(role) {
		return nil
	}
	return &common.SystemRoleItem{
		RoleId:    role.GetID(),
		Name:      role.GetName(),
		Remark:    role.GetRemark(),
		Status:    common.GlobalStatus(role.GetStatus().GetValue()),
		CreatedAt: timex.Format(role.GetCreatedAt()),
		UpdatedAt: timex.Format(role.GetUpdatedAt()),
		Resources: nil,
		Users:     nil,
		Creator:   ToUserBaseItem(role.GetCreator()),
	}
}

func ToSystemRoleItems(roles []do.Role) []*common.SystemRoleItem {
	return slices.Map(roles, ToSystemRoleItem)
}

type SaveTeamRoleRequest interface {
	GetRoleId() uint32
	GetName() string
	GetRemark() string
	GetMenuIds() []uint32
}

func ToSaveTeamRoleRequest(req SaveTeamRoleRequest) *bo.SaveTeamRoleReq {
	return &bo.SaveTeamRoleReq{
		ID:      req.GetRoleId(),
		Name:    req.GetName(),
		Remark:  req.GetRemark(),
		MenuIds: req.GetMenuIds(),
	}
}

func ToSaveRoleRequest(req SaveTeamRoleRequest) *bo.SaveRoleReq {
	return &bo.SaveRoleReq{
		ID:      req.GetRoleId(),
		Name:    req.GetName(),
		Remark:  req.GetRemark(),
		MenuIds: req.GetMenuIds(),
	}
}
