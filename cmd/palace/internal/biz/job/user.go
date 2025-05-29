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

var _ cron_server.CronJob = (*userJob)(nil)

func NewUserJob(
	userRepo repository.User,
	cacheRepo repository.Cache,
	logger log.Logger,
) cron_server.CronJob {
	return &userJob{
		index:     "cache.user",
		id:        0,
		spec:      cron_server.CronSpecEvery(10 * time.Minute),
		helper:    log.NewHelper(log.With(logger, "module", "job.user")),
		userRepo:  userRepo,
		cacheRepo: cacheRepo,
	}
}

type userJob struct {
	index     string
	id        cron.EntryID
	spec      cron_server.CronSpec
	helper    *log.Helper
	userRepo  repository.User
	cacheRepo repository.Cache
}

func (u *userJob) Run() {
	pageReq := &bo.UserListRequest{
		PaginationRequest: &bo.PaginationRequest{
			Page:  1,
			Limit: 100,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for {
		userListReply, err := u.userRepo.List(ctx, pageReq)
		if err != nil {
			u.helper.Warnw("msg", "list user fail", "err", err)
			continue
		}
		if len(userListReply.Items) == 0 {
			break
		}
		if err := u.cacheRepo.CacheUsers(ctx, userListReply.Items...); err != nil {
			u.helper.Warnw("msg", "cache user fail", "err", err)
			break
		}
		if userListReply.Total < pageReq.Limit {
			break
		}
		pageReq.Page++
	}

}

func (u *userJob) ID() cron.EntryID {
	return u.id
}

func (u *userJob) Index() string {
	return u.index
}

func (u *userJob) Spec() cron_server.CronSpec {
	return u.spec
}

func (u *userJob) WithID(id cron.EntryID) cron_server.CronJob {
	u.id = id
	return u
}

// IsImmediate implements server.CronJob.
func (u *userJob) IsImmediate() bool {
	return false
}
