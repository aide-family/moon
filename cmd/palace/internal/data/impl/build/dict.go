package build

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/team"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func ToDict(ctx context.Context, dict do.TeamDict) *team.Dict {
	if validate.IsNil(dict) {
		return nil
	}
	if dict, ok := dict.(*team.Dict); ok {
		dict.WithContext(ctx)
		return dict
	}
	return &team.Dict{
		TeamModel: ToTeamModel(ctx, dict),
		Key:       dict.GetKey(),
		Value:     dict.GetValue(),
		Lang:      dict.GetLang(),
		Color:     dict.GetColor(),
		DictType:  dict.GetType(),
		Status:    dict.GetStatus(),
	}
}

func ToDicts(ctx context.Context, dicts []do.TeamDict) []*team.Dict {
	return slices.Map(dicts, func(v do.TeamDict) *team.Dict {
		return ToDict(ctx, v)
	})
}
