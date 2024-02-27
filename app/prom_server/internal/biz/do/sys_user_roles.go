package do

import (
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/cache"

	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/util/slices"
)

var _ encoding.BinaryMarshaler = (*UserRoles)(nil)
var _ encoding.BinaryUnmarshaler = (*UserRoles)(nil)

const TableNameUserRoles = "sys_user_roles"

type UserRole struct {
	UserID uint32 `json:"sys_user_id" gorm:"column:sys_user_id"`
	RoleID uint32 `json:"sys_role_id" gorm:"column:sys_role_id"`
}

func (*UserRole) TableName() string {
	return TableNameUserRoles
}

type UserRoles struct {
	UserID uint32   `json:"user_id"`
	Roles  []uint32 `json:"roles"`
}

// Bytes binary marshaling
func (l *UserRoles) Bytes() []byte {
	if l == nil {
		return nil
	}
	bs, _ := json.Marshal(l)
	return bs
}

func (l *UserRoles) UnmarshalBinary(data []byte) error {
	if l == nil {
		return nil
	}
	return json.Unmarshal(data, l)
}

func (l *UserRoles) MarshalBinary() (data []byte, err error) {
	if l == nil {
		return nil, nil
	}
	return json.Marshal(*l)
}

// CacheUserRoles 缓存用户角色关联关系
func CacheUserRoles(db *gorm.DB, cacheClient cache.GlobalCache) error {
	var uRoles []*UserRole
	if err := db.Find(&uRoles).Error; err != nil {
		return err
	}

	if len(uRoles) == 0 {
		return nil
	}

	roleMap := make(map[uint32]*UserRoles)
	for _, ur := range uRoles {
		if _, ok := roleMap[ur.UserID]; !ok {
			roleMap[ur.UserID] = &UserRoles{
				UserID: ur.UserID,
				Roles:  make([]uint32, 0),
			}
		}
		roleMap[ur.UserID].Roles = append(roleMap[ur.UserID].Roles, ur.RoleID)
	}
	// 写入redis hash表中
	args := make([][]byte, 0, len(roleMap))
	for userId, ur := range roleMap {
		key := generateUserCacheKey(userId)
		args = append(args, []byte(key), ur.Bytes())
	}

	key := consts.UserRolesKey
	return cacheClient.HSet(context.Background(), key.String(), args...)
}

// CacheUserRole 缓存用户角色关联关系
func CacheUserRole(db *gorm.DB, cacheClient cache.GlobalCache, userID uint32) error {
	if userID == 0 {
		return nil
	}

	// 查询这个用户的全部角色
	var uRoles []*UserRole
	if err := db.Where("sys_user_id =?", userID).Find(&uRoles).Error; err != nil {
		return err
	}

	if len(uRoles) == 0 {
		// 清除缓存
		if err := cacheClient.HDel(context.Background(), consts.UserRolesKey.String(), generateUserCacheKey(userID)); err != nil {
			return err
		}
		return nil
	}

	roleIds := make([]uint32, 0, len(uRoles))
	for _, ur := range uRoles {
		roleIds = append(roleIds, ur.RoleID)
	}

	userRole := &UserRoles{
		UserID: userID,
		Roles:  roleIds,
	}
	fieldKey := generateUserCacheKey(userID)
	if err := cacheClient.HSet(context.Background(), consts.UserRolesKey.String(), []byte(fieldKey), userRole.Bytes()); err != nil {
		return err
	}

	return nil
}

func generateUserCacheKey(userID uint32) string {
	return strconv.FormatUint(uint64(userID), 10)
}

// CheckUserRoleExist 检查用户角色是否存在
func CheckUserRoleExist(ctx context.Context, cacheClient cache.GlobalCache, userID, roleID uint32) error {
	if userID == 0 || roleID == 0 {
		return nil
	}

	// 从redis hash表中获取
	key := generateUserCacheKey(userID)
	result, err := cacheClient.HGet(ctx, consts.UserRolesKey.String(), key)
	if err != nil {
		return perrors.ErrorPermissionDenied("用户角色关系已变化, 请重新登录")
	}

	if len(result) == 0 {
		return perrors.ErrorPermissionDenied("用户角色关系已变化, 请重新登录")
	}

	var ur UserRoles
	if err = json.Unmarshal(result, &ur); err != nil {
		return err
	}
	if ur.UserID != userID || !slices.Contains(ur.Roles, roleID) {
		return perrors.ErrorPermissionDenied("用户角色关系已变化, 请重新登录")
	}

	// 判断角色是否存在, 且状态为启用状态
	_, err = cacheClient.HGet(ctx, consts.RoleDisabledKey.String(), strconv.FormatUint(uint64(roleID), 10))
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	return nil
}

// CacheDisabledRoles 缓存角色禁用列表
func CacheDisabledRoles(db *gorm.DB, cacheClient cache.GlobalCache, roleIds ...uint32) error {
	wheres := []func(tx *gorm.DB) *gorm.DB{
		basescopes.StatusNotEQ(vo.StatusEnabled),
		basescopes.DeleteAtGT0(),
	}
	if len(roleIds) > 0 {
		wheres = append(wheres, basescopes.WhereInColumn("id", roleIds...))
	}
	var roles []*SysRole
	if err := db.Unscoped().Select("id,status,deleted_at").Scopes(wheres...).Find(&roles).Error; err != nil {
		return err
	}

	if len(roles) == 0 {
		// 删除找不到的角色
		if len(roleIds) > 0 {
			idsString := slices.To(roleIds, func(v uint32) string { return strconv.FormatUint(uint64(v), 10) })
			if err := cacheClient.HDel(context.Background(), consts.RoleDisabledKey.String(), idsString...); err != nil {
				return err
			}
		}
		return cacheClient.Del(context.Background(), consts.RoleDisabledKey.String())
	}

	args := make([][]byte, 0, len(roles))
	for _, role := range roles {
		args = append(args, []byte(strconv.FormatUint(uint64(role.ID), 10)), []byte("true"))
	}

	return cacheClient.HSet(context.Background(), consts.RoleDisabledKey.String(), args...)
}
