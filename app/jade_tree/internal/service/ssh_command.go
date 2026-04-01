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
	listBo := bo.NewListSSHCommandAuditsBo(req.GetPage(), req.GetPageSize(), req.GetStatusFilter())
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

var _ apiv1.SSHCommandServer = (*SSHCommandService)(nil)
