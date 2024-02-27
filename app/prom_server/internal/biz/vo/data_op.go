package vo

type Op int32

const (
	// OpCreate 创建
	OpCreate Op = iota + 1
	// OpUpdate 更新
	OpUpdate
	// OpDelete 删除
	OpDelete
	// OpRead 读取
	OpRead
	// OpAll 所有
	OpAll
)

// String 转换为字符串
func (o Op) String() string {
	switch o {
	case OpCreate:
		return "创建"
	case OpUpdate:
		return "更新"
	case OpDelete:
		return "删除"
	case OpRead:
		return "读取"
	case OpAll:
		return "所有"
	default:
		return "未知"
	}
}

// Value 转换为int32
func (o Op) Value() int32 {
	return int32(o)
}
