package service

import (
	"github.com/aide-family/rabbit/internal/biz"
)

func NewSelfService(selfBiz *biz.Self) (*SelfService, error) {
	return &SelfService{Self: selfBiz}, nil
}

type SelfService struct {
	*biz.Self
}
