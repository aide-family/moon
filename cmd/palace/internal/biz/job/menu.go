package job

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/plugin/server"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"
)

var _ server.CronJob = (*menuJob)(nil)

func NewMenuJob(
	menuRepo repository.Menu,
	cacheRepo repository.Cache,
	logger log.Logger,
) server.CronJob {
	return &menuJob{
		menuRepo:  menuRepo,
		cacheRepo: cacheRepo,
		helper:    log.NewHelper(log.With(logger, "module", "job.menu")),
		index:     "cache.menu",
		id:        0,
		spec:      server.CronSpecEvery(10 * time.Minute),
	}
}

type menuJob struct {
	index     string
	id        cron.EntryID
	spec      server.CronSpec
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
func (m *menuJob) Spec() server.CronSpec {
	return m.spec
}

// WithID implements server.CronJob.
func (m *menuJob) WithID(id cron.EntryID) server.CronJob {
	m.id = id
	return m
}

// IsImmediate implements server.CronJob.
func (m *menuJob) IsImmediate() bool {
	return true
}
