// Package service provides transport-facing handlers.
package service

import (
	"context"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"

	"github.com/aide-family/jade_tree/internal/biz"
	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/magicbox/merr"
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

func (s *MachineInfoService) ReportMachineInfos(ctx context.Context, req *apiv1.ReportMachineInfosRequest) (*apiv1.ReportMachineInfosReply, error) {
	if req == nil {
		return nil, merr.ErrorInvalidArgument("request is required")
	}

	incoming := make([]*bo.MachineInfoBo, 0, len(req.GetMachines()))
	for _, mi := range req.GetMachines() {
		incoming = append(incoming, bo.FromAPIV1MachineInfoReply(mi))
	}

	err := s.machineInfo.ReportMachineInfos(ctx, incoming)
	if err != nil {
		return nil, err
	}
	return &apiv1.ReportMachineInfosReply{}, nil
}

func (s *MachineInfoService) GetClusterMachineInfos(ctx context.Context, req *apiv1.GetClusterMachineInfosRequest) (*apiv1.GetClusterMachineInfosReply, error) {
	if req == nil {
		return nil, merr.ErrorInvalidArgument("request is required")
	}

	listBo := bo.NewListMachineInfosBo(req.GetPage(), req.GetPageSize())
	page, err := s.machineInfo.ListClusterMachineInfos(ctx, listBo)
	if err != nil {
		return nil, err
	}

	out := &apiv1.GetClusterMachineInfosReply{
		Machines: make([]*apiv1.GetMachineInfoReply, 0, len(page.GetItems())),
	}
	for _, mi := range page.GetItems() {
		out.Machines = append(out.Machines, bo.ToAPIV1MachineInfoReply(mi))
	}
	out.Total = page.GetTotal()
	out.Page = page.GetPage()
	out.PageSize = page.GetPageSize()
	return out, nil
}
