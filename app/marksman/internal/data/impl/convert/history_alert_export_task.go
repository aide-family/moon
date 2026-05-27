package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToHistoryAlertExportFilterDo(filter *bo.HistoryAlertExportFilterBo) *do.HistoryAlertExportFilterConfig {
	if filter == nil {
		return nil
	}
	return &do.HistoryAlertExportFilterConfig{
		StartAtUnix:       filter.StartAtUnix,
		EndAtUnix:         filter.EndAtUnix,
		Status:            filter.Status,
		StrategyGroupUIDs: filter.StrategyGroupUIDs,
		LevelUIDs:         filter.LevelUIDs,
		StrategyUIDs:      filter.StrategyUIDs,
		DatasourceUIDs:    filter.DatasourceUIDs,
		Keyword:           filter.Keyword,
	}
}

func ToHistoryAlertExportFilterBo(filter *do.HistoryAlertExportFilterConfig) *bo.HistoryAlertExportFilterBo {
	if filter == nil {
		return nil
	}
	return &bo.HistoryAlertExportFilterBo{
		StartAtUnix:       filter.StartAtUnix,
		EndAtUnix:         filter.EndAtUnix,
		Status:            filter.Status,
		StrategyGroupUIDs: filter.StrategyGroupUIDs,
		LevelUIDs:         filter.LevelUIDs,
		StrategyUIDs:      filter.StrategyUIDs,
		DatasourceUIDs:    filter.DatasourceUIDs,
		Keyword:           filter.Keyword,
	}
}

func ToHistoryAlertExportTaskDo(ctx context.Context, req *bo.CreateHistoryAlertExportTaskBo) *do.HistoryAlertExportTask {
	m := &do.HistoryAlertExportTask{
		NamespaceUID: contextx.GetNamespace(ctx),
		Status:       int32(bo.HistoryAlertExportTaskStatusPending),
		FilterConfig: ToHistoryAlertExportFilterDo(req.Filter),
	}
	m.WithCreator(contextx.GetUserUID(ctx))
	return m
}

func ToHistoryAlertExportTaskItemBo(m *do.HistoryAlertExportTask) *bo.HistoryAlertExportTaskItemBo {
	if m == nil {
		return nil
	}
	return &bo.HistoryAlertExportTaskItemBo{
		UID:           m.ID,
		NamespaceUID:  m.NamespaceUID,
		Creator:       m.Creator,
		Status:        bo.HistoryAlertExportTaskStatus(m.Status),
		Filter:        ToHistoryAlertExportFilterBo(m.FilterConfig),
		TotalRows:     m.TotalRows,
		ProcessedRows: m.ProcessedRows,
		FileName:      m.FileName,
		FilePath:      m.FilePath,
		ErrorMessage:  m.ErrorMessage,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		CompletedAt:   m.CompletedAt,
	}
}