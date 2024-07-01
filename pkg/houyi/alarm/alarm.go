package alarm

import (
	"sort"
	"strings"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"

	"golang.org/x/exp/maps"
)

var _ watch.Indexer = (*EventMsg)(nil)

func NewWatchAlarm(opts ...watch.WatcherOption) *WatchAlarm {
	return &WatchAlarm{
		Watcher: watch.NewWatcher(opts...),
	}
}

type (
	WatchAlarm struct {
		*watch.Watcher
	}

	// EventMsg 满足策略条件的事件数据
	EventMsg struct {
		// 告警状态
		Status vobj.AlertStatus `json:"status"`
		// 标签
		Labels vobj.Labels `json:"labels"`
		// 注解
		Annotations vobj.Annotations `json:"annotations"`
		// 开始时间
		StartsAt *types.Time `json:"startsAt"`
		// 结束时间, 如果为空, 则表示未结束
		EndsAt *types.Time `json:"endsAt"`
		// 告警生成链接
		GeneratorURL string `json:"generatorURL"`
		// 指纹
		Fingerprint string `json:"fingerprint"`
	}
)

func (e *EventMsg) Index() string {
	str := strings.Builder{}
	str.WriteString("{")
	keys := maps.Keys(e.Labels)
	// 排序
	sort.Strings(keys)
	for _, key := range keys {
		k := key
		v := e.Labels[key]
		str.WriteString(`"` + k + `"`)
		str.WriteString(":")
		str.WriteString(`"` + v + `"`)
		str.WriteString(",")
	}
	return strings.TrimRight(str.String(), ",") + "}"
}
