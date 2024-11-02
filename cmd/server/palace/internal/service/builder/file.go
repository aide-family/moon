package builder

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/log"
	"gopkg.in/yaml.v3"
)

var _ IFileModuleBuild = (*fileModuleBuild)(nil)

type (
	fileModuleBuild struct {
		ctx context.Context
	}

	IFileModuleBuild interface {
		WithUpdateFileRequest(bytes []byte) IUploadFileRequest
	}

	IUploadFileRequest interface {
		ToJson() string
		ToYaml() string
	}
	updateFileRequest struct {
		ctx   context.Context
		bytes []byte
	}
)

func (u *updateFileRequest) ToYaml() string {
	var content map[string]interface{}
	if types.IsNil(u) || types.IsNil(u.bytes) {
		return ""
	}
	if err := json.Unmarshal(u.bytes, &content); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	yamlData, err := yaml.Marshal(content)
	if err != nil {
		log.Fatalf("Failed to marshal to YAML: %v", err)
	}

	return string(yamlData)
}

func (u *updateFileRequest) ToJson() string {
	var content map[string]interface{}
	if types.IsNil(u) || types.IsNil(u.bytes) {
		return ""
	}

	if err := yaml.Unmarshal(u.bytes, &content); err != nil {
		log.Fatalf("file unmarshal error: %v", err)
		return ""
	}

	yamlData, err := yaml.Marshal(content)
	if err != nil {
		log.Fatalf("Failed to marshal to json: %v", err)
	}
	return string(yamlData)
}

func (u *fileModuleBuild) WithUpdateFileRequest(bytes []byte) IUploadFileRequest {
	return &updateFileRequest{
		ctx:   u.ctx,
		bytes: bytes,
	}
}
