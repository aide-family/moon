package bo

import (
	"fmt"
	"time"
)

// 一个「规则 × 数据源」组合在本轮领取里的状态。
const (
	// WorkerTaskRunning 表示这个组合正在被评测。
	WorkerTaskRunning = "RUNNING"
	// WorkerTaskNew 表示本轮新加入、上一批没有这个组合。
	WorkerTaskNew = "NEW"
	// WorkerTaskPending 表示本轮已领取、还在排队等待评测。
	WorkerTaskPending = "PENDING"
	// WorkerTaskDone 表示本轮已经回写完成。
	WorkerTaskDone = "DONE"
	// WorkerTaskLeft 表示上一批领取过、本批已不再包含它。
	WorkerTaskLeft = "LEFT"
)

var workerTaskStates = map[string]struct{}{
	WorkerTaskRunning: {},
	WorkerTaskNew:     {},
	WorkerTaskPending: {},
	WorkerTaskDone:    {},
	WorkerTaskLeft:    {},
}

// ValidWorkerTaskState 判断状态是否是后台工作进程支持的取值。
func ValidWorkerTaskState(state string) bool {
	_, ok := workerTaskStates[state]
	return ok
}

// WorkerTask is one rule × datasource combination claimed by a worker round.
type WorkerTask struct {
	State          string    `json:"state"`
	RuleUID        string    `json:"ruleUid"`
	RuleName       string    `json:"ruleName"`
	DatasourceUID  string    `json:"datasourceUid"`
	DatasourceName string    `json:"datasourceName"`
	EvalSlot       time.Time `json:"evalSlot"`
}

// IdempotencyKey 返回玉衡回写去重用的批次键：规则、数据源与批次三者一起唯一。
func (task WorkerTask) IdempotencyKey() string {
	return fmt.Sprintf("%s/%s@%s", task.RuleUID, task.DatasourceUID, task.EvalSlot.UTC().Format(time.RFC3339))
}

// WorkerInstance describes one live or recently-seen background worker process.
type WorkerInstance struct {
	Service       string
	InstanceUID   string
	Hostname      string
	Busy          bool
	TaskUID       string
	TaskName      string
	TaskStartedAt *time.Time
	LastSeenAt    time.Time
	Online        bool
	// RoundStartedAt 是当前这批组合的领取时间，没有领取批次时为空。
	RoundStartedAt *time.Time
	// NextPollAt 是本实例自己上报的下一次轮询时间，没有上报时为空。
	NextPollAt *time.Time
	// PollIntervalSeconds 是本实例的轮询周期秒数。
	PollIntervalSeconds int32
	// Tasks 是本批领取的「规则 × 数据源」组合快照。
	Tasks []WorkerTask
}

// WorkerStatus is the heartbeat payload accepted from an internal worker.
type WorkerStatus struct {
	Service       string
	InstanceUID   string
	Hostname      string
	Busy          bool
	TaskUID       string
	TaskName      string
	TaskStartedAt *time.Time
	// RoundStartedAt 是当前这批组合的领取时间。
	RoundStartedAt *time.Time
	// NextPollAt 是本实例下一次轮询时间。
	NextPollAt *time.Time
	// PollIntervalSeconds 是本实例的轮询周期秒数。
	PollIntervalSeconds int32
	// Tasks 是本批领取的「规则 × 数据源」组合快照。
	Tasks []WorkerTask
}
