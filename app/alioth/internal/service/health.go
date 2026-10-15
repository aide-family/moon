package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/health"
)

type HealthService struct {
	v1.UnimplementedHealthServiceServer
	uc *health.Usecase
}

func NewHealthService(uc *health.Usecase) *HealthService {
	return &HealthService{uc: uc}
}

func (s *HealthService) Check(context.Context, *v1.CheckRequest) (*v1.CheckResponse, error) {
	return &v1.CheckResponse{Status: s.uc.Status()}, nil
}
