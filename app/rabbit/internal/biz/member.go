package biz

import (
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewMember(
	memberRepo repository.Member,
	helper *klog.Helper,
) *Member {
	return &Member{
		Member: memberRepo,
		helper: klog.NewHelper(klog.With(helper.Logger(), "biz", "member")),
	}
}

type Member struct {
	helper *klog.Helper
	repository.Member
}
