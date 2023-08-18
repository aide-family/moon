package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"

	"prometheus-manager/api"
	pb "prometheus-manager/api/node"
	"prometheus-manager/api/perrors"
	"prometheus-manager/pkg/conn"

	"prometheus-manager/apps/master/internal/service"
)

type (
	IPushRepo interface {
		V1Repo
		GRPCPushCall(ctx context.Context, server conn.INodeServer) error
		HTTPPushCall(ctx context.Context, server conn.INodeServer) error
		DeleteGroupSyncNode(ctx context.Context, server conn.INodeServer) error
	}

	PushLogic struct {
		logger *log.Helper
		repo   IPushRepo
	}
)

var _ service.IPushLogic = (*PushLogic)(nil)

func NewPushLogic(repo IPushRepo, logger log.Logger) *PushLogic {
	return &PushLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Push"))}
}

// Call TODO 限制该方法并发, 同一时段内, 只允许执行一次, 如果请求该方法, 监测到正在执行, 则返回正在执行的结果
func (s *PushLogic) Call(ctx context.Context, req *pb.CallRequest) (*pb.CallResponse, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PushLogic.Call")
	defer span.End()

	grpcNodeServers := make([]*pb.NodeServer, 0, len(req.GetServers()))
	httpNodeServers := make([]*pb.NodeServer, 0, len(req.GetServers()))
	for _, server := range req.GetServers() {
		switch server.GetNetwork() {
		case conn.NetworkHttp, conn.NetworkHttps:
			httpNodeServers = append(httpNodeServers, server)
		default:
			grpcNodeServers = append(grpcNodeServers, server)
		}
	}

	var eg errgroup.Group

	for _, srv := range grpcNodeServers {
		server := srv
		eg.Go(func() error {
			if err := s.repo.GRPCPushCall(ctx, server); err != nil {
				// TODO 加入重试队列, 后续重试
			}
			return nil
		})
		eg.Go(func() error {
			if err := s.repo.DeleteGroupSyncNode(ctx, server); err != nil {
				//TODO 加入重试队列, 后续重试
			}
			return nil
		})
	}
	for _, srv := range httpNodeServers {
		server := srv
		eg.Go(func() error {
			if err := s.repo.HTTPPushCall(ctx, server); err != nil {
				// TODO 加入重试队列, 后续重试
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		s.logger.WithContext(ctx).Errorw("Call", req, "err", err)
		return nil, perrors.ErrorServerUnknown("push call err").WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	return &pb.CallResponse{Response: &api.Response{Message: s.repo.V1(ctx)}}, nil
}
