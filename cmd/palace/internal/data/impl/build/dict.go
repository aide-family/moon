package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToDict(ctx context.Context, dict do.TeamDict) *team.Dict {
	if validate.IsNil(dict) {
		return nil
	}
	if dict, ok := dict.(*team.Dict); ok {
		dict.WithContext(ctx)
		return dict
	}
	dictDo := &team.Dict{
		TeamModel: ToTeamModel(ctx, dict),
		Key:       dict.GetKey(),
		Value:     dict.GetValue(),
		Lang:      dict.GetLang(),
		Color:     dict.GetColor(),
		DictType:  dict.GetType(),
		Status:    dict.GetStatus(),
	}
	dictDo.WithContext(ctx)
	return dictDo
}

func ToDicts(ctx context.Context, dicts []do.TeamDict) []*team.Dict {
	return slices.MapFilter(dicts, func(v do.TeamDict) (*team.Dict, bool) {
		if validate.IsNil(v) {
			return nil, false
		}
		return ToDict(ctx, v), true
	})
}
