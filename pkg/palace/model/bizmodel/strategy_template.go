package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"

	"google.golang.org/protobuf/types/known/durationpb"
)

const TableNameStrategyTemplate = "strategy_template"

// StrategyTemplate mapped from table <StrategyTemplate>
type StrategyTemplate struct {
	model.AllFieldModel
	Alert       string              `gorm:"column:alert;type:varchar(64);not null;comment:策略名称" json:"alert"`
	Expr        string              `gorm:"column:expr;type:text;not null;comment:告警表达式" json:"expr"`
	For         durationpb.Duration `gorm:"column:for;type:varchar(64);not null;comment:告警持续时间" json:"for"`
	Count       uint32              `gorm:"column:count;type:int unsigned;not null;comment:持续次数" json:"count"`
	SustainType vobj.Sustain        `gorm:"column:sustain_type;type:int(11);not null;comment:持续类型" json:"sustain_type"`
	Labels      vobj.Labels         `gorm:"column:labels;type:JSON;not null;comment:标签" json:"labels"`
	Annotations vobj.Annotations    `gorm:"column:annotations;type:JSON;not null;comment:注解" json:"annotations"`
	Interval    durationpb.Duration `gorm:"column:interval;type:varchar(64);not null;comment:执行频率" json:"interval"`
	Remark      string              `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status      vobj.Status         `gorm:"column:status;type:int;not null;comment:策略状态" json:"status"`
}

// String json string
func (c *StrategyTemplate) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *StrategyTemplate) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *StrategyTemplate) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName StrategyTemplate's table name
func (*StrategyTemplate) TableName() string {
	return TableNameStrategyTemplate
}
