package portal_service

import (
	"context"

	"github.com/aide-family/moon/pkg/api/palace/common"
	portalapi "github.com/aide-family/moon/pkg/api/palace/portal"
)

func NewHomeService() *HomeService {
	return &HomeService{}
}

type HomeService struct {
	portalapi.UnimplementedHomeServer
}

func (s *HomeService) Features(ctx context.Context, req *common.EmptyRequest) (*portalapi.FeaturesReply, error) {
	return &portalapi.FeaturesReply{}, nil
}

func (s *HomeService) Partners(ctx context.Context, req *common.EmptyRequest) (*portalapi.PartnersReply, error) {
	return &portalapi.PartnersReply{}, nil
}

func (s *HomeService) Footer(ctx context.Context, req *common.EmptyRequest) (*portalapi.FooterReply, error) {
	return &portalapi.FooterReply{}, nil
}
