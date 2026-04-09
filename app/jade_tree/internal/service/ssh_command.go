package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"

	"github.com/aide-family/jade_tree/internal/biz"
	"github.com/aide-family/jade_tree/internal/biz/bo"
)

func NewSSHCommandService(b *biz.SSHCommand) *SSHCommandService {
	return &SSHCommandService{bizSSH: b}
}

type SSHCommandService struct {
	apiv1.UnimplementedSSHCommandServer
	bizSSH *biz.SSHCommand
}

func (s *SSHCommandService) SubmitCreateSSHCommand(ctx context.Context, req *apiv1.SubmitCreateSSHCommandRequest) (*apiv1.SubmitSSHCommandAuditReply, error) {
	fields := &bo.SSHCommandFields{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Content:     req.GetContent(),
		WorkDir:     req.GetWorkDir(),
		Env:         req.GetEnv(),
	}
	audit, err := s.bizSSH.SubmitCreate(ctx, fields)
	if err != nil {
		return nil, err
	}
	return &apiv1.SubmitSSHCommandAuditReply{Audit: bo.ToAPIV1SSHCommandAuditItem(audit)}, nil
}

func (s *SSHCommandService) SubmitUpdateSSHCommand(ctx context.Context, req *apiv1.SubmitUpdateSSHCommandRequest) (*apiv1.SubmitSSHCommandAuditReply, error) {
	in := &bo.SubmitSSHCommandUpdateInput{
		CommandUID: snowflake.ID(req.GetCommandUid()),
		Fields: bo.SSHCommandFields{
			Name:        req.GetName(),
			Description: req.GetDescription(),
			Content:     req.GetContent(),
			WorkDir:     req.GetWorkDir(),
			Env:         req.GetEnv(),
		},
	}
	audit, err := s.bizSSH.SubmitUpdate(ctx, in)
	if err != nil {
		return nil, err
	}
	return &apiv1.SubmitSSHCommandAuditReply{Audit: bo.ToAPIV1SSHCommandAuditItem(audit)}, nil
}

func (s *SSHCommandService) ListSSHCommands(ctx context.Context, req *apiv1.ListSSHCommandsRequest) (*apiv1.ListSSHCommandsReply, error) {
	listBo := bo.NewListSSHCommandsBo(req.GetPage(), req.GetPageSize(), req.GetKeyword())
	page, err := s.bizSSH.ListCommands(ctx, listBo)
	if err != nil {
		return nil, err
	}
	items := make([]*apiv1.SSHCommandItem, 0, len(page.GetItems()))
	for _, it := range page.GetItems() {
		items = append(items, bo.ToAPIV1SSHCommandItem(it))
	}
	return &apiv1.ListSSHCommandsReply{
		Items:    items,
		Total:    page.GetTotal(),
		Page:     page.GetPage(),
		PageSize: page.GetPageSize(),
	}, nil
}

func (s *SSHCommandService) GetSSHCommand(ctx context.Context, req *apiv1.GetSSHCommandRequest) (*apiv1.SSHCommandItem, error) {
	item, err := s.bizSSH.GetCommand(ctx, snowflake.ID(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SSHCommandItem(item), nil
}

func (s *SSHCommandService) ListSSHCommandAudits(ctx context.Context, req *apiv1.ListSSHCommandAuditsRequest) (*apiv1.ListSSHCommandAuditsReply, error) {
	listBo := bo.NewListSSHCommandAuditsBo(req.GetPage(), req.GetPageSize(), req.GetStatusFilter(), req.GetKeyword(), req.GetKind())
	page, err := s.bizSSH.ListAudits(ctx, listBo)
	if err != nil {
		return nil, err
	}
	items := make([]*apiv1.SSHCommandAuditItem, 0, len(page.GetItems()))
	for _, it := range page.GetItems() {
		items = append(items, bo.ToAPIV1SSHCommandAuditItem(it))
	}
	return &apiv1.ListSSHCommandAuditsReply{
		Items:    items,
		Total:    page.GetTotal(),
		Page:     page.GetPage(),
		PageSize: page.GetPageSize(),
	}, nil
}

func (s *SSHCommandService) GetSSHCommandAudit(ctx context.Context, req *apiv1.GetSSHCommandAuditRequest) (*apiv1.SSHCommandAuditItem, error) {
	item, err := s.bizSSH.GetAudit(ctx, snowflake.ID(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SSHCommandAuditItem(item), nil
}

func (s *SSHCommandService) ApproveSSHCommandAudit(ctx context.Context, req *apiv1.ApproveSSHCommandAuditRequest) (*apiv1.ApproveSSHCommandAuditReply, error) {
	audit, cmd, err := s.bizSSH.ApproveAudit(ctx, snowflake.ID(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return &apiv1.ApproveSSHCommandAuditReply{
		Audit:   bo.ToAPIV1SSHCommandAuditItem(audit),
		Command: bo.ToAPIV1SSHCommandItem(cmd),
	}, nil
}

func (s *SSHCommandService) RejectSSHCommandAudit(ctx context.Context, req *apiv1.RejectSSHCommandAuditRequest) (*apiv1.RejectSSHCommandAuditReply, error) {
	audit, err := s.bizSSH.RejectAudit(ctx, &bo.RejectSSHCommandAuditInput{
		AuditUID: snowflake.ID(req.GetUid()),
		Reason:   req.GetReason(),
	})
	if err != nil {
		return nil, err
	}
	return &apiv1.RejectSSHCommandAuditReply{Audit: bo.ToAPIV1SSHCommandAuditItem(audit)}, nil
}

func (s *SSHCommandService) ExecuteSSHCommand(ctx context.Context, req *apiv1.ExecuteSSHCommandRequest) (*apiv1.ExecuteSSHCommandReply, error) {
	reply, err := s.bizSSH.Execute(ctx, &bo.ExecuteStoredSSHCommandBo{
		CommandUID:     snowflake.ID(req.GetCommandUid()),
		Host:           req.GetHost(),
		Port:           int(req.GetPort()),
		Username:       req.GetUsername(),
		Password:       req.GetPassword(),
		PrivateKey:     req.GetPrivateKey(),
		TimeoutSeconds: req.GetTimeoutSeconds(),
	})
	if err != nil {
		return nil, err
	}
	return &apiv1.ExecuteSSHCommandReply{
		Stdout:   reply.Stdout,
		Stderr:   reply.Stderr,
		ExitCode: int32(reply.ExitCode),
	}, nil
}

func (s *SSHCommandService) BatchExecuteSSHCommands(ctx context.Context, req *apiv1.BatchExecuteSSHCommandsRequest) (*apiv1.BatchExecuteSSHCommandsReply, error) {
	in := &bo.BatchExecuteSSHCommandsBo{
		Requests: make([]*bo.ExecuteStoredSSHCommandBo, 0, len(req.GetRequests())),
	}
	for _, item := range req.GetRequests() {
		if item == nil {
			continue
		}
		in.Requests = append(in.Requests, &bo.ExecuteStoredSSHCommandBo{
			CommandUID:     snowflake.ID(item.GetCommandUid()),
			Host:           item.GetHost(),
			Port:           int(item.GetPort()),
			Username:       item.GetUsername(),
			Password:       item.GetPassword(),
			PrivateKey:     item.GetPrivateKey(),
			TimeoutSeconds: item.GetTimeoutSeconds(),
		})
	}
	replyItems, err := s.bizSSH.BatchExecute(ctx, in)
	if err != nil {
		return nil, err
	}
	out := &apiv1.BatchExecuteSSHCommandsReply{
		Items: make([]*apiv1.BatchExecuteSSHCommandsItem, 0, len(replyItems)),
	}
	for _, item := range replyItems {
		if item == nil {
			continue
		}
		row := &apiv1.BatchExecuteSSHCommandsItem{
			Index: item.Index,
			Error: item.Error,
		}
		if item.Reply != nil {
			row.Reply = &apiv1.ExecuteSSHCommandReply{
				Stdout:   item.Reply.Stdout,
				Stderr:   item.Reply.Stderr,
				ExitCode: int32(item.Reply.ExitCode),
			}
		}
		out.Items = append(out.Items, row)
	}
	return out, nil
}

func (s *SSHCommandService) CountDispatchSSHCommandTargets(ctx context.Context, req *apiv1.CountDispatchSSHCommandTargetsRequest) (*apiv1.CountDispatchSSHCommandTargetsReply, error) {
	count, err := s.bizSSH.CountDispatchTargets(ctx, bo.NewDispatchSSHCommandFilterBo(req.GetFilter()))
	if err != nil {
		return nil, err
	}
	return &apiv1.CountDispatchSSHCommandTargetsReply{Count: count}, nil
}

func (s *SSHCommandService) DispatchSSHCommandToAgents(ctx context.Context, req *apiv1.DispatchSSHCommandToAgentsRequest) (*apiv1.DispatchSSHCommandToAgentsReply, error) {
	in := &bo.DispatchSSHCommandToAgentsInput{
		CommandUID:     snowflake.ID(req.GetCommandUid()),
		Username:       req.GetUsername(),
		Password:       req.GetPassword(),
		PrivateKey:     req.GetPrivateKey(),
		Port:           int(req.GetPort()),
		TimeoutSeconds: req.GetTimeoutSeconds(),
		Filter:         bo.NewDispatchSSHCommandFilterBo(req.GetFilter()),
	}
	reply, err := s.bizSSH.DispatchToAgents(ctx, in)
	if err != nil {
		return nil, err
	}
	out := &apiv1.DispatchSSHCommandToAgentsReply{
		Total:   reply.Total,
		Success: reply.Success,
		Failed:  reply.Failed,
		Items:   make([]*apiv1.DispatchSSHCommandTargetResult, 0, len(reply.Items)),
	}
	for _, item := range reply.Items {
		if item == nil {
			continue
		}
		row := &apiv1.DispatchSSHCommandTargetResult{
			Endpoint: item.Endpoint,
			Error:    item.Error,
		}
		if item.Machine != nil {
			row.MachineUid = item.Machine.ID.Int64()
			row.MachineUuid = item.Machine.MachineUUID
			row.HostName = item.Machine.HostName
			if item.Machine.Network != nil {
				row.LocalIp = item.Machine.Network.LocalIP
			}
		}
		if item.Reply != nil {
			row.Reply = &apiv1.ExecuteSSHCommandReply{
				Stdout:   item.Reply.Stdout,
				Stderr:   item.Reply.Stderr,
				ExitCode: int32(item.Reply.ExitCode),
			}
		}
		out.Items = append(out.Items, row)
	}
	return out, nil
}

var _ apiv1.SSHCommandServer = (*SSHCommandService)(nil)
