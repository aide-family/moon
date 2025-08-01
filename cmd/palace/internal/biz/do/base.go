// Package do is a data object package for kratos.
package do

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/validate"
)

type userDoContextKey struct{}

func WithUserDoContext(ctx context.Context, userDo User) context.Context {
	return context.WithValue(ctx, userDoContextKey{}, userDo)
}

func GetUserDoContext(ctx context.Context) (userDo User, ok bool) {
	userDo, ok = ctx.Value(userDoContextKey{}).(User)
	return
}

type menuDoContextKey struct{}

func WithMenuDoContext(ctx context.Context, menuDo Menu) context.Context {
	return context.WithValue(ctx, menuDoContextKey{}, menuDo)
}

func GetMenuDoContext(ctx context.Context) (menuDo Menu, ok bool) {
	menuDo, ok = ctx.Value(menuDoContextKey{}).(Menu)
	return
}

type GetUserFun func(id uint32) User

type GetTeamFun func(id uint32) Team

type GetTeamMemberFun func(id uint32) TeamMember

type GetTeamMembersFun func(ids []uint32) []TeamMember

type HasTableFun func(teamId uint32, tableName string) bool

type CacheTableFlag func(teamId uint32, tableName string) error

var registerGetUserFuncOnce sync.Once
var getUser GetUserFun

func RegisterGetUserFunc(getUserFunc GetUserFun) {
	registerGetUserFuncOnce.Do(func() {
		getUser = getUserFunc
	})
}

func GetUser(id uint32) User {
	return getUser(id)
}

var registerGetTeamFuncOnce sync.Once
var getTeam GetTeamFun

func RegisterGetTeamFunc(getTeamFunc GetTeamFun) {
	registerGetTeamFuncOnce.Do(func() {
		getTeam = getTeamFunc
	})
}

func GetTeam(id uint32) Team {
	return getTeam(id)
}

var registerGetTeamMemberFuncOnce sync.Once
var getTeamMember GetTeamMemberFun
var getTeamMembers GetTeamMembersFun

func RegisterGetTeamMemberFunc(getTeamMemberFunc GetTeamMemberFun, getTeamMembersFunc GetTeamMembersFun) {
	registerGetTeamMemberFuncOnce.Do(func() {
		getTeamMember = getTeamMemberFunc
		getTeamMembers = getTeamMembersFunc
	})
}

func GetTeamMember(id uint32) TeamMember {
	return getTeamMember(id)
}

func GetTeamMembers(ids []uint32) []TeamMember {
	return getTeamMembers(ids)
}

var registerHasTableFuncOnce sync.Once
var hasTable HasTableFun = nil

func RegisterHasTableFunc(hasTableFunc HasTableFun) {
	registerHasTableFuncOnce.Do(func() {
		hasTable = hasTableFunc
	})
}

var registerCacheTableFlagOnce sync.Once
var cacheTableFlag CacheTableFlag = nil

func RegisterCacheTableFlag(cacheTableFlagFunc CacheTableFlag) {
	registerCacheTableFlagOnce.Do(func() {
		cacheTableFlag = cacheTableFlagFunc
	})
}

type ORMModel interface {
	sql.Scanner
	driver.Valuer
}

type Base interface {
	GetID() uint32
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() soft_delete.DeletedAt
	GetContext() context.Context
	WithContext(context.Context)
}

type Creator interface {
	Base
	GetCreatorID() uint32
	GetCreator() User
	GetCreatorMember() TeamMember
}

type TeamBase interface {
	Creator
	GetTeamID() uint32
	GetTeam() Team
}

var _ Base = (*BaseModel)(nil)

// BaseModel represents the base model for GORM
type BaseModel struct {
	ctx context.Context `gorm:"-"`

	ID        uint32                `gorm:"column:id;primaryKey;autoIncrement" json:"id,omitempty"`
	CreatedAt time.Time             `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:creation time" json:"created_at,omitempty"`
	UpdatedAt time.Time             `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:update time" json:"updated_at,omitempty"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;" json:"deleted_at,omitempty"`
}

func (u *BaseModel) GetID() uint32 {
	if u == nil {
		return 0
	}
	return u.ID
}

func (u *BaseModel) GetCreatedAt() time.Time {
	if u == nil {
		return time.Time{}
	}
	return u.CreatedAt
}

func (u *BaseModel) GetUpdatedAt() time.Time {
	if u == nil {
		return time.Time{}
	}
	return u.UpdatedAt
}

func (u *BaseModel) GetDeletedAt() soft_delete.DeletedAt {
	if u == nil {
		return 0
	}
	return u.DeletedAt
}

// WithContext sets the context
func (u *BaseModel) WithContext(ctx context.Context) {
	u.ctx = ctx
}

// GetContext gets the context
func (u *BaseModel) GetContext() context.Context {
	if validate.IsNil(u.ctx) {
		panic("context is nil")
	}
	return u.ctx
}

var _ Creator = (*CreatorModel)(nil)

type CreatorModel struct {
	BaseModel
	CreatorID uint32 `gorm:"column:creator;type:int unsigned;not null;comment:creator" json:"creator_id,omitempty"`
}

func (u *CreatorModel) GetCreatorMember() TeamMember {
	if u == nil {
		return nil
	}
	return getTeamMember(u.GetCreatorID())
}

func (u *CreatorModel) GetCreatorID() uint32 {
	if u == nil {
		return 0
	}
	return u.CreatorID
}

func (u *CreatorModel) GetCreator() (user User) {
	if u == nil {
		return nil
	}
	return getUser(u.GetCreatorID())
}

func (u *CreatorModel) BeforeCreate(tx *gorm.DB) (err error) {
	var exist bool
	if u.CreatorID <= 0 {
		u.CreatorID, exist = permission.GetUserIDByContext(u.GetContext())
		if !exist || u.CreatorID == 0 {
			return merr.ErrorInternalServer("user id not found")
		}
	}
	tx.WithContext(u.GetContext())
	return
}

var _ TeamBase = (*TeamModel)(nil)

type TeamModel struct {
	CreatorModel
	TeamID uint32 `gorm:"column:team_id;type:int unsigned;not null;comment:team ID" json:"team_id,omitempty"`
}

func (u *TeamModel) GetTeam() Team {
	if u == nil {
		return nil
	}
	return getTeam(u.GetTeamID())
}

func (u *TeamModel) GetTeamID() uint32 {
	if u == nil {
		return 0
	}
	return u.TeamID
}

func (u *TeamModel) BeforeCreate(tx *gorm.DB) (err error) {
	var exist bool
	if err := u.CreatorModel.BeforeCreate(tx); err != nil {
		return err
	}
	if u.TeamID <= 0 {
		u.TeamID, exist = permission.GetTeamIDByContext(u.GetContext())
		if !exist || u.TeamID == 0 {
			return merr.ErrorInternalServer("team id not found")
		}
	}
	return
}

func HasTable(teamID uint32, tx *gorm.DB, tableName string) bool {
	if validate.IsNil(hasTable) {
		if validate.IsNil(tx) {
			return false
		}
		if !tx.Migrator().HasTable(tableName) {
			return false
		}
		if cacheTableFlag != nil {
			_ = cacheTableFlag(teamID, tableName)
		}
		return true
	}
	return hasTable(teamID, tableName)
}

func CreateTable(teamID uint32, tx *gorm.DB, tableName string, model any) error {
	if err := tx.Table(tableName).AutoMigrate(model); err != nil {
		return err
	}
	if validate.IsNil(cacheTableFlag) {
		return nil
	}
	return cacheTableFlag(teamID, tableName)
}

// GetPreviousMonday returns the Monday of the week containing the given date
func GetPreviousMonday(t time.Time) time.Time {
	// Calculate the offset to Monday
	offset := int(time.Monday - t.Weekday())
	if offset > 0 { // If the current day is Sunday (Weekday=0), offset=1, need to subtract 7 days
		offset -= 7
	}
	return t.AddDate(0, 0, offset)
}
