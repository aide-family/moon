package biz

import (
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewMember(memberRepo repository.Member) *Member {
	return &Member{
		Member: memberRepo,
	}
}

type Member struct {
	repository.Member
}
