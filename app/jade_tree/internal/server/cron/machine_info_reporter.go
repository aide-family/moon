// Package cron provides cron-based background servers for Jade Tree.
package cron

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	mcron "github.com/aide-family/magicbox/server/cron"
	"github.com/aide-family/magicbox/strutil"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/aide-family/jade_tree/internal/biz"
	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/conf"
	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

var _ transport.Server = (*MachineInfoReporterServer)(nil)

func NewMachineInfoReporterServer(bc *conf.Bootstrap, machineInfoBiz *biz.MachineInfo, helper *klog.Helper) *MachineInfoReporterServer {
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

	if !enabled {
		return &MachineInfoReporterServer{
			disabled: true,
			helper:   helper,
		}
	}

	reportJob := &machineInfoReportJob{
		index:       "jade-tree-machine-info-reporter",
		spec:        mcron.CronSpecEvery(interval),
		isImmediate: true,
		endpoints:   endpoints,
		headers:     reportCfg.GetHeaders(),
		client:      &http.Client{Timeout: timeout},
		machineInfo: machineInfoBiz,
		helper:      helper,
	}

	return &MachineInfoReporterServer{
		Server: mcron.New("jade-tree-machine-info-reporter", helper.Logger(), mcron.WithCronJobs(reportJob)),
		helper: helper,
	}
}

type MachineInfoReporterServer struct {
	*mcron.Server
	disabled bool
	helper   *klog.Helper
}

func (s *MachineInfoReporterServer) Start(ctx context.Context) error {
	if s.disabled {
		s.helper.WithContext(ctx).Warnf("[MachineInfoReporter] is not enabled")
		return nil
	}
	s.helper.WithContext(ctx).Infow("msg", "[MachineInfoReporter] started")
	return s.Server.Start(ctx)
}

func (s *MachineInfoReporterServer) Stop(ctx context.Context) error {
	if s.disabled {
		return nil
	}
	defer s.helper.WithContext(ctx).Infow("msg", "[MachineInfoReporter] stopped")
	return s.Server.Stop(ctx)
}

type machineInfoReportJob struct {
	index       string
	spec        mcron.CronSpec
	isImmediate bool

	endpoints   []string
	headers     map[string]string
	client      *http.Client
	machineInfo *biz.MachineInfo
	helper      *klog.Helper
}

func (j *machineInfoReportJob) Index() string        { return j.index }
func (j *machineInfoReportJob) Spec() mcron.CronSpec { return j.spec }
func (j *machineInfoReportJob) IsImmediate() bool    { return j.isImmediate }
func (j *machineInfoReportJob) Run()                 { j.reportOnce(context.Background()) }

func (j *machineInfoReportJob) reportOnce(ctx context.Context) {
	local, err := j.machineInfo.RefreshLocalMachineInfo(ctx)
	if err != nil {
		j.helper.Errorw("msg", "refresh local machine info failed", "error", err)
		return
	}
	if local == nil || strutil.IsEmpty(local.MachineUUID) {
		j.helper.Warnw("msg", "local machine uuid is empty, skip report")
		return
	}

	req := &apiv1.ReportMachineInfosRequest{
		Machines: []*apiv1.GetMachineInfoReply{bo.ToAPIV1MachineInfoReply(local)},
	}
	payload, err := protojson.Marshal(req)
	if err != nil {
		j.helper.Errorw("msg", "marshal machine info payload failed", "error", err)
		return
	}

	for _, endpoint := range j.endpoints {
		respPayload, postErr := j.post(ctx, endpoint, payload)
		if postErr != nil {
			j.helper.Errorw("msg", "report machine info failed", "endpoint", endpoint, "error", postErr)
			continue
		}
		var resp apiv1.ReportMachineInfosReply
		if err := protojson.Unmarshal(respPayload, &resp); err != nil {
			j.helper.Errorw("msg", "unmarshal report machine info response failed", "endpoint", endpoint, "error", err)
		}
	}
}

func (j *machineInfoReportJob) post(ctx context.Context, endpoint string, payload []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range j.headers {
		if strings.TrimSpace(key) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	resp, err := j.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("machine info report endpoint status=%d, body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}
