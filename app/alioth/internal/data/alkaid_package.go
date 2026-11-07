package data

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/buildinfo"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	defaultPackageTimeout = 2 * time.Minute
	// maxPackageBytes 防止被错误地址指到超大文件，安装包通常在几十 MB 以内。
	maxPackageBytes = 256 << 20
)

// agentPackageSource 按机器架构取摇光安装包。
// 优先使用本地文件（离线部署），否则按模板从发布地址下载。
type agentPackageSource struct {
	version     string
	downloadURL string
	localPath   string
	client      *http.Client
	log         *slog.Logger
}

// NewAgentPackageSource 读取 deploy 配置里的安装包来源。
func NewAgentPackageSource(cfg *conf.Deploy, logger *slog.Logger) repository.AgentPackageSource {
	// 安装包通常比一条 SSH 命令大，下载超时至少保留两分钟。
	timeout := defaultPackageTimeout
	if cfg.GetTimeout() != nil && cfg.GetTimeout().AsDuration() > timeout {
		timeout = cfg.GetTimeout().AsDuration()
	}
	version := strings.TrimSpace(cfg.GetAlkaidVersion())
	if version == "" {
		version = strings.TrimSpace(buildinfo.Version)
	}
	return &agentPackageSource{
		version:     version,
		downloadURL: strings.TrimSpace(cfg.GetAlkaidDownloadUrl()),
		localPath:   strings.TrimSpace(cfg.GetAlkaidPackagePath()),
		client:      &http.Client{Timeout: timeout},
		log:         logger,
	}
}

func (source *agentPackageSource) Fetch(ctx context.Context, arch string) (string, io.ReadCloser, error) {
	if source.localPath != "" {
		file, err := os.Open(source.localPath)
		if err != nil {
			return "", nil, merr.Invalid(fmt.Sprintf("alkaid package is unavailable: %v", err))
		}
		return filepath.Base(source.localPath), file, nil
	}
	if source.downloadURL == "" {
		return "", nil, merr.Invalid("alkaid download url is empty")
	}
	if source.version == "" || source.version == "dev" {
		return "", nil, merr.Invalid("alkaid version is unknown; set deploy.alkaid_version or deploy.alkaid_package_path")
	}
	address := strings.NewReplacer("{version}", source.version, "{arch}", arch).Replace(source.downloadURL)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return "", nil, merr.Invalid(fmt.Sprintf("alkaid download request failed: %v", err))
	}
	response, err := source.client.Do(request)
	if err != nil {
		return "", nil, merr.Server(fmt.Sprintf("alkaid download failed: %v", err))
	}
	if response.StatusCode != http.StatusOK {
		_ = response.Body.Close()
		return "", nil, merr.Server(fmt.Sprintf("alkaid download failed: %s", response.Status))
	}
	if response.ContentLength > maxPackageBytes {
		_ = response.Body.Close()
		return "", nil, merr.Server(fmt.Sprintf("alkaid package is too large: %d bytes", response.ContentLength))
	}
	name := filepath.Base(address)
	if name == "" || name == "." || name == "/" {
		name = "alkaid-package.tar.gz"
	}
	if source.log != nil {
		source.log.Info(fmt.Sprintf("alkaid package: %s", address))
	}
	return name, response.Body, nil
}
