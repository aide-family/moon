package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/api/perrors"
	nodeV1Push "prometheus-manager/api/strategy/v1/push"
	"prometheus-manager/pkg/conn"

	"prometheus-manager/apps/master/internal/biz"
)

type (
	PushRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IPushRepo = (*PushRepo)(nil)

func NewPushRepo(data *Data, logger log.Logger) *PushRepo {
	return &PushRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Push"))}
}

func (l *PushRepo) GRPCPushCall(ctx context.Context, server conn.INodeServer) error {
	ctx, span := otel.Tracer(pushModuleName).Start(ctx, "PushRepo.GRPCPushCall")
	defer span.End()
	rpcConn, err := conn.GetNodeGrpcClient(ctx, server, conn.GetDiscovery())
	if err != nil {
		l.logger.WithContext(ctx).Warnw("GRPCPushCall", server, "err", err)
		return perrors.ErrorServerGrpcError("GRPCPushCall").WithCause(err).WithMetadata(map[string]string{
			"server": server.GetServerName(),
		})
	}

	strategiesResp, err := nodeV1Push.NewPushClient(rpcConn).Strategies(ctx, &nodeV1Push.StrategiesRequest{
		Node:         nil,
		StrategyDirs: nil,
	})
	if err != nil {
		l.logger.WithContext(ctx).Warnw("GRPCPushCall", server, "err", err)
		return perrors.ErrorServerGrpcError("GRPCPushCall").WithCause(err).WithMetadata(map[string]string{
			"server": server.GetServerName(),
		})
	}

	l.logger.WithContext(ctx).Infow("GRPCPushCall", server, "resp", strategiesResp)

	return nil
}

func (l *PushRepo) HTTPPushCall(ctx context.Context, server conn.INodeServer) error {
	ctx, span := otel.Tracer(pushModuleName).Start(ctx, "PushRepo.GRPCPushCall")
	defer span.End()
	return perrors.ErrorServerHttpError("HTTPPushCall not implement").WithMetadata(map[string]string{
		"server": server.GetServerName(),
	})
}

func (l *PushRepo) V1(ctx context.Context) string {
	_, span := otel.Tracer("data").Start(ctx, "PushRepo.V1")
	defer span.End()
	return "PushRepo.V1"
}
