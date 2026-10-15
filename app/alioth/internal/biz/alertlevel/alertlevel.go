package alertlevel

import (
	"context"
	"strings"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
)

var ProviderSet = wire.NewSet(NewUsecase)

type Usecase struct {
	levels repository.AlertLevelRepo
	access access.Checker
}

func NewUsecase(levels repository.AlertLevelRepo, checker access.Checker) *Usecase {
	return &Usecase{levels: levels, access: checker}
}

func (uc *Usecase) CreateAlertLevel(ctx context.Context, in bo.AlertLevel) (*bo.AlertLevel, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.levels.CreateAlertLevel(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateAlertLevel(ctx context.Context, in bo.AlertLevel) (*bo.AlertLevel, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.levels.UpdateAlertLevel(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateAlertLevelStatus(ctx context.Context, uid, status string) (*bo.AlertLevel, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.levels.UpdateAlertLevelStatus(ctx, ns, uid, status)
	return item, uc.access.Wrap(err)
}

// DeleteAlertLevel 拒绝删除仍挂在规则上的级别，避免条件指向一条已经不存在的级别。
func (uc *Usecase) DeleteAlertLevel(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.levels.DeleteAlertLevel(ctx, ns, uid))
}

func (uc *Usecase) ListAlertLevels(ctx context.Context, pageNo, pageSize int32, kind, name string) ([]bo.AlertLevel, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.levels.ListAlertLevels(ctx, ns, p, s, kind, strings.TrimSpace(name))
	return items, total, p, s, uc.access.Wrap(err)
}
