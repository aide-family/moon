// Package grpctls 从环境变量装载 gRPC 的 TLS 与 mTLS 配置。
package grpctls

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aide-family/moon/pkg/merr"
)

// ClientFromEnv 读取 <prefix>_CA、<prefix>_CERT、<prefix>_KEY、<prefix>_SERVER_NAME、<prefix>_INSECURE。
// 没有配置任何一项时返回 (nil, nil)，调用方继续使用明文连接。
func ClientFromEnv(prefix string) (*tls.Config, error) {
	prefix = strings.TrimSuffix(prefix, "_")
	ca := strings.TrimSpace(os.Getenv(prefix + "_CA"))
	cert := strings.TrimSpace(os.Getenv(prefix + "_CERT"))
	key := strings.TrimSpace(os.Getenv(prefix + "_KEY"))
	serverName := strings.TrimSpace(os.Getenv(prefix + "_SERVER_NAME"))
	insecure := strings.EqualFold(os.Getenv(prefix+"_INSECURE"), "true")
	if ca == "" && cert == "" && key == "" && !insecure {
		return nil, nil
	}
	if (cert == "") != (key == "") {
		return nil, merr.Server(prefix + "_CERT and _KEY must be set together")
	}
	cfg := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: serverName}
	if insecure {
		cfg.InsecureSkipVerify = true
	}
	if ca != "" {
		pool, err := loadPool(ca)
		if err != nil {
			return nil, err
		}
		cfg.RootCAs = pool
	}
	if cert != "" {
		pair, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, merr.Server("failed to load grpc client certificate")
		}
		cfg.Certificates = []tls.Certificate{pair}
	}
	return cfg, nil
}

// ClientCredentials 返回客户端凭证；未配置 TLS 时返回 (nil, nil)。
func ClientCredentials(prefix string) (credentials.TransportCredentials, error) {
	cfg, err := ClientFromEnv(prefix)
	if err != nil || cfg == nil {
		return nil, err
	}
	return credentials.NewTLS(cfg), nil
}

// ClientDialOption 返回带 TLS 或明文的 gRPC 拨号选项。
func ClientDialOption(prefix string) (grpc.DialOption, error) {
	creds, err := ClientCredentials(prefix)
	if err != nil {
		return nil, err
	}
	if creds == nil {
		return grpc.WithTransportCredentials(insecure.NewCredentials()), nil
	}
	return grpc.WithTransportCredentials(creds), nil
}

// ServerFromEnv 读取 ALIOTH_GRPC_TLS_CERT、ALIOTH_GRPC_TLS_KEY、ALIOTH_GRPC_TLS_CLIENT_CA。
// 配置了 CLIENT_CA 时要求并校验客户端证书，即 mTLS。
func ServerFromEnv() (*tls.Config, error) {
	cert := strings.TrimSpace(os.Getenv("ALIOTH_GRPC_TLS_CERT"))
	key := strings.TrimSpace(os.Getenv("ALIOTH_GRPC_TLS_KEY"))
	clientCA := strings.TrimSpace(os.Getenv("ALIOTH_GRPC_TLS_CLIENT_CA"))
	if cert == "" && key == "" {
		return nil, nil
	}
	if cert == "" || key == "" {
		return nil, merr.Server("ALIOTH_GRPC_TLS_CERT and ALIOTH_GRPC_TLS_KEY must be set together")
	}
	pair, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, merr.Server("failed to load grpc server certificate")
	}
	cfg := &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{pair}}
	if clientCA != "" {
		pool, err := loadPool(clientCA)
		if err != nil {
			return nil, err
		}
		cfg.ClientCAs = pool
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return cfg, nil
}

func loadPool(path string) (*x509.CertPool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, merr.Server("failed to read grpc CA certificate")
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(raw) {
		return nil, merr.Server("invalid grpc CA certificate")
	}
	return pool, nil
}
