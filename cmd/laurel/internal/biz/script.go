package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
)

func NewScript(scriptRepo repository.Script) *Script {
	return &Script{scriptRepo: scriptRepo}
}

type Script struct {
	scriptRepo repository.Script
}

func (s *Script) GetScripts(ctx context.Context) ([]*bo.TaskScript, error) {
	return s.scriptRepo.GetScripts(ctx)
}
