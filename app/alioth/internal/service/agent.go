package service

import (
	"context"
	"encoding/json"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/agent"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

type AgentService struct {
	v1.UnimplementedAgentServiceServer
	registry *agent.Registry
	deployer *agent.Deployer
}

func NewAgentService(registry *agent.Registry, deployer *agent.Deployer) *AgentService {
	return &AgentService{registry: registry, deployer: deployer}
}

func (s *AgentService) CreateAgent(ctx context.Context, req *v1.CreateAgentRequest) (*v1.CreateAgentResponse, error) {
	item, err := s.registry.Create(ctx, agentInput("", req.GetName(), req.GetSshHost(), req.GetSshUser(), req.GetSshPassword(), req.GetSshPrivateKey(), req.GetRemark(), req.GetLabels(), int(req.GetSshPort())))
	if err != nil {
		return nil, err
	}
	return &v1.CreateAgentResponse{Agent: agentItem(*item), Token: item.Token}, nil
}

func (s *AgentService) UpdateAgent(ctx context.Context, req *v1.UpdateAgentRequest) (*v1.UpdateAgentResponse, error) {
	item, err := s.registry.Update(ctx, agentInput(req.GetUid(), req.GetName(), req.GetSshHost(), req.GetSshUser(), req.GetSshPassword(), req.GetSshPrivateKey(), req.GetRemark(), req.GetLabels(), int(req.GetSshPort())))
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAgentResponse{Agent: agentItem(*item)}, nil
}

func (s *AgentService) RotateAgentToken(ctx context.Context, req *v1.RotateAgentTokenRequest) (*v1.RotateAgentTokenResponse, error) {
	item, err := s.registry.Rotate(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.RotateAgentTokenResponse{Agent: agentItem(*item), Token: item.Token}, nil
}

func (s *AgentService) RevokeAgent(ctx context.Context, req *v1.RevokeAgentRequest) (*v1.RevokeAgentResponse, error) {
	item, err := s.registry.Revoke(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.RevokeAgentResponse{Agent: agentItem(*item)}, nil
}

func (s *AgentService) DeleteAgent(ctx context.Context, req *v1.DeleteAgentRequest) (*v1.DeleteAgentResponse, error) {
	if err := s.registry.Delete(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteAgentResponse{}, nil
}

func (s *AgentService) GetAgent(ctx context.Context, req *v1.GetAgentRequest) (*v1.GetAgentResponse, error) {
	item, err := s.registry.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetAgentResponse{Agent: agentItem(*item)}, nil
}

func (s *AgentService) ListAgents(ctx context.Context, req *v1.ListAgentsRequest) (*v1.ListAgentsResponse, error) {
	items, total, page, pageSize, err := s.registry.List(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListAgentsResponse{Items: agentItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *AgentService) Connect(stream v1.AgentService_ConnectServer) error {
	return s.registry.Connect(stream.Context(), stream)
}

func (s *AgentService) DeployAgent(ctx context.Context, req *v1.DeployAgentRequest) (*v1.DeployAgentResponse, error) {
	item, err := s.deployer.Start(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.DeployAgentResponse{Deploy: agentDeployItem(*item)}, nil
}

func (s *AgentService) GetAgentDeploy(ctx context.Context, req *v1.GetAgentDeployRequest) (*v1.GetAgentDeployResponse, error) {
	item, err := s.deployer.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &v1.GetAgentDeployResponse{}, nil
	}
	return &v1.GetAgentDeployResponse{Deploy: agentDeployItem(*item)}, nil
}

func (s *AgentService) MatchAgents(ctx context.Context, req *v1.MatchAgentsRequest) (*v1.MatchAgentsResponse, error) {
	items, err := s.registry.Match(ctx, req.GetLabels())
	if err != nil {
		return nil, err
	}
	return &v1.MatchAgentsResponse{Items: agentItems(items), Total: int32(len(items))}, nil
}

func agentInput(uid, name, host, user, password, privateKey, remark, labels string, port int) bo.Agent {
	return bo.Agent{
		UID: uid, Name: name, SSHHost: host, SSHPort: port, SSHUser: user,
		SSHPassword: password, SSHPrivateKey: privateKey, Remark: remark, Labels: labels,
	}
}

func agentItem(item bo.Agent) *v1.AgentItem {
	out := &v1.AgentItem{
		Uid: item.UID, Name: item.Name, Hostname: item.Hostname, MachineId: item.MachineID,
		LocalIp: item.LocalIP, Version: item.Version, Online: item.Online, Status: item.Status,
		SshHost: item.SSHHost, SshPort: int32(item.SSHPort), SshUser: item.SSHUser,
		HasPassword: item.HasPassword, HasPrivateKey: item.HasPrivateKey, Remark: item.Remark,
		Labels: item.Labels,
	}
	applyCollect(out, item.CollectJSON)
	return out
}

func agentItems(items []bo.Agent) []*v1.AgentItem {
	out := make([]*v1.AgentItem, 0, len(items))
	for _, item := range items {
		out = append(out, agentItem(item))
	}
	return out
}

func agentDeployItem(item bo.AgentDeploy) *v1.AgentDeployItem {
	steps := make([]*v1.AgentDeployStep, 0, len(item.Steps))
	for _, step := range item.Steps {
		steps = append(steps, &v1.AgentDeployStep{
			Key: step.Key, Title: step.Title, Status: step.Status, Message: step.Message,
		})
	}
	finished := ""
	if !item.FinishedAt.IsZero() {
		finished = item.FinishedAt.Format(time.RFC3339)
	}
	return &v1.AgentDeployItem{
		Uid: item.UID, AgentUid: item.AgentUID, Status: item.Status, Step: item.Step,
		Steps: steps, Log: item.Log, LogTruncated: item.LogTruncated,
		Error: item.Error, StartedAt: item.StartedAt.Format(time.RFC3339), FinishedAt: finished,
	}
}

func applyCollect(item *v1.AgentItem, raw string) {
	if raw == "" {
		return
	}
	var payload struct {
		OS         string `json:"os"`
		Arch       string `json:"arch"`
		OSVersion  string `json:"osVersion"`
		CPU        string `json:"cpu"`
		Disk       string `json:"disk"`
		Memory     string `json:"memory"`
		BootTime   string `json:"bootTime"`
		TimeServer string `json:"timeServer"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return
	}
	item.Os = payload.OS
	item.Arch = payload.Arch
	item.OsVersion = payload.OSVersion
	item.Cpu = payload.CPU
	item.Disk = payload.Disk
	item.Memory = payload.Memory
	item.BootTime = payload.BootTime
	item.TimeServer = payload.TimeServer
}
