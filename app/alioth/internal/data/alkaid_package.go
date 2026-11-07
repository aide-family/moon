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

// expandPackageTemplate 展开 <version>/<arch>（兼容早期的 {version}/{arch} 写法）。
func expandPackageTemplate(template, version, arch string) string {
	return strings.NewReplacer(
		"<version>", version, "<arch>", arch,
		"{version}", version, "{arch}", arch,
	).Replace(template)
}

func (source *agentPackageSource) Fetch(ctx context.Context, arch string) (string, io.ReadCloser, error) {
	if source.localPath != "" {
		path := expandPackageTemplate(source.localPath, source.version, arch)
		file, err := os.Open(path)
		if err == nil {
			return filepath.Base(path), file, nil
		}
		// 镜像内置的是 <version>/<arch> 模板路径；本地开发没有这个目录时继续走下载。
		// 用户显式配置的固定路径缺失则直接报错，避免悄悄换成远程包。
		if !strings.Contains(source.localPath, "<") || source.downloadURL == "" {
			return "", nil, merr.Invalid(fmt.Sprintf("alkaid package is unavailable: %v", err))
		}
		if source.log != nil {
			source.log.Warn(fmt.Sprintf("bundled alkaid package is missing, falling back to download: %s", path))
		}
	}
	if source.downloadURL == "" {
		return "", nil, merr.Invalid("alkaid download url is empty")
	}
	if source.version == "" || source.version == "dev" {
		return "", nil, merr.Invalid("alkaid version is unknown; set deploy.alkaid_version or deploy.alkaid_package_path")
	}
	address := expandPackageTemplate(source.downloadURL, source.version, arch)
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
