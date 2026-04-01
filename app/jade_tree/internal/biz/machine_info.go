// Package biz provides core business use cases.
package biz

import (
	"context"
	"sync"
	"time"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
)

type MachineInfo struct {
	repo   repository.MachineInfoProvider
	helper *klog.Helper

	mu         sync.Mutex
	cond       *sync.Cond
	refreshing bool
	cache      *bo.MachineInfoBo
	expireAt   time.Time
	ttl        time.Duration
	nowFn      func() time.Time
}

func NewMachineInfo(repo repository.MachineInfoProvider, helper *klog.Helper) *MachineInfo {
	m := &MachineInfo{
		repo:   repo,
		helper: helper,
		ttl:    10 * time.Minute,
		nowFn:  time.Now,
	}
	m.cond = sync.NewCond(&m.mu)
	return m
}

func (m *MachineInfo) GetMachineInfo(ctx context.Context) (*bo.MachineInfoBo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.nowFn()
	if m.cache != nil && now.Before(m.expireAt) {
		data := m.cache
		return data, nil
	}
	for m.refreshing {
		m.cond.Wait()
		now = m.nowFn()
		if m.cache != nil && now.Before(m.expireAt) {
			data := m.cache
			return data, nil
		}
	}
	m.refreshing = true

	data, err := m.repo.Collect(ctx)

	m.refreshing = false
	m.cond.Broadcast()
	if err != nil {
		return nil, err
	}
	m.cache = data
	m.expireAt = m.nowFn().Add(m.ttl)
	return m.cache, nil
}
