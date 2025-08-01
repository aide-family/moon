package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/server"
)

func NewServerRepo(data *data.Data, logger log.Logger) repository.Server {
	return &serverRepoImpl{
		data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.server")),
	}
}

type serverRepoImpl struct {
	data   *data.Data
	helper *log.Helper
}

func (s *serverRepoImpl) DeregisterServer(ctx context.Context, req *bo.ServerRegisterReq) error {
	s.helper.WithContext(ctx).Debugf("deregister %s server: %v", req.ServerType, req)
	serverConn, ok := s.data.GetServerConn(req.ServerType, req.UUID)
	if !ok {
		return nil
	}
	defer s.data.RemoveServerConn(req.ServerType, req.UUID)
	if err := serverConn.Close(); err != nil {
		return err
	}
	return nil
}

func (s *serverRepoImpl) RegisterServer(ctx context.Context, req *bo.ServerRegisterReq) error {
	s.helper.WithContext(ctx).Debugf("register %s server: %s", req.ServerType, req.UUID)
	initConfig := &server.InitConfig{
		MicroConfig: req.Server,
		Registry:    (*config.Registry)(req.Discovery),
	}
	serverBo := &bo.Server{Config: req}
	switch req.Server.GetNetwork() {
	case config.Network_GRPC:
		conn, err := server.InitGRPCClient(initConfig)
		if err != nil {
			return err
		}
		serverBo.Conn = conn
	case config.Network_HTTP:
		client, err := server.InitHTTPClient(initConfig)
		if err != nil {
			return err
		}
		serverBo.Client = client
	default:
		return merr.ErrorInvalidArgument("unsupported network: %s", req.Server.GetNetwork())
	}
	s.data.SetServerConn(req.ServerType, req.UUID, serverBo)
	return nil
}
