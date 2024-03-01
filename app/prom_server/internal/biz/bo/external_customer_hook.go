package bo

import (
	"encoding/json"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type (
	ExternalCustomerHookBO struct {
		Id         uint32              `json:"id"`
		Hook       string              `json:"hook"`
		HookName   string              `json:"hookName"`
		NotifyApp  vo.NotifyApp        `json:"notifyApp"`
		Status     vo.Status           `json:"status"`
		Remark     string              `json:"remark"`
		CustomerId uint32              `json:"customerId"`
		Customer   *ExternalCustomerBO `json:"externalCustomer"`
		CreatedAt  int64               `json:"createdAt"`
		UpdatedAt  int64               `json:"updatedAt"`
		DeletedAt  int64               `json:"deletedAt"`
	}
)

// String json string
func (d *ExternalCustomerHookBO) String() string {
	if d == nil {
		return "{}"
	}
	marshal, err := json.Marshal(d)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

// GetCustomer 获取外部客户
func (d *ExternalCustomerHookBO) GetCustomer() *ExternalCustomerBO {
	if d == nil {
		return nil
	}
	return d.Customer
}

// ToModel 转换为模型
func (d *ExternalCustomerHookBO) ToModel() *do.ExternalCustomerHook {
	if d == nil {
		return nil
	}
	return &do.ExternalCustomerHook{
		BaseModel:  do.BaseModel{ID: d.Id},
		Hook:       d.Hook,
		HookName:   d.HookName,
		NotifyApp:  d.NotifyApp,
		Status:     d.Status,
		Remark:     d.Remark,
		CustomerId: d.CustomerId,
		Customer:   d.GetCustomer().ToModel(),
	}
}

// ToApi 转换为API
func (d *ExternalCustomerHookBO) ToApi() *api.ExternalCustomerHook {
	if d == nil {
		return nil
	}
	return &api.ExternalCustomerHook{
		Id:         d.Id,
		HookName:   d.HookName,
		Remark:     d.Remark,
		Status:     d.Status.Value(),
		CustomerId: d.CustomerId,
		Hook:       d.Hook,
		NotifyApp:  d.NotifyApp.Value(),
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
		DeletedAt:  d.DeletedAt,
	}
}

// ExternalCustomerHookModelToBO 模型转换为BO
func ExternalCustomerHookModelToBO(m *do.ExternalCustomerHook) *ExternalCustomerHookBO {
	if m == nil {
		return nil
	}
	return &ExternalCustomerHookBO{
		Id:         m.ID,
		Hook:       m.Hook,
		HookName:   m.HookName,
		NotifyApp:  m.NotifyApp,
		Status:     m.Status,
		Remark:     m.Remark,
		CustomerId: m.CustomerId,
		Customer:   ExternalCustomerModelToBO(m.Customer),
		CreatedAt:  m.CreatedAt.Unix(),
		UpdatedAt:  m.UpdatedAt.Unix(),
		DeletedAt:  int64(m.DeletedAt),
	}
}
