package bo

import (
	"encoding/json"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	ExternalNotifyObjBO struct {
		Id               uint32                    `json:"id"`
		Name             string                    `json:"name"`
		Remark           string                    `json:"remark"`
		Status           vo.Status                 `json:"status"`
		CustomerList     []*ExternalCustomerBO     `json:"externalCustomerList"`
		CustomerHookList []*ExternalCustomerHookBO `json:"externalCustomerHookList"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// String json string
func (d *ExternalNotifyObjBO) String() string {
	if d == nil {
		return "{}"
	}
	marshal, err := json.Marshal(d)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

// GetCustomerList 获取客户列表
func (d *ExternalNotifyObjBO) GetCustomerList() []*ExternalCustomerBO {
	if d == nil {
		return nil
	}
	return d.CustomerList
}

// GetCustomerHookList 获取客户钩子列表
func (d *ExternalNotifyObjBO) GetCustomerHookList() []*ExternalCustomerHookBO {
	if d == nil {
		return nil
	}
	return d.CustomerHookList
}

// ToModel 将对象转换为模型
func (d *ExternalNotifyObjBO) ToModel() *do.ExternalNotifyObj {
	if d == nil {
		return nil
	}
	return &do.ExternalNotifyObj{
		BaseModel: do.BaseModel{
			ID: d.Id,
		},
		Name:   d.Name,
		Remark: d.Remark,
		Status: d.Status,
		CustomerList: slices.To(d.GetCustomerList(), func(item *ExternalCustomerBO) *do.ExternalCustomer {
			return item.ToModel()
		}),
		CustomerHookList: slices.To(d.GetCustomerHookList(), func(item *ExternalCustomerHookBO) *do.ExternalCustomerHook {
			return item.ToModel()
		}),
	}
}

// ToApi 将对象转换为API
func (d *ExternalNotifyObjBO) ToApi() *api.ExternalNotifyObj {
	if d == nil {
		return nil
	}
	return &api.ExternalNotifyObj{
		Id:     d.Id,
		Name:   d.Name,
		Remark: d.Remark,
		Status: d.Status.Value(),
		ExternalCustomerList: slices.To(d.GetCustomerList(), func(item *ExternalCustomerBO) *api.ExternalCustomer {
			return item.ToApi()
		}),
		ExternalCustomerHookList: slices.To(d.GetCustomerHookList(), func(item *ExternalCustomerHookBO) *api.ExternalCustomerHook {
			return item.ToApi()
		}),
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		DeletedAt: d.DeletedAt,
	}
}

// ExternalNotifyObjModelToBO 将模型转换为对象
func ExternalNotifyObjModelToBO(m *do.ExternalNotifyObj) *ExternalNotifyObjBO {
	if m == nil {
		return nil
	}
	return &ExternalNotifyObjBO{
		Id:     m.ID,
		Name:   m.Name,
		Remark: m.Remark,
		Status: m.Status,
		CustomerList: slices.To(m.GetCustomerList(), func(item *do.ExternalCustomer) *ExternalCustomerBO {
			return ExternalCustomerModelToBO(item)
		}),
		CustomerHookList: slices.To(m.GetCustomerHookList(), func(item *do.ExternalCustomerHook) *ExternalCustomerHookBO {
			return ExternalCustomerHookModelToBO(item)
		}),
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
	}
}
