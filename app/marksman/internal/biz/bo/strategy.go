package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/timex"
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

func (c *CreateStrategyBo) Validate() error {
	return validateDatasourceTypeDriver(c.Type, c.Driver)
}

func NewCreateStrategyBo(req *apiv1.CreateStrategyRequest) (*CreateStrategyBo, error) {
	bo := &CreateStrategyBo{
		Name:             req.GetName(),
		Remark:           req.GetRemark(),
		Type:             req.GetType(),
		Driver:           req.GetDriver(),
		StrategyGroupUID: snowflake.ParseInt64(req.GetStrategyGroupUID()),
		Metadata:         req.GetMetadata(),
		Status:           req.GetStatus(),
	}
	if err := bo.Validate(); err != nil {
		return nil, err
	}
	return bo, nil
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

func (u *UpdateStrategyBo) Validate() error {
	return validateDatasourceTypeDriver(u.Type, u.Driver)
}

func NewUpdateStrategyBo(req *apiv1.UpdateStrategyRequest) (*UpdateStrategyBo, error) {
	bo := &UpdateStrategyBo{
		UID:              snowflake.ParseInt64(req.GetUid()),
		Name:             req.GetName(),
		Remark:           req.GetRemark(),
		StrategyGroupUID: snowflake.ParseInt64(req.GetStrategyGroupUID()),
		Type:             req.GetType(),
		Driver:           req.GetDriver(),
		Metadata:         req.GetMetadata(),
	}
	if err := bo.Validate(); err != nil {
		return nil, err
	}
	return bo, nil
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
	StrategyGroup    *StrategyGroupItemBo
}

func ToAPIV1StrategyItem(b *StrategyItemBo) *apiv1.StrategyItem {
	if b == nil {
		return nil
	}
	return &apiv1.StrategyItem{
		Uid:           b.UID.Int64(),
		Name:          b.Name,
		Remark:        b.Remark,
		Type:          b.Type,
		Driver:        b.Driver,
		Status:        b.Status,
		Metadata:      b.Metadata,
		StrategyGroup: ToAPIV1StrategyGroupItem(b.StrategyGroup),
		CreatedAt:     timex.FormatTime(&b.CreatedAt),
		UpdatedAt:     timex.FormatTime(&b.UpdatedAt),
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
		items = append(items, ToAPIV1StrategyItem(item))
	}
	return &apiv1.ListStrategyReply{
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
		Items:    items,
	}
}

type SelectStrategyBo struct {
	Keyword           string
	Limit             int32
	LastUID           snowflake.ID
	Status            enum.GlobalStatus
	StrategyGroupUIDs []int64
}

func NewSelectStrategyBo(req *apiv1.SelectStrategyRequest) *SelectStrategyBo {
	return &SelectStrategyBo{
		Keyword:           req.GetKeyword(),
		Limit:             req.GetLimit(),
		LastUID:           snowflake.ParseInt64(req.GetLastUid()),
		Status:            req.GetStatus(),
		StrategyGroupUIDs: req.GetStrategyGroupUids(),
	}
}

type StrategyItemSelectBo struct {
	Value    int64
	Label    string
	Disabled bool
	Tooltip  string
}

func ToAPIV1StrategyItemSelect(b *StrategyItemSelectBo) *apiv1.StrategyItemSelect {
	if b == nil {
		return nil
	}
	return &apiv1.StrategyItemSelect{
		Value:    b.Value,
		Label:    b.Label,
		Disabled: b.Disabled,
		Tooltip:  b.Tooltip,
	}
}

type SelectStrategyBoResult struct {
	Items   []*StrategyItemSelectBo
	Total   int64
	LastUID snowflake.ID
	HasMore bool
}

func ToAPIV1SelectStrategyReply(result *SelectStrategyBoResult) *apiv1.SelectStrategyReply {
	items := make([]*apiv1.StrategyItemSelect, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, ToAPIV1StrategyItemSelect(item))
	}
	return &apiv1.SelectStrategyReply{
		Items:   items,
		Total:   result.Total,
		LastUid: result.LastUID.Int64(),
		HasMore: result.HasMore,
	}
}

// Driver value ranges by datasource type (proto convention): METRICS 1-1000, LOGS 1001-2000, TRACE 2001-3000.
const (
	gap                   = 999
	driverRangeMETRICSMin = int32(enum.DatasourceType_METRICS)
	driverRangeMETRICSMax = int32(enum.DatasourceType_METRICS) + gap
	driverRangeLOGSMin    = int32(enum.DatasourceType_LOGS)
	driverRangeLOGSMax    = int32(enum.DatasourceType_LOGS) + gap
	driverRangeTRACEMin   = int32(enum.DatasourceType_TRACE)
	driverRangeTRACEMax   = int32(enum.DatasourceType_TRACE) + gap
)

// validateDatasourceTypeDriver ensures driver falls in the range for the given type so new drivers need no code change.
func validateDatasourceTypeDriver(dsType enum.DatasourceType, driver enum.DatasourceDriver) error {
	d := int32(driver)
	switch dsType {
	case enum.DatasourceType_METRICS:
		if d < driverRangeMETRICSMin || d > driverRangeMETRICSMax {
			return merr.ErrorParams("when type is METRICS, driver must be a METRICS driver (value 1-1000)")
		}
		return nil
	case enum.DatasourceType_LOGS:
		if d < driverRangeLOGSMin || d > driverRangeLOGSMax {
			return merr.ErrorParams("when type is LOGS, driver must be a LOGS driver (value 1001-2000)")
		}
		return nil
	case enum.DatasourceType_TRACE:
		if d < driverRangeTRACEMin || d > driverRangeTRACEMax {
			return merr.ErrorParams("when type is TRACE, driver must be a TRACE driver (value 2001-3000)")
		}
		return nil
	default:
		return merr.ErrorParams("unknown datasource type, driver must match type")
	}
}
