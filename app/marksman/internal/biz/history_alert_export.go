package biz

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

const historyAlertExportPageSize int32 = 200

type HistoryAlertExportTaskEvent struct {
	UID           int64  `json:"uid"`
	Status        int32  `json:"status"`
	TotalRows     int64  `json:"totalRows"`
	ProcessedRows int64  `json:"processedRows"`
	FileName      string `json:"fileName"`
	ErrorMessage  string `json:"errorMessage"`
	CompletedAt   string `json:"completedAt"`
}

type HistoryAlertExportNotifier struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan HistoryAlertExportTaskEvent]struct{}
}

func NewHistoryAlertExportNotifier() *HistoryAlertExportNotifier {
	return &HistoryAlertExportNotifier{
		subscribers: make(map[string]map[chan HistoryAlertExportTaskEvent]struct{}),
	}
}

func exportSubscriberKey(namespaceUID, userUID snowflake.ID) string {
	return fmt.Sprintf("%d:%d", namespaceUID.Int64(), userUID.Int64())
}

func (n *HistoryAlertExportNotifier) Subscribe(namespaceUID, userUID snowflake.ID) (<-chan HistoryAlertExportTaskEvent, func()) {
	key := exportSubscriberKey(namespaceUID, userUID)
	ch := make(chan HistoryAlertExportTaskEvent, 8)
	n.mu.Lock()
	if n.subscribers[key] == nil {
		n.subscribers[key] = make(map[chan HistoryAlertExportTaskEvent]struct{})
	}
	n.subscribers[key][ch] = struct{}{}
	n.mu.Unlock()
	unsubscribe := func() {
		n.mu.Lock()
		delete(n.subscribers[key], ch)
		if len(n.subscribers[key]) == 0 {
			delete(n.subscribers, key)
		}
		n.mu.Unlock()
		close(ch)
	}
	return ch, unsubscribe
}

func (n *HistoryAlertExportNotifier) Publish(namespaceUID, userUID snowflake.ID, event HistoryAlertExportTaskEvent) {
	key := exportSubscriberKey(namespaceUID, userUID)
	n.mu.RLock()
	subs := n.subscribers[key]
	channels := make([]chan HistoryAlertExportTaskEvent, 0, len(subs))
	for ch := range subs {
		channels = append(channels, ch)
	}
	n.mu.RUnlock()
	for _, ch := range channels {
		select {
		case ch <- event:
		default:
		}
	}
}

type runningExportTask struct {
	cancel context.CancelFunc
}

func NewHistoryAlertExportBiz(
	exportTaskRepo repository.HistoryAlertExportTask,
	alertEventRepo repository.AlertEvent,
	notifier *HistoryAlertExportNotifier,
	helper *klog.Helper,
) *HistoryAlertExportBiz {
	return &HistoryAlertExportBiz{
		exportTaskRepo: exportTaskRepo,
		alertEventRepo: alertEventRepo,
		notifier:       notifier,
		running:        sync.Map{},
		exportDir:      filepath.Join(os.TempDir(), "marksman-history-alert-exports"),
		helper:         klog.NewHelper(klog.With(helper.Logger(), "biz", "history_alert_export")),
	}
}

type HistoryAlertExportBiz struct {
	exportTaskRepo repository.HistoryAlertExportTask
	alertEventRepo repository.AlertEvent
	notifier       *HistoryAlertExportNotifier
	running        sync.Map
	exportDir      string
	helper         *klog.Helper
}

func (b *HistoryAlertExportBiz) Notifier() *HistoryAlertExportNotifier {
	return b.notifier
}

func (b *HistoryAlertExportBiz) CreateHistoryAlertExportTask(ctx context.Context, req *bo.CreateHistoryAlertExportTaskBo) (snowflake.ID, error) {
	if req == nil || req.Filter == nil {
		return 0, merr.ErrorParams("filter is required")
	}
	if err := validateHistoryAlertExportFilter(req.Filter); err != nil {
		return 0, err
	}
	uid, err := b.exportTaskRepo.CreateHistoryAlertExportTask(ctx, req)
	if err != nil {
		return 0, err
	}
	runCtx := context.WithoutCancel(ctx)
	go b.runExportTask(runCtx, uid, req.Filter)
	return uid, nil
}

func (b *HistoryAlertExportBiz) ListHistoryAlertExportTask(ctx context.Context, req *bo.ListHistoryAlertExportTaskBo) (*bo.PageResponseBo[*bo.HistoryAlertExportTaskItemBo], error) {
	return b.exportTaskRepo.ListHistoryAlertExportTask(ctx, req)
}

func (b *HistoryAlertExportBiz) CancelHistoryAlertExportTask(ctx context.Context, uid snowflake.ID) error {
	item, err := b.exportTaskRepo.GetHistoryAlertExportTask(ctx, uid)
	if err != nil {
		return err
	}
	if item.Creator != contextx.GetUserUID(ctx) {
		return merr.ErrorForbidden("cannot cancel export task created by another user")
	}
	switch item.Status {
	case bo.HistoryAlertExportTaskStatusCompleted,
		bo.HistoryAlertExportTaskStatusFailed,
		bo.HistoryAlertExportTaskStatusCancelled:
		return merr.ErrorInvalidArgument("export task is already finished")
	}
	if v, ok := b.running.Load(uid.Int64()); ok {
		if task, ok := v.(*runningExportTask); ok && task.cancel != nil {
			task.cancel()
		}
	}
	now := time.Now()
	status := bo.HistoryAlertExportTaskStatusCancelled
	if err := b.exportTaskRepo.UpdateHistoryAlertExportTask(ctx, &bo.UpdateHistoryAlertExportTaskBo{
		UID:         uid,
		Status:      &status,
		CompletedAt: &now,
	}); err != nil {
		return err
	}
	b.publishTask(ctx, &bo.HistoryAlertExportTaskItemBo{
		UID:           uid,
		Status:        status,
		TotalRows:     item.TotalRows,
		ProcessedRows: item.ProcessedRows,
		FileName:      item.FileName,
		CompletedAt:   &now,
	})
	if item.FilePath != "" {
		_ = os.Remove(item.FilePath)
	}
	return nil
}

func (b *HistoryAlertExportBiz) GetHistoryAlertExportTask(ctx context.Context, uid snowflake.ID) (*bo.HistoryAlertExportTaskItemBo, error) {
	item, err := b.exportTaskRepo.GetHistoryAlertExportTask(ctx, uid)
	if err != nil {
		return nil, err
	}
	if item.Creator != contextx.GetUserUID(ctx) {
		return nil, merr.ErrorForbidden("export task not found")
	}
	return item, nil
}

func (b *HistoryAlertExportBiz) runExportTask(ctx context.Context, uid snowflake.ID, filter *bo.HistoryAlertExportFilterBo) {
	taskCtx, cancel := context.WithCancel(ctx)
	b.running.Store(uid.Int64(), &runningExportTask{cancel: cancel})
	defer func() {
		cancel()
		b.running.Delete(uid.Int64())
	}()

	runningStatus := bo.HistoryAlertExportTaskStatusRunning
	if err := b.exportTaskRepo.UpdateHistoryAlertExportTask(ctx, &bo.UpdateHistoryAlertExportTaskBo{
		UID:    uid,
		Status: &runningStatus,
	}); err != nil {
		b.helper.Errorw("msg", "update export task running failed", "error", err, "uid", uid.Int64())
		return
	}
	b.publishByUID(ctx, uid)

	firstPage, err := b.alertEventRepo.ListRealtimeAlert(ctx, filter.ToListRealtimeAlertBo(1, historyAlertExportPageSize), nil)
	if err != nil {
		b.failTask(ctx, uid, err)
		return
	}
	total := firstPage.GetTotal()
	totalRows := total
	if err := b.exportTaskRepo.UpdateHistoryAlertExportTask(ctx, &bo.UpdateHistoryAlertExportTaskBo{
		UID:       uid,
		TotalRows: &totalRows,
	}); err != nil {
		b.failTask(ctx, uid, err)
		return
	}
	b.publishByUID(ctx, uid)

	if total == 0 {
		b.completeTask(ctx, uid, "", "", 0)
		return
	}

	if err := os.MkdirAll(b.exportDir, 0o755); err != nil {
		b.failTask(ctx, uid, err)
		return
	}
	fileName := fmt.Sprintf("history-alerts-%d.csv", uid.Int64())
	filePath := filepath.Join(b.exportDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		b.failTask(ctx, uid, err)
		return
	}
	defer file.Close()

	// UTF-8 BOM for Excel
	if _, err := file.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		b.failTask(ctx, uid, err)
		return
	}
	writer := csv.NewWriter(file)
	if err := writer.Write(historyAlertExportCSVHeader()); err != nil {
		b.failTask(ctx, uid, err)
		return
	}

	writeItems := func(items []*bo.AlertEventItemBo) error {
		for _, item := range items {
			if err := writer.Write(historyAlertExportCSVRow(item)); err != nil {
				return err
			}
		}
		writer.Flush()
		return writer.Error()
	}

	processed := int64(len(firstPage.GetItems()))
	if err := writeItems(firstPage.GetItems()); err != nil {
		b.failTask(ctx, uid, err)
		_ = os.Remove(filePath)
		return
	}
	if err := b.updateProcessedRows(ctx, uid, processed); err != nil {
		b.failTask(ctx, uid, err)
		_ = os.Remove(filePath)
		return
	}

	totalPages := int((total + int64(historyAlertExportPageSize) - 1) / int64(historyAlertExportPageSize))
	for page := int32(2); page <= int32(totalPages); page++ {
		if taskCtx.Err() != nil {
			b.cancelTask(ctx, uid, filePath)
			return
		}
		res, err := b.alertEventRepo.ListRealtimeAlert(taskCtx, filter.ToListRealtimeAlertBo(page, historyAlertExportPageSize), nil)
		if err != nil {
			b.failTask(ctx, uid, err)
			_ = os.Remove(filePath)
			return
		}
		if err := writeItems(res.GetItems()); err != nil {
			b.failTask(ctx, uid, err)
			_ = os.Remove(filePath)
			return
		}
		processed += int64(len(res.GetItems()))
		if err := b.updateProcessedRows(ctx, uid, processed); err != nil {
			b.failTask(ctx, uid, err)
			_ = os.Remove(filePath)
			return
		}
	}

	b.completeTask(ctx, uid, fileName, filePath, processed)
}

func (b *HistoryAlertExportBiz) updateProcessedRows(ctx context.Context, uid snowflake.ID, processed int64) error {
	if err := b.exportTaskRepo.UpdateHistoryAlertExportTask(ctx, &bo.UpdateHistoryAlertExportTaskBo{
		UID:           uid,
		ProcessedRows: &processed,
	}); err != nil {
		return err
	}
	b.publishByUID(ctx, uid)
	return nil
}

func (b *HistoryAlertExportBiz) completeTask(ctx context.Context, uid snowflake.ID, fileName, filePath string, processed int64) {
	now := time.Now()
	status := bo.HistoryAlertExportTaskStatusCompleted
	update := &bo.UpdateHistoryAlertExportTaskBo{
		UID:           uid,
		Status:        &status,
		ProcessedRows: &processed,
		CompletedAt:   &now,
	}
	if fileName != "" {
		update.FileName = &fileName
	}
	if filePath != "" {
		update.FilePath = &filePath
	}
	if err := b.exportTaskRepo.UpdateHistoryAlertExportTask(ctx, update); err != nil {
		b.helper.Errorw("msg", "complete export task failed", "error", err, "uid", uid.Int64())
		return
	}
	b.publishByUID(ctx, uid)
}

func (b *HistoryAlertExportBiz) failTask(ctx context.Context, uid snowflake.ID, err error) {
	now := time.Now()
	status := bo.HistoryAlertExportTaskStatusFailed
	msg := err.Error()
	update := &bo.UpdateHistoryAlertExportTaskBo{
		UID:          uid,
		Status:       &status,
		ErrorMessage: &msg,
		CompletedAt:  &now,
	}
	if uerr := b.exportTaskRepo.UpdateHistoryAlertExportTask(ctx, update); uerr != nil {
		b.helper.Errorw("msg", "fail export task update failed", "error", uerr, "uid", uid.Int64())
		return
	}
	b.publishByUID(ctx, uid)
}

func (b *HistoryAlertExportBiz) cancelTask(ctx context.Context, uid snowflake.ID, filePath string) {
	item, err := b.exportTaskRepo.GetHistoryAlertExportTask(ctx, uid)
	if err != nil {
		return
	}
	if item.Status == bo.HistoryAlertExportTaskStatusCancelled {
		if filePath != "" {
			_ = os.Remove(filePath)
		}
		return
	}
	now := time.Now()
	status := bo.HistoryAlertExportTaskStatusCancelled
	if err := b.exportTaskRepo.UpdateHistoryAlertExportTask(ctx, &bo.UpdateHistoryAlertExportTaskBo{
		UID:         uid,
		Status:      &status,
		CompletedAt: &now,
	}); err != nil {
		b.helper.Errorw("msg", "cancel export task update failed", "error", err, "uid", uid.Int64())
		return
	}
	if filePath != "" {
		_ = os.Remove(filePath)
	}
	b.publishByUID(ctx, uid)
}

func (b *HistoryAlertExportBiz) publishByUID(ctx context.Context, uid snowflake.ID) {
	item, err := b.exportTaskRepo.GetHistoryAlertExportTask(ctx, uid)
	if err != nil {
		return
	}
	b.publishTask(ctx, item)
}

func (b *HistoryAlertExportBiz) publishTask(ctx context.Context, item *bo.HistoryAlertExportTaskItemBo) {
	if item == nil {
		return
	}
	event := HistoryAlertExportTaskEvent{
		UID:           item.UID.Int64(),
		Status:        int32(item.Status),
		TotalRows:     item.TotalRows,
		ProcessedRows: item.ProcessedRows,
		FileName:      item.FileName,
		ErrorMessage:  item.ErrorMessage,
	}
	if item.CompletedAt != nil {
		event.CompletedAt = item.CompletedAt.Format(time.RFC3339)
	}
	b.notifier.Publish(contextx.GetNamespace(ctx), contextx.GetUserUID(ctx), event)
}

func validateHistoryAlertExportFilter(filter *bo.HistoryAlertExportFilterBo) error {
	if filter == nil {
		return merr.ErrorParams("filter is required")
	}
	if filter.StartAtUnix > 0 && filter.EndAtUnix > 0 {
		if filter.StartAtUnix > filter.EndAtUnix {
			return merr.ErrorParams("startAtUnix must be less than endAtUnix")
		}
		if filter.EndAtUnix-filter.StartAtUnix > 31*24*60*60 {
			return merr.ErrorParams("time range must be less than or equal to 31 days")
		}
	}
	return nil
}

func historyAlertExportCSVHeader() []string {
	return []string{
		"uid", "firedAt", "status", "strategyGroupName", "strategyName", "levelName",
		"datasourceName", "summary", "description", "duration", "value",
		"intervenedAt", "intervenedByName", "recoveredAt", "recoveredByName", "labels",
	}
}

func historyAlertExportCSVRow(item *bo.AlertEventItemBo) []string {
	if item == nil {
		return make([]string, len(historyAlertExportCSVHeader()))
	}
	labels := ""
	if len(item.Labels) > 0 {
		if b, err := json.Marshal(item.Labels); err == nil {
			labels = string(b)
		}
	}
	return []string{
		item.UID.String(),
		item.FiredAt.Format(time.RFC3339),
		strconv.Itoa(int(item.Status)),
		item.StrategyGroupName,
		item.StrategyName,
		item.LevelName,
		item.DatasourceName,
		item.Summary,
		item.Description,
		formatAlertDuration(item),
		strconv.FormatFloat(item.Value, 'f', -1, 64),
		formatOptionalTime(item.IntervenedAt),
		item.IntervenedByName,
		formatOptionalTime(item.RecoveredAt),
		item.RecoveredByName,
		labels,
	}
}

func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func formatAlertDuration(item *bo.AlertEventItemBo) string {
	if item == nil {
		return ""
	}
	end := time.Now()
	if item.RecoveredAt != nil {
		end = *item.RecoveredAt
	}
	if item.FiredAt.IsZero() {
		return ""
	}
	return end.Sub(item.FiredAt).Round(time.Second).String()
}
