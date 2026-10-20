package data

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	merakv1 "github.com/aide-family/Alioth/api/merak/v1"
	"github.com/aide-family/Alioth/app/merak/internal/conf"
	"github.com/aide-family/Alioth/app/merak/internal/data/repository"
	"github.com/aide-family/Alioth/app/merak/pkg/bo"
	"github.com/aide-family/Alioth/pkg/grpctls"
	"github.com/aide-family/Alioth/pkg/merr"
)

type gateway struct {
	client merakv1.NotifyServiceClient
	key    string
}

func NewGateway(c *conf.Alioth) (repository.NotifyGateway, func(), error) {
	if c == nil || c.GetGrpcAddr() == "" || c.GetServiceKey() == "" {
		return nil, nil, merr.Server("alioth grpc is empty")
	}
	dialOption, err := grpctls.ClientDialOption("ALIOTH_GRPC_TLS")
	if err != nil {
		return nil, nil, err
	}
	conn, err := grpc.NewClient(c.GetGrpcAddr(), dialOption)
	if err != nil {
		return nil, nil, merr.Server("alioth grpc is empty")
	}
	cleanup := func() { _ = conn.Close() }
	return &gateway{client: merakv1.NewNotifyServiceClient(conn), key: c.GetServiceKey()}, cleanup, nil
}

func (g *gateway) Claim(ctx context.Context) ([]bo.TaskSnapshot, error) {
	reply, err := g.client.ClaimTasks(g.auth(ctx), &merakv1.ClaimTasksRequest{})
	if err != nil {
		return nil, merr.Server("claim failed")
	}
	items := make([]bo.TaskSnapshot, 0, len(reply.GetTasks()))
	for _, task := range reply.GetTasks() {
		items = append(items, bo.TaskSnapshot{
			UID: task.GetUid(), NamespaceUID: task.GetNamespaceUid(), Phase: task.GetPhase(),
			ChannelUID: task.GetChannelUid(), ChannelType: task.GetChannelType(), ChannelApp: task.GetChannelApp(),
			ChannelEndpoint: task.GetChannelEndpoint(), ChannelUsername: task.GetChannelUsername(),
			ChannelPassword: task.GetChannelPassword(), ChannelSecret: task.GetChannelSecret(),
			ChannelFrom: task.GetChannelFrom(), TemplateUID: task.GetTemplateUid(), TemplateType: task.GetTemplateType(),
			TemplateSubject: task.GetTemplateSubject(), TemplateBody: task.GetTemplateBody(),
			Receivers: task.GetReceivers(), Params: task.GetParams(), Subject: task.GetSubject(), Body: task.GetBody(),
		})
	}
	return items, nil
}

func (g *gateway) MarkSending(ctx context.Context, taskUID string) error {
	_, err := g.client.MarkSending(g.auth(ctx), &merakv1.MarkSendingRequest{TaskUid: taskUID})
	if err != nil {
		return merr.Server("mark sending failed")
	}
	return nil
}

func (g *gateway) Report(ctx context.Context, report bo.NotifyReport) error {
	_, err := g.client.ReportNotify(g.auth(ctx), &merakv1.ReportNotifyRequest{
		TaskUid: report.TaskUID, Ok: report.OK, Detail: report.Detail,
	})
	if err != nil {
		return merr.Server("report failed")
	}
	return nil
}

func (g *gateway) auth(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+g.key))
}
