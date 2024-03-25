package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

const TableNamePromNotifyExternal = "prom_alarm_external_notify_objs"

const (
	ExternalNotifyObjFieldName                    = "name"
	ExternalNotifyObjFieldRemark                  = "remark"
	ExternalNotifyObjFieldStatus                  = "status"
	ExternalNotifyObjPreloadFieldCustomerList     = "CustomerList"
	ExternalNotifyObjPreloadFieldCustomerHookList = "CustomerHookList"
)

type ExternalNotifyObj struct {
	BaseModel
	// 外部通知对象名称
	Name string `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__eno__name,priority:1;comment:外部通知对象名称"`
	// 外部通知对象说明
	Remark string `gorm:"column:remark;type:varchar(255);not null;comment:外部通知对象说明"`
	// 状态
	Status vobj.Status `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	// 外部客户列表
	CustomerList []*ExternalCustomer `gorm:"many2many:prom_alarm_external_notify_obj_external_customers;comment:外部客户列表"`
	// 外部客户hook列表
	CustomerHookList []*ExternalCustomerHook `gorm:"many2many:prom_alarm_external_notify_obj_external_customer_hooks;comment:外部客户hook列表"`
}

func (*ExternalNotifyObj) TableName() string {
	return TableNamePromNotifyExternal
}

// GetCustomerList 获取外部客户列表
func (e *ExternalNotifyObj) GetCustomerList() []*ExternalCustomer {
	if e == nil {
		return nil
	}
	return e.CustomerList
}

// GetCustomerHookList 获取外部客户hook列表
func (e *ExternalNotifyObj) GetCustomerHookList() []*ExternalCustomerHook {
	if e == nil {
		return nil
	}
	return e.CustomerHookList
}
