package microserver

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc"

	metadataapi "github.com/aide-cloud/moon/api/metadata"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-cloud/moon/pkg/utils/httpx"
	"github.com/aide-cloud/moon/pkg/vobj"
)

// NewHouYiConn 创建一个HouYi rpc连接
func NewHouYiConn(c *palaceconf.Bootstrap) (*HouYiConn, func(), error) {
	var err error
	var grpcConn *grpc.ClientConn
	microServer := c.GetMicroServer()
	HouYiServer := microServer.GetHouYiServer()
	endpoint := HouYiServer.GetEndpoint()
	network := vobj.NetworkRpc
	httpEndpoint := endpoint
	if strings.HasPrefix(endpoint, "http") {
		network = vobj.NetworkHttp
	} else if strings.HasPrefix(endpoint, "https") {
		network = vobj.NetworkHttps
	} else {
		network = vobj.NetworkRpc
		grpcConn, err = newRpcConn(HouYiServer, microServer.GetDiscovery())
		if err != nil {
			log.Errorw("连接HouYi rpc失败：", err)
			return nil, nil, err
		}
	}

	// 退出时清理资源
	cleanup := func() {
		if grpcConn != nil {
			if err = grpcConn.Close(); err != nil {
				log.Errorw("关闭 reseller rpc 连接失败：", err)
			}
		}
		log.Info("关闭 reseller rpc连接已完成")
	}

	return &HouYiConn{
		rpcClient:    grpcConn,
		httpEndpoint: httpEndpoint,
		network:      network,
	}, cleanup, nil
}

var _ metadataapi.MetricClient = (*HouYiConn)(nil)

const (
	// 此路由保持和proto文件中定义的一直， 有变更记得修改
	syncApi = "/metric/metadata/sync"
)

type HouYiConn struct {
	// rpc连接
	rpcClient *grpc.ClientConn
	// http请求地址
	httpEndpoint string
	// 网络请求类型
	network vobj.Network
}

func (l *HouYiConn) Sync(ctx context.Context, in *metadataapi.SyncRequest, opts ...grpc.CallOption) (*metadataapi.SyncReply, error) {
	switch l.network {
	case vobj.NetworkHttp, vobj.NetworkHttps:
		return l.httpSync(ctx, in)
	default:
		return metadataapi.NewMetricClient(l.rpcClient).Sync(ctx, in, opts...)
	}
}

func (l *HouYiConn) httpSync(ctx context.Context, in *metadataapi.SyncRequest) (*metadataapi.SyncReply, error) {
	requestBody, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	apiAddr, err := url.JoinPath(l.httpEndpoint, syncApi)
	if err != nil {
		return nil, err
	}
	response, err := httpx.NewHttpX().POSTWithContext(ctx, apiAddr, requestBody)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var reply metadataapi.SyncReply
	if err = json.Unmarshal(responseBody, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}
