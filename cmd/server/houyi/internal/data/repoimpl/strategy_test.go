package repoimpl

import (
	"testing"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/vobj"
)

var eqValues = []*datasource.Value{
	{
		Timestamp: 1680000000000,
		Value:     2,
	},
	{
		Timestamp: 1680000000000,
		Value:     1,
	},
	{
		Timestamp: 1680000000000,
		Value:     1,
	},
	{
		Timestamp: 1680000000000,
		Value:     1,
	},
	{
		Timestamp: 1680000000000,
		Value:     1,
	},
	{
		Timestamp: 1680000000000,
		Value:     1,
	},
	{
		Timestamp: 1680000000000,
		Value:     1,
	},
	{
		Timestamp: 1680000000000,
		Value:     1,
	},
	{
		Timestamp: 1680000000000,
		Value:     1,
	},
	{
		Timestamp: 1680000000000,
		Value:     1,
	},
	{
		Timestamp: 1680000000000,
		Value:     1,
	},
}

var valuesLen = uint32(len(eqValues))

func Test_isCompletelyMeet(t *testing.T) {
	type args struct {
		pointValues []*datasource.Value
		strategy    *bo.StrategyMetric
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "最新数据连续满足等于1成功",
			args: args{
				pointValues: eqValues,
				strategy: &bo.StrategyMetric{
					Condition:   vobj.ConditionEQ,
					Count:       valuesLen - 1,
					Threshold:   1,
					SustainType: vobj.SustainFor,
				},
			},
			want: true,
		},
		{
			name: "最新数据连续满足等于1成功",
			args: args{
				pointValues: eqValues,
				strategy: &bo.StrategyMetric{
					Condition:   vobj.ConditionEQ,
					Count:       valuesLen - 2,
					Threshold:   1,
					SustainType: vobj.SustainFor,
				},
			},
			want: true,
		},
		{
			name: "全部数据连续等于1失败",
			args: args{
				pointValues: eqValues,
				strategy: &bo.StrategyMetric{
					Condition:   vobj.ConditionEQ,
					Count:       valuesLen,
					Threshold:   1,
					SustainType: vobj.SustainFor,
				},
			},
			want: false,
		},
		{
			name: "最多n-1次满足条件",
			args: args{
				pointValues: eqValues,
				strategy: &bo.StrategyMetric{
					Condition:   vobj.ConditionEQ,
					Count:       valuesLen - 1,
					Threshold:   1,
					SustainType: vobj.SustainMax,
				},
			},
			want: true,
		},
		{
			name: "最多n-2次满足条件",
			args: args{
				pointValues: eqValues,
				strategy: &bo.StrategyMetric{
					Condition:   vobj.ConditionEQ,
					Count:       valuesLen - 2,
					Threshold:   1,
					SustainType: vobj.SustainMax,
				},
			},
			want: false,
		},
		{
			name: "最多n次满足条件",
			args: args{
				pointValues: eqValues,
				strategy: &bo.StrategyMetric{
					Condition:   vobj.ConditionEQ,
					Count:       valuesLen,
					Threshold:   1,
					SustainType: vobj.SustainMax,
				},
			},
			want: true,
		},
		{
			name: "最少n次满足条件",
			args: args{
				pointValues: eqValues,
				strategy: &bo.StrategyMetric{
					Condition:   vobj.ConditionEQ,
					Count:       valuesLen,
					Threshold:   1,
					SustainType: vobj.SustainMin,
				},
			},
			want: false,
		},
		{
			name: "最少n-1次满足条件",
			args: args{
				pointValues: eqValues,
				strategy: &bo.StrategyMetric{
					Condition:   vobj.ConditionEQ,
					Count:       valuesLen - 1,
					Threshold:   1,
					SustainType: vobj.SustainMin,
				},
			},
			want: true,
		},
		{
			name: "最少n-2次满足条件",
			args: args{
				pointValues: eqValues,
				strategy: &bo.StrategyMetric{
					Condition:   vobj.ConditionEQ,
					Count:       valuesLen - 2,
					Threshold:   1,
					SustainType: vobj.SustainMin,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCompletelyMeet(tt.args.pointValues, tt.args.strategy); got != tt.want {
				t.Errorf("isCompletelyMeet() = %v, want %v", got, tt.want)
			}
		})
	}
}
