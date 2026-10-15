package data

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/permission"
	"github.com/aide-family/Alioth/app/alioth/internal/perm"
)

func seed(ctx context.Context, client *ent.Client) error {
	for _, item := range perm.Catalog() {
		exists, err := client.Permission.Query().Where(permission.NameEQ(item.Name)).Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err := client.Permission.Create().
			SetUID(newID()).
			SetName(item.Name).
			SetResource(item.Resource).
			SetAction(item.Action).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}
