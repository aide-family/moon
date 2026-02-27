// Package conf provides configuration loading and management functionality.
package conf

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
	kconfig "github.com/go-kratos/kratos/v2/config"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
)

var (
	_ kconfig.Source  = (*bytesSource)(nil)
	_ kconfig.Watcher = (*noOpWatcher)(nil)
)

// NewBytesSource creates a new bytes source from the given data.
func NewBytesSource(data []byte) kconfig.Source {
	d := bytesSource(data)
	return &d
}

type bytesSource []byte

// Load implements config.Source.
func (b *bytesSource) Load() ([]*kconfig.KeyValue, error) {
	// Make a copy of the data to avoid external modifications
	data := make([]byte, len(*b))
	copy(data, *b)
	return []*kconfig.KeyValue{
		{
			Key:    "server",
			Value:  data,
			Format: format(*b),
		},
	}, nil
}

// format detects the format from the data content.
func format(data []byte) string {
	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
		return "json"
	}
	return "yaml"
}

// Watch implements config.Source.
func (b *bytesSource) Watch() (kconfig.Watcher, error) {
	return newNoOpWatcher(), nil
}

type noOpWatcher struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func newNoOpWatcher() kconfig.Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &noOpWatcher{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Next implements config.Watcher.
func (w *noOpWatcher) Next() ([]*kconfig.KeyValue, error) {
	<-w.ctx.Done()
	return nil, w.ctx.Err()
}

// Stop implements config.Watcher.
func (w *noOpWatcher) Stop() error {
	w.cancel()
	return nil
}

func Load(bc any, sources ...kconfig.Source) error {
	c := kconfig.New(kconfig.WithSource(sources...))
	if err := c.Load(); err != nil {
		return merr.ErrorInternalServer("load config failed").WithCause(err)
	}
	if err := c.Scan(bc); err != nil {
		return merr.ErrorInternalServer("scan config failed").WithCause(err)
	}
	return nil
}

type ServerConfig interface {
	GetAddress() string
	GetNetwork() string
	GetTimeout() *durationpb.Duration
	GetProtocol() config.Protocol
}

type JWTConfig interface {
	GetSecret() string
	GetExpire() *durationpb.Duration
	GetIssuer() string
}
