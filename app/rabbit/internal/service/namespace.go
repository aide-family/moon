package service

import (
	"github.com/aide-family/rabbit/internal/biz"
)

func NewNamespaceService(namespaceBiz *biz.Namespace) (*NamespaceService, error) {
	return &NamespaceService{Namespace: namespaceBiz}, nil
}

type NamespaceService struct {
	*biz.Namespace
}
