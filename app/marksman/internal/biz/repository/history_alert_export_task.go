package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type HistoryAlertExportTask interface {
	CreateHistoryAlertExportTask(ctx context.Context, req *bo.CreateHistoryAlertExportTaskBo) (snowflake.ID, error)
	UpdateHistoryAlertExportTask(ctx context.Context, req *bo.UpdateHistoryAlertExportTaskBo) error
	GetHistoryAlertExportTask(ctx context.Context, uid snowflake.ID) (*bo.HistoryAlertExportTaskItemBo, error)
	ListHistoryAlertExportTask(ctx context.Context, req *bo.ListHistoryAlertExportTaskBo) (*bo.PageResponseBo[*bo.HistoryAlertExportTaskItemBo], error)
}
