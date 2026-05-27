package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

type HistoryAlertExportTaskStatus int32

const (
	HistoryAlertExportTaskStatusUnknown HistoryAlertExportTaskStatus = iota
	HistoryAlertExportTaskStatusPending
	HistoryAlertExportTaskStatusRunning
	HistoryAlertExportTaskStatusCompleted
	HistoryAlertExportTaskStatusFailed
	HistoryAlertExportTaskStatusCancelled
)

type HistoryAlertExportFilterBo struct {
	StartAtUnix       int64
	EndAtUnix         int64
	Status            enum.AlertEventStatus
	StrategyGroupUIDs []int64
	LevelUIDs         []int64
	StrategyUIDs      []int64
	DatasourceUIDs    []int64
	Keyword           string
}

type CreateHistoryAlertExportTaskBo struct {
	Filter *HistoryAlertExportFilterBo
}

type UpdateHistoryAlertExportTaskBo struct {
	UID           snowflake.ID
	Status        *HistoryAlertExportTaskStatus
	TotalRows     *int64
	ProcessedRows *int64
	FileName      *string
	FilePath      *string
	ErrorMessage  *string
	CompletedAt   *time.Time
}

type HistoryAlertExportTaskItemBo struct {
	UID           snowflake.ID
	NamespaceUID  snowflake.ID
	Creator       snowflake.ID
	Status        HistoryAlertExportTaskStatus
	Filter        *HistoryAlertExportFilterBo
	TotalRows     int64
	ProcessedRows int64
	FileName      string
	FilePath      string
	ErrorMessage  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CompletedAt   *time.Time
}

type ListHistoryAlertExportTaskBo struct {
	*PageRequestBo
	Status *HistoryAlertExportTaskStatus
}

func NewHistoryAlertExportFilterBo(req *apiv1.HistoryAlertExportFilter) *HistoryAlertExportFilterBo {
	if req == nil {
		return &HistoryAlertExportFilterBo{}
	}
	return &HistoryAlertExportFilterBo{
		StartAtUnix:       req.GetStartAtUnix(),
		EndAtUnix:         req.GetEndAtUnix(),
		Status:            req.GetStatus(),
		StrategyGroupUIDs: req.GetStrategyGroupUids(),
		LevelUIDs:         req.GetLevelUids(),
		StrategyUIDs:      req.GetStrategyUids(),
		DatasourceUIDs:    req.GetDatasourceUids(),
		Keyword:           req.GetKeyword(),
	}
}

func NewCreateHistoryAlertExportTaskBo(req *apiv1.CreateHistoryAlertExportTaskRequest) *CreateHistoryAlertExportTaskBo {
	return &CreateHistoryAlertExportTaskBo{
		Filter: NewHistoryAlertExportFilterBo(req.GetFilter()),
	}
}

func NewListHistoryAlertExportTaskBo(req *apiv1.ListHistoryAlertExportTaskRequest) *ListHistoryAlertExportTaskBo {
	var status *HistoryAlertExportTaskStatus
	if req.GetStatus() != apiv1.HistoryAlertExportTaskStatus_HISTORY_ALERT_EXPORT_TASK_STATUS_UNKNOWN {
		s := HistoryAlertExportTaskStatus(req.GetStatus())
		status = &s
	}
	return &ListHistoryAlertExportTaskBo{
		PageRequestBo: NewPageRequestBo(req.GetPage(), req.GetPageSize()),
		Status:        status,
	}
}

func (f *HistoryAlertExportFilterBo) ToListRealtimeAlertBo(page, pageSize int32) *ListRealtimeAlertBo {
	var startAt, endAt time.Time
	if f != nil {
		if f.StartAtUnix != 0 {
			startAt = time.Unix(f.StartAtUnix, 0)
		}
		if f.EndAtUnix != 0 {
			endAt = time.Unix(f.EndAtUnix, 0)
		}
	}
	if f == nil {
		return &ListRealtimeAlertBo{
			PageRequestBo: NewPageRequestBo(page, pageSize),
		}
	}
	return &ListRealtimeAlertBo{
		PageRequestBo:     NewPageRequestBo(page, pageSize),
		StartAt:           startAt,
		EndAt:             endAt,
		Status:            f.Status,
		StrategyGroupUids: f.StrategyGroupUIDs,
		LevelUids:         f.LevelUIDs,
		StrategyUids:      f.StrategyUIDs,
		DatasourceUids:    f.DatasourceUIDs,
		Keyword:           f.Keyword,
	}
}

func ToAPIV1HistoryAlertExportTaskItem(b *HistoryAlertExportTaskItemBo) *apiv1.HistoryAlertExportTaskItem {
	if b == nil {
		return nil
	}
	return &apiv1.HistoryAlertExportTaskItem{
		Uid:           b.UID.Int64(),
		Status:        apiv1.HistoryAlertExportTaskStatus(b.Status),
		TotalRows:     b.TotalRows,
		ProcessedRows: b.ProcessedRows,
		FileName:      b.FileName,
		ErrorMessage:  b.ErrorMessage,
		CreatedAt:     timex.FormatTime(&b.CreatedAt),
		UpdatedAt:     timex.FormatTime(&b.UpdatedAt),
		CompletedAt:   timex.FormatTime(b.CompletedAt),
	}
}

func ToAPIV1ListHistoryAlertExportTaskReply(page *PageResponseBo[*HistoryAlertExportTaskItemBo]) *apiv1.ListHistoryAlertExportTaskReply {
	items := make([]*apiv1.HistoryAlertExportTaskItem, 0, len(page.GetItems()))
	for _, item := range page.GetItems() {
		items = append(items, ToAPIV1HistoryAlertExportTaskItem(item))
	}
	return &apiv1.ListHistoryAlertExportTaskReply{
		Total:    page.GetTotal(),
		Page:     page.GetPage(),
		PageSize: page.GetPageSize(),
		Items:    items,
	}
}
