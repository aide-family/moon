package manager

import (
	"context"
	"github.com/aide-family/moon/pkg/rabbit"
)

type SenderManager struct {
}

func (s *SenderManager) Get(context context.Context, name string) (rabbit.Sender, error) {
	// TODO: implement me
	return nil, nil
}
