package service

import (
	"github.com/aide-family/rabbit/internal/biz"
)

func NewMemberService(memberBiz *biz.Member) (*MemberService, error) {
	return &MemberService{Member: memberBiz}, nil
}

type MemberService struct {
	*biz.Member
}
