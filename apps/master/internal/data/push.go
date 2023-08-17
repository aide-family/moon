package data

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/api/perrors"
	"prometheus-manager/api/strategy"
	nodeV1Push "prometheus-manager/api/strategy/v1/push"
	"prometheus-manager/apps/master/internal/conf"
	"prometheus-manager/dal/model"
	"prometheus-manager/pkg/conn"
	"prometheus-manager/pkg/helper"
	"prometheus-manager/pkg/util/dir"
	"prometheus-manager/pkg/util/hash"

	"prometheus-manager/apps/master/internal/biz"
)

type (
	PushRepo struct {
		logger     *log.Helper
		data       *Data
		promV1Repo *PromV1Repo
	}
)

var _ biz.IPushRepo = (*PushRepo)(nil)

func NewPushRepo(data *Data, logger log.Logger) *PushRepo {
	return &PushRepo{
		data:       data,
		promV1Repo: NewPromV1Repo(data, logger),
		logger:     log.NewHelper(log.With(logger, "module", "data/Push")),
	}
}

func (l *PushRepo) GRPCPushCall(ctx context.Context, server conn.INodeServer) error {
	ctx, span := otel.Tracer(pushModuleName).Start(ctx, "PushRepo.GRPCPushCall")
	defer span.End()

	groups, err := l.promV1Repo.AllGroups(ctx)
	if err != nil {
		l.logger.WithContext(ctx).Errorf("GRPCPushCall err: %v", err)
		return perrors.ErrorServerGrpcError("GRPCPushCall").WithCause(err)
	}

	if len(groups) == 0 {
		return nil
	}

	strategies := make([]*strategy.Strategy, 0, len(groups))
	for _, group := range groups {
		if len(group.PromStrategies) == 0 {
			continue
		}
		strategies = append(strategies, &strategy.Strategy{
			Filename: dir.BuildYamlFilename(hash.MD5([]byte(strconv.Itoa(int(group.ID))))),
			Groups: []*strategy.Group{
				{
					Name: group.Name,
					Rules: func(rs []*model.PromStrategy) []*strategy.Rule {
						rules := make([]*strategy.Rule, 0, len(rs))
						for _, rule := range rs {
							labels := helper.BuildLabels(rule.Labels)
							labels["__strategy_id__"] = strconv.Itoa(int(rule.ID))

							rules = append(rules, &strategy.Rule{
								Alert:       rule.Alert,
								Expr:        rule.Expr,
								For:         rule.For,
								Labels:      labels,
								Annotations: helper.BuildAnnotations(rule.Annotations),
							})
						}
						return rules
					}(group.PromStrategies),
				},
			},
		})
	}

	rpcConn, ok := l.data.nodeGrpcClients[server]
	if !ok {
		rpcConn, err = conn.GetNodeGrpcClient(ctx, server, conn.GetDiscovery())
		if err != nil {
			l.logger.WithContext(ctx).Warnw("GRPCPushCall", server, "err", err)
			return perrors.ErrorServerGrpcError("GRPCPushCall").WithCause(err).WithMetadata(map[string]string{
				"server": server.GetServerName(),
			})
		}
	}

	strategiesResp, err := nodeV1Push.NewPushClient(rpcConn).Strategies(ctx, &nodeV1Push.StrategiesRequest{
		Node: server.GetServerName(),
		StrategyDirs: []*strategy.StrategyDir{
			{
				Dir:        conf.Get().GetPushStrategy().GetDir(),
				Strategies: strategies,
			},
		},
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
