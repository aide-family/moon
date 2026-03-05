package biz

import (
	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewSelf(selfRepo repository.Self) *Self {
	return &Self{
		Self: selfRepo,
	}
}

type Self struct {
	repository.Self
}
