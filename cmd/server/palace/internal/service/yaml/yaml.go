package service

import (
	"context"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"

	yamlapi "github.com/aide-family/moon/api/admin/yaml"
)

// YamlService yaml操作相关服务
type YamlService struct {
	yamlapi.UnimplementedYamlServer
}

func NewYamlService() *YamlService {
	return &YamlService{}
}

func (s *YamlService) YamlToJson(ctx context.Context, req *yamlapi.UploadYamlFileRequest) (*yamlapi.UploadYamlFileReply, error) {
	builder.NewParamsBuild().WithContext(ctx).YamlModuleBuilder().WithUploadYamlRequest(req).ToJson()
	return &yamlapi.UploadYamlFileReply{}, nil
}
func (s *YamlService) JsonToYaml(ctx context.Context, req *yamlapi.UploadJsonRequest) (*yamlapi.UploadJsonReply, error) {
	builder.NewParamsBuild().WithContext(ctx).YamlModuleBuilder().WithUpdateJsonRequest(req).ToYaml()
	return &yamlapi.UploadJsonReply{}, nil
}
