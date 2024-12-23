package config

const (
	// MainCode 主库
	MainCode = "main" // 主库
	// BizModelCode 业务库
	BizModelCode = "biz" // 业务库
	// AlarmModelBizCode 告警库
	AlarmModelBizCode = "alarm" // 告警库
)

const (
	// MainPath 主库路径
	MainPath = "./pkg/palace/model/query"
	// BizModelPath 业务库路径
	BizModelPath = "./pkg/palace/model/bizmodel/bizquery"
	// AlarmModelBizPath 告警库路径
	AlarmModelBizPath = "./pkg/palace/model/alarmmodel/alarmquery"
)

// GetModelPath  获取模型路径
func GetModelPath(typeCode string) string {
	switch typeCode {
	case MainCode:
		return MainPath
	case BizModelCode:
		return BizModelPath
	case AlarmModelBizCode:
		return AlarmModelBizPath
	default:
		return MainPath
	}
}
