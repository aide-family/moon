package impl

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
	"github.com/aide-family/jade_tree/pkg/machine"
)

func NewAgentCommandDispatcher() repository.AgentCommandDispatcher {
	return &agentCommandDispatcher{}
}

type agentCommandDispatcher struct{}

const maxConcurrentProbeTaskDispatch = 10

func (d *agentCommandDispatcher) BatchExecute(ctx context.Context, agent *machine.MachineAgent, req *bo.BatchExecuteSSHCommandsBo) ([]*bo.BatchExecuteSSHCommandItemBo, error) {
	if req == nil {
		return []*bo.BatchExecuteSSHCommandItemBo{}, nil
	}
	apiReq := &apiv1.BatchExecuteSSHCommandsRequest{
		Requests: make([]*apiv1.ExecuteSSHCommandRequest, 0, len(req.Requests)),
	}
	for _, item := range req.Requests {
		if item == nil {
			continue
		}
		apiReq.Requests = append(apiReq.Requests, &apiv1.ExecuteSSHCommandRequest{
			CommandUid:     item.CommandUID.Int64(),
			Host:           item.Host,
			Port:           int32(item.Port),
			Username:       item.Username,
			Password:       item.Password,
			PrivateKey:     item.PrivateKey,
			TimeoutSeconds: item.TimeoutSeconds,
		})
	}

	caller, err := d.newSSHCommandCaller(agent)
	if err != nil {
		return nil, err
	}
	defer func() { _ = caller.Close() }()
	reply, err := caller.BatchExecuteSSHCommands(ctx, apiReq)
	if err != nil {
		return nil, err
	}
	items := make([]*bo.BatchExecuteSSHCommandItemBo, 0, len(reply.GetItems()))
	for _, item := range reply.GetItems() {
		if item == nil {
			continue
		}
		boItem := &bo.BatchExecuteSSHCommandItemBo{
			Index: item.GetIndex(),
			Error: item.GetError(),
		}
		if item.GetReply() != nil {
			boItem.Reply = &bo.SSHExecReply{
				Stdout:   item.GetReply().GetStdout(),
				Stderr:   item.GetReply().GetStderr(),
				ExitCode: int(item.GetReply().GetExitCode()),
			}
		}
		items = append(items, boItem)
	}
	return items, nil
}

func (d *agentCommandDispatcher) BatchCreateProbeTasks(ctx context.Context, agent *machine.MachineAgent, req *bo.BatchCreateProbeTasksBo) (*bo.BatchCreateProbeTasksReplyBo, error) {
	if req == nil {
		return &bo.BatchCreateProbeTasksReplyBo{}, nil
	}
	caller, err := d.newProbeTaskCaller(agent)
	if err != nil {
		return nil, err
	}
	defer func() { _ = caller.Close() }()
	results := safety.NewSlice(make([]*bo.BatchCreateProbeTaskItemResultBo, len(req.Requests)))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrentProbeTaskDispatch)
	for idx, item := range req.Requests {
		idx := idx
		item := item
		g.Go(func() error {
			row := &bo.BatchCreateProbeTaskItemResultBo{Index: int32(idx)}
			if item == nil {
				row.Error = "request is nil"
				results.Set(idx, row)
				return nil
			}
			created, err := caller.CreateProbeTask(gctx, &apiv1.CreateProbeTaskRequest{
				Type:           item.Type,
				Host:           item.Host,
				Port:           item.Port,
				Url:            item.URL,
				Name:           item.Name,
				Status:         item.Status,
				TimeoutSeconds: item.TimeoutSeconds,
			})
			if err != nil {
				row.Error = fmt.Sprintf("agent dispatch error: %v", err)
				results.Set(idx, row)
				return nil
			}
			row.UID = snowflake.ID(created.GetUid())
			results.Set(idx, row)
			return nil
		})
	}
	_ = g.Wait()
	resultItems := results.List()
	out := &bo.BatchCreateProbeTasksReplyBo{
		Items: make([]*bo.BatchCreateProbeTaskItemResultBo, 0, len(resultItems)),
	}
	for _, row := range resultItems {
		if row == nil {
			continue
		}
		out.Items = append(out.Items, row)
	}
	return out, nil
}

type sshCommandCaller struct {
	grpcClient apiv1.SSHCommandClient
	httpClient apiv1.SSHCommandHTTPClient
	closeFn    func() error
}

func (s *sshCommandCaller) BatchExecuteSSHCommands(ctx context.Context, req *apiv1.BatchExecuteSSHCommandsRequest) (*apiv1.BatchExecuteSSHCommandsReply, error) {
	if s.httpClient != nil {
		return s.httpClient.BatchExecuteSSHCommands(ctx, req)
	}
	return s.grpcClient.BatchExecuteSSHCommands(ctx, req)
}

func (s *sshCommandCaller) Close() error {
	if s.closeFn == nil {
		return nil
	}
	return s.closeFn()
}

type probeTaskCaller struct {
	grpcClient apiv1.ProbeTaskClient
	httpClient apiv1.ProbeTaskHTTPClient
	closeFn    func() error
}

func (p *probeTaskCaller) CreateProbeTask(ctx context.Context, req *apiv1.CreateProbeTaskRequest) (*apiv1.ProbeTaskItem, error) {
	if p.httpClient != nil {
		return p.httpClient.CreateProbeTask(ctx, req)
	}
	return p.grpcClient.CreateProbeTask(ctx, req)
}

func (p *probeTaskCaller) Close() error {
	if p.closeFn == nil {
		return nil
	}
	return p.closeFn()
}

func (d *agentCommandDispatcher) newSSHCommandCaller(agent *machine.MachineAgent) (*sshCommandCaller, error) {
	protocol, grpcConn, httpClient, closeFn, err := d.initAgentClient("jade_tree.ssh_command", agent)
	if err != nil {
		return nil, err
	}
	switch protocol {
	case connect.ProtocolGRPC:
		return &sshCommandCaller{grpcClient: apiv1.NewSSHCommandClient(grpcConn), closeFn: closeFn}, nil
	case connect.ProtocolHTTP:
		return &sshCommandCaller{httpClient: apiv1.NewSSHCommandHTTPClient(httpClient), closeFn: closeFn}, nil
	default:
		return nil, merr.ErrorInvalidArgument("agent endpoint is required")
	}
}

func (d *agentCommandDispatcher) newProbeTaskCaller(agent *machine.MachineAgent) (*probeTaskCaller, error) {
	protocol, grpcConn, httpClient, closeFn, err := d.initAgentClient("jade_tree.probe_task", agent)
	if err != nil {
		return nil, err
	}
	switch protocol {
	case connect.ProtocolGRPC:
		return &probeTaskCaller{grpcClient: apiv1.NewProbeTaskClient(grpcConn), closeFn: closeFn}, nil
	case connect.ProtocolHTTP:
		return &probeTaskCaller{httpClient: apiv1.NewProbeTaskHTTPClient(httpClient), closeFn: closeFn}, nil
	default:
		return nil, merr.ErrorInvalidArgument("agent endpoint is required")
	}
}

func (d *agentCommandDispatcher) initAgentClient(
	name string,
	agent *machine.MachineAgent,
) (string, *grpc.ClientConn, *khttp.Client, func() error, error) {
	if agent == nil {
		return "", nil, nil, nil, merr.ErrorInvalidArgument("agent endpoint is required")
	}
	protocol, endpoint := selectAgentProtocolEndpoint(agent)
	initCfg := connect.NewDefaultConfig(name, endpoint, 10*time.Second, protocol)
	switch protocol {
	case connect.ProtocolGRPC:
		conn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return "", nil, nil, nil, err
		}
		return protocol, conn, nil, conn.Close, nil
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return "", nil, nil, nil, err
		}
		return protocol, nil, httpClient, httpClient.Close, nil
	default:
		return "", nil, nil, nil, merr.ErrorInvalidArgument("agent endpoint is required")
	}
}

func selectAgentProtocolEndpoint(agent *machine.MachineAgent) (string, string) {
	if agent == nil {
		return "", ""
	}
	grpcEndpoint := normalizeEndpoint(strings.TrimSpace(agent.GRPCEndpoint))
	httpEndpoint := normalizeEndpoint(strings.TrimSpace(agent.HTTPEndpoint))
	if grpcEndpoint != "" {
		return connect.ProtocolGRPC, grpcEndpoint
	}
	if httpEndpoint != "" {
		return connect.ProtocolHTTP, httpEndpoint
	}
	return "", ""
}

func normalizeEndpoint(endpoint string) string {
	if endpoint == "" {
		return ""
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	return parsed.Host
}
