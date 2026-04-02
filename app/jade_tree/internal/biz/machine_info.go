// Package biz provides core business use cases.
package biz

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
)

type MachineInfo struct {
	machineInfoRepo repository.MachineInfoProvider
	helper          *klog.Helper

	localUpdateInterval time.Duration
}

func NewMachineInfo(machineInfoRepo repository.MachineInfoProvider, helper *klog.Helper) *MachineInfo {
	m := &MachineInfo{
		machineInfoRepo:     machineInfoRepo,
		helper:              helper,
		localUpdateInterval: 10 * time.Minute,
	}
	go m.localRefreshLoop()
	return m
}

func (m *MachineInfo) GetMachineInfo(ctx context.Context) (*bo.MachineInfoBo, error) {
	mi, err := m.machineInfoRepo.GetMachineInfoByMachineUUID(ctx, m.machineInfoRepo.GetLocalMachineUUID())
	if err == nil {
		return mi, nil
	}
	if !merr.IsNotFound(err) {
		return nil, err
	}
	local, err := m.machineInfoRepo.Collect(ctx)
	if err != nil {
		return nil, err
	}

	go m.machineInfoRepo.UpdateLocalMachineInfo(safety.CopyValueCtx(ctx), local)

	return local, nil
}

func (m *MachineInfo) ListClusterMachineInfos(ctx context.Context, req *bo.ListMachineInfosBo) (*bo.PageResponseBo[*bo.MachineInfoBo], error) {
	return m.machineInfoRepo.ListMachineInfos(ctx, req)
}

func (m *MachineInfo) ReportMachineInfos(ctx context.Context, incoming []*bo.MachineInfoBo) error {
	if len(incoming) == 0 {
		return nil
	}

	uuids := make([]string, 0, len(incoming))
	for _, mi := range incoming {
		if mi == nil {
			continue
		}
		uuids = append(uuids, mi.MachineUUID)
	}
	existing, err := m.machineInfoRepo.GetMachineInfosByMachineUUIDs(ctx, uuids)
	if err != nil {
		return err
	}
	existingMap := make(map[string]*bo.MachineInfoBo, len(existing))
	for _, mi := range existing {
		if mi == nil {
			continue
		}
		existingMap[mi.MachineUUID] = mi
	}

	// Deduplicate incoming payload by machine UUID and merge duplicates.
	mergedIncoming := make(map[string]struct{}, len(incoming))
	toUpsert := make([]*bo.MachineInfoBo, 0, len(incoming))
	for _, mi := range incoming {
		if mi == nil {
			continue
		}
		if mi.MachineUUID == "" {
			continue
		}

		if _, ok := mergedIncoming[mi.MachineUUID]; ok {
			continue
		}
		mi.Source = enum.MachineInfoSource_MachineInfoSource_ORIGIN
		if existing, ok := existingMap[mi.MachineUUID]; ok {
			mi.ID = existing.ID
			mi.Source = existing.Source
		}
		mergedIncoming[mi.MachineUUID] = struct{}{}
		toUpsert = append(toUpsert, mi)
	}

	return m.machineInfoRepo.UpsertMachineInfos(ctx, toUpsert)
}

func (m *MachineInfo) localRefreshLoop() {
	ticker := time.NewTicker(m.localUpdateInterval)
	defer ticker.Stop()
	for range ticker.C {
		if err := m.refreshLocalNow(context.Background()); err != nil && m.helper != nil {
			m.helper.Errorw("msg", "refresh local machine info failed", "error", err)
		}
	}
}

func (m *MachineInfo) refreshLocalNow(ctx context.Context) error {
	local, err := m.machineInfoRepo.Collect(ctx)
	if err != nil {
		return err
	}
	if local == nil || local.MachineUUID == "" {
		return merr.ErrorInvalidArgument("local machine info is invalid")
	}

	local.Source = enum.MachineInfoSource_MachineInfoSource_LOCAL
	return m.machineInfoRepo.UpdateLocalMachineInfo(ctx, local)
}
