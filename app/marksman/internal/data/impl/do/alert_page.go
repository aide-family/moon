package do

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// AlertPageFilterConfig is the JSON shape stored in alert_pages.filter_config.
type AlertPageFilterConfig struct {
	StrategyGroupUIDs []int64 `json:"strategy_group_uids"`
	LevelUIDs         []int64 `json:"level_uids"`
	StrategyUIDs      []int64 `json:"strategy_uids"`
}

// Value implements driver.Valuer for JSON storage.
func (c *AlertPageFilterConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan implements sql.Scanner for JSON storage.
func (c *AlertPageFilterConfig) Scan(value any) error {
	if value == nil {
		*c = AlertPageFilterConfig{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("alert_page_filter_config: expected []byte or string")
	}
	if len(b) == 0 {
		*c = AlertPageFilterConfig{}
		return nil
	}
	return json.Unmarshal(b, c)
}

type AlertPage struct {
	BaseModel
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
	NamespaceUID snowflake.ID  `gorm:"column:namespace_uid;index"`
	Name        string        `gorm:"column:name;type:varchar(100)"`
	Color       string        `gorm:"column:color;type:varchar(32);default:''"`
	SortOrder   int32         `gorm:"column:sort_order;default:0"`
	FilterConfig *AlertPageFilterConfig `gorm:"column:filter_config;type:json"`
}

func (AlertPage) TableName() string {
	return "alert_pages"
}

func (a *AlertPage) WithNamespace(namespace snowflake.ID) *AlertPage {
	a.NamespaceUID = namespace
	return a
}
