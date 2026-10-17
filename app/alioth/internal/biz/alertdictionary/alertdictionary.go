package alertdictionary

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
	dictionaries repository.AlertDictionaryRepo
	access       access.Checker
}

func NewUsecase(dictionaries repository.AlertDictionaryRepo, checker access.Checker) *Usecase {
	return &Usecase{dictionaries: dictionaries, access: checker}
}

func (uc *Usecase) CreateAlertDictionary(ctx context.Context, in bo.AlertDictionary) (*bo.AlertDictionary, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.dictionaries.CreateAlertDictionary(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateAlertDictionary(ctx context.Context, in bo.AlertDictionary) (*bo.AlertDictionary, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.dictionaries.UpdateAlertDictionary(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateAlertDictionaryStatus(ctx context.Context, uid, status string) (*bo.AlertDictionary, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.dictionaries.UpdateAlertDictionaryStatus(ctx, ns, uid, status)
	return item, uc.access.Wrap(err)
}

// DeleteAlertDictionary 拒绝删除仍挂在规则上的字典项，避免条件指向一条已经不存在的级别。
func (uc *Usecase) DeleteAlertDictionary(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.dictionaries.DeleteAlertDictionary(ctx, ns, uid))
}

func (uc *Usecase) ListAlertDictionaries(ctx context.Context, pageNo, pageSize int32, kind, name string) ([]bo.AlertDictionary, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.dictionaries.ListAlertDictionaries(ctx, ns, p, s, kind, strings.TrimSpace(name))
	return items, total, p, s, uc.access.Wrap(err)
}
