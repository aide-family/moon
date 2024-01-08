package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNameExternalCustomerHook = "external_customer_hooks"

type ExternalCustomerHook struct {
	BaseModel
	Hook      string       `gorm:"column:hook;type:varchar(255);not null;comment:钩子地址"`
	HookName  string       `gorm:"column:hook_name;type:varchar(64);not null;comment:钩子名称"`
	NotifyApp vo.NotifyApp `gorm:"column:notify_app;type:tinyint;not null;default:1;comment:通知方式"`
	Status    vo.Status    `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark    string       `gorm:"column:remark;type:varchar(255);not null;comment:备注"`

	CustomerId uint32            `gorm:"column:customer_id;type:int unsigned;not null;comment:外部客户ID"`
	Customer   *ExternalCustomer `gorm:"foreignKey:CustomerId;comment:外部客户"`
}

func (*ExternalCustomerHook) TableName() string {
	return TableNameExternalCustomerHook
}

// GetCustomer 获取外部客户
func (e *ExternalCustomerHook) GetCustomer() *ExternalCustomer {
	if e == nil {
		return nil
	}
	return e.Customer
}
