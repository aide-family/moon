package sender

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

type JsonProvider struct {
}

func (d *JsonProvider) Provider(in []byte, out any) error {
	return json.Unmarshal(in, out)
}

type YamlProvider struct {
}

func (d *YamlProvider) Provider(in []byte, out any) error {
	return yaml.Unmarshal(in, out)
}
