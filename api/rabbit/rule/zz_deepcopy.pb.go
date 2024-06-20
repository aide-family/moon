package rule

import (
	"github.com/aide-family/moon/pkg/rabbit"
)

func (x *FilterRule) DeepCopyRule() rabbit.Rule {
	return &FilterRule{
		FilterType: x.FilterType,
		MatchType:  x.MatchType,
		MatchLabel: x.MatchLabel,
		Extra:      x.Extra,
	}
}

func (x *SendRule) DeepCopyRule() rabbit.Rule {
	return &SendRule{
		Config: x.Config,
	}
}

func (x *TemplateRule) DeepCopyRule() rabbit.Rule {
	return &TemplateRule{
		Template: x.Template,
	}
}

func (x *AggregationRule) DeepCopyRule() rabbit.Rule {
	return &AggregationRule{
		Count:    x.Count,
		Interval: x.Interval,
		GroupBy:  x.GroupBy,
	}
}
