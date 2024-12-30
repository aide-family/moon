package helper

import (
	"context"
	"testing"

	"github.com/aide-family/moon/api"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
)

/**
ollama:
  type: 'openai'
  model: 'gpt-4o-mini'
  url: 'https://free.v36.cm/v1/chat/completions'
  auth: 'sk-lSlotX3nG97FMfX5346e1fC139C6486aBdFc94B3Be129e9d'
  contextSize: 10
*/

func TestOllama_responseFromOllama(t *testing.T) {
	opts := []OllamaOption{
		WithOllamaModel("gpt-4o-mini"),
		WithOllamaAuth("sk-lSlotX3nG97FMfX5346e1fC139C6486aBdFc94B3Be129e9d"),
		WithOllamaType("openai"),
	}
	ollamaClient := NewOllama("https://free.v36.cm/v1/chat/completions", opts...)

	messages := []Message{
		{
			Content: "你好",
			Role:    "user",
		},
	}
	fromOllama, err := ollamaClient.responseFromOllama(context.Background(), messages)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(fromOllama)
}

func TestOllama_GetAnnotation(t *testing.T) {
	opts := []OllamaOption{
		WithOllamaModel("gpt-4o-mini"),
		WithOllamaAuth("sk-lSlotX3nG97FMfX5346e1fC139C6486aBdFc94B3Be129e9d"),
		WithOllamaType("openai"),
	}
	ollamaClient := NewOllama("https://free.v36.cm/v1/chat/completions", opts...)

	strategyItem := &strategyapi.CreateStrategyRequest{
		GroupId:       1,
		TemplateId:    0,
		Remark:        "",
		Status:        api.Status_StatusEnable,
		DatasourceIds: []uint32{1, 2, 3},
		SourceType:    api.TemplateSourceType(api.DatasourceType_DatasourceTypeMetric),
		Name:          "test",
		StrategyType:  0,
		Labels:        map[string]string{"test": "test"},
		Annotations:   nil,
		Expr:          "load1",
		CategoriesIds: nil,
		AlarmGroupIds: nil,
		StrategyMetricLevels: []*strategyapi.CreateStrategyMetricLevelRequest{{
			Duration:      10,
			Count:         1,
			SustainType:   api.SustainType_SustainTypeFor,
			LevelId:       1,
			Threshold:     10,
			Condition:     api.Condition_ConditionGTE,
			AlarmPageIds:  []uint32{1, 2, 3},
			AlarmGroupIds: []uint32{1, 2, 3},
			LabelNotices:  []*strategyapi.CreateStrategyLabelNoticeRequest{},
		}},
		StrategyEventLevels:  nil,
		StrategyDomainLevels: nil,
		StrategyPortLevels:   nil,
		StrategyHTTPLevels:   nil,
	}

	annotation, err := ollamaClient.GetAnnotation(context.Background(), strategyItem)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(annotation)
}
