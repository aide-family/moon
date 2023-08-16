package biz

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

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

	var Err struct {
		errs []error
		lock sync.Mutex
	}
	var wg sync.WaitGroup
	wg.Add(len(grpcNodeServers) + len(httpNodeServers))
	for _, server := range grpcNodeServers {
		go func(srv conn.INodeServer) {
			if err := s.repo.GRPCPushCall(ctx, srv); err != nil {
				Err.lock.Lock()
				Err.errs = append(Err.errs, err)
				Err.lock.Unlock()
			}
		}(server)
	}
	for _, server := range httpNodeServers {
		go func(srv conn.INodeServer) {
			if err := s.repo.HTTPPushCall(ctx, srv); err != nil {
				Err.lock.Lock()
				Err.errs = append(Err.errs, err)
				Err.lock.Unlock()
			}
		}(server)
	}

	wg.Wait()

	if len(Err.errs) > 0 {
		err := perrors.ErrorServerUnknown("push call err").WithMetadata(map[string]string{
			"req": req.String(),
		})
		for _, e := range Err.errs {
			err = err.WithCause(e)
		}
		return nil, err
	}

	return &pb.CallResponse{Response: &api.Response{Message: s.repo.V1(ctx)}}, nil
}
