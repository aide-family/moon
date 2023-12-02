package valueobj

type Gender int32

const (
	// GenderUnknown 未知性别
	GenderUnknown Gender = iota
	// GenderMale 男性
	GenderMale
	// GenderFemale 女性
	GenderFemale
)

// String 获取性别字符串
func (g Gender) String() string {
	switch g {
	case GenderUnknown:
		return "未知"
	case GenderMale:
		return "男"
	case GenderFemale:
		return "女"
	default:
		return "未知"
	}
}

// Key 获取性别key
func (g Gender) Key() string {
	switch g {
	case GenderUnknown:
		return "unknown"
	case GenderMale:
		return "male"
	case GenderFemale:
		return "female"
	default:
		return "unknown"
	}
}

// Value 获取性别value
func (g Gender) Value() int32 {
	return int32(g)
}

// ToGender 转换性别
func ToGender(gender string) Gender {
	switch gender {
	case "男":
		return GenderMale
	case "女":
		return GenderFemale
	default:
		return GenderUnknown
	}
}
