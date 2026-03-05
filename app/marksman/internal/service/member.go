package service

import (
	"github.com/aide-family/marksman/internal/biz"
)

func NewMemberService(memberBiz *biz.Member) *MemberService {
	return &MemberService{
		Member: memberBiz,
	}
}

type MemberService struct {
	*biz.Member
}
