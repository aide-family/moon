package vo

type Action int32

/**

// 系统日志操作类型
enum SysLogAction {
  // 未知操作类型
  SysLogActionUnknown = 0;
  // 创建
  SysLogActionCreate = 1;
  // 更新
  SysLogActionUpdate = 2;
  // 删除
  SysLogActionDelete = 3;
  // 查询
  SysLogActionQuery = 4;
  // 导入
  SysLogActionImport = 5;
  // 导出
  SysLogActionExport = 6;
}
*/

const (
	// ActionUnknown 未知操作类型
	ActionUnknown Action = iota
	// ActionCreate 创建
	ActionCreate
	// ActionUpdate 更新
	ActionUpdate
	// ActionDelete  删除
	ActionDelete
	// ActionQuery 查询
	ActionQuery
	// ActionImport 导入
	ActionImport
	// ActionExport 导出
	ActionExport
)

// String stringer
func (a Action) String() string {
	switch a {
	case ActionCreate:
		return "创建"
	case ActionUpdate:
		return "更新"
	case ActionDelete:
		return "删除"
	case ActionQuery:
		return "查询"
	case ActionImport:
		return "导入"
	case ActionExport:
		return "导出"
	case ActionUnknown:
		fallthrough
	default:
		return "未知"
	}
}

// Value int32
func (a Action) Value() int32 {
	return int32(a)
}
