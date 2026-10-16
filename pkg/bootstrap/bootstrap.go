package bootstrap

import (
	"context"
	"os"
	"regexp"

	"github.com/go-kratos/kratos/v3/config"
)

var envPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::([^}]*))?\}`)

// Expand replaces ${VAR:default} placeholders. An empty environment value overrides the default.
func Expand(content string) string {
	return envPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := envPattern.FindStringSubmatch(match)
		if value, ok := os.LookupEnv(parts[1]); ok {
			return value
		}
		return parts[2]
	})
}

// NewSource loads a Bootstrap YAML file and expands environment placeholders.
func NewSource(path string) config.Source {
	return &fileSource{path: path}
}

type fileSource struct {
	path string
}

func (s *fileSource) Load() ([]*config.KeyValue, error) {
	body, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	return []*config.KeyValue{{
		Key:    s.path,
		Value:  []byte(Expand(string(body))),
		Format: "yaml",
	}}, nil
}

func (s *fileSource) Watch() (config.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &watcher{ctx: ctx, cancel: cancel}, nil
}

type watcher struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (w *watcher) Next() ([]*config.KeyValue, error) {
	<-w.ctx.Done()
	return nil, w.ctx.Err()
}

func (w *watcher) Stop() error {
	w.cancel()
	return nil
}
