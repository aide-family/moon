package impl

import (
	"context"
	"slices"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/data/impl/do"
	"github.com/aide-family/marksman/internal/data/impl/query"
)

func NewUserAlertPageRepository(d *data.Data) (repository.UserAlertPage, error) {
	query.SetDefault(d.DB())
	return &userAlertPageRepository{}, nil
}

type userAlertPageRepository struct{}

func (r *userAlertPageRepository) GetUserAlertPageUIDs(ctx context.Context, userUID snowflake.ID) ([]snowflake.ID, error) {
	namespaceUID := contextx.GetNamespace(ctx)
	u := query.UserAlertPage
	list, err := u.WithContext(ctx).
		Where(u.NamespaceUID.Eq(namespaceUID.Int64()), u.UserUID.Eq(userUID.Int64())).
		Order(u.SortOrder.Desc()).
		Find()
	if err != nil {
		return nil, err
	}
	out := make([]snowflake.ID, 0, len(list))
	for _, row := range list {
		out = append(out, row.AlertPageUID)
	}
	return out, nil
}

func (r *userAlertPageRepository) SaveUserAlertPages(ctx context.Context, userUID snowflake.ID, alertPageUIDs []snowflake.ID) error {
	namespaceUID := contextx.GetNamespace(ctx)
	uidsReversed := slices.Clone(alertPageUIDs)
	slices.Reverse(uidsReversed)
	rows := make([]*do.UserAlertPage, 0, len(uidsReversed))
	for i, uid := range uidsReversed {
		rows = append(rows, &do.UserAlertPage{
			NamespaceUID: namespaceUID,
			UserUID:      userUID,
			AlertPageUID: uid,
			SortOrder:    int32(i),
		})
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		u := tx.UserAlertPage
		_, err := u.WithContext(ctx).
			Where(u.NamespaceUID.Eq(namespaceUID.Int64()), u.UserUID.Eq(userUID.Int64())).
			Delete()
		if err != nil {
			return err
		}
		if len(alertPageUIDs) == 0 {
			return nil
		}

		if err := u.WithContext(ctx).CreateInBatches(rows, len(rows)); err != nil {
			return merr.ErrorInternalServer("save user alert pages failed").WithCause(err)
		}
		return nil
	})
}
