package impl

import (
	memberv1 "github.com/aide-family/goddess/domain/member/v1"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

func NewMemberRepository(c *conf.Bootstrap, d *data.Data) (repository.Member, error) {
	repoConfig := c.GetMemberDomain()
	if repoConfig == nil {
		return nil, merr.ErrorInternalServer("memberDomain is required")
	}
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := memberv1.GetMemberV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("member repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig)
		if err != nil {
			return nil, err
		}
		d.AppendClose("memberRepo", close)
		return &memberRepository{MemberServer: repoImpl}, nil
	}
}

type memberRepository struct {
	goddessv1.MemberServer
}
