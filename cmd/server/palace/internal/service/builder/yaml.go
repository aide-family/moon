package builder

import (
	"context"
	"encoding/json"

	yamlapi "github.com/aide-family/moon/api/admin/yaml"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"gopkg.in/yaml.v3"
)

var _ IYamlModuleBuild = (*yamlModuleBuild)(nil)

type (
	yamlModuleBuild struct {
		ctx context.Context
	}

	IYamlModuleBuild interface {
		WithUploadYamlRequest(*yamlapi.UploadYamlFileRequest) IUploadYamlRequest
		WithUpdateJsonRequest(*yamlapi.UploadJsonRequest) IUpdateJsonRequest
	}

	IUploadYamlRequest interface {
		ToJson() map[string]interface{}
	}

	uploadYamlRequest struct {
		ctx context.Context
		*yamlapi.UploadYamlFileRequest
	}

	IUpdateJsonRequest interface {
		ToYaml() string
	}

	updateJsonRequest struct {
		ctx context.Context
		*yamlapi.UploadJsonRequest
	}
)

func (u *updateJsonRequest) ToYaml() string {
	var content map[string]interface{}
	if types.IsNil(u) || types.IsNil(u.File) {
		return ""
	}
	if err := json.Unmarshal(u.File, &content); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	yamlData, err := yaml.Marshal(content)
	if err != nil {
		log.Fatalf("Failed to marshal to YAML: %v", err)
	}

	return string(yamlData)
}

func (u *uploadYamlRequest) ToJson() map[string]interface{} {
	var content map[string]interface{}
	if types.IsNil(u) || types.IsNil(u.File) {
		return content
	}

	if err := yaml.Unmarshal(u.File, &content); err != nil {
		log.Fatalf("yaml unmarshal error: %v", err)
		return nil
	}

	return content
}

func (u *yamlModuleBuild) WithUploadYamlRequest(req *yamlapi.UploadYamlFileRequest) IUploadYamlRequest {
	return &uploadYamlRequest{
		ctx:                   u.ctx,
		UploadYamlFileRequest: req,
	}
}

func (u *yamlModuleBuild) WithUpdateJsonRequest(req *yamlapi.UploadJsonRequest) IUpdateJsonRequest {
	return &updateJsonRequest{
		ctx:               u.ctx,
		UploadJsonRequest: req,
	}
}
