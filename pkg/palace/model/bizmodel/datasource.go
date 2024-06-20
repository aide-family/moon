package bizmodel

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

const TableNameDatasource = "datasource"

// Datasource mapped from table <datasource>
type Datasource struct {
	ID          uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	Name        string                `gorm:"column:name;type:varchar(64);not null;comment:数据源名称" json:"name"` // 数据源名称
	Category    vobj.DatasourceType   `gorm:"column:category;type:int;not null;comment:数据源类型" json:"category"` // 数据源类型
	StorageType vobj.StorageType      `gorm:"column:storage_type;type:int;not null;comment:存储类型" json:"storage_type"`
	Config      string                `gorm:"column:config;type:varchar(255);not null;comment:数据源配置参数" json:"config"`                            // 数据源配置参数
	Endpoint    string                `gorm:"column:endpoint;type:varchar(255);not null;comment:数据源地址" json:"endpoint"`                          // 数据源地址
	Status      vobj.Status           `gorm:"column:status;type:int;not null;comment:数据源状态" json:"status"`                                       // 数据源状态
	CreatedAt   types.Time            `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt   types.Time            `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt   soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间" json:"deleted_at"`                             // 删除时间
	Remark      string                `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark"`                               // 描述信息
	Metrics     []*DatasourceMetric   `gorm:"foreignKey:DatasourceID" json:"metrics"`
}

// String json string
func (c *Datasource) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *Datasource) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *Datasource) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// Create func
func (c *Datasource) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *Datasource) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *Datasource) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName Datasource's table name
func (*Datasource) TableName() string {
	return TableNameDatasource
}
