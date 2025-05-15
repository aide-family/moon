package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToBaseModel(ctx context.Context, model do.Base) do.BaseModel {
	if validate.IsNil(model) {
		m := do.BaseModel{}
		m.WithContext(ctx)
		return m
	}
	m := do.BaseModel{
		ID:        model.GetID(),
		CreatedAt: model.GetCreatedAt(),
		UpdatedAt: model.GetUpdatedAt(),
		DeletedAt: model.GetDeletedAt(),
	}
	m.WithContext(ctx)
	return m
}

func ToCreatorModel(ctx context.Context, model do.Creator) do.CreatorModel {
	if validate.IsNil(model) {
		return do.CreatorModel{}
	}
	return do.CreatorModel{
		BaseModel: ToBaseModel(ctx, model),
		CreatorID: model.GetCreatorID(),
	}
}

func ToTeamModel(ctx context.Context, model do.TeamBase) do.TeamModel {
	if validate.IsNil(model) {
		return do.TeamModel{}
	}
	return do.TeamModel{
		CreatorModel: ToCreatorModel(ctx, model),
		TeamID:       model.GetTeamID(),
	}
}
