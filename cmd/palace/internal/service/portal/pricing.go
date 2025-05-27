package portal_service

import (
	"context"

	"github.com/aide-family/moon/pkg/api/palace/common"
	portalapi "github.com/aide-family/moon/pkg/api/palace/portal"
)

func NewPricingService() *PricingService {
	return &PricingService{}
}

type PricingService struct {
	portalapi.UnimplementedPricingServer
}

func (s *PricingService) ListPackage(ctx context.Context, req *common.EmptyRequest) (*portalapi.ListPackageReply, error) {
	return &portalapi.ListPackageReply{}, nil
}
