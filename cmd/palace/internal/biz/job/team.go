package job

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
)

var _ cron_server.CronJob = (*teamJob)(nil)

func NewTeamJob(
	teamRepo repository.Team,
	cacheRepo repository.Cache,
	logger log.Logger,
) cron_server.CronJob {
	return &teamJob{
		index:     "cache.team",
		id:        0,
		spec:      cron_server.CronSpecEvery(10 * time.Minute),
		helper:    log.NewHelper(log.With(logger, "module", "job.team")),
		teamRepo:  teamRepo,
		cacheRepo: cacheRepo,
	}
}

type teamJob struct {
	index     string
	id        cron.EntryID
	spec      cron_server.CronSpec
	helper    *log.Helper
	teamRepo  repository.Team
	cacheRepo repository.Cache
}

func (t *teamJob) Run() {
	pageReq := &bo.TeamListRequest{
		PaginationRequest: &bo.PaginationRequest{
			Page:  1,
			Limit: 100,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for {
		teamListReply, err := t.teamRepo.List(ctx, pageReq)
		if err != nil {
			t.helper.Warnw("msg", "list team fail", "err", err)
			continue
		}
		if len(teamListReply.Items) == 0 {
			break
		}
		if err := t.cacheRepo.CacheTeams(ctx, teamListReply.Items...); err != nil {
			t.helper.Warnw("msg", "cache team fail", "err", err)
			break
		}
		if teamListReply.Total < pageReq.Limit {
			break
		}
		pageReq.Page++
	}
}

func (t *teamJob) ID() cron.EntryID {
	return t.id
}

func (t *teamJob) Index() string {
	return t.index
}

func (t *teamJob) Spec() cron_server.CronSpec {
	return t.spec
}

func (t *teamJob) WithID(id cron.EntryID) cron_server.CronJob {
	t.id = id
	return t
}

// IsImmediate implements server.CronJob.
func (t *teamJob) IsImmediate() bool {
	return false
}
