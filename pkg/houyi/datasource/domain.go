package datasource

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

type CertInfo struct {
	Domain    string `json:"domain"`
	Subject   string `json:"subject"`
	ExpiresOn string `json:"expires_on"`
	DaysLeft  int    `json:"days_left"`
}

// DomainEval 函数，传递一个域名和超时时间，返回一个切片和错误信息
func DomainEval(_ context.Context, domain string, timeout time.Duration) (map[watch.Indexer]*Point, error) {
	now := time.Now()
	// 创建 TCP 连接
	conn, err := net.DialTimeout("tcp", domain+":443", timeout*time.Second)
	if err != nil {
		return nil, err
	}
	// 函数执行完后关闭连接
	defer conn.Close()

	// 配置 TLS 的参数，ServerName 为域名，也就是我们调用函数时传递的参数
	config := &tls.Config{
		ServerName: domain,
	}

	// 创建一个 TLS 的连接
	tlsConn := tls.Client(conn, config)
	// 函数执行完后关闭连接
	defer tlsConn.Close()

	// 创建一个 TLS 的握手
	err = tlsConn.Handshake()
	if err != nil {
		return nil, err
	}

	// 获取证书信息，返回的是一个切片
	certs := tlsConn.ConnectionState().PeerCertificates
	points := make(map[watch.Indexer]*Point)
	for _, cert := range certs {
		labels := vobj.NewLabels(map[string]string{
			vobj.Domain:          domain,
			vobj.DomainSubject:   cert.Subject.CommonName,
			vobj.DomainExpiresOn: cert.NotAfter.Format("2006-01-02 15:04:05"),
		})
		points[labels] = &Point{
			Labels: labels.Map(),
			Values: []*Value{
				{
					Value:     float64(int(cert.NotAfter.Sub(now).Hours() / 24)),
					Timestamp: now.Unix(),
				},
			},
		}
		break // 只取第一个证书信息
	}

	return points, nil
}
