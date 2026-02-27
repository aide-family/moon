package impl

import (
	magicboxapiv1 "github.com/aide-family/magicbox/api/v1"
	memberv1 "github.com/aide-family/magicbox/domain/member/v1"
	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
)

func NewMemberRepository(c *conf.Bootstrap, d *data.Data) (repository.Member, error) {
	repoConfig := c.GetMemberConfig()
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
	magicboxapiv1.MemberServer
}
