package do

import (
	"context"
	"encoding"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"

	"prometheus-manager/api/perrors"
	"prometheus-manager/pkg/helper/consts"
)

var _ schema.Tabler = (*SysAPI)(nil)
var _ encoding.BinaryMarshaler = (*ApiSimple)(nil)
var _ encoding.BinaryUnmarshaler = (*ApiSimple)(nil)

const TableNameSysApi = "sys_apis"

// SysAPI 系统api
type SysAPI struct {
	BaseModel
	Name   string     `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:api名称"`
	Path   string     `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__path,priority:1;comment:api路径"`
	Method string     `gorm:"column:method;type:varchar(16);not null;default:POST;comment:请求方法"`
	Status vo.Status  `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark string     `gorm:"column:remark;type:varchar(255);not null;default:这个API没有说明, 赶紧补充吧;comment:备注"`
	Module vo.Module  `gorm:"column:module;type:int;not null;default:0;comment:模块"`
	Domain vo.Domain  `gorm:"column:domain;type:int;not null;default:0;comment:领域"`
	Roles  []*SysRole `gorm:"many2many:sys_role_apis;comment:角色api"`
}

// TableName 表名
func (SysAPI) TableName() string {
	return TableNameSysApi
}

type ApiSimple struct {
	ID     uint   `gorm:"primary_key" json:"id"`
	Path   string `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__path,priority:1;comment:api路径"`
	Method string `gorm:"column:method;type:varchar(16);not null;comment:请求方法"`
}

func (l *ApiSimple) UnmarshalBinary(data []byte) error {
	if l == nil {
		return nil
	}

	return json.Unmarshal(data, l)
}

func (l *ApiSimple) MarshalBinary() (data []byte, err error) {
	if l == nil {
		return nil, nil
	}

	return json.Marshal(l)
}

// CacheAllApiSimple 缓存所有api简单信息
func CacheAllApiSimple(db *gorm.DB, cacheClient *redis.Client) error {
	var apiList []*ApiSimple
	if err := db.Model(&SysAPI{}).Where("status", vo.StatusEnabled).Find(&apiList).Error; err != nil {
		return err
	}

	apiCacheKey := consts.APICacheKey.String()
	// 删除redis hash表中所有数据
	if err := cacheClient.Del(context.Background(), apiCacheKey).Err(); err != nil {
		return err
	}

	if len(apiList) == 0 {
		return nil
	}

	// 写入redis hash表中
	args := make([]interface{}, 0, len(apiList))
	for _, api := range apiList {
		key := generateApiCacheKey(api.Path, api.Method)
		args = append(args, key, api.ID)
	}

	return cacheClient.HSet(context.Background(), apiCacheKey, args).Err()
}

// CacheApiSimple 缓存单个api简单信息
func CacheApiSimple(db *gorm.DB, cacheClient *redis.Client, apiIds ...uint32) error {
	if len(apiIds) == 0 {
		return nil
	}

	var apiList []*ApiSimple
	if err := db.Model(&SysAPI{}).Where("status", vo.StatusEnabled).Scopes(basescopes.InIds(apiIds...)).Find(&apiList).Error; err != nil {
		return err
	}

	if len(apiList) == 0 {
		return CacheAllApiSimple(db, cacheClient)
	}

	// 写入redis hash表中
	args := make([]interface{}, 0, len(apiList))
	for _, api := range apiList {
		key := generateApiCacheKey(api.Path, api.Method)
		args = append(args, key, api.ID)
	}

	key := consts.APICacheKey.String()
	return cacheClient.HSet(context.Background(), key, args).Err()
}

// GetApiIDByPathAndMethod 根据api路径和请求方法获取api id
func GetApiIDByPathAndMethod(cacheClient *redis.Client, path, method string) (uint64, error) {
	key := generateApiCacheKey(path, method)
	id, err := cacheClient.HGet(context.Background(), consts.APICacheKey.String(), key).Uint64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, perrors.ErrorUnauthorized("API暂未授权, 请联系管理员开通")
		}
		return 0, perrors.ErrorUnknown("系统错误")
	}
	return id, nil
}

// generateApiCacheKey 生成api缓存key
func generateApiCacheKey(path, method string) string {
	return method + ":" + path
}
