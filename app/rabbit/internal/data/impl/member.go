package impl

import (
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"

	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
)

func NewMemberRepository(c *conf.Bootstrap, d *data.Data) (repository.Member, error) {
	repoImpl, close, err := newGoddessMember(c.GetGoddessDomain())
	if err != nil {
		return nil, err
	}
	d.AppendClose("memberRepo", close)
	return &memberRepository{MemberServer: repoImpl}, nil
}

type memberRepository struct {
	goddessv1.MemberServer
}
