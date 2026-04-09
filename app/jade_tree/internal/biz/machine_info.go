// Package biz provides core business use cases.
package biz

import (
	"context"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/pkg/machine"
)

type MachineInfo struct {
	machineInfoRepo repository.MachineInfoProvider
	helper          *klog.Helper
}

func NewMachineInfo(machineInfoRepo repository.MachineInfoProvider, helper *klog.Helper) *MachineInfo {
	return &MachineInfo{
		machineInfoRepo: machineInfoRepo,
		helper:          helper,
	}
}

func (m *MachineInfo) GetMachineInfo(ctx context.Context) (*machine.MachineInfo, error) {
	mi, err := m.machineInfoRepo.GetMachineInfoByIdentity(ctx, m.machineInfoRepo.GetLocalMachineIdentity())
	if err == nil {
		m.fillLocalAgentInfo(ctx, mi)
		return mi, nil
	}
	if !merr.IsNotFound(err) {
		return nil, err
	}
	local, err := m.machineInfoRepo.Collect(ctx)
	if err != nil {
		return nil, err
	}
	if local != nil && local.Network != nil {
		local.Agent = m.machineInfoRepo.GetLocalAgentInfo(local.Network.LocalIP)
	}

	go m.machineInfoRepo.UpdateLocalMachineInfo(safety.CopyValueCtx(ctx), local)

	return local, nil
}

func (m *MachineInfo) fillLocalAgentInfo(ctx context.Context, mi *machine.MachineInfo) {
	if mi == nil || mi.Network == nil {
		return
	}
	localAgent := m.machineInfoRepo.GetLocalAgentInfo(mi.Network.LocalIP)
	if localAgent == nil {
		return
	}
	needUpdate := mi.Agent == nil || mi.Agent.Version == "" || mi.Agent.HTTPEndpoint == "" || mi.Agent.GRPCEndpoint == "" || mi.Agent.Endpoint == ""
	if !needUpdate {
		return
	}
	mi.Agent = localAgent
	go func() {
		if err := m.machineInfoRepo.UpdateLocalMachineInfo(safety.CopyValueCtx(ctx), mi); err != nil {
			m.helper.Errorw("msg", "backfill local agent info failed", "error", err)
		}
	}()
}

func (m *MachineInfo) ListClusterMachineInfos(ctx context.Context, req *bo.ListMachineInfosBo) (*bo.PageResponseBo[*machine.MachineInfo], error) {
	return m.machineInfoRepo.ListMachineInfos(ctx, req)
}

func (m *MachineInfo) ReportMachineInfos(ctx context.Context, incoming []*machine.MachineInfo) error {
	if len(incoming) == 0 {
		return nil
	}

	// Deduplicate incoming payload by machine UUID + hostname + local IP; persistence upserts by the same composite key.
	mergedIncoming := make(map[string]struct{}, len(incoming))
	toUpsert := make([]*machine.MachineInfo, 0, len(incoming))
	for _, mi := range incoming {
		if mi == nil {
			continue
		}
		if mi.MachineUUID == "" {
			continue
		}

		key := bo.NewMachineInfoIdentityBo(mi).DedupKey()
		if _, ok := mergedIncoming[key]; ok {
			continue
		}
		mi.Source = enum.MachineInfoSource_MachineInfoSource_ORIGIN
		mergedIncoming[key] = struct{}{}
		toUpsert = append(toUpsert, mi)
	}

	return m.machineInfoRepo.UpsertMachineInfos(ctx, toUpsert)
}

func (m *MachineInfo) RefreshLocalMachineInfo(ctx context.Context) (*machine.MachineInfo, error) {
	local, err := m.machineInfoRepo.Collect(ctx)
	if err != nil {
		return nil, err
	}
	if local == nil || local.MachineUUID == "" {
		return nil, merr.ErrorInvalidArgument("local machine info is invalid")
	}
	if local.Network != nil {
		local.Agent = m.machineInfoRepo.GetLocalAgentInfo(local.Network.LocalIP)
	}

	copyCtx := safety.CopyValueCtx(ctx)
	go func() {
		local.Source = enum.MachineInfoSource_MachineInfoSource_LOCAL
		if err := m.machineInfoRepo.UpdateLocalMachineInfo(copyCtx, local); err != nil {
			m.helper.Errorw("msg", "update local machine info failed", "error", err)
		}
	}()
	return local, nil
}
