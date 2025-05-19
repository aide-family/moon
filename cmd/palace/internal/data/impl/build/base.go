package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToBaseModel(ctx context.Context, model do.Base) do.BaseModel {
	m := do.BaseModel{}
	if validate.IsNotNil(model) {
		m = do.BaseModel{
			ID:        model.GetID(),
			CreatedAt: model.GetCreatedAt(),
			UpdatedAt: model.GetUpdatedAt(),
			DeletedAt: model.GetDeletedAt(),
		}
	}
	m.WithContext(ctx)
	return m
}

func ToCreatorModel(ctx context.Context, model do.Creator) do.CreatorModel {
	item := do.CreatorModel{}
	if validate.IsNotNil(model) {
		item = do.CreatorModel{
			BaseModel: ToBaseModel(ctx, model),
			CreatorID: model.GetCreatorID(),
		}
	}
	item.WithContext(ctx)
	return item
}

func ToTeamModel(ctx context.Context, model do.TeamBase) do.TeamModel {
	item := do.TeamModel{}
	if validate.IsNil(model) {
		item = do.TeamModel{
			CreatorModel: ToCreatorModel(ctx, model),
			TeamID:       model.GetTeamID(),
		}
	}
	item.WithContext(ctx)
	return item
}
