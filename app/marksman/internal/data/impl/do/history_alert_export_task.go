package do

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"
)

const TableNameHistoryAlertExportTask = "history_alert_export_tasks"

// HistoryAlertExportFilterConfig stores export query filters as JSON.
type HistoryAlertExportFilterConfig struct {
	StartAtUnix       int64                  `json:"start_at_unix"`
	EndAtUnix         int64                  `json:"end_at_unix"`
	Status            enum.AlertEventStatus  `json:"status"`
	StrategyGroupUIDs []int64                `json:"strategy_group_uids"`
	LevelUIDs         []int64                `json:"level_uids"`
	StrategyUIDs      []int64                `json:"strategy_uids"`
	DatasourceUIDs    []int64                `json:"datasource_uids"`
	Keyword           string                 `json:"keyword"`
}

func (c *HistoryAlertExportFilterConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

func (c *HistoryAlertExportFilterConfig) Scan(value any) error {
	if value == nil {
		*c = HistoryAlertExportFilterConfig{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("history_alert_export_filter_config: expected []byte or string")
	}
	if len(b) == 0 {
		*c = HistoryAlertExportFilterConfig{}
		return nil
	}
	return json.Unmarshal(b, c)
}

type HistoryAlertExportTask struct {
	BaseModel
	NamespaceUID  snowflake.ID                    `gorm:"column:namespace_uid;index"`
	Status        int32                           `gorm:"column:status;index"`
	FilterConfig  *HistoryAlertExportFilterConfig `gorm:"column:filter_config;type:json"`
	TotalRows     int64                           `gorm:"column:total_rows"`
	ProcessedRows int64                           `gorm:"column:processed_rows"`
	FileName      string                          `gorm:"column:file_name;type:varchar(255)"`
	FilePath      string                          `gorm:"column:file_path;type:varchar(512)"`
	ErrorMessage  string                          `gorm:"column:error_message;type:text"`
	CompletedAt   *time.Time                      `gorm:"column:completed_at"`
}

func (HistoryAlertExportTask) TableName() string {
	return TableNameHistoryAlertExportTask
}

func (t *HistoryAlertExportTask) WithNamespace(namespace snowflake.ID) *HistoryAlertExportTask {
	t.NamespaceUID = namespace
	return t
}
