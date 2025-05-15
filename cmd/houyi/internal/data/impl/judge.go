package impl

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/houyi/internal/conf"
	"github.com/moon-monitor/moon/cmd/houyi/internal/data"
	"github.com/moon-monitor/moon/cmd/houyi/internal/data/impl/judge"
	"github.com/moon-monitor/moon/cmd/houyi/internal/data/impl/judge/condition"
	"github.com/moon-monitor/moon/pkg/api/common"
	"github.com/moon-monitor/moon/pkg/util/cnst"
	"github.com/moon-monitor/moon/pkg/util/hash"
	"github.com/moon-monitor/moon/pkg/util/kv"
	"github.com/moon-monitor/moon/pkg/util/pointer"
	"github.com/moon-monitor/moon/pkg/util/template"
	"github.com/moon-monitor/moon/pkg/util/timex"
)

func NewJudgeRepo(bc *conf.Bootstrap, data *data.Data, logger log.Logger) repository.Judge {
	return &judgeImpl{
		Data:             data,
		evaluateInterval: bc.GetEvaluate().GetInterval().AsDuration() * 2,
		helper:           log.NewHelper(log.With(logger, "module", "data.repo.judge")),
	}
}

type judgeImpl struct {
	*data.Data
	evaluateInterval time.Duration
	helper           *log.Helper
}

func (j *judgeImpl) Metric(_ context.Context, data *bo.MetricJudgeRequest) ([]bo.Alert, error) {
	rule := data.Strategy
	judgeData := data.JudgeData
	conditionType := condition.NewMetricCondition(rule.GetCondition())
	opts := []judge.MetricJudgeOption{
		judge.WithMetricJudgeCondition(conditionType),
		judge.WithMetricJudgeConditionValues(rule.GetValues()),
		judge.WithMetricJudgeConditionCount(rule.GetCount()),
		judge.WithMetricJudgeStep(data.Step),
	}
	judgeInstance := judge.NewMetricJudge(rule.GetSampleMode(), opts...)
	alerts := make([]bo.Alert, 0, len(judgeData))
	for _, datum := range judgeData {
		value, ok := judgeInstance.Judge(datum.GetValues())
		if !ok {
			continue
		}
		alert := j.generateAlert(rule, value, datum.GetLabels())
		alerts = append(alerts, alert)
	}
	return alerts, nil
}

func (j *judgeImpl) generateAlert(rule bo.MetricJudgeRule, value bo.MetricJudgeDataValue, originLabels map[string]string) bo.Alert {
	ext := rule.GetExt()
	ext.Set(cnst.ExtKeyValues, value.GetValue())
	ext.Set(cnst.ExtKeyTimestamp, value.GetTimestamp())

	labels := rule.GetLabels().Copy()
	labelsMap := labels.ToMap()
	for k, v := range labelsMap {
		labelsMap[k] = template.TextFormatterX(v, ext)
	}
	labels = labels.Appends(labelsMap).Appends(originLabels)
	ext.Set(cnst.ExtKeyLabels, labels.ToMap())

	annotations := rule.GetAnnotations().Copy()
	summary := template.TextFormatterX(annotations.GetSummary(), ext)
	description := template.TextFormatterX(annotations.GetDescription(), ext)
	annotations.SetSummary(summary)
	annotations.SetDescription(description)

	stringMap := kv.NewStringMap(originLabels)
	fingerprint := hash.MD5(kv.SortString(stringMap))

	now := timex.Now()
	alert := &do.Alert{
		Status:       common.AlertStatus_pending,
		Labels:       labels,
		Annotations:  annotations,
		StartsAt:     pointer.Of(time.Unix(value.GetTimestamp(), 0)),
		EndsAt:       nil,
		GeneratorURL: "",
		Fingerprint:  fingerprint,
		Value:        value.GetValue(),
		LastUpdated:  now,
		Duration:     j.evaluateInterval,
	}
	return alert
}
