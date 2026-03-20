package do

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
)

const (
	TableNameNotificationGroupSubscription = "notification_group_subscriptions"
)

// StrategyLevelPairDO is the JSON element for (strategy_uid, level_uid) in subscription.
type StrategyLevelPairDO struct {
	StrategyUID int64 `json:"strategy_uid"`
	LevelUID    int64 `json:"level_uid"`
}

// StrategyLevelPairsDO is the JSON array stored in notification_group_subscriptions.strategy_levels.
type StrategyLevelPairsDO []StrategyLevelPairDO

// Value implements driver.Valuer for JSON storage.
func (c StrategyLevelPairsDO) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan implements sql.Scanner for JSON storage.
func (c *StrategyLevelPairsDO) Scan(value any) error {
	if value == nil {
		*c = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("strategy_level_pairs_do: expected []byte or string")
	}
	if len(b) == 0 {
		*c = nil
		return nil
	}
	return json.Unmarshal(b, c)
}

type NotificationGroupSubscription struct {
	BaseModel
	NotificationGroupUID snowflake.ID                `gorm:"column:notification_group_uid;uniqueIndex:idx__notification_group_subscriptions__notification_group_uid"`
	StrategyGroupUIDs    *safety.Slice[int64]        `gorm:"column:strategy_group_uids;type:json;"`
	StrategyUIDs         *safety.Slice[int64]        `gorm:"column:strategy_uids;type:json;"`
	StrategyLevels       StrategyLevelPairsDO        `gorm:"column:strategy_levels;type:json;"`
	Labels               *safety.Map[string, string] `gorm:"column:labels;type:json;"`
	ExcludeLabels        *safety.Map[string, string] `gorm:"column:exclude_labels;type:json;"`
	DatasourceUIDs       *safety.Slice[int64]        `gorm:"column:datasource_uids;type:json;"`
}

func (NotificationGroupSubscription) TableName() string {
	return TableNameNotificationGroupSubscription
}
