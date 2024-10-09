package vobj

// ModelType 来源类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=ModelType -linecomment
type ModelType int

const (
	ModelCodeUnknown  ModelType = iota // 未知
	ModelCode                          // 普通模型
	BizModelCode                       // 业务模型
	AlarmModelBizCode                  // 告警模型
)

const (
	ModelPath         = "./pkg/palace/model/query"
	BizModelPath      = "./pkg/palace/model/bizmodel/bizquery"
	AlarmModelBizPath = "./pkg/palace/model/alarmmodel/alarmquery"
)

// GetModelPath  获取模型路径
func GetModelPath(typeCode ModelType) string {
	switch typeCode {
	case ModelCode:
		return ModelPath
	case BizModelCode:
		return BizModelPath
	case AlarmModelBizCode:
		return AlarmModelBizPath
	default:
		return ModelPath
	}
}
