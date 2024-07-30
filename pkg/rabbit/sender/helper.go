package sender

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// JsonProvider json
type JsonProvider struct {
}

// Provider provider json
func (d *JsonProvider) Provider(in []byte, out any) error {
	return json.Unmarshal(in, out)
}

// YamlProvider yaml
type YamlProvider struct {
}

// Provider provider yaml
func (d *YamlProvider) Provider(in []byte, out any) error {
	return yaml.Unmarshal(in, out)
}
