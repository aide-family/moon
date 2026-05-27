package service

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/bwmarrin/snowflake"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/biz/bo"
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

const (
	OperationHistoryAlertExportTaskEvents   = "/marksman.api.v1.Alert/HistoryAlertExportTaskEvents"
	OperationHistoryAlertExportTaskDownload = "/marksman.api.v1.Alert/HistoryAlertExportTaskDownload"
)

func (s *AlertService) CreateHistoryAlertExportTask(ctx context.Context, req *apiv1.CreateHistoryAlertExportTaskRequest) (*apiv1.CreateHistoryAlertExportTaskReply, error) {
	uid, err := s.historyAlertExportBiz.CreateHistoryAlertExportTask(ctx, bo.NewCreateHistoryAlertExportTaskBo(req))
	if err != nil {
		return nil, err
	}
	return &apiv1.CreateHistoryAlertExportTaskReply{Uid: uid.Int64()}, nil
}

func (s *AlertService) ListHistoryAlertExportTask(ctx context.Context, req *apiv1.ListHistoryAlertExportTaskRequest) (*apiv1.ListHistoryAlertExportTaskReply, error) {
	result, err := s.historyAlertExportBiz.ListHistoryAlertExportTask(ctx, bo.NewListHistoryAlertExportTaskBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListHistoryAlertExportTaskReply(result), nil
}

func (s *AlertService) CancelHistoryAlertExportTask(ctx context.Context, req *apiv1.CancelHistoryAlertExportTaskRequest) (*apiv1.CancelHistoryAlertExportTaskReply, error) {
	if err := s.historyAlertExportBiz.CancelHistoryAlertExportTask(ctx, snowflake.ParseInt64(req.GetUid())); err != nil {
		return nil, err
	}
	return &apiv1.CancelHistoryAlertExportTaskReply{}, nil
}

func (s *AlertService) HistoryAlertExportTaskEventsHandler(ctx kratoshttp.Context) error {
	kratoshttp.SetOperation(ctx, OperationHistoryAlertExportTaskEvents)
	h := ctx.Middleware(func(c context.Context, _ interface{}) (interface{}, error) {
		return nil, s.streamHistoryAlertExportTaskEvents(c, ctx.Response())
	})
	_, err := h(ctx, struct{}{})
	return err
}

func (s *AlertService) streamHistoryAlertExportTaskEvents(ctx context.Context, w nethttp.ResponseWriter) error {
	w = unwrapResponseWriter(w)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(nethttp.StatusOK)

	namespaceUID := contextx.GetNamespace(ctx)
	userUID := contextx.GetUserUID(ctx)
	ch, unsubscribe := s.historyAlertExportBiz.Notifier().Subscribe(namespaceUID, userUID)
	defer unsubscribe()

	if _, err := fmt.Fprintf(w, "event: connected\ndata: {}\n\n"); err != nil {
		return nil
	}
	flushSSE(w)

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-heartbeat.C:
			if _, err := fmt.Fprintf(w, "event: ping\ndata: {}\n\n"); err != nil {
				return nil
			}
			flushSSE(w)
		case event, ok := <-ch:
			if !ok {
				return nil
			}
			if _, err := fmt.Fprintf(w, "event: export-task\ndata: %s\n\n", marshalExportTaskEvent(event)); err != nil {
				return nil
			}
			flushSSE(w)
		}
	}
}

func (s *AlertService) HistoryAlertExportTaskDownloadHandler(ctx kratoshttp.Context) error {
	var in apiv1.CancelHistoryAlertExportTaskRequest
	if err := ctx.BindVars(&in); err != nil {
		return err
	}
	kratoshttp.SetOperation(ctx, OperationHistoryAlertExportTaskDownload)
	h := ctx.Middleware(func(c context.Context, _ interface{}) (interface{}, error) {
		return nil, s.serveHistoryAlertExportTaskDownload(c, ctx.Response(), ctx.Request(), &in)
	})
	_, err := h(ctx, &in)
	return err
}

func (s *AlertService) serveHistoryAlertExportTaskDownload(ctx context.Context, w nethttp.ResponseWriter, r *nethttp.Request, in *apiv1.CancelHistoryAlertExportTaskRequest) error {
	item, err := s.historyAlertExportBiz.GetHistoryAlertExportTask(ctx, snowflake.ParseInt64(in.GetUid()))
	if err != nil {
		return err
	}
	if item.Status != bo.HistoryAlertExportTaskStatusCompleted || item.FilePath == "" {
		return fmt.Errorf("export file is not ready")
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", item.FileName))
	nethttp.ServeFile(w, r, item.FilePath)
	return nil
}

func unwrapResponseWriter(w nethttp.ResponseWriter) nethttp.ResponseWriter {
	for {
		u, ok := w.(interface{ Unwrap() nethttp.ResponseWriter })
		if !ok {
			break
		}
		w = u.Unwrap()
	}
	return w
}

func flushSSE(w nethttp.ResponseWriter) {
	w = unwrapResponseWriter(w)
	if f, ok := w.(nethttp.Flusher); ok {
		f.Flush()
		return
	}
	_ = nethttp.NewResponseController(w).Flush()
}

func marshalExportTaskEvent(event biz.HistoryAlertExportTaskEvent) string {
	b, err := json.Marshal(event)
	if err != nil {
		return `{}`
	}
	return string(b)
}
