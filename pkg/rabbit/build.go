package rabbit

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/runtime"
	"github.com/aide-family/moon/pkg/runtime/schema"
	"github.com/aide-family/moon/pkg/util/types"
)

func NewFilterRuleBuilder(r *api.FilterRule) Rule {
	return &FilterRuleBuilder{
		FilterRule: r,
	}
}

func NewSendRuleBuilder(r *api.SendRule) Rule {
	return &SendRuleBuilder{
		SendRule: r,
	}
}

func NewTemplateRuleBuilder(r *api.TemplateRule) Rule {
	return &TemplateRuleBuilder{
		TemplateRule: r,
	}
}

func NewAggregationRuleBuilder(r *api.AggregationRule) Rule {
	return &AggregationRuleBuilder{
		AggregationRule: r,
	}
}

type (
	TypeMetaBuilder struct {
		*api.TypeMeta
	}

	FilterRuleBuilder struct {
		*api.FilterRule
	}

	SendRuleBuilder struct {
		*api.SendRule
	}

	TemplateRuleBuilder struct {
		*api.TemplateRule
	}

	AggregationRuleBuilder struct {
		*api.AggregationRule
	}

	MessageFilterRule api.MessageFilterRule

	MessageAggregationRule api.MessageAggregationRule

	MessageTemplateRule api.MessageTemplateRule

	MessageSendRule api.MessageSendRule

	RuleGroup api.RuleGroup
)

func (r *RuleGroup) GetObjectKind() schema.ObjectKind {
	return r
}

func (r *RuleGroup) DeepCopyObject() runtime.Object {
	if types.IsNil(r) {
		return nil
	}
	return &(*r)
}

func (r *MessageSendRule) GetObjectKind() schema.ObjectKind {
	return r
}

func (r *MessageSendRule) DeepCopyObject() runtime.Object {
	if types.IsNil(r) {
		return nil
	}
	return &(*r)
}

func (r *MessageTemplateRule) GetObjectKind() schema.ObjectKind {
	return r
}

func (r *MessageTemplateRule) DeepCopyObject() runtime.Object {
	if types.IsNil(r) {
		return nil
	}
	return &(*r)
}

func (r *MessageAggregationRule) GetObjectKind() schema.ObjectKind {
	return r
}

func (r *MessageAggregationRule) DeepCopyObject() runtime.Object {
	if types.IsNil(r) {
		return nil
	}
	return &(*r)
}

func (r *MessageFilterRule) GetObjectKind() schema.ObjectKind {
	return r
}

func (r *MessageFilterRule) DeepCopyObject() runtime.Object {
	if types.IsNil(r) {
		return nil
	}
	return &(*r)
}

func (r *AggregationRuleBuilder) DeepCopyRule() Rule {
	if types.IsNil(r) || types.IsNil(r.AggregationRule) {
		return nil
	}
	return &(*r)
}

func (r *TemplateRuleBuilder) DeepCopyRule() Rule {
	if types.IsNil(r) || types.IsNil(r.TemplateRule) {
		return nil
	}
	return &(*r)
}

func (r *SendRuleBuilder) DeepCopyRule() Rule {
	if types.IsNil(r) || types.IsNil(r.SendRule) {
		return nil
	}
	return &(*r)
}

func (r *FilterRuleBuilder) DeepCopyRule() Rule {
	if types.IsNil(r) || types.IsNil(r.FilterRule) {
		return nil
	}
	return &(*r)
}
