package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

type CreateDatasourceBo struct {
	Name     string
	Type     enum.DatasourceType
	Driver   enum.DatasourceDriver
	Metadata map[string]string
	URL      string
	Remark   string
	LevelUID snowflake.ID
}

func NewCreateDatasourceBo(req *apiv1.CreateDatasourceRequest) *CreateDatasourceBo {
	return &CreateDatasourceBo{
		Name:     req.GetName(),
		Type:     req.GetType(),
		Driver:   req.GetDriver(),
		Metadata: req.GetMetadata(),
		URL:      req.GetUrl(),
		Remark:   req.GetRemark(),
		LevelUID: snowflake.ParseInt64(req.GetLevelUid()),
	}
}

type UpdateDatasourceBo struct {
	UID      snowflake.ID
	Name     string
	Type     enum.DatasourceType
	Driver   enum.DatasourceDriver
	Metadata map[string]string
	URL      string
	Remark   string
	LevelUID snowflake.ID
}

func NewUpdateDatasourceBo(req *apiv1.UpdateDatasourceRequest) *UpdateDatasourceBo {
	return &UpdateDatasourceBo{
		UID:      snowflake.ParseInt64(req.GetUid()),
		Name:     req.GetName(),
		Type:     req.GetType(),
		Driver:   req.GetDriver(),
		Metadata: req.GetMetadata(),
		URL:      req.GetUrl(),
		Remark:   req.GetRemark(),
		LevelUID: snowflake.ParseInt64(req.GetLevelUid()),
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
	URL       string
	Remark    string
	LevelUID  snowflake.ID
	LevelName string
}

func ToAPIV1DatasourceItem(b *DatasourceItemBo) *apiv1.DatasourceItem {
	if b == nil {
		return nil
	}
	return &apiv1.DatasourceItem{
		Uid:       b.UID.Int64(),
		Name:      b.Name,
		Type:      b.Type,
		Driver:    b.Driver,
		Metadata:  b.Metadata,
		Status:    b.Status,
		CreatedAt: timex.FormatTime(&b.CreatedAt),
		UpdatedAt: timex.FormatTime(&b.UpdatedAt),
		Url:       b.URL,
		Remark:    b.Remark,
		LevelUid:  b.LevelUID.Int64(),
		LevelName: b.LevelName,
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
		items = append(items, ToAPIV1DatasourceItem(item))
	}
	return &apiv1.ListDatasourceReply{
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
		Items:    items,
	}
}

type SelectDatasourceBo struct {
	Keyword string
	Limit   int32
	LastUID snowflake.ID
	Type    enum.DatasourceType
	Driver  enum.DatasourceDriver
	Status  enum.GlobalStatus
	Uids    []int64
}

func NewSelectDatasourceBo(req *apiv1.SelectDatasourceRequest) *SelectDatasourceBo {
	return &SelectDatasourceBo{
		Keyword: req.GetKeyword(),
		Limit:   req.GetLimit(),
		LastUID: snowflake.ParseInt64(req.GetLastUID()),
		Type:    req.GetType(),
		Driver:  req.GetDriver(),
		Status:  req.GetStatus(),
		Uids:    req.GetUids(),
	}
}

type SelectDatasourceItemBo struct {
	Value    snowflake.ID
	Label    string
	Disabled bool
	Tooltip  string
	Type     enum.DatasourceType
	Driver   enum.DatasourceDriver
	URL      string
}

func ToAPIV1SelectDatasourceItem(b *SelectDatasourceItemBo) *apiv1.SelectDatasourceItem {
	if b == nil {
		return nil
	}
	return &apiv1.SelectDatasourceItem{
		Value:    b.Value.Int64(),
		Label:    b.Label,
		Disabled: b.Disabled,
		Tooltip:  b.Tooltip,
		Type:     b.Type,
		Driver:   b.Driver,
		Url:      b.URL,
	}
}

type SelectDatasourceReplyBo struct {
	Items   []*SelectDatasourceItemBo
	Total   int64
	LastUID snowflake.ID
	HasMore bool
}

func ToAPIV1SelectDatasourceReply(replyBo *SelectDatasourceReplyBo) *apiv1.SelectDatasourceReply {
	items := make([]*apiv1.SelectDatasourceItem, 0, len(replyBo.Items))
	for _, item := range replyBo.Items {
		items = append(items, ToAPIV1SelectDatasourceItem(item))
	}
	return &apiv1.SelectDatasourceReply{
		Items:   items,
		Total:   replyBo.Total,
		LastUID: replyBo.LastUID.Int64(),
		HasMore: replyBo.HasMore,
	}
}
