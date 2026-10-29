package service

import (
	"context"
	"encoding/json"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	probebiz "github.com/aide-family/moon/app/alioth/internal/biz/probe"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/probe"
)

type ProbeTargetService struct {
	v1.UnimplementedProbeTargetServiceServer
	gateway *probebiz.Gateway
}

func NewProbeTargetService(gateway *probebiz.Gateway) *ProbeTargetService {
	return &ProbeTargetService{gateway: gateway}
}

func (s *ProbeTargetService) CreateProbeTargetGroup(ctx context.Context, req *v1.CreateProbeTargetGroupRequest) (*v1.CreateProbeTargetGroupResponse, error) {
	item, err := s.gateway.CreateProbeTargetGroup(ctx, probeGroupInput(req.GetName(), req.GetDriver(), req.GetLabels(), req.GetExecutorLabels(), req.GetRequest(), req.GetRemark(), req.GetTargets()))
	if err != nil {
		return nil, err
	}
	return &v1.CreateProbeTargetGroupResponse{Group: probeGroupItem(*item)}, nil
}

func (s *ProbeTargetService) UpdateProbeTargetGroup(ctx context.Context, req *v1.UpdateProbeTargetGroupRequest) (*v1.UpdateProbeTargetGroupResponse, error) {
	in := probeGroupInput(req.GetName(), req.GetDriver(), req.GetLabels(), req.GetExecutorLabels(), req.GetRequest(), req.GetRemark(), req.GetTargets())
	in.UID = req.GetUid()
	item, err := s.gateway.UpdateProbeTargetGroup(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateProbeTargetGroupResponse{Group: probeGroupItem(*item)}, nil
}

func (s *ProbeTargetService) UpdateProbeTargetGroupStatus(ctx context.Context, req *v1.UpdateProbeTargetGroupStatusRequest) (*v1.UpdateProbeTargetGroupStatusResponse, error) {
	item, err := s.gateway.UpdateProbeTargetGroupStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateProbeTargetGroupStatusResponse{Group: probeGroupItem(*item)}, nil
}

func (s *ProbeTargetService) DeleteProbeTargetGroup(ctx context.Context, req *v1.DeleteProbeTargetGroupRequest) (*v1.DeleteProbeTargetGroupResponse, error) {
	if err := s.gateway.DeleteProbeTargetGroup(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteProbeTargetGroupResponse{}, nil
}

func (s *ProbeTargetService) GetProbeTargetGroup(ctx context.Context, req *v1.GetProbeTargetGroupRequest) (*v1.GetProbeTargetGroupResponse, error) {
	item, err := s.gateway.GetProbeTargetGroup(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetProbeTargetGroupResponse{Group: probeGroupItem(*item)}, nil
}

func (s *ProbeTargetService) ListProbeTargetGroups(ctx context.Context, req *v1.ListProbeTargetGroupsRequest) (*v1.ListProbeTargetGroupsResponse, error) {
	items, total, pageNo, pageSize, err := s.gateway.ListProbeTargetGroups(ctx, req.GetPage(), req.GetPageSize(), req.GetName(), req.GetDriver())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ProbeTargetGroupItem, 0, len(items))
	for i := range items {
		item := probeGroupItem(items[i])
		out = append(out, item)
	}
	return &v1.ListProbeTargetGroupsResponse{Items: out, Total: int32(total), Page: int32(pageNo), PageSize: int32(pageSize)}, nil
}

func (s *ProbeTargetService) ListProbeTargetResults(ctx context.Context, req *v1.ListProbeTargetResultsRequest) (*v1.ListProbeTargetResultsResponse, error) {
	items, total, pageNo, pageSize, err := s.gateway.ListProbeTargetResults(ctx, req.GetUid(), req.GetTargetUid(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ProbeResultItem, 0, len(items))
	for i := range items {
		item := probeResultItem(items[i])
		out = append(out, item)
	}
	return &v1.ListProbeTargetResultsResponse{Items: out, Total: int32(total), Page: int32(pageNo), PageSize: int32(pageSize)}, nil
}

func (s *ProbeTargetService) ProbeTargetGroupNow(ctx context.Context, req *v1.ProbeTargetGroupNowRequest) (*v1.ProbeTargetGroupNowResponse, error) {
	items, err := s.gateway.ProbeTargetGroupNow(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ProbeResultItem, 0, len(items))
	for i := range items {
		out = append(out, probeResultItem(items[i]))
	}
	return &v1.ProbeTargetGroupNowResponse{Results: out}, nil
}

func (s *ProbeTargetService) ProbeTrial(ctx context.Context, req *v1.ProbeTrialRequest) (*v1.ProbeTrialResponse, error) {
	item, err := s.gateway.ProbeTrial(ctx, req.GetSpec(), probeRequest(req.GetRequest()))
	if err != nil {
		return nil, err
	}
	return &v1.ProbeTrialResponse{Result: probeResultItem(*item)}, nil
}

// probeGroupInput 把 proto 请求转成业务对象。
func probeGroupInput(name, driver, labels, executorLabels string, request *v1.ProbeRequestConfig, remark string, targets []*v1.ProbeTargetInput) bo.ProbeTargetGroup {
	in := bo.ProbeTargetGroup{
		Name:           name,
		Driver:         driver,
		Labels:         labels,
		ExecutorLabels: executorLabels,
		Request:        probeRequestJSON(request),
		Remark:         remark,
	}
	for _, target := range targets {
		in.Targets = append(in.Targets, bo.ProbeTarget{
			UID:    target.GetUid(),
			Spec:   target.GetSpec(),
			Labels: target.GetLabels(),
			Status: target.GetStatus(),
			Sort:   int(target.GetSort()),
		})
	}
	return in
}

// probeRequestJSON 把 proto 的请求参数转成入库 JSON。
func probeRequestJSON(in *v1.ProbeRequestConfig) string {
	return probeRequest(in).Marshal()
}

// probeRequest 把 proto 的请求参数转成执行器结构。
func probeRequest(in *v1.ProbeRequestConfig) probe.Request {
	if in == nil {
		return probe.DefaultRequest()
	}
	return probe.Request{
		Method:          in.GetMethod(),
		ExpectStatus:    in.GetExpectStatus(),
		TimeoutMS:       int(in.GetTimeoutMs()),
		FollowRedirects: in.GetFollowRedirects(),
		SkipTLSVerify:   in.GetSkipTlsVerify(),
		VerifyCert:      in.GetVerifyCert(),
		Headers:         parseHeaders(in.GetHeaders()),
		BodyKeyword:     in.GetBodyKeyword(),
	}
}

func parseHeaders(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	out := map[string]string{}
	if err := jsonUnmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

// probeGroupItem 把业务对象转成响应结构。
func probeGroupItem(in bo.ProbeTargetGroup) *v1.ProbeTargetGroupItem {
	request, err := probe.ParseRequest(in.Request)
	if err != nil {
		request = probe.DefaultRequest()
	}
	targets := make([]*v1.ProbeTargetItem, 0, len(in.Targets))
	for _, target := range in.Targets {
		item := &v1.ProbeTargetItem{
			Uid:    target.UID,
			Spec:   target.Spec,
			Labels: target.Labels,
			Status: target.Status,
			Sort:   int32(target.Sort),
		}
		if target.LastResult != nil {
			item.LastResult = probeResultItem(*target.LastResult)
		}
		targets = append(targets, item)
	}
	item := &v1.ProbeTargetGroupItem{
		Uid:               in.UID,
		Name:              in.Name,
		Driver:            in.Driver,
		Status:            in.Status,
		Labels:            in.Labels,
		ExecutorLabels:    in.ExecutorLabels,
		Request:           probeRequestItem(request),
		Remark:            in.Remark,
		Targets:           targets,
		TargetCount:       int32(in.TargetCount),
		ActiveTargetCount: int32(in.ActiveTargetCount),
	}
	if in.LastProbedAt != nil {
		item.LastProbedAt = in.LastProbedAt.UTC().Format(time.RFC3339)
	}
	return item
}

// probeRequestItem 把执行器结构转回 proto。
func probeRequestItem(in probe.Request) *v1.ProbeRequestConfig {
	headers := ""
	if len(in.Headers) > 0 {
		encoded, err := jsonMarshal(in.Headers)
		if err == nil {
			headers = string(encoded)
		}
	}
	return &v1.ProbeRequestConfig{
		Method:          in.Method,
		ExpectStatus:    in.ExpectStatus,
		TimeoutMs:       int32(in.TimeoutMS),
		FollowRedirects: in.FollowRedirects,
		SkipTlsVerify:   in.SkipTLSVerify,
		VerifyCert:      in.VerifyCert,
		Headers:         headers,
		BodyKeyword:     in.BodyKeyword,
	}
}

// probeResultItem 把探测记录转成响应结构。
func probeResultItem(in bo.ProbeResult) *v1.ProbeResultItem {
	item := &v1.ProbeResultItem{
		TargetUid:    in.TargetUID,
		TargetSpec:   in.TargetSpec,
		RuleUid:      in.RuleUID,
		RuleName:     in.RuleName,
		Ok:           in.OK,
		Value:        in.Value,
		DurationMs:   int32(in.DurationMS),
		StatusCode:   int32(in.StatusCode),
		ErrorKind:    in.ErrorKind,
		ErrorMessage: in.ErrorMessage,
		DnsMs:        int32(in.DNSMS),
		TcpMs:        int32(in.TCPMS),
		TlsMs:        int32(in.TLSMS),
		TtfbMs:       int32(in.TTFBMS),
		CertDays:     int32(in.CertDays),
		ResolvedIp:   in.ResolvedIP,
		Executor:     in.Executor,
	}
	if !in.Slot.IsZero() {
		item.Slot = in.Slot.UTC().Format(time.RFC3339)
	}
	if !in.ProbedAt.IsZero() {
		item.ProbedAt = in.ProbedAt.UTC().Format(time.RFC3339)
	}
	return item
}

func jsonUnmarshal(raw string, out any) error {
	return json.Unmarshal([]byte(raw), out)
}

func jsonMarshal(in any) ([]byte, error) {
	return json.Marshal(in)
}
