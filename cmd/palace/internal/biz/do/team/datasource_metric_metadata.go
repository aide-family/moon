package team

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/pkg/util/kv"
)

var _ do.DatasourceMetricMetadata = (*DatasourceMetricMetadata)(nil)

const metricDatasourceMetadataTableName = "team_datasource_metric_metadata"

type DatasourceMetricMetadata struct {
	do.TeamModel
	DatasourceMetricID uint32            `gorm:"column:datasource_metric_id;type:int unsigned;not null;comment:数据源ID;uniqueIndex:uk__datasource_metric_id__name" json:"datasource_metric_id"`
	DatasourceMetric   *DatasourceMetric `gorm:"foreignKey:DatasourceMetricID;references:ID" json:"datasource_metric"`
	Name               string            `gorm:"column:name;type:varchar(255);not null;comment:名称;uniqueIndex:uk__datasource_metric_id__name" json:"name"`
	Help               string            `gorm:"column:help;type:text;not null;comment:帮助" json:"help"`
	Type               string            `gorm:"column:type;type:varchar(32);not null;comment:类型" json:"type"`
	Labels             kv.StringMap      `gorm:"column:labels;type:text;not null;comment:标签" json:"labels"`
	Unit               string            `gorm:"column:unit;type:varchar(32);not null;comment:单位" json:"unit"`
}

func (d *DatasourceMetricMetadata) TableName() string {
	return metricDatasourceMetadataTableName
}

func (d *DatasourceMetricMetadata) GetDatasourceMetricID() uint32 {
	return d.DatasourceMetricID
}

func (d *DatasourceMetricMetadata) GetDatasourceMetric() do.DatasourceMetric {
	return d.DatasourceMetric
}

func (d *DatasourceMetricMetadata) GetName() string {
	return d.Name
}

func (d *DatasourceMetricMetadata) GetHelp() string {
	return d.Help
}

func (d *DatasourceMetricMetadata) GetType() string {
	return d.Type
}

func (d *DatasourceMetricMetadata) GetLabels() map[string]string {
	return d.Labels
}

func (d *DatasourceMetricMetadata) GetUnit() string {
	return d.Unit
}
