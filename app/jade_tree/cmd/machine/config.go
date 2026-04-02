package machine

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultClientConfigPath = "~/.jade_tree/client.yaml"

type clientConfig struct {
	Endpoint  string   `yaml:"endpoint"`
	Endpoints []string `yaml:"endpoints"`
	JWT       string   `yaml:"jwt"`
}

func resolveConfigPath(path string) (string, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		p = defaultClientConfigPath
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, strings.TrimPrefix(p, "~/"))
	}
	return p, nil
}

func loadClientConfig(path string) (*clientConfig, error) {
	absPath, err := resolveConfigPath(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &clientConfig{}, nil
		}
		return nil, err
	}
	cfg := &clientConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
