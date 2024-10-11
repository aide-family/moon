package vobj

const (
	ModelCodeUnknown  = "unknown" // 未知
	MainCode          = "main"    // 主库
	BizModelCode      = "biz"     // 业务库
	AlarmModelBizCode = "alarm"   // 告警库
)

const (
	MainPath          = "./pkg/palace/model/query"
	BizModelPath      = "./pkg/palace/model/bizmodel/bizquery"
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
