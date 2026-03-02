package biz

import (
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewNamespace(
	namespaceRepo repository.Namespace,
) *Namespace {
	return &Namespace{
		Namespace: namespaceRepo,
	}
}

type Namespace struct {
	repository.Namespace
}
