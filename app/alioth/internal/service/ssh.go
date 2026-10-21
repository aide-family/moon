package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/ssh"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type SSHCommandService struct {
	v1.UnimplementedSSHCommandServiceServer
	commands *ssh.Usecase
}

func NewSSHCommandService(uc *ssh.Usecase) *SSHCommandService {
	return &SSHCommandService{commands: uc}
}

func (s *SSHCommandService) CreateSSHCommand(ctx context.Context, req *v1.CreateSSHCommandRequest) (*v1.CreateSSHCommandResponse, error) {
	item, err := s.commands.Create(ctx, sshInput("", req.GetName(), req.GetCommand(), req.GetWorkdir(), req.GetEnv(), req.GetRemark()))
	if err != nil {
		return nil, err
	}
	return &v1.CreateSSHCommandResponse{Command: sshItem(*item)}, nil
}

func (s *SSHCommandService) UpdateSSHCommand(ctx context.Context, req *v1.UpdateSSHCommandRequest) (*v1.UpdateSSHCommandResponse, error) {
	item, err := s.commands.Update(ctx, sshInput(req.GetUid(), req.GetName(), req.GetCommand(), req.GetWorkdir(), req.GetEnv(), req.GetRemark()))
	if err != nil {
		return nil, err
	}
	return &v1.UpdateSSHCommandResponse{Command: sshItem(*item)}, nil
}

func (s *SSHCommandService) ApproveSSHCommand(ctx context.Context, req *v1.ApproveSSHCommandRequest) (*v1.ApproveSSHCommandResponse, error) {
	item, err := s.commands.Approve(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.ApproveSSHCommandResponse{Command: sshItem(*item)}, nil
}

func (s *SSHCommandService) RejectSSHCommand(ctx context.Context, req *v1.RejectSSHCommandRequest) (*v1.RejectSSHCommandResponse, error) {
	item, err := s.commands.Reject(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.RejectSSHCommandResponse{Command: sshItem(*item)}, nil
}

func (s *SSHCommandService) DeleteSSHCommand(ctx context.Context, req *v1.DeleteSSHCommandRequest) (*v1.DeleteSSHCommandResponse, error) {
	if err := s.commands.Delete(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteSSHCommandResponse{}, nil
}

func (s *SSHCommandService) GetSSHCommand(ctx context.Context, req *v1.GetSSHCommandRequest) (*v1.GetSSHCommandResponse, error) {
	item, err := s.commands.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetSSHCommandResponse{Command: sshItem(*item)}, nil
}

func (s *SSHCommandService) ListSSHCommands(ctx context.Context, req *v1.ListSSHCommandsRequest) (*v1.ListSSHCommandsResponse, error) {
	items, total, page, pageSize, err := s.commands.List(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListSSHCommandsResponse{Items: sshItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *SSHCommandService) ExecuteSSHCommand(ctx context.Context, req *v1.ExecuteSSHCommandRequest) (*v1.ExecuteSSHCommandResponse, error) {
	uids := req.GetAgentUids()
	if len(uids) == 0 && req.GetAgentUid() != "" {
		uids = []string{req.GetAgentUid()}
	}
	audits, err := s.commands.ExecuteOnAgents(ctx, req.GetUid(), uids)
	if err != nil {
		return nil, err
	}
	items := sshAuditItems(audits)
	resp := &v1.ExecuteSSHCommandResponse{Audits: items}
	if len(items) == 1 {
		resp.Audit = items[0]
	}
	return resp, nil
}

func (s *SSHCommandService) ExecuteSSHCommands(ctx context.Context, req *v1.ExecuteSSHCommandsRequest) (*v1.ExecuteSSHCommandsResponse, error) {
	audits, err := s.commands.ExecuteMany(ctx, req.GetUids(), req.GetAgentUid())
	if err != nil {
		return nil, err
	}
	return &v1.ExecuteSSHCommandsResponse{Items: sshAuditItems(audits)}, nil
}

func (s *SSHCommandService) ListSSHCommandAudits(ctx context.Context, req *v1.ListSSHCommandAuditsRequest) (*v1.ListSSHCommandAuditsResponse, error) {
	items, total, page, pageSize, err := s.commands.ListAudits(ctx, req.GetCommandUid(), req.GetAgentUid(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	return &v1.ListSSHCommandAuditsResponse{Items: sshAuditItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func sshInput(uid, name, command, workdir, env, remark string) bo.SSHCommand {
	return bo.SSHCommand{UID: uid, Name: name, Command: command, Workdir: workdir, Env: env, Remark: remark}
}

func sshItem(item bo.SSHCommand) *v1.SSHCommandItem {
	return &v1.SSHCommandItem{
		Uid: item.UID, Name: item.Name, Command: item.Command, Workdir: item.Workdir,
		Env: item.Env, ReviewStatus: item.ReviewStatus, Remark: item.Remark,
	}
}

func sshItems(items []bo.SSHCommand) []*v1.SSHCommandItem {
	out := make([]*v1.SSHCommandItem, 0, len(items))
	for _, item := range items {
		out = append(out, sshItem(item))
	}
	return out
}

func sshAuditItem(item bo.SSHCommandAudit) *v1.SSHCommandAuditItem {
	created := ""
	if !item.CreatedAt.IsZero() {
		created = item.CreatedAt.UTC().Format(time.RFC3339)
	}
	return &v1.SSHCommandAuditItem{
		Uid: item.UID, CommandUid: item.CommandUID, AgentUid: item.AgentUID, OperatorUid: item.OperatorUID,
		OperatorName: item.OperatorName, OperatorNickname: item.OperatorNickname, OperatorAvatar: item.OperatorAvatar,
		Kind: item.Kind, Status: item.Status, Command: item.Command, ExitCode: int32(item.ExitCode),
		Output: item.Output, Truncated: item.Truncated, Detail: item.Detail, CreatedAt: created,
	}
}

func sshAuditItems(items []bo.SSHCommandAudit) []*v1.SSHCommandAuditItem {
	out := make([]*v1.SSHCommandAuditItem, 0, len(items))
	for _, item := range items {
		out = append(out, sshAuditItem(item))
	}
	return out
}
