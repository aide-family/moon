package rpc

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/data"
	"github.com/moon-monitor/moon/pkg/api/common"
	houyiv1 "github.com/moon-monitor/moon/pkg/api/houyi/v1"
	"github.com/moon-monitor/moon/pkg/config"
	"github.com/moon-monitor/moon/pkg/merr"
)

func NewHouyiServer(data *data.Data, logger log.Logger) repository.Houyi {
	return &houyiServer{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.houyi")),
	}
}

type houyiServer struct {
	*data.Data
	helper *log.Helper
}

type houyiSyncClient struct {
	server *bo.Server
}

func (s *houyiServer) Sync() (repository.HouyiSyncClient, bool) {
	server, ok := s.FirstServerConn(vobj.ServerTypeHouyi)
	if !ok {
		return nil, false
	}
	return &houyiSyncClient{server: server}, true
}

func (s *houyiSyncClient) MetricMetadata(ctx context.Context, req *houyiv1.MetricMetadataRequest) (*houyiv1.SyncReply, error) {
	switch s.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return houyiv1.NewSyncClient(s.server.Conn).MetricMetadata(ctx, req)
	case config.Network_HTTP:
		return houyiv1.NewSyncHTTPClient(s.server.Client).MetricMetadata(ctx, req)
	default:
		return nil, merr.ErrorInternalServerError("network is not supported")
	}
}

type houyiQueryClient struct {
	server *bo.Server
}

func (s *houyiServer) Query() (repository.HouyiQueryClient, bool) {
	server, ok := s.FirstServerConn(vobj.ServerTypeHouyi)
	if !ok {
		return nil, false
	}
	return &houyiQueryClient{server: server}, true
}

func (s *houyiQueryClient) MetricDatasourceQuery(ctx context.Context, req *houyiv1.MetricDatasourceQueryRequest) (*common.MetricDatasourceQueryReply, error) {
	switch s.server.Config.Server.GetNetwork() {
	case config.Network_GRPC:
		return houyiv1.NewQueryClient(s.server.Conn).MetricDatasourceQuery(ctx, req)
	case config.Network_HTTP:
		return houyiv1.NewQueryHTTPClient(s.server.Client).MetricDatasourceQuery(ctx, req)
	default:
		return nil, merr.ErrorInternalServerError("network is not supported")
	}
}
