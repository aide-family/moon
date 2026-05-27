package impl

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/data/impl/convert"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func NewHistoryAlertExportTaskRepository(d *data.Data) repository.HistoryAlertExportTask {
	return &historyAlertExportTaskRepository{db: d.DB()}
}

type historyAlertExportTaskRepository struct {
	db *gorm.DB
}

func (r *historyAlertExportTaskRepository) CreateHistoryAlertExportTask(ctx context.Context, req *bo.CreateHistoryAlertExportTaskBo) (snowflake.ID, error) {
	m := convert.ToHistoryAlertExportTaskDo(ctx, req)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return 0, err
	}
	return m.ID, nil
}

func (r *historyAlertExportTaskRepository) UpdateHistoryAlertExportTask(ctx context.Context, req *bo.UpdateHistoryAlertExportTaskBo) error {
	if req == nil || req.UID.Int64() <= 0 {
		return merr.ErrorInvalidArgument("uid is required")
	}
	updates := map[string]any{}
	if req.Status != nil {
		updates["status"] = int32(*req.Status)
	}
	if req.TotalRows != nil {
		updates["total_rows"] = *req.TotalRows
	}
	if req.ProcessedRows != nil {
		updates["processed_rows"] = *req.ProcessedRows
	}
	if req.FileName != nil {
		updates["file_name"] = *req.FileName
	}
	if req.FilePath != nil {
		updates["file_path"] = *req.FilePath
	}
	if req.ErrorMessage != nil {
		updates["error_message"] = *req.ErrorMessage
	}
	if req.CompletedAt != nil {
		updates["completed_at"] = req.CompletedAt
	}
	if len(updates) == 0 {
		return nil
	}
	result := r.db.WithContext(ctx).Model(&do.HistoryAlertExportTask{}).Where(
		"namespace_uid = ? AND id = ?",
		contextx.GetNamespace(ctx).Int64(),
		req.UID.Int64(),
	).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return merr.ErrorNotFound("history alert export task not found")
	}
	return nil
}

func (r *historyAlertExportTaskRepository) GetHistoryAlertExportTask(ctx context.Context, uid snowflake.ID) (*bo.HistoryAlertExportTaskItemBo, error) {
	var m do.HistoryAlertExportTask
	err := r.db.WithContext(ctx).Where(
		"namespace_uid = ? AND id = ?",
		contextx.GetNamespace(ctx).Int64(),
		uid.Int64(),
	).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("history alert export task not found")
		}
		return nil, err
	}
	return convert.ToHistoryAlertExportTaskItemBo(&m), nil
}

func (r *historyAlertExportTaskRepository) ListHistoryAlertExportTask(ctx context.Context, req *bo.ListHistoryAlertExportTaskBo) (*bo.PageResponseBo[*bo.HistoryAlertExportTaskItemBo], error) {
	wrappers := r.db.WithContext(ctx).Model(&do.HistoryAlertExportTask{}).Where(
		"namespace_uid = ? AND creator = ?",
		contextx.GetNamespace(ctx).Int64(),
		contextx.GetUserUID(ctx).Int64(),
	)
	if req.Status != nil {
		wrappers = wrappers.Where("status = ?", int32(*req.Status))
	}
	var total int64
	if err := wrappers.Count(&total).Error; err != nil {
		return nil, err
	}
	req.WithTotal(total)
	if req.Page > 0 && req.PageSize > 0 {
		wrappers = wrappers.Offset(req.Offset()).Limit(req.Limit())
	}
	var list []*do.HistoryAlertExportTask
	if err := wrappers.Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	items := make([]*bo.HistoryAlertExportTaskItemBo, 0, len(list))
	for _, m := range list {
		items = append(items, convert.ToHistoryAlertExportTaskItemBo(m))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}
