package agent

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/aide-family/Alioth/pkg/merr"
)

// k8sConnection 是代理请求要用的 API Server 连接信息。
// 来源可以是 ALKAID_KUBECONFIG，也可以是集群内 ServiceAccount。
type k8sConnection struct {
	base     *url.URL
	client   *http.Client
	token    string
	username string
	password string
}

type kubeConfigFile struct {
	CurrentContext string             `yaml:"current-context"`
	Clusters       []kubeNamedCluster `yaml:"clusters"`
	Users          []kubeNamedUser    `yaml:"users"`
	Contexts       []kubeNamedContext `yaml:"contexts"`
}

type kubeNamedCluster struct {
	Name    string      `yaml:"name"`
	Cluster kubeCluster `yaml:"cluster"`
}

type kubeCluster struct {
	Server                   string `yaml:"server"`
	CertificateAuthority     string `yaml:"certificate-authority"`
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	InsecureSkipTLSVerify    bool   `yaml:"insecure-skip-tls-verify"`
}

type kubeNamedUser struct {
	Name string   `yaml:"name"`
	User kubeUser `yaml:"user"`
}

type kubeUser struct {
	Token                 string `yaml:"token"`
	TokenFile             string `yaml:"tokenFile"`
	ClientCertificate     string `yaml:"client-certificate"`
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKey             string `yaml:"client-key"`
	ClientKeyData         string `yaml:"client-key-data"`
	Username              string `yaml:"username"`
	Password              string `yaml:"password"`
}

type kubeNamedContext struct {
	Name    string      `yaml:"name"`
	Context kubeContext `yaml:"context"`
}

type kubeContext struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

// k8sConnectionFromEnv 优先使用 kubeconfig，其次使用集群内 ServiceAccount。
func k8sConnectionFromEnv() (*k8sConnection, error) {
	path := strings.TrimSpace(os.Getenv("ALKAID_KUBECONFIG"))
	if path == "" {
		return defaultK8sConnection()
	}
	return kubeConfigConnection(path)
}

func defaultK8sConnection() (*k8sConnection, error) {
	base, err := url.Parse(k8sAPIBase())
	if err != nil || base.Host == "" {
		return nil, merr.Server("invalid kubernetes api address")
	}
	token := strings.TrimSpace(os.Getenv("ALKAID_K8S_TOKEN"))
	if token == "" {
		if raw, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
			token = strings.TrimSpace(string(raw))
		}
	}
	return &k8sConnection{
		base: base,
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: &http.Transport{TLSClientConfig: k8sTLSConfig(base.Hostname())},
		},
		token: token,
	}, nil
}

func kubeConfigConnection(path string) (*k8sConnection, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, merr.Server("failed to read kubeconfig")
	}
	var cfg kubeConfigFile
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, merr.Server("invalid kubeconfig")
	}
	dir := filepath.Dir(path)
	clusterName, userName := kubeTargets(cfg)
	cluster, ok := findCluster(cfg, clusterName)
	if !ok {
		return nil, merr.Server("kubeconfig has no matching cluster")
	}
	base, err := url.Parse(cluster.Cluster.Server)
	if err != nil || base.Host == "" {
		return nil, merr.Server("invalid kubeconfig server")
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if cluster.Cluster.InsecureSkipTLSVerify {
		tlsConfig.InsecureSkipVerify = true
	}
	rawCA, err := kubeDataOrFile(dir, cluster.Cluster.CertificateAuthorityData, cluster.Cluster.CertificateAuthority)
	if err != nil {
		return nil, err
	}
	if len(rawCA) > 0 {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(rawCA) {
			return nil, merr.Server("invalid kubeconfig certificate authority")
		}
		tlsConfig.RootCAs = pool
	}
	user, _ := findUser(cfg, userName)
	cert, err := kubeDataOrFile(dir, user.User.ClientCertificateData, user.User.ClientCertificate)
	if err != nil {
		return nil, err
	}
	key, err := kubeDataOrFile(dir, user.User.ClientKeyData, user.User.ClientKey)
	if err != nil {
		return nil, err
	}
	if len(cert) > 0 && len(key) > 0 {
		pair, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, merr.Server("invalid kubeconfig client certificate")
		}
		tlsConfig.Certificates = []tls.Certificate{pair}
	}
	token := strings.TrimSpace(user.User.Token)
	if token == "" && user.User.TokenFile != "" {
		if rawToken, err := os.ReadFile(resolveKubePath(dir, user.User.TokenFile)); err == nil {
			token = strings.TrimSpace(string(rawToken))
		}
	}
	return &k8sConnection{
		base: base,
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: &http.Transport{TLSClientConfig: tlsConfig},
		},
		token:    token,
		username: user.User.Username,
		password: user.User.Password,
	}, nil
}

func kubeTargets(cfg kubeConfigFile) (cluster, user string) {
	name := cfg.CurrentContext
	if name == "" && len(cfg.Contexts) > 0 {
		name = cfg.Contexts[0].Name
	}
	for _, item := range cfg.Contexts {
		if item.Name == name {
			return item.Context.Cluster, item.Context.User
		}
	}
	if len(cfg.Clusters) > 0 {
		cluster = cfg.Clusters[0].Name
	}
	if len(cfg.Users) > 0 {
		user = cfg.Users[0].Name
	}
	return cluster, user
}

func findCluster(cfg kubeConfigFile, name string) (kubeNamedCluster, bool) {
	for _, item := range cfg.Clusters {
		if item.Name == name {
			return item, true
		}
	}
	if len(cfg.Clusters) > 0 {
		return cfg.Clusters[0], true
	}
	return kubeNamedCluster{}, false
}

func findUser(cfg kubeConfigFile, name string) (kubeNamedUser, bool) {
	for _, item := range cfg.Users {
		if item.Name == name {
			return item, true
		}
	}
	if len(cfg.Users) > 0 {
		return cfg.Users[0], true
	}
	return kubeNamedUser{}, false
}

func kubeDataOrFile(dir, data, path string) ([]byte, error) {
	if data != "" {
		raw, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, merr.Server("invalid kubeconfig embedded data")
		}
		return raw, nil
	}
	if path == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(resolveKubePath(dir, path))
	if err != nil {
		return nil, merr.Server("failed to read kubeconfig file")
	}
	return raw, nil
}

func resolveKubePath(dir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(dir, path)
}
