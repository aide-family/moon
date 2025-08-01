// Package job is a job package for kratos.
package job

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
)

var _ cron_server.CronJob = (*menuJob)(nil)

func NewMenuJob(
	menuRepo repository.Menu,
	cacheRepo repository.Cache,
	logger log.Logger,
) cron_server.CronJob {
	return &menuJob{
		menuRepo:  menuRepo,
		cacheRepo: cacheRepo,
		helper:    log.NewHelper(log.With(logger, "module", "job.menu")),
		index:     "cache.menu",
		id:        0,
		spec:      cron_server.CronSpecEvery(10 * time.Minute),
	}
}

type menuJob struct {
	index     string
	id        cron.EntryID
	spec      cron_server.CronSpec
	helper    *log.Helper
	menuRepo  repository.Menu
	cacheRepo repository.Cache
}

// ID implements server.CronJob.
func (m *menuJob) ID() cron.EntryID {
	return m.id
}

// Index implements server.CronJob.
func (m *menuJob) Index() string {
	return m.index
}

// Run implements server.CronJob.
func (m *menuJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	menus, err := m.menuRepo.FindAll(ctx)
	if err != nil {
		m.helper.Errorw("method", "Run", "err", err)
		return
	}
	if len(menus) == 0 {
		return
	}
	if err := m.cacheRepo.CacheMenus(ctx, menus...); err != nil {
		m.helper.Errorw("method", "Run", "err", err)
		return
	}
}

// Spec implements server.CronJob.
func (m *menuJob) Spec() cron_server.CronSpec {
	return m.spec
}

// WithID implements server.CronJob.
func (m *menuJob) WithID(id cron.EntryID) cron_server.CronJob {
	m.id = id
	return m
}

// IsImmediate implements server.CronJob.
func (m *menuJob) IsImmediate() bool {
	return true
}
