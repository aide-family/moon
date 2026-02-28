package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

type CreateStrategyBo struct {
	Name             string
	Remark           string
	Type             enum.DatasourceType
	Driver           enum.DatasourceDriver
	StrategyGroupUID snowflake.ID
	Metadata         map[string]string
	Status           enum.GlobalStatus
}

func NewCreateStrategyBo(req *apiv1.CreateStrategyRequest) *CreateStrategyBo {
	return &CreateStrategyBo{
		Name:             req.GetName(),
		Remark:           req.GetRemark(),
		Type:             req.GetType(),
		Driver:           req.GetDriver(),
		StrategyGroupUID: snowflake.ParseInt64(req.GetStrategyGroupUID()),
		Metadata:         req.GetMetadata(),
		Status:           req.GetStatus(),
	}
}

type UpdateStrategyBo struct {
	UID              snowflake.ID
	Name             string
	Remark           string
	StrategyGroupUID snowflake.ID
	Type             enum.DatasourceType
	Driver           enum.DatasourceDriver
	Metadata         map[string]string
}

func NewUpdateStrategyBo(req *apiv1.UpdateStrategyRequest) *UpdateStrategyBo {
	return &UpdateStrategyBo{
		UID:              snowflake.ParseInt64(req.GetUid()),
		Name:             req.GetName(),
		Remark:           req.GetRemark(),
		StrategyGroupUID: snowflake.ParseInt64(req.GetStrategyGroupUID()),
		Type:             req.GetType(),
		Driver:           req.GetDriver(),
		Metadata:         req.GetMetadata(),
	}
}

type UpdateStrategyStatusBo struct {
	UID    snowflake.ID
	Status enum.GlobalStatus
}

func NewUpdateStrategyStatusBo(req *apiv1.UpdateStrategyStatusRequest) *UpdateStrategyStatusBo {
	return &UpdateStrategyStatusBo{
		UID:    snowflake.ParseInt64(req.GetUid()),
		Status: req.GetStatus(),
	}
}

type StrategyItemBo struct {
	UID              snowflake.ID
	Name             string
	Remark           string
	Type             enum.DatasourceType
	Driver           enum.DatasourceDriver
	Status           enum.GlobalStatus
	Metadata         map[string]string
	StrategyGroupUID snowflake.ID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (b *StrategyItemBo) ToAPIV1StrategyItem() *apiv1.StrategyItem {
	return &apiv1.StrategyItem{
		Uid:       b.UID.Int64(),
		Name:      b.Name,
		Remark:    b.Remark,
		Type:      b.Type,
		Driver:    b.Driver,
		Status:    b.Status,
		Metadata:  b.Metadata,
		CreatedAt: b.CreatedAt.Format(time.DateTime),
		UpdatedAt: b.UpdatedAt.Format(time.DateTime),
	}
}

type ListStrategyBo struct {
	*PageRequestBo
	Keyword          string
	Status           enum.GlobalStatus
	StrategyGroupUID snowflake.ID
	Type             enum.DatasourceType
	Driver           enum.DatasourceDriver
}

func NewListStrategyBo(req *apiv1.ListStrategyRequest) *ListStrategyBo {
	return &ListStrategyBo{
		PageRequestBo:    NewPageRequestBo(req.GetPage(), req.GetPageSize()),
		Keyword:          req.GetKeyword(),
		Status:           req.GetStatus(),
		StrategyGroupUID: snowflake.ParseInt64(req.GetStrategyGroupUID()),
		Type:             req.GetType(),
		Driver:           req.GetDriver(),
	}
}

func ToAPIV1ListStrategyReply(pageResponseBo *PageResponseBo[*StrategyItemBo]) *apiv1.ListStrategyReply {
	items := make([]*apiv1.StrategyItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1StrategyItem())
	}
	return &apiv1.ListStrategyReply{
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
		Items:    items,
	}
}
