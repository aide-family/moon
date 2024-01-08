package bo

import (
	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	ExternalCustomerBO struct {
		Id      uint32                    `json:"id"`
		Name    string                    `json:"name"`
		Address string                    `json:"address"`
		Contact string                    `json:"contact"`
		Phone   string                    `json:"phone"`
		Email   string                    `json:"email"`
		Remark  string                    `json:"remark"`
		Status  vo.Status                 `json:"status"`
		Hooks   []*ExternalCustomerHookBO `json:"hooks"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// GetHooks 获取钩子列表
func (d *ExternalCustomerBO) GetHooks() []*ExternalCustomerHookBO {
	if d == nil {
		return nil
	}
	return d.Hooks
}

// ToModel 转换为模型
func (d *ExternalCustomerBO) ToModel() *do.ExternalCustomer {
	if d == nil {
		return nil
	}
	return &do.ExternalCustomer{
		BaseModel: do.BaseModel{ID: d.Id},
		Name:      d.Name,
		Address:   d.Address,
		Contact:   d.Contact,
		Phone:     d.Phone,
		Email:     d.Email,
		Remark:    d.Remark,
		Status:    d.Status,
		Hooks: slices.To(d.GetHooks(), func(item *ExternalCustomerHookBO) *do.ExternalCustomerHook {
			return item.ToModel()
		}),
	}
}

func (d *ExternalCustomerBO) ToApi() *api.ExternalCustomer {
	if d == nil {
		return nil
	}
	return &api.ExternalCustomer{
		Id:        d.Id,
		Name:      d.Name,
		Remark:    d.Remark,
		Status:    d.Status.Value(),
		Addr:      d.Address,
		Contact:   d.Contact,
		Phone:     d.Phone,
		Email:     d.Email,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		DeletedAt: d.DeletedAt,
		ExternalCustomerHookList: slices.To(d.GetHooks(), func(item *ExternalCustomerHookBO) *api.ExternalCustomerHook {
			return item.ToApi()
		}),
	}
}

// ExternalCustomerModelToBO 模型转换为BO
func ExternalCustomerModelToBO(m *do.ExternalCustomer) *ExternalCustomerBO {
	if m == nil {
		return nil
	}
	return &ExternalCustomerBO{
		Id:      m.ID,
		Name:    m.Name,
		Address: m.Address,
		Contact: m.Contact,
		Phone:   m.Phone,
		Email:   m.Email,
		Remark:  m.Remark,
		Status:  m.Status,
		Hooks: slices.To(m.GetHooks(), func(item *do.ExternalCustomerHook) *ExternalCustomerHookBO {
			return ExternalCustomerHookModelToBO(item)
		}),
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
	}
}
