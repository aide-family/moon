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

	// IFileModuleBuild 文件模块构造器
	IFileModuleBuild interface {
		// WithUpdateFileRequest 更新文件请求参数构造器
		WithUpdateFileRequest(bytes []byte) IUploadFileRequest
	}

	// IUploadFileRequest 上传文件请求参数构造器
	IUploadFileRequest interface {
		// ToJson 转换为JSON字符串
		ToJSON() string
		// ToYaml 转换为YAML字符串
		ToYAML() string
	}
	updateFileRequest struct {
		ctx   context.Context
		bytes []byte
	}
)

// ToYAML 转换为YAML字符串
func (u *updateFileRequest) ToYAML() string {
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

// ToJSON 转换为JSON字符串
func (u *updateFileRequest) ToJSON() string {
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

// WithUpdateFileRequest 更新文件请求参数构造器
func (u *fileModuleBuild) WithUpdateFileRequest(bytes []byte) IUploadFileRequest {
	return &updateFileRequest{
		ctx:   u.ctx,
		bytes: bytes,
	}
}
