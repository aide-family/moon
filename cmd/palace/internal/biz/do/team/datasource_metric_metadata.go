package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/util/kv"
)

var _ do.DatasourceMetricMetadata = (*DatasourceMetricMetadata)(nil)

const metricDatasourceMetadataTableName = "team_datasource_metric_metadata"

type DatasourceMetricMetadata struct {
	do.TeamModel
	DatasourceMetricID uint32                   `gorm:"column:datasource_metric_id;type:int unsigned;not null;comment:datasource ID;uniqueIndex:uk__datasource_metric_id__name" json:"datasource_metric_id"`
	DatasourceMetric   *DatasourceMetric        `gorm:"foreignKey:DatasourceMetricID;references:ID" json:"datasource_metric"`
	Name               string                   `gorm:"column:name;type:varchar(255);not null;comment:name;uniqueIndex:uk__datasource_metric_id__name" json:"name"`
	Help               string                   `gorm:"column:help;type:text;not null;comment:help" json:"help"`
	Type               string                   `gorm:"column:type;type:varchar(32);not null;comment:type" json:"type"`
	Labels             kv.Map[string, []string] `gorm:"column:labels;type:text;not null;comment:labels" json:"labels"`
	Unit               string                   `gorm:"column:unit;type:varchar(32);not null;comment:unit" json:"unit"`
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

func (d *DatasourceMetricMetadata) GetLabels() map[string][]string {
	return d.Labels
}

func (d *DatasourceMetricMetadata) GetUnit() string {
	return d.Unit
}
