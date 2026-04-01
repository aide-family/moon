// Package service provides transport-facing handlers.
package service

import (
	"context"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"

	"github.com/aide-family/jade_tree/internal/biz"
	"github.com/aide-family/jade_tree/internal/biz/bo"
)

func NewMachineInfoService(machineInfo *biz.MachineInfo) *MachineInfoService {
	return &MachineInfoService{machineInfo: machineInfo}
}

type MachineInfoService struct {
	apiv1.UnimplementedMachineInfoServer
	machineInfo *biz.MachineInfo
}

func (s *MachineInfoService) GetMachineInfo(ctx context.Context, req *apiv1.GetMachineInfoRequest) (*apiv1.GetMachineInfoReply, error) {
	info, err := s.machineInfo.GetMachineInfo(ctx)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1MachineInfoReply(info), nil
}
