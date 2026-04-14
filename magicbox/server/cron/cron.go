// Package cron provides a cron job server.
package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/magicbox/log"
	"github.com/aide-family/magicbox/safety"
)

var _ transport.Server = (*Server)(nil)

type Option func(*Server)

func WithCronJobs(jobs ...CronJob) Option {
	return func(c *Server) {
		for _, job := range jobs {
			wrappedJob := WrapCronJobWithMetrics(job)
			id, ok := c.runner.Get(wrappedJob.Index())
			if ok {
				c.cron.Remove(id)
			}
			id, err := c.cron.AddJob(string(wrappedJob.Spec()), wrappedJob)
			if err != nil {
				c.helper.Errorf("[CronJob] %s add job %s error: %v", c.name, wrappedJob.Index(), err)
				continue
			}
			c.jobs.Set(wrappedJob.Index(), wrappedJob)
			c.runner.Set(wrappedJob.Index(), id)
			if wrappedJob.IsImmediate() {
				safety.Go(c.ctx, fmt.Sprintf("cron-job-immediate-%s", wrappedJob.Index()), func(ctx context.Context) error {
					if err := ctx.Err(); err != nil {
						return err
					}
					wrappedJob.Run()
					return nil
				})
			}
		}
	}
}

func WithCronJobChannel(ch <-chan CronJob) Option {
	return func(c *Server) {
		c.wg.Add(1)
		safety.Go(c.ctx, "cron-job-channel", func(ctx context.Context) error {
			defer c.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return nil
				case job, ok := <-ch:
					if !ok {
						return nil
					}
					WithCronJobs(job)(c)
				}
			}
		})
	}
}

func WithRemoveJobChannel(ch <-chan string) Option {
	return func(c *Server) {
		c.wg.Add(1)
		safety.Go(c.ctx, "cron-job-remove-channel", func(ctx context.Context) error {
			defer c.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return nil
				case jobKey, ok := <-ch:
					if !ok {
						return nil
					}
					if id, ok := c.runner.Get(jobKey); ok {
						c.cron.Remove(id)
						c.runner.Delete(jobKey)
					}
					if _, ok := c.jobs.Get(jobKey); ok {
						c.jobs.Delete(jobKey)
					}
				}
			}
		})
	}
}

func New(name string, logger log.Interface, opts ...Option) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Server{
		name:   name,
		helper: klog.NewHelper(logger),
		jobs:   safety.NewMap(map[string]CronJob{}),
		runner: safety.NewMap(map[string]cron.EntryID{}),
		cron:   cron.New(cron.WithSeconds()),
		ctx:    ctx,
		cancel: cancel,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Server struct {
	name   string
	helper *klog.Helper
	jobs   *safety.Map[string, CronJob]
	runner *safety.Map[string, cron.EntryID]
	cron   *cron.Cron
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Start implements transport.Server.
func (c *Server) Start(ctx context.Context) error {
	defer c.helper.WithContext(ctx).Infof("[CronJob] %s server started", c.name)
	c.cron.Start()
	return nil
}

// Stop implements transport.Server.
func (c *Server) Stop(ctx context.Context) error {
	defer c.helper.WithContext(ctx).Infof("[CronJob] %s server stopped", c.name)
	c.cancel()
	waitDone := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(waitDone)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-waitDone:
	}
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
