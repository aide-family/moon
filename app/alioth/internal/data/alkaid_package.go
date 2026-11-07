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
	// 版本没注入时（例如 go run 或本地裸二进制）拼不出下载地址，但本机可能已经用
	// make package-alkaid 打好了包。这时把本地路径里的 <version> 换成通配符，
	// 命中 dist 下的实际包；有版本号时仍然是精确路径。
	versionKnown := source.version != "" && source.version != "dev"
	localVersion := source.version
	if !versionKnown {
		localVersion = "*"
	}
	var localPattern string
	if source.localPath != "" {
		localPattern = expandPackageTemplate(source.localPath, localVersion, arch)
		file, name, err := openLocalPackage(localPattern)
		if err == nil {
			return name, file, nil
		}
		// 镜像内置的是 <version>/<arch> 模板路径；本地开发没有这个目录时继续走下载。
		// 用户显式配置的固定路径缺失则直接报错，避免悄悄换成远程包。
		if !strings.Contains(source.localPath, "<") {
			return "", nil, merr.Invalid(fmt.Sprintf("alkaid package is unavailable: %v", err))
		}
		if source.log != nil {
			source.log.Warn(fmt.Sprintf("bundled alkaid package is missing, falling back to download: %s", localPattern))
		}
	}
	if source.downloadURL == "" {
		return "", nil, merr.Invalid("alkaid download url is empty")
	}
	if !versionKnown {
		return "", nil, merr.Invalid(unavailableMessage(source.version, arch, localPattern))
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

// openLocalPackage 打开本地安装包。模式里带通配符（版本未知时的 <version>）时，
// 取最近修改的那个文件，避免同名目录里堆积多个版本时随机命中旧的包。
func openLocalPackage(pattern string) (io.ReadCloser, string, error) {
	if !strings.ContainsAny(pattern, "*?[") {
		file, err := os.Open(pattern)
		if err != nil {
			return nil, "", err
		}
		return file, filepath.Base(pattern), nil
	}
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, "", err
	}
	newest := ""
	var newestTime time.Time
	for _, match := range matches {
		info, statErr := os.Stat(match)
		if statErr != nil || info.IsDir() {
			continue
		}
		if info.ModTime().After(newestTime) {
			newest, newestTime = match, info.ModTime()
		}
	}
	if newest == "" {
		return nil, "", merr.Invalid("no package matched " + pattern)
	}
	file, err := os.Open(newest)
	if err != nil {
		return nil, "", err
	}
	return file, filepath.Base(newest), nil
}

// unavailableMessage 说明为什么既没有本地包也没法下载，并给出下一步动作。
func unavailableMessage(version, arch, localPattern string) string {
	looked := ""
	if localPattern != "" {
		looked = fmt.Sprintf("no file matched %s, and ", localPattern)
	}
	return fmt.Sprintf(
		"alkaid package is unavailable for linux/%s: %sthe control plane version is unknown (%q), so the download URL cannot be expanded; run make package-alkaid ARCHES=%s and set deploy.alkaid_package_path (ALIOTH_DEPLOY_ALKAID_PACKAGE_PATH) to the built file, or set deploy.alkaid_version",
		arch, looked, version, arch,
	)
}
