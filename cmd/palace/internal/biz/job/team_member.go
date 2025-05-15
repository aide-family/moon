package job

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/plugin/server"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"
)

var _ server.CronJob = (*teamMemberJob)(nil)

func NewTeamMemberJob(
	teamMemberRepo repository.Member,
	cacheRepo repository.Cache,
	logger log.Logger,
) server.CronJob {
	return &teamMemberJob{
		index:          "cache.team.member",
		id:             0,
		spec:           server.CronSpecEvery(10 * time.Minute),
		helper:         log.NewHelper(log.With(logger, "module", "job.team.member")),
		teamMemberRepo: teamMemberRepo,
		cacheRepo:      cacheRepo,
	}
}

type teamMemberJob struct {
	index          string
	id             cron.EntryID
	spec           server.CronSpec
	helper         *log.Helper
	teamMemberRepo repository.Member
	cacheRepo      repository.Cache
}

func (t *teamMemberJob) Run() {
	pageReq := &bo.TeamMemberListRequest{
		PaginationRequest: &bo.PaginationRequest{
			Page:  1,
			Limit: 100,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for {
		teamMemberListReply, err := t.teamMemberRepo.List(ctx, pageReq)
		if err != nil {
			t.helper.Warnw("msg", "list team member fail", "err", err)
			continue
		}
		if len(teamMemberListReply.Items) == 0 {
			break
		}
		if err := t.cacheRepo.CacheTeamMembers(ctx, teamMemberListReply.Items...); err != nil {
			t.helper.Warnw("msg", "cache team member fail", "err", err)
			break
		}
		if teamMemberListReply.Total < pageReq.Limit {
			break
		}
		pageReq.Page++
	}
}

func (t *teamMemberJob) ID() cron.EntryID {
	if t == nil {
		return 0
	}
	return t.id
}

func (t *teamMemberJob) Index() string {
	if t == nil {
		return ""
	}
	return t.index
}

func (t *teamMemberJob) Spec() server.CronSpec {
	if t == nil {
		return server.CronSpecEvery(1 * time.Minute)
	}
	return t.spec
}

func (t *teamMemberJob) WithID(id cron.EntryID) server.CronJob {
	if t == nil {
		return nil
	}
	t.id = id
	return t
}
