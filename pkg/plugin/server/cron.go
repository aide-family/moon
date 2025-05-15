package server

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/moon/pkg/util/safety"
)

var _ transport.Server = (*CronJobServer)(nil)

type CronSpec string

const (
	CronSpecYearly CronSpec = "@yearly"

	CronSpecAnnually CronSpec = "@annually"

	CronSpecMonthly CronSpec = "@monthly"

	CronSpecWeekly CronSpec = "@weekly"

	CronSpecDaily CronSpec = "@daily"

	CronSpecMidnight CronSpec = "@midnight"

	CronSpecHourly CronSpec = "@hourly"
)

func CronSpecEvery(duration time.Duration) CronSpec {
	return CronSpec("@every " + duration.String())
}

func CronSpecCustom(s, m, h, d, M, w string) CronSpec {
	return CronSpec(s + " " + m + " " + h + " " + d + " " + M + " " + w)
}

type CronJob interface {
	cron.Job

	ID() cron.EntryID
	Index() string
	Spec() CronSpec
	WithID(id cron.EntryID) CronJob
}

type CronJobServer struct {
	name   string
	cron   *cron.Cron
	tasks  *safety.Map[string, CronJob]
	helper *log.Helper
}

func NewCronJobServer(name string, logger log.Logger, jobs ...CronJob) *CronJobServer {
	c := &CronJobServer{
		name:   name,
		cron:   cron.New(cron.WithSeconds()),
		tasks:  safety.NewMap[string, CronJob](),
		helper: log.NewHelper(logger),
	}
	for _, job := range jobs {
		c.AddJob(job)
	}
	return c
}

func (c *CronJobServer) AddJob(job CronJob) {
	if _, ok := c.tasks.Get(job.Index()); ok {
		return
	}
	id, err := c.cron.AddJob(string(job.Spec()), job)
	if err != nil {
		c.helper.Warnw("method", "add job", "err", err)
		return
	}
	job.WithID(id)
	c.tasks.Set(job.Index(), job)
}

func (c *CronJobServer) AddJobForce(job CronJob) {
	if oldJob, ok := c.tasks.Get(job.Index()); ok {
		defer c.cron.Remove(oldJob.ID())
	}
	id, err := c.cron.AddJob(string(job.Spec()), job)
	if err != nil {
		c.helper.Warnw("method", "add job", "err", err)
		return
	}
	job.WithID(id)
	c.tasks.Set(job.Index(), job)
}

func (c *CronJobServer) RemoveJob(job CronJob) {
	task, ok := c.tasks.Get(job.Index())
	if ok {
		c.cron.Remove(task.ID())
		c.tasks.Delete(job.Index())
	}
}

func (c *CronJobServer) Start(ctx context.Context) error {
	defer c.helper.WithContext(ctx).Infof("[CronJob] %s server started", c.name)
	c.cron.Start()
	return nil
}

func (c *CronJobServer) Stop(ctx context.Context) error {
	defer c.helper.WithContext(ctx).Infof("[CronJob] %s server stopped", c.name)
	c.cron.Stop()
	c.tasks.Clear()
	return nil
}
