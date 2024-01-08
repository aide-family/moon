package do

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNameExternalCustomer = "external_customers"

type ExternalCustomer struct {
	query.BaseModel
	// 外部客户名称
	Name string `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx_name,priority:1;comment:外部客户名称"`
	// 外部客户地址
	Address string `gorm:"column:address;type:varchar(255);not null;comment:外部客户地址"`
	// 外部客户联系人
	Contact string `gorm:"column:contact;type:varchar(64);not null;comment:外部客户联系人"`
	// 外部客户联系电话
	Phone string `gorm:"column:phone;type:varchar(64);not null;comment:外部客户联系电话"`
	// 外部客户联系邮箱
	Email string `gorm:"column:email;type:varchar(64);not null;comment:外部客户联系邮箱"`
	// 外部客户备注
	Remark string `gorm:"column:remark;type:varchar(255);not null;comment:外部客户备注"`
	// 外部客户状态
	Status vo.Status `gorm:"column:status;type:tinyint;not null;default:1;comment:外部客户状态"`
	// 钩子列表
	Hooks []*ExternalCustomerHook `gorm:"foreignKey:CustomerId;comment:外部客户钩子"`
}

func (*ExternalCustomer) TableName() string {
	return TableNameExternalCustomer
}

// GetHooks 获取外部客户钩子列表
func (e *ExternalCustomer) GetHooks() []*ExternalCustomerHook {
	if e == nil {
		return nil
	}
	return e.Hooks
}
