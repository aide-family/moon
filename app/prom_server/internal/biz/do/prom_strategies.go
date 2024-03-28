package do

import (
	"encoding/json"
	"strings"

	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/pkg/strategy"
)

const TableNamePromStrategy = "prom_strategies"

const (
	PromStrategyFieldGroupID                  = "group_id"
	PromStrategyFieldAlert                    = "alert"
	PromStrategyFieldExpr                     = "expr"
	PromStrategyFieldFor                      = "for"
	PromStrategyFieldLabels                   = "labels"
	PromStrategyFieldAnnotations              = "annotations"
	PromStrategyFieldAlertLevelID             = "alert_level_id"
	PromStrategyFieldStatus                   = "status"
	PromStrategyFieldRemark                   = "remark"
	PromStrategyFieldMaxSuppress              = "max_suppress"
	PromStrategyFieldSendRecover              = "send_recover"
	PromStrategyFieldSendInterval             = "send_interval"
	PromStrategyFieldEndpointID               = "endpoint_id"
	PromStrategyFieldCreateBy                 = "create_by"
	PromStrategyPreloadFieldAlarmPages        = "AlarmPages"
	PromStrategyPreloadFieldCategories        = "Categories"
	PromStrategyPreloadFieldAlertLevel        = "AlertLevel"
	PromStrategyPreloadFieldGroupInfo         = "GroupInfo"
	PromStrategyPreloadFieldPromNotifies      = "PromNotifies"
	PromStrategyPreloadFieldPromNotifyUpgrade = "PromNotifyUpgrade"
	PromStrategyPreloadFieldEndpoint          = "Endpoint"
	PromStrategyPreloadFieldCreateByUser      = "CreateByUser"
	PromStrategyPreloadFieldTemplate          = "Templates"
)

// StrategyInGroupIds 策略组ID
func StrategyInGroupIds(ids ...uint32) basescopes.ScopeMethod {
	// 过滤0值
	newIds := slices.Filter(ids, func(id uint32) bool { return id > 0 })
	return basescopes.WhereInColumn(PromStrategyFieldGroupID, newIds...)
}

// StrategyAlertLike 策略名称匹配
func StrategyAlertLike(keyword string) basescopes.ScopeMethod {
	return basescopes.WhereLikePrefixKeyword(keyword, PromStrategyFieldAlert)
}

// StrategyPreloadEndpoint 预加载endpoint
func StrategyPreloadEndpoint() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromStrategyPreloadFieldEndpoint)
	}
}

// StrategyPreloadAlarmPages 预加载alarm_pages
func StrategyPreloadAlarmPages() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromStrategyPreloadFieldAlarmPages)
	}
}

// StrategyPreloadCategories 预加载categories
func StrategyPreloadCategories() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromStrategyPreloadFieldCategories)
	}
}

// StrategyPreloadAlertLevel 预加载alert_level
func StrategyPreloadAlertLevel() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromStrategyPreloadFieldAlertLevel)
	}
}

// StrategyPreloadPromNotifies 预加载prom_notifies
func StrategyPreloadPromNotifies(preloadKeys ...string) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(preloadKeys) == 0 {
			return db.Preload(PromStrategyPreloadFieldPromNotifies)
		}
		tx := db
		for _, key := range preloadKeys {
			tx = tx.Preload(strings.Join([]string{PromStrategyPreloadFieldPromNotifies, key}, "."))
		}
		return tx
	}
}

// StrategyPreloadPromNotifyUpgrade 预加载prom_notify_upgrade
func StrategyPreloadPromNotifyUpgrade() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromStrategyPreloadFieldPromNotifyUpgrade)
	}
}

// StrategyPreloadGroupInfo 预加载group_info
func StrategyPreloadGroupInfo() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromStrategyPreloadFieldGroupInfo)
	}
}

// StrategyPreloadTemplate 预加载template
func StrategyPreloadTemplate() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromStrategyPreloadFieldTemplate)
	}
}

// PromStrategy mapped from table <prom_strategies>
type PromStrategy struct {
	BaseModel
	GroupID      uint32                `gorm:"column:group_id;type:int unsigned;not null;comment:所属规则组ID;uniqueIndex:idx__prom_strategy__group_id__alert" json:"group_id"`
	Alert        string                `gorm:"column:alert;type:varchar(64);not null;comment:规则名称;uniqueIndex:idx__prom_strategy__group_id__alert" json:"alert"`
	Expr         string                `gorm:"column:expr;type:text;not null;comment:prom ql" json:"expr"`
	For          string                `gorm:"column:for;type:varchar(64);not null;default:10s;comment:持续时间" json:"for"`
	Labels       *strategy.Labels      `gorm:"column:labels;type:json;not null;comment:标签" json:"labels"`
	Annotations  *strategy.Annotations `gorm:"column:annotations;type:json;not null;comment:告警文案" json:"annotations"`
	AlertLevelID uint32                `gorm:"column:alert_level_id;type:int;not null;index:idx__alert_level_id,priority:1;comment:告警等级dict ID" json:"alert_level_id"`
	Status       vobj.Status           `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态: 1启用;2禁用" json:"status"`
	Remark       string                `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark"`

	AlarmPages []*SysDict         `gorm:"References:ID;foreignKey:ID;joinForeignKey:PromStrategyID;joinReferences:AlarmPageID;many2many:prom_strategy_alarm_pages" json:"-"`
	Categories []*SysDict         `gorm:"References:ID;foreignKey:ID;joinForeignKey:PromStrategyID;joinReferences:DictID;many2many:prom_strategy_categories" json:"-"`
	AlertLevel *SysDict           `gorm:"foreignKey:AlertLevelID" json:"-"`
	GroupInfo  *PromStrategyGroup `gorm:"foreignKey:GroupID" json:"-"`

	// 通知对象
	PromNotifies []*PromAlarmNotify `gorm:"many2many:prom_strategy_notifies;comment:通知对象" json:"-"`
	// 告警升级后的通知对象
	PromNotifyUpgrade []*PromAlarmNotify `gorm:"many2many:prom_strategy_notify_upgrades;comment:告警升级后的通知对象" json:"-"`
	// 最大抑制时长(s)
	MaxSuppress string `gorm:"column:max_suppress;type:varchar(255);not null;default:1m;comment:最大抑制时长(s)" json:"max_suppress"`
	// 是否发送告警恢复通知
	SendRecover vobj.IsSendRecover `gorm:"column:send_recover;type:tinyint;not null;default:0;comment:是否发送告警恢复通知" json:"send_recover"`
	// 发送告警时间间隔(s), 默认为for的10倍时间分钟, 用于长时间未消警情况
	SendInterval string `gorm:"column:send_interval;type:varchar(255);not null;default:1m;comment:发送告警时间间隔(s), 默认为for的10倍时间分钟, 用于长时间未消警情况" json:"send_interval"`

	// 数据源
	EndpointID uint32    `gorm:"column:endpoint_id;type:int unsigned;not null;default:0;comment:数据源ID" json:"endpoint_id"`
	Endpoint   *Endpoint `gorm:"foreignKey:EndpointID" json:"-"`

	// 创建人ID
	CreateBy     uint32   `gorm:"column:create_by;type:int;not null;comment:创建人ID"`
	CreateByUser *SysUser `gorm:"foreignKey:CreateBy" json:"-"`

	Templates []*PromStrategyNotifyTemplate `gorm:"foreignKey:StrategyID" json:"-"`
}

// TableName PromStrategy's table name
func (*PromStrategy) TableName() string {
	return TableNamePromStrategy
}

// GetAlertLevel 获取告警等级
func (p *PromStrategy) GetAlertLevel() *SysDict {
	if p == nil {
		return nil
	}
	return p.AlertLevel
}

// GetAlarmPages 获取告警页面
func (p *PromStrategy) GetAlarmPages() []*SysDict {
	if p == nil {
		return nil
	}
	return p.AlarmPages
}

// GetLabels 获取标签
func (p *PromStrategy) GetLabels() *strategy.Labels {
	if p == nil {
		return nil
	}
	return p.Labels
}

// GetAnnotations 获取告警文案
func (p *PromStrategy) GetAnnotations() *strategy.Annotations {
	if p == nil {
		return nil
	}
	return p.Annotations
}

// GetCategories 获取分类
func (p *PromStrategy) GetCategories() []*SysDict {
	if p == nil {
		return nil
	}
	return p.Categories
}

// GetPromNotifies 获取通知对象
func (p *PromStrategy) GetPromNotifies() []*PromAlarmNotify {
	if p == nil {
		return nil
	}
	return p.PromNotifies
}

// GetPromNotifyUpgrade 获取告警升级后的通知对象
func (p *PromStrategy) GetPromNotifyUpgrade() []*PromAlarmNotify {
	if p == nil {
		return nil
	}
	return p.PromNotifyUpgrade
}

// GetGroupInfo 获取所属规则组
func (p *PromStrategy) GetGroupInfo() *PromStrategyGroup {
	if p == nil {
		return nil
	}
	return p.GroupInfo
}

// GetEndpoint 获取数据源
func (p *PromStrategy) GetEndpoint() *Endpoint {
	if p == nil {
		return nil
	}
	return p.Endpoint
}

// GetTemplates 获取模板
func (p *PromStrategy) GetTemplates() []*PromStrategyNotifyTemplate {
	if p == nil {
		return nil
	}
	return p.Templates
}

// ToMap to map[string]any
func (p *PromStrategy) ToMap() map[string]any {
	if p == nil {
		return nil
	}
	bs, _ := json.Marshal(p)
	pMap := make(map[string]any)
	_ = json.Unmarshal(bs, &pMap)
	pMap["labels"] = p.GetLabels().String()
	pMap["annotations"] = p.GetAnnotations().String()
	return pMap
}
