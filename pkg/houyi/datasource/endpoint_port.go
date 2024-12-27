package datasource

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

// EndpointPortEval 函数用于获取指定 endpoint 端口的监控数据， 用于检查端口是否正常
func EndpointPortEval(_ context.Context, endpoint string, port uint32, timeout time.Duration) (map[watch.Indexer]*Point, error) {
	now := time.Now()
	points := make(map[watch.Indexer]*Point)
	// 超时或者连接失败，返回空切片和错误信息
	labels := vobj.NewLabels(map[string]string{vobj.Domain: endpoint, vobj.DomainPort: strconv.FormatUint(uint64(port), 10)})
	// 创建 TCP 连接
	conn, err := net.DialTimeout("tcp", endpoint+":"+strconv.FormatUint(uint64(port), 10), timeout)
	if err != nil {
		points[labels] = &Point{
			Labels: labels.Map(),
			Values: []*Value{
				{
					Value:     1,
					Timestamp: now.Unix(),
				},
			},
		}
		return points, nil
	}
	// 连接成功，关闭连接
	defer conn.Close()
	// 记录监控数据
	points[labels] = &Point{
		Labels: labels.Map(),
		Values: []*Value{
			{
				Value:     0,
				Timestamp: now.Unix(),
			},
		},
	}

	return points, nil
}
