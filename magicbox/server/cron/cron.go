// Package cron provides a cron job server.
package cron

import (
	"context"
	"fmt"

	"time"

	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/magicbox/log"
	"github.com/aide-family/magicbox/safety"
)

var _ transport.Server = (*cronJobServer)(nil)

type Option func(*cronJobServer)

func WithCronJobs(jobs ...CronJob) Option {
	return func(c *cronJobServer) {
		for _, job := range jobs {
			wrappedJob := WrapCronJobWithMetrics(job)
			id, err := c.cron.AddJob(string(wrappedJob.Spec()), wrappedJob)
			if err != nil {
				c.helper.Errorf("[CronJob] %s add job %s error: %v", c.name, wrappedJob.Index(), err)
				continue
			}
			c.jobs.Set(wrappedJob.Index(), wrappedJob)
			c.runner.Set(wrappedJob.Index(), id)
			if wrappedJob.IsImmediate() {
				safety.Go(context.Background(), fmt.Sprintf("cron-job-immediate-%s", wrappedJob.Index()), func(ctx context.Context) error {
					wrappedJob.Run()
					return nil
				})
			}
		}
	}
}

func WithCronJobChannel(ch <-chan CronJob) Option {
	return func(c *cronJobServer) {
		ctx := context.Background()
		safety.Go(ctx, "cron-job-channel", func(ctx context.Context) error {
			for job := range ch {
				WithCronJobs(job)(c)
			}
			return nil
		})
	}
}

func WithRemoveJobChannel(ch <-chan string) Option {
	return func(c *cronJobServer) {
		ctx := context.Background()
		safety.Go(ctx, "cron-job-remove-channel", func(ctx context.Context) error {
			for jobKey := range ch {
				if id, ok := c.runner.Get(jobKey); ok {
					c.cron.Remove(id)
					c.runner.Delete(jobKey)
				}
				if _, ok := c.jobs.Get(jobKey); ok {
					c.jobs.Delete(jobKey)
				}
			}
			return nil
		})
	}
}

func NewCronJobServer(name string, logger log.Interface, opts ...Option) transport.Server {
	c := &cronJobServer{
		name:   name,
		helper: klog.NewHelper(klog.With(logger, "module", name)),
		jobs:   safety.NewMap(map[string]CronJob{}),
		runner: safety.NewMap(map[string]cron.EntryID{}),
		cron:   cron.New(cron.WithSeconds()),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type cronJobServer struct {
	name   string
	helper *klog.Helper
	jobs   *safety.Map[string, CronJob]
	runner *safety.Map[string, cron.EntryID]
	cron   *cron.Cron
}

// Start implements transport.Server.
func (c *cronJobServer) Start(ctx context.Context) error {
	defer c.helper.WithContext(ctx).Infof("[CronJob] %s server started", c.name)
	c.cron.Start()
	return nil
}

// Stop implements transport.Server.
func (c *cronJobServer) Stop(ctx context.Context) error {
	defer c.helper.WithContext(ctx).Infof("[CronJob] %s server stopped", c.name)
	c.cron.Stop()
	c.jobs.Clear()
	c.runner.Clear()
	return nil
}

type CronJob interface {
	cron.Job
	Index() string
	Spec() CronSpec
	IsImmediate() bool
}

type CronSpec string

const (
	CronSpecYearly   CronSpec = "@yearly"
	CronSpecAnnually CronSpec = "@annually"
	CronSpecMonthly  CronSpec = "@monthly"
	CronSpecWeekly   CronSpec = "@weekly"
	CronSpecDaily    CronSpec = "@daily"
	CronSpecMidnight CronSpec = "@midnight"
	CronSpecHourly   CronSpec = "@hourly"
)

func CronSpecEvery(duration time.Duration) CronSpec {
	return CronSpec("@every " + duration.String())
}

func CronSpecCustom(s, m, h, d, M, w string) CronSpec {
	return CronSpec(s + " " + m + " " + h + " " + d + " " + M + " " + w)
}
