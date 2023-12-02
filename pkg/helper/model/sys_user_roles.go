package model

import (
	"context"
	"encoding"
	"encoding/json"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"prometheus-manager/api/perrors"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/util/slices"
)

const TableNameUserRoles = "sys_user_roles"

type UserRole struct {
	UserID uint `json:"sys_role_id" gorm:"column:sys_role_id"`
	RoleID uint `json:"sys_user_id" gorm:"column:sys_user_id"`
}

func (*UserRole) TableName() string {
	return TableNameUserRoles
}

type UserRoles struct {
	UserID uint   `json:"user_id"`
	Roles  []uint `json:"roles"`
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

var _ encoding.BinaryMarshaler = (*UserRoles)(nil)
var _ encoding.BinaryUnmarshaler = (*UserRoles)(nil)

// CacheUserRoles 缓存用户角色关联关系
func CacheUserRoles(db *gorm.DB, cacheClient *redis.Client) error {
	var uRoles []*UserRole
	if err := db.Find(&uRoles).Error; err != nil {
		return err
	}

	if len(uRoles) == 0 {
		return nil
	}

	roleMap := make(map[uint]*UserRoles)
	for _, ur := range uRoles {
		if _, ok := roleMap[ur.UserID]; !ok {
			roleMap[ur.UserID] = &UserRoles{
				UserID: ur.UserID,
				Roles:  make([]uint, 0),
			}
		}
		roleMap[ur.UserID].Roles = append(roleMap[ur.UserID].Roles, ur.RoleID)
	}
	// 写入redis hash表中
	args := make([]interface{}, 0, len(roleMap))
	for userId, ur := range roleMap {
		key := generateKey(userId)
		args = append(args, key, ur)
	}

	key := consts.UserRolesKey
	return cacheClient.HSet(context.Background(), key.String(), args).Err()
}

func generateKey(userID uint) string {
	return strconv.FormatUint(uint64(userID), 10)
}

// CheckUserRoleExist 检查用户角色是否存在
func CheckUserRoleExist(ctx context.Context, cacheClient *redis.Client, userID uint, roleID string) error {
	if userID == 0 || roleID == "" {
		return nil
	}

	// 从redis hash表中获取
	key := generateKey(userID)
	result, err := cacheClient.HGet(ctx, consts.UserRolesKey.String(), key).Result()
	if err != nil {
		return perrors.ErrorPermissionDenied("用户角色关系已变化, 请重新登录")
	}

	if result == "" {
		return perrors.ErrorPermissionDenied("用户角色关系已变化, 请重新登录")
	}

	rID, err := strconv.ParseUint(roleID, 10, 32)
	if err != nil {
		return err
	}
	var ur UserRoles
	if err = json.Unmarshal([]byte(result), &ur); err != nil {
		return err
	}
	if ur.UserID != userID || !slices.Contains(ur.Roles, uint(rID)) {
		return perrors.ErrorPermissionDenied("用户角色关系已变化, 请重新登录")
	}

	return nil
}
