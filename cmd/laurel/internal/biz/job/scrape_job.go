package job

import (
	"context"
	"io"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
	"github.com/aide-family/moon/pkg/util/hash"
)

var _ cron_server.CronJob = (*scrapeJob)(nil)

func NewScrapeJob(target *bo.ScrapeTarget, logger log.Logger) cron_server.CronJob {
	return &scrapeJob{
		target: target,
		helper: log.NewHelper(log.With(logger, "module", "job.scrape", "target", target.Target)),
	}
}

type scrapeJob struct {
	target *bo.ScrapeTarget
	helper *log.Helper
	id     cron.EntryID
}

// ID implements cron_server.CronJob.
func (s *scrapeJob) ID() cron.EntryID {
	return s.id
}

// Index implements cron_server.CronJob.
func (s *scrapeJob) Index() string {
	index := s.target.JobName + s.target.Target + s.target.MetricsPath + s.target.Scheme
	return hash.MD5(index)
}

// IsImmediate implements cron_server.CronJob.
func (s *scrapeJob) IsImmediate() bool {
	return false
}

// Run implements cron_server.CronJob.
func (s *scrapeJob) Run() {
	s.helper.Infof("scrape job running")
	resp, err := s.target.Do(context.Background())
	if err != nil {
		s.helper.Errorf("scrape job failed: %v", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.helper.Errorf("scrape job failed: %v", err)
		return
	}
	s.helper.Infof("scrape job success: %s", string(body))
}

// Spec implements cron_server.CronJob.
func (s *scrapeJob) Spec() cron_server.CronSpec {
	if s.target.Interval <= 0 {
		return cron_server.CronSpecEvery(15 * time.Second)
	}
	return cron_server.CronSpecEvery(s.target.Interval)
}

// WithID implements cron_server.CronJob.
func (s *scrapeJob) WithID(id cron.EntryID) cron_server.CronJob {
	s.id = id
	return s
}
