package run

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/aide-family/jade_tree/internal/biz"
	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/conf"
)

type machineInfoReporter struct {
	enabled   bool
	endpoints []string
	headers   map[string]string
	interval  time.Duration

	machineInfo *biz.MachineInfo
	helper      *klog.Helper
	client      *http.Client

	stopCh chan struct{}
	doneCh chan struct{}
	once   sync.Once
}

var _ transport.Server = (*machineInfoReporter)(nil)

func newMachineInfoReporter(bc *conf.Bootstrap, machineInfo *biz.MachineInfo, helper *klog.Helper) *machineInfoReporter {
	reportCfg := bc.GetMachineInfoReport()
	endpoints := make([]string, 0, len(reportCfg.GetEndpoints()))
	for _, endpoint := range reportCfg.GetEndpoints() {
		if s := strings.TrimSpace(endpoint); s != "" {
			endpoints = append(endpoints, s)
		}
	}
	enabled := strings.EqualFold(reportCfg.GetEnabled(), "true") && len(endpoints) > 0
	timeout := reportCfg.GetTimeout().AsDuration()
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	interval := reportCfg.GetInterval().AsDuration()
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	return &machineInfoReporter{
		enabled:     enabled,
		endpoints:   endpoints,
		headers:     reportCfg.GetHeaders(),
		interval:    interval,
		machineInfo: machineInfo,
		helper:      helper,
		client:      &http.Client{Timeout: timeout},
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
	}
}

func (r *machineInfoReporter) Start(ctx context.Context) error {
	if !r.enabled {
		r.helper.Warnf("machine info reporter is not enabled")
		return nil
	}
	go r.loop()
	r.helper.Infof("machine info reporter started, endpoints: %v", r.endpoints)
	return nil
}

func (r *machineInfoReporter) Stop(_ context.Context) error {
	if !r.enabled {
		return nil
	}
	r.once.Do(func() {
		close(r.stopCh)
		<-r.doneCh
	})
	r.helper.Infof("machine info reporter stopped")
	return nil
}

func (r *machineInfoReporter) loop() {
	defer close(r.doneCh)
	r.reportOnce(context.Background())
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.reportOnce(context.Background())
		}
	}
}

func (r *machineInfoReporter) reportOnce(ctx context.Context) {
	info, err := r.machineInfo.GetMachineInfo(ctx)
	if err != nil {
		r.helper.Errorw("msg", "collect machine info for report failed", "error", err)
		return
	}
	payload, err := protojson.Marshal(bo.ToAPIV1MachineInfoReply(info))
	if err != nil {
		r.helper.Errorw("msg", "marshal machine info payload failed", "error", err)
		return
	}
	for _, endpoint := range r.endpoints {
		if err := r.post(ctx, endpoint, payload); err != nil {
			r.helper.Errorw("msg", "report machine info failed", "endpoint", endpoint, "error", err)
		}
	}
}

func (r *machineInfoReporter) post(ctx context.Context, endpoint string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range r.headers {
		if strings.TrimSpace(key) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("machine info report endpoint status=%d", resp.StatusCode)
	}
	return nil
}
