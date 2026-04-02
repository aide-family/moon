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
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/aide-family/jade_tree/internal/biz"
	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/conf"
	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

func NewMachineInfoReporterJob(bc *conf.Bootstrap, machineInfoBiz *biz.MachineInfo, helper *klog.Helper) *machineInfoReportJob {
	reportCfg := bc.GetMachineInfoReport()
	enabledReport, reportInterval, reportTimeout := false, 60*time.Second, 60*time.Second
	var endpoints []string
	if reportCfg := bc.GetMachineInfoReport(); reportCfg != nil {
		enabledReport = strings.EqualFold(reportCfg.GetEnabled(), "true")
		reportInterval = reportCfg.GetInterval().AsDuration()
		reportTimeout = reportCfg.GetTimeout().AsDuration()
		endpoints = make([]string, 0, len(reportCfg.GetEndpoints()))
		for _, endpoint := range reportCfg.GetEndpoints() {
			if s := strings.TrimSpace(endpoint); s != "" {
				endpoints = append(endpoints, s)
			}
		}
	}

	return &machineInfoReportJob{
		enabled:     enabledReport && len(endpoints) > 0,
		index:       "jade-tree-machine-info-reporter",
		spec:        mcron.CronSpecEvery(reportInterval),
		isImmediate: true,
		endpoints:   endpoints,
		headers:     reportCfg.GetHeaders(),
		client:      &http.Client{Timeout: reportTimeout},
		machineInfo: machineInfoBiz,
		helper:      helper,
	}
}

type machineInfoReportJob struct {
	enabled     bool
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
func (j *machineInfoReportJob) Run() {
	if !j.enabled {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), j.client.Timeout)
	defer cancel()
	j.reportOnce(ctx)
}

func (j *machineInfoReportJob) reportOnce(ctx context.Context) {
	req := &bo.ListMachineInfosBo{
		PageRequestBo: bo.NewPageRequestBo(1, 100),
	}
	for {
		pages, err := j.machineInfo.ListClusterMachineInfos(ctx, req)
		if err != nil {
			j.helper.Errorw("msg", "list cluster machine infos failed", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}
		machines := make([]*apiv1.GetMachineInfoReply, 0, len(pages.GetItems()))
		for _, machine := range pages.GetItems() {
			machines = append(machines, bo.ToAPIV1MachineInfoReply(machine))
		}

		if len(machines) == 0 {
			break
		}

		payload, err := protojson.Marshal(&apiv1.ReportMachineInfosRequest{
			Machines: machines,
		})
		if err != nil {
			j.helper.Errorw("msg", "marshal report machine info payload failed", "error", err)
			break
		}
		for _, endpoint := range j.endpoints {
			respPayload, postErr := j.post(ctx, endpoint, payload)
			if postErr != nil {
				j.helper.Errorw("msg", "report machine info failed", "endpoint", endpoint, "error", postErr)
				continue
			}
			j.helper.Debugw("msg", "report machine info success", "endpoint", endpoint, "response", string(respPayload))
		}
		if len(machines) < int(pages.GetPageSize()) {
			break
		}
		req.Page++
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
