package convert

import (
	"strconv"
	"time"

	"github.com/aide-family/magicbox/timex"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/do"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// SysUserRoleToBo 将数据库模型转换为用户角色业务对象
func SysUserRoleToBo(ur *do.SysUserRole) *bo.SysUserRoleItemBo {
	if ur == nil {
		return nil
	}
	return &bo.SysUserRoleItemBo{
		RoleCode:  ur.RoleCode,
		CreatedAt: ur.CreatedAt,
		Grantor:   int64(ur.Creator),
	}
}

// SysUserRoleJoinRow 用于接收 JOIN 查询结果的结构体
type SysUserRoleJoinRow struct {
	RoleCode      string
	RoleName      string
	RoleDataScope int32
	RoleStatus    int32
	CreatedAt     time.Time
	Grantor       int64
}

// SysUserRoleJoinToBo 将 JOIN 查询结果转换为业务对象
func SysUserRoleJoinToBo(row *SysUserRoleJoinRow) *bo.SysUserRoleItemBo {
	if row == nil {
		return nil
	}
	return &bo.SysUserRoleItemBo{
		RoleCode:      row.RoleCode,
		RoleName:      row.RoleName,
		RoleDataScope: row.RoleDataScope,
		RoleStatus:    row.RoleStatus,
		CreatedAt:     row.CreatedAt,
		Grantor:       row.Grantor,
	}
}

// SysUserRoleItemBotoAPIV1 将业务对象转换为 proto 响应对象
func SysUserRoleItemBotoAPIV1(b *bo.SysUserRoleItemBo) *goddessv1.SysUserRoleItem {
	return &goddessv1.SysUserRoleItem{
		RoleCode:      b.RoleCode,
		RoleName:      b.RoleName,
		RoleDataScope: b.RoleDataScope,
		RoleStatus:    b.RoleStatus,
		GrantedAt:     timex.FormatTime(&b.CreatedAt),
		Grantor:       strconv.FormatInt(b.Grantor, 10),
	}
}
