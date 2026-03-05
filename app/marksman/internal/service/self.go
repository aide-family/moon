package service

import (
	"github.com/aide-family/marksman/internal/biz"
)

func NewSelfService(selfBiz *biz.Self) *SelfService {
	return &SelfService{
		Self: selfBiz,
	}
}

type SelfService struct {
	*biz.Self
}
