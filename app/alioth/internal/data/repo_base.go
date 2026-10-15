package data

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

func (r *repoBase) fail(err error) error {
	if err == nil {
		return nil
	}
	if ent.IsNotFound(err) {
		return merr.NotFound("not found")
	}
	if ent.IsConstraintError(err) {
		return merr.Conflict("name already exists")
	}
	r.log.Errorf("database: %v", err)
	return merr.Server("request failed")
}

func (r *repoBase) GetNamespace(ctx context.Context, uid string) (*bo.Namespace, error) {
	row, err := r.client.Namespace.Query().Where(namespace.UIDEQ(uid), namespace.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNamespace(row)
	return &item, nil
}
