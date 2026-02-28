package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

type CreateDatasourceBo struct {
	Name     string
	Type     enum.DatasourceType
	Driver   enum.DatasourceDriver
	Metadata map[string]string
}

func NewCreateDatasourceBo(req *apiv1.CreateDatasourceRequest) *CreateDatasourceBo {
	return &CreateDatasourceBo{
		Name:     req.GetName(),
		Type:     req.GetType(),
		Driver:   req.GetDriver(),
		Metadata: req.GetMetadata(),
	}
}

type UpdateDatasourceBo struct {
	UID      snowflake.ID
	Name     string
	Type     enum.DatasourceType
	Driver   enum.DatasourceDriver
	Metadata map[string]string
}

func NewUpdateDatasourceBo(req *apiv1.UpdateDatasourceRequest) *UpdateDatasourceBo {
	return &UpdateDatasourceBo{
		UID:      snowflake.ParseInt64(req.GetUid()),
		Name:     req.GetName(),
		Type:     req.GetType(),
		Driver:   req.GetDriver(),
		Metadata: req.GetMetadata(),
	}
}

type DatasourceItemBo struct {
	UID       snowflake.ID
	Name      string
	Type      enum.DatasourceType
	Driver    enum.DatasourceDriver
	Metadata  map[string]string
	Status    enum.GlobalStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (b *DatasourceItemBo) ToAPIV1DatasourceItem() *apiv1.DatasourceItem {
	return &apiv1.DatasourceItem{
		Uid:       b.UID.Int64(),
		Name:      b.Name,
		Type:      b.Type,
		Driver:    b.Driver,
		Metadata:  b.Metadata,
		Status:    b.Status,
		CreatedAt: b.CreatedAt.Format(time.DateTime),
		UpdatedAt: b.UpdatedAt.Format(time.DateTime),
	}
}

type ListDatasourceBo struct {
	*PageRequestBo
	Keyword string
	Type    enum.DatasourceType
	Driver  enum.DatasourceDriver
	Status  enum.GlobalStatus
}

func NewListDatasourceBo(req *apiv1.ListDatasourceRequest) *ListDatasourceBo {
	return &ListDatasourceBo{
		PageRequestBo: NewPageRequestBo(req.GetPage(), req.GetPageSize()),
		Keyword:       req.GetKeyword(),
		Type:          req.GetType(),
		Driver:        req.GetDriver(),
		Status:        req.GetStatus(),
	}
}

func ToAPIV1ListDatasourceReply(pageResponseBo *PageResponseBo[*DatasourceItemBo]) *apiv1.ListDatasourceReply {
	items := make([]*apiv1.DatasourceItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1DatasourceItem())
	}
	return &apiv1.ListDatasourceReply{
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
		Items:    items,
	}
}
