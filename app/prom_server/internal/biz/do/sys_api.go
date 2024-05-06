package do

import (
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/soft_delete"

	"github.com/aide-family/moon/api/perrors"
	"github.com/aide-family/moon/pkg/helper/consts"
)

var _ schema.Tabler = (*SysAPI)(nil)
var _ encoding.BinaryMarshaler = (*ApiSimple)(nil)
var _ encoding.BinaryUnmarshaler = (*ApiSimple)(nil)

const TableNameSysApi = "sys_apis"

const (
	SysAPIFieldName         = "name"
	SysAPIFieldPath         = "path"
	SysAPIFieldMethod       = "method"
	SysAPIFieldStatus       = "status"
	SysAPIFieldRemark       = "remark"
	SysAPIFieldModule       = "module"
	SysAPIFieldDomain       = "domain"
	SysAPIPreloadFieldRoles = "Roles"
)

type SysAPIField string
type SysAPIWithField string

// SysAPI 系统api
type SysAPI struct {
	BaseModel
	Name   string      `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__sa__name,priority:1;comment:api名称"`
	Path   string      `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__sa__path,priority:1;comment:api路径"`
	Method string      `gorm:"column:method;type:varchar(16);not null;default:POST;comment:请求方法"`
	Status vobj.Status `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark string      `gorm:"column:remark;type:varchar(255);not null;default:这个API没有说明, 赶紧补充吧;comment:备注"`
	Module vobj.Module `gorm:"column:module;type:int;not null;default:0;comment:模块"`
	Domain vobj.Domain `gorm:"column:domain;type:int;not null;default:0;comment:领域"`
	Roles  []*SysRole  `gorm:"many2many:sys_role_apis;comment:角色api"`
}

// TableName 表名
func (s *SysAPI) TableName() string {
	return TableNameSysApi
}

// SysAPIPreloadRoles 预加载角色
func SysAPIPreloadRoles() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(SysAPIPreloadFieldRoles)
	}
}

// SysApiLike like name or path
func SysApiLike(keyword string) basescopes.ScopeMethod {
	return basescopes.WhereLikePrefixKeyword(keyword, SysAPIFieldName, SysAPIFieldPath)
}

type ApiSimple struct {
	ID     uint   `gorm:"primary_key" json:"id"`
	Path   string `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__as__path,priority:1;comment:api路径"`
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
func CacheAllApiSimple(db *gorm.DB, cacheClient cache.GlobalCache) error {
	var apiList []*ApiSimple
	if err := db.Model(&SysAPI{}).Where("status", vobj.StatusEnabled).Find(&apiList).Error; err != nil {
		return err
	}

	apiCacheKey := consts.APICacheKey.String()
	// 删除redis hash表中所有数据
	if err := cacheClient.Del(context.Background(), apiCacheKey); err != nil {
		return err
	}

	if len(apiList) == 0 {
		return nil
	}

	// 写入redis hash表中
	args := make([][]byte, 0, len(apiList))
	for _, api := range apiList {
		key := generateApiCacheKey(api.Path, api.Method)
		args = append(args, []byte(key), []byte(strconv.FormatUint(uint64(api.ID), 10)))
	}

	return cacheClient.HSet(context.Background(), apiCacheKey, args...)
}

// CacheApiSimple 缓存单个api简单信息
func CacheApiSimple(db *gorm.DB, cacheClient cache.GlobalCache, apiIds ...uint32) error {
	if len(apiIds) == 0 {
		return nil
	}

	var apiList []*ApiSimple
	if err := db.Model(&SysAPI{}).Where("status", vobj.StatusEnabled).Scopes(basescopes.InIds(apiIds...)).Find(&apiList).Error; err != nil {
		return err
	}

	if len(apiList) == 0 {
		return CacheAllApiSimple(db, cacheClient)
	}

	// 写入redis hash表中
	args := make([][]byte, 0, len(apiList))
	for _, api := range apiList {
		key := generateApiCacheKey(api.Path, api.Method)
		args = append(args, []byte(key), []byte(strconv.FormatUint(uint64(api.ID), 10)))
	}

	key := consts.APICacheKey.String()
	return cacheClient.HSet(context.Background(), key, args...)
}

// GetApiIDByPathAndMethod 根据api路径和请求方法获取api id
func GetApiIDByPathAndMethod(cacheClient cache.GlobalCache, path, method string) (uint64, error) {
	key := generateApiCacheKey(path, method)
	idBytes, err := cacheClient.HGet(context.Background(), consts.APICacheKey.String(), key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, perrors.ErrorUnauthorized("API暂未授权, 请联系管理员开通")
		}
		return 0, perrors.ErrorUnknown("系统错误")
	}

	return strconv.ParseUint(string(idBytes), 10, 64)
}

// generateApiCacheKey 生成api缓存key
func generateApiCacheKey(path, method string) string {
	return method + ":" + path
}

// SetID 设置ID字段
func (s *SysAPI) SetID(id uint32) *SysAPI {
	if s == nil {
		return nil
	}
	s.ID = id
	return s
}

// SetCreatedAt 设置创建时间字段
func (s *SysAPI) SetCreatedAt(createdAt time.Time) *SysAPI {
	if s == nil {
		return nil
	}
	s.CreatedAt = createdAt
	return s
}

// SetUpdatedAt 设置更新时间字段
func (s *SysAPI) SetUpdatedAt(updatedAt time.Time) *SysAPI {
	if s == nil {
		return nil
	}
	s.UpdatedAt = updatedAt
	return s
}

// SetDeletedAt 设置删除时间字段
func (s *SysAPI) SetDeletedAt(deletedAt soft_delete.DeletedAt) *SysAPI {
	if s == nil {
		return nil
	}
	s.DeletedAt = deletedAt
	return s
}

// GetName 获取API名称
func (s *SysAPI) GetName() string {
	if s == nil {
		return ""
	}
	return s.Name
}

// SetName 设置API名称
func (s *SysAPI) SetName(name string) *SysAPI {
	if s == nil {
		return nil
	}
	s.Name = name
	return s
}

// GetPath 获取API路径
func (s *SysAPI) GetPath() string {
	if s == nil {
		return ""
	}
	return s.Path
}

// SetPath 设置API路径
func (s *SysAPI) SetPath(path string) *SysAPI {
	if s == nil {
		return nil
	}
	s.Path = path
	return s
}

// GetMethod 获取请求方法
func (s *SysAPI) GetMethod() string {
	if s == nil {
		return ""
	}
	return s.Method
}

// SetMethod 设置请求方法
func (s *SysAPI) SetMethod(method string) *SysAPI {
	if s == nil {
		return nil
	}
	s.Method = method
	return s
}

// GetStatus 获取状态
func (s *SysAPI) GetStatus() vobj.Status {
	if s == nil {
		return vobj.StatusUnknown
	}
	return s.Status
}

// SetStatus 设置状态
func (s *SysAPI) SetStatus(status vobj.Status) *SysAPI {
	if s == nil {
		return nil
	}
	s.Status = status
	return s
}

// GetRemark 获取备注
func (s *SysAPI) GetRemark() string {
	if s == nil {
		return ""
	}
	return s.Remark
}

// SetRemark 设置备注
func (s *SysAPI) SetRemark(remark string) *SysAPI {
	if s == nil {
		return nil
	}
	s.Remark = remark
	return s
}

// GetModule 获取模块
func (s *SysAPI) GetModule() vobj.Module {
	if s == nil {
		return vobj.ModuleOther
	}
	return s.Module
}

// SetModule 设置模块
func (s *SysAPI) SetModule(module vobj.Module) *SysAPI {
	if s == nil {
		return nil
	}
	s.Module = module
	return s
}

// GetDomain 获取领域
func (s *SysAPI) GetDomain() vobj.Domain {
	if s == nil {
		return vobj.DomainOther
	}
	return s.Domain
}

// SetDomain 设置领域
func (s *SysAPI) SetDomain(domain vobj.Domain) *SysAPI {
	if s == nil {
		return nil
	}
	s.Domain = domain
	return s
}

// GetRoles 获取角色
func (s *SysAPI) GetRoles() []*SysRole {
	if s == nil {
		return nil
	}
	return s.Roles
}

// SetRoles 设置角色
// 注意：设置角色可能涉及数据库操作，这里仅简单赋值，不考虑关联关系处理
func (s *SysAPI) SetRoles(roles []*SysRole) *SysAPI {
	if s == nil {
		return nil
	}
	s.Roles = roles
	return s
}
