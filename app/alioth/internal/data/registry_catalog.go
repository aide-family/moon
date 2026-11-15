package data

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/imagesync"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/registry"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

var _ repository.ImageRegistryRepo = (*registryCatalog)(nil)

const registryCallTimeout = 15 * time.Second

func NewImageRegistryRepo(deps *Dependencies, logger *slog.Logger) repository.ImageRegistryRepo {
	return &registryCatalog{repoBase: newBase(deps, logger)}
}

type registryCatalog struct{ *repoBase }

func (c *registryCatalog) Probe(ctx context.Context, driver, endpoint, username, password string, allowInsecureTLS bool) (*bo.RegistryProbe, error) {
	reg := bo.Registry{Driver: driver, Endpoint: endpoint, Username: username, AllowInsecureTLS: allowInsecureTLS}
	repos, err := c.catalog(ctx, reg, password)
	if err != nil {
		return nil, err
	}
	return &bo.RegistryProbe{Endpoint: endpoint, Driver: driver, RepositoryCount: int32(len(repos))}, nil
}

func (c *registryCatalog) ProbeSaved(ctx context.Context, registryUID string) (*bo.RegistryProbe, error) {
	reg, password, err := c.savedRegistry(ctx, registryUID)
	if err != nil {
		return nil, err
	}
	return c.Probe(ctx, reg.Driver, reg.Endpoint, reg.Username, password, reg.AllowInsecureTLS)
}

func (c *registryCatalog) ListProjects(ctx context.Context, reg bo.Registry, nameFilter string) ([]bo.RegistryProject, error) {
	repos, err := c.catalog(ctx, reg, "")
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int32)
	for _, repo := range repos {
		project := projectOf(repo)
		if project == "" {
			continue
		}
		if nameFilter != "" && !strings.Contains(strings.ToLower(project), strings.ToLower(nameFilter)) {
			continue
		}
		counts[project]++
	}
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	items := make([]bo.RegistryProject, 0, len(names))
	for _, name := range names {
		items = append(items, bo.RegistryProject{Name: name, RepositoryCount: counts[name]})
	}
	return items, nil
}

func (c *registryCatalog) ListRepositories(ctx context.Context, reg bo.Registry, project, nameFilter string, page, pageSize int) ([]bo.RegistryRepository, int, error) {
	repos, err := c.catalog(ctx, reg, "")
	if err != nil {
		return nil, 0, err
	}
	filtered := make([]string, 0, len(repos))
	for _, repo := range repos {
		if project != "" && !strings.HasPrefix(repo, project+"/") {
			continue
		}
		if nameFilter != "" && !strings.Contains(strings.ToLower(repo), strings.ToLower(nameFilter)) {
			continue
		}
		filtered = append(filtered, repo)
	}
	sort.Strings(filtered)
	total := len(filtered)
	start := (page - 1) * pageSize
	if start > total {
		return []bo.RegistryRepository{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	items := make([]bo.RegistryRepository, 0, end-start)
	for _, repo := range filtered[start:end] {
		items = append(items, bo.RegistryRepository{Name: repo, Project: projectOf(repo)})
	}
	return items, total, nil
}

func (c *registryCatalog) ListTags(ctx context.Context, reg bo.Registry, project, repositoryName, architecture, nameFilter string, page, pageSize int) ([]bo.RegistryTag, int, error) {
	password, err := c.savedPassword(ctx, reg.UID)
	if err != nil {
		return nil, 0, err
	}
	full := fullRepository(project, repositoryName)
	tags, err := c.listTagNames(ctx, reg, full, password)
	if err != nil {
		return nil, 0, err
	}
	items := make([]bo.RegistryTag, 0, len(tags))
	callCtx, cancel := context.WithTimeout(ctx, registryCallTimeout)
	defer cancel()
	for _, tag := range tags {
		if nameFilter != "" && !strings.Contains(strings.ToLower(tag), strings.ToLower(nameFilter)) {
			continue
		}
		item, err := c.tagItem(callCtx, reg, full, tag, password)
		if err != nil {
			// 单个 Tag 的 manifest 读取失败不拖垮整页，保留可读的失败项。
			items = append(items, bo.RegistryTag{Tag: tag, Digest: "", Size: 0})
			continue
		}
		if architecture != "" && !hasArchitecture(item.Platforms, architecture) {
			continue
		}
		items = append(items, *item)
	}
	total := len(items)
	start := (page - 1) * pageSize
	if start > total {
		return []bo.RegistryTag{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return items[start:end], total, nil
}

func (c *registryCatalog) GetTag(ctx context.Context, reg bo.Registry, project, repositoryName, tag string) (*bo.RegistryTagDetail, error) {
	password, err := c.savedPassword(ctx, reg.UID)
	if err != nil {
		return nil, err
	}
	full := fullRepository(project, repositoryName)
	ref, err := c.reference(reg, full+":"+tag)
	if err != nil {
		return nil, err
	}
	desc, err := remote.Get(ref, c.options(ctx, reg, password)...)
	if err != nil {
		return nil, registryError(err)
	}
	detail := &bo.RegistryTagDetail{Tag: tag, Digest: desc.Digest.String(), Size: desc.Size, MediaType: string(desc.MediaType)}
	if desc.MediaType.IsIndex() {
		index, err := desc.ImageIndex()
		if err != nil {
			return nil, registryError(err)
		}
		manifest, err := index.IndexManifest()
		if err != nil {
			return nil, registryError(err)
		}
		for _, child := range manifest.Manifests {
			if child.Platform == nil {
				continue
			}
			detail.Platforms = append(detail.Platforms, bo.RegistryPlatform{
				OS: child.Platform.OS, Architecture: child.Platform.Architecture,
				Variant: child.Platform.Variant, Digest: child.Digest.String(), Size: child.Size,
			})
		}
	} else {
		image, err := desc.Image()
		if err != nil {
			return nil, registryError(err)
		}
		config, err := image.ConfigFile()
		if err != nil {
			return nil, registryError(err)
		}
		detail.Platforms = append(detail.Platforms, bo.RegistryPlatform{
			OS: config.OS, Architecture: config.Architecture, Variant: config.Variant, Digest: desc.Digest.String(),
		})
		detail.CreatedAt = config.Created.Time.Unix()
		detail.Labels = make([]string, 0, len(config.Config.Labels))
		for key := range config.Config.Labels {
			detail.Labels = append(detail.Labels, key)
		}
		sort.Strings(detail.Labels)
	}
	return detail, nil
}

func (c *registryCatalog) CopyTag(ctx context.Context, reg bo.Registry, project, repositoryName, tag, digest, targetRepository, targetTag string) (*bo.RegistryTag, error) {
	password, err := c.savedPassword(ctx, reg.UID)
	if err != nil {
		return nil, err
	}
	sourceRepo := fullRepository(project, repositoryName)
	sourceRef, err := c.reference(reg, sourceRepo+":"+tag)
	if err != nil {
		return nil, err
	}
	desc, err := remote.Get(sourceRef, c.options(ctx, reg, password)...)
	if err != nil {
		return nil, registryError(err)
	}
	if digest != "" && digest != desc.Digest.String() {
		sourceRef, err = c.reference(reg, sourceRepo+"@"+digest)
		if err != nil {
			return nil, err
		}
		desc, err = remote.Get(sourceRef, c.options(ctx, reg, password)...)
		if err != nil {
			return nil, registryError(err)
		}
	}
	targetRef, err := c.reference(reg, targetRepository+":"+targetTag)
	if err != nil {
		return nil, err
	}
	if desc.MediaType.IsIndex() {
		index, err := desc.ImageIndex()
		if err != nil {
			return nil, registryError(err)
		}
		if err := remote.WriteIndex(targetRef, index, c.options(ctx, reg, password)...); err != nil {
			return nil, registryError(err)
		}
	} else {
		image, err := desc.Image()
		if err != nil {
			return nil, registryError(err)
		}
		if err := remote.Write(targetRef, image, c.options(ctx, reg, password)...); err != nil {
			return nil, registryError(err)
		}
	}
	return c.tagItem(ctx, reg, targetRepository, targetTag, password)
}

func (c *registryCatalog) DeleteTag(ctx context.Context, reg bo.Registry, project, repositoryName, tag, digest string) error {
	password, err := c.savedPassword(ctx, reg.UID)
	if err != nil {
		return err
	}
	full := fullRepository(project, repositoryName)
	refText := full + ":" + tag
	if digest != "" {
		refText = full + "@" + digest
	}
	ref, err := c.reference(reg, refText)
	if err != nil {
		return err
	}
	if err := remote.Delete(ref, c.options(ctx, reg, password)...); err != nil {
		return registryError(err)
	}
	return nil
}

func (c *registryCatalog) ResolveSource(ctx context.Context, source bo.ImageSyncSource, reg *bo.Registry) (*bo.ImageSyncPreview, error) {
	var (
		ref      name.Reference
		password string
		regValue bo.Registry
		err      error
	)
	switch source.Kind {
	case "CONNECTION":
		if reg == nil {
			regValue, password, err = c.savedRegistry(ctx, source.RegistryUID)
			if err != nil {
				return nil, err
			}
		} else {
			regValue = *reg
			password, err = c.savedPassword(ctx, reg.UID)
			if err != nil {
				return nil, err
			}
		}
		ref, err = c.reference(regValue, source.Image)
		if err != nil {
			return nil, merr.Invalid("invalid source image")
		}
	default:
		regValue = bo.Registry{Endpoint: "https://" + sourceImageHost(source.Image), Driver: "DOCKER_V2", AllowInsecureTLS: source.AllowInsecureTLS}
		password = source.Password
		ref, err = parseImageReference(source.Image, source.AllowInsecureTLS)
		if err != nil {
			return nil, merr.Invalid("invalid source image")
		}
	}
	desc, err := remote.Get(ref, c.options(ctx, regValue, password)...)
	if err != nil {
		return nil, registryError(err)
	}
	preview := &bo.ImageSyncPreview{SourceImage: ref.Name(), SourceDigest: desc.Digest.String()}
	preview.SourcePlatforms, err = platformsOfDescriptor(ctx, desc)
	if err != nil {
		return nil, registryError(err)
	}
	return preview, nil
}

func (c *registryCatalog) TagExists(ctx context.Context, reg bo.Registry, repositoryName, tag string) (bool, string, error) {
	password, err := c.savedPassword(ctx, reg.UID)
	if err != nil {
		return false, "", err
	}
	ref, err := c.reference(reg, fullRepository("", repositoryName)+":"+tag)
	if err != nil {
		return false, "", err
	}
	desc, err := remote.Head(ref, c.options(ctx, reg, password)...)
	if err != nil {
		var transportErr *transport.Error
		if errors.As(err, &transportErr) && transportErr.StatusCode == http.StatusNotFound {
			return false, "", nil
		}
		return false, "", registryError(err)
	}
	return true, desc.Digest.String(), nil
}

func (c *registryCatalog) CopyImage(ctx context.Context, in bo.ImageSync) (string, error) {
	sourceReg, sourcePassword, err := c.sourceRegistry(ctx, in)
	if err != nil {
		return "", err
	}
	targetReg, targetPassword, err := c.savedRegistry(ctx, in.TargetRegistryUID)
	if err != nil {
		return "", err
	}
	sourceRef, err := c.reference(sourceReg, in.SourceRepository+":"+in.SourceTag)
	if err != nil {
		return "", merr.Invalid("invalid source image")
	}
	if in.SourceDigest != "" {
		sourceRef, _ = c.reference(sourceReg, in.SourceRepository+"@"+in.SourceDigest)
	}
	desc, err := remote.Get(sourceRef, c.options(ctx, sourceReg, sourcePassword)...)
	if err != nil {
		return "", registryError(err)
	}
	targetRef, err := c.reference(targetReg, in.TargetRepository+":"+in.TargetTag)
	if err != nil {
		return "", merr.Invalid("invalid target image")
	}
	selected := in.Mode == "SELECTED_PLATFORMS"
	if !selected {
		if desc.MediaType.IsIndex() {
			index, err := desc.ImageIndex()
			if err != nil {
				return "", registryError(err)
			}
			if err := remote.WriteIndex(targetRef, index, c.options(ctx, targetReg, targetPassword)...); err != nil {
				return "", registryError(err)
			}
		} else {
			image, err := desc.Image()
			if err != nil {
				return "", registryError(err)
			}
			if err := remote.Write(targetRef, image, c.options(ctx, targetReg, targetPassword)...); err != nil {
				return "", registryError(err)
			}
		}
	} else {
		if !desc.MediaType.IsIndex() {
			image, err := desc.Image()
			if err != nil {
				return "", registryError(err)
			}
			if err := remote.Write(targetRef, image, c.options(ctx, targetReg, targetPassword)...); err != nil {
				return "", registryError(err)
			}
		} else {
			index, err := desc.ImageIndex()
			if err != nil {
				return "", registryError(err)
			}
			manifest, err := index.IndexManifest()
			if err != nil {
				return "", registryError(err)
			}
			wanted := make(map[string]struct{}, len(in.SourcePlatforms))
			for _, platform := range in.SourcePlatforms {
				wanted[platform] = struct{}{}
			}
			children := make([]mutate.IndexAddendum, 0, len(in.SourcePlatforms))
			for _, child := range manifest.Manifests {
				if child.Platform == nil {
					continue
				}
				if _, ok := wanted[platformString(child.Platform.OS, child.Platform.Architecture, child.Platform.Variant)]; !ok {
					continue
				}
				image, err := index.Image(child.Digest)
				if err != nil {
					return "", registryError(err)
				}
				children = append(children, mutate.IndexAddendum{Add: image, Descriptor: v1.Descriptor{Platform: child.Platform}})
			}
			if len(children) == 0 {
				return "", merr.Invalid("selected platform not found")
			}
			if len(children) == 1 {
				if err := remote.Write(targetRef, children[0].Add.(v1.Image), c.options(ctx, targetReg, targetPassword)...); err != nil {
					return "", registryError(err)
				}
			} else if err := remote.WriteIndex(targetRef, mutate.AppendManifests(empty.Index, children...), c.options(ctx, targetReg, targetPassword)...); err != nil {
				return "", registryError(err)
			}
		}
	}
	head, err := remote.Head(targetRef, c.options(ctx, targetReg, targetPassword)...)
	if err != nil {
		return "", registryError(err)
	}
	return head.Digest.String(), nil
}

func (c *registryCatalog) catalog(ctx context.Context, reg bo.Registry, password string) ([]string, error) {
	host, insecure, err := endpointHost(reg.Endpoint)
	if err != nil {
		return nil, err
	}
	options := []name.Option{}
	if insecure || reg.AllowInsecureTLS {
		options = append(options, name.Insecure)
	}
	registryName, err := name.NewRegistry(host, options...)
	if err != nil {
		return nil, merr.Invalid("invalid registry endpoint")
	}
	repos, err := remote.Catalog(ctx, registryName, c.options(ctx, reg, password)...)
	if err != nil {
		var transportErr *transport.Error
		if errors.As(err, &transportErr) && (transportErr.StatusCode == http.StatusNotFound || transportErr.StatusCode == http.StatusMethodNotAllowed) {
			return []string{}, nil
		}
		return nil, registryError(err)
	}
	return repos, nil
}

func (c *registryCatalog) listTagNames(ctx context.Context, reg bo.Registry, repositoryName, password string) ([]string, error) {
	repo, err := c.repository(reg, repositoryName)
	if err != nil {
		return nil, err
	}
	tags, err := remote.List(repo, c.options(ctx, reg, password)...)
	if err != nil {
		return nil, registryError(err)
	}
	sort.Strings(tags)
	return tags, nil
}

func (c *registryCatalog) tagItem(ctx context.Context, reg bo.Registry, repositoryName, tag, password string) (*bo.RegistryTag, error) {
	ref, err := c.reference(reg, repositoryName+":"+tag)
	if err != nil {
		return nil, err
	}
	desc, err := remote.Get(ref, c.options(ctx, reg, password)...)
	if err != nil {
		return nil, err
	}
	platforms, err := platformsOfDescriptor(ctx, desc)
	if err != nil {
		return nil, err
	}
	item := &bo.RegistryTag{Tag: tag, Digest: desc.Digest.String(), Size: desc.Size, Platforms: platforms}
	if !desc.MediaType.IsIndex() {
		if image, err := desc.Image(); err == nil {
			if config, err := image.ConfigFile(); err == nil {
				item.PushedAt = config.Created.Time.Unix()
			}
		}
	}
	return item, nil
}

func (c *registryCatalog) options(ctx context.Context, reg bo.Registry, password string) []remote.Option {
	auth := authn.Authenticator(authn.Anonymous)
	if reg.Username != "" || password != "" {
		auth = &authn.Basic{Username: reg.Username, Password: password}
	}
	options := []remote.Option{remote.WithContext(ctx), remote.WithAuth(auth)}
	if reg.AllowInsecureTLS {
		options = append(options, remote.WithTransport(insecureTransport()))
	}
	return options
}

func (c *registryCatalog) reference(reg bo.Registry, refText string) (name.Reference, error) {
	host, insecure, err := endpointHost(reg.Endpoint)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(refText, "/") {
		return nil, merr.Invalid("invalid image reference")
	}
	if !strings.Contains(refText, host+"/") {
		refText = host + "/" + strings.TrimPrefix(refText, "/")
	}
	options := []name.Option{}
	if insecure || reg.AllowInsecureTLS {
		options = append(options, name.Insecure)
	}
	ref, err := name.ParseReference(refText, options...)
	if err != nil {
		return nil, merr.Invalid("invalid image reference")
	}
	return ref, nil
}

func (c *registryCatalog) repository(reg bo.Registry, repositoryName string) (name.Repository, error) {
	host, insecure, err := endpointHost(reg.Endpoint)
	if err != nil {
		return name.Repository{}, err
	}
	refText := strings.TrimPrefix(strings.TrimSpace(repositoryName), "/")
	if !strings.HasPrefix(refText, host+"/") {
		refText = host + "/" + refText
	}
	options := []name.Option{}
	if insecure || reg.AllowInsecureTLS {
		options = append(options, name.Insecure)
	}
	return name.NewRepository(refText, options...)
}

func (c *registryCatalog) savedRegistry(ctx context.Context, uid string) (bo.Registry, string, error) {
	row, err := c.client.Registry.Query().Where(registry.UIDEQ(uid), registry.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return bo.Registry{}, "", merr.NotFound("registry not found")
		}
		return bo.Registry{}, "", c.fail(err)
	}
	plain, err := c.box.decrypt(row.PasswordCipher)
	if err != nil {
		c.log.Error(fmt.Sprintf("registry credential: %v", err))
		return bo.Registry{}, "", merr.Server("request failed")
	}
	return toRegistry(row), plain, nil
}

func (c *registryCatalog) savedPassword(ctx context.Context, uid string) (string, error) {
	_, plain, err := c.savedRegistry(ctx, uid)
	return plain, err
}

func (c *registryCatalog) sourceRegistry(ctx context.Context, in bo.ImageSync) (bo.Registry, string, error) {
	if in.SourceKind == "CONNECTION" {
		return c.savedRegistry(ctx, in.SourceRegistryUID)
	}
	row, err := c.client.ImageSync.Query().Where(imagesync.UIDEQ(in.UID)).Only(ctx)
	if err != nil {
		return bo.Registry{}, "", c.fail(err)
	}
	password := ""
	if row.SourceCredentialCipher != "" {
		password, err = c.box.decrypt(row.SourceCredentialCipher)
		if err != nil {
			return bo.Registry{}, "", merr.Server("request failed")
		}
	}
	ref, err := parseImageReference(row.SourceImage, row.SourceAllowInsecureTLS)
	if err != nil {
		return bo.Registry{}, "", merr.Invalid("invalid source image")
	}
	return bo.Registry{
		Endpoint: "https://" + ref.Context().RegistryStr(),
		Driver:   "DOCKER_V2", Username: row.SourceUsername,
		AllowInsecureTLS: row.SourceAllowInsecureTLS,
	}, password, nil
}

func endpointHost(endpoint string) (string, bool, error) {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || u.Host == "" {
		return "", false, merr.Invalid("invalid registry endpoint")
	}
	return u.Host, strings.EqualFold(u.Scheme, "http"), nil
}

func fullRepository(project, repositoryName string) string {
	repositoryName = strings.Trim(strings.TrimSpace(repositoryName), "/")
	project = strings.Trim(strings.TrimSpace(project), "/")
	if project == "" || repositoryName == "" || strings.HasPrefix(repositoryName, project+"/") {
		return repositoryName
	}
	return project + "/" + repositoryName
}

func projectOf(repositoryName string) string {
	index := strings.Index(repositoryName, "/")
	if index <= 0 {
		return ""
	}
	return repositoryName[:index]
}

func hasArchitecture(platforms []string, architecture string) bool {
	for _, platform := range platforms {
		if strings.Contains(platform, "/"+architecture) || strings.HasSuffix(platform, architecture) {
			return true
		}
	}
	return false
}

func platformString(osName, architecture, variant string) string {
	value := osName + "/" + architecture
	if variant != "" {
		value += "/" + variant
	}
	return value
}

func platformsOfDescriptor(_ context.Context, desc *remote.Descriptor) ([]string, error) {
	if !desc.MediaType.IsIndex() {
		image, err := desc.Image()
		if err != nil {
			return nil, err
		}
		config, err := image.ConfigFile()
		if err != nil {
			return nil, err
		}
		return []string{platformString(config.OS, config.Architecture, config.Variant)}, nil
	}
	index, err := desc.ImageIndex()
	if err != nil {
		return nil, err
	}
	manifest, err := index.IndexManifest()
	if err != nil {
		return nil, err
	}
	platforms := make([]string, 0, len(manifest.Manifests))
	for _, child := range manifest.Manifests {
		if child.Platform == nil {
			continue
		}
		platforms = append(platforms, platformString(child.Platform.OS, child.Platform.Architecture, child.Platform.Variant))
	}
	sort.Strings(platforms)
	return platforms, nil
}

func parseImageReference(image string, insecure bool) (name.Reference, error) {
	image = strings.TrimSpace(image)
	if image == "" {
		return nil, merr.Invalid("image is required")
	}
	options := []name.Option{}
	if insecure {
		options = append(options, name.Insecure)
	}
	return name.ParseReference(image, options...)
}

func sourceImageHost(image string) string {
	ref, err := parseImageReference(image, false)
	if err != nil {
		return "docker.io"
	}
	return ref.Context().RegistryStr()
}

func insecureTransport() http.RoundTripper {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // 用户显式打开
	return transport
}

func registryError(err error) error {
	var transportErr *transport.Error
	if errors.As(err, &transportErr) {
		switch transportErr.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return merr.Forbidden("registry credentials were rejected")
		case http.StatusNotFound:
			return merr.NotFound("image not found")
		}
	}
	var timeout interface{ Timeout() bool }
	if errors.As(err, &timeout) && timeout.Timeout() {
		return merr.Invalid("registry is unreachable")
	}
	if strings.Contains(strings.ToLower(err.Error()), "no such host") ||
		strings.Contains(strings.ToLower(err.Error()), "connection refused") {
		return merr.Invalid("registry is unreachable")
	}
	return merr.Server("registry request failed")
}

var _ = strconv.Itoa
