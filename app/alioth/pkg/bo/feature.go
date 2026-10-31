package bo

import "time"

// 套餐定价支持的币种。价格按对应币种的最小单位（分）存储，展示时配符号。
const (
	CurrencyCNY = "CNY"
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
	CurrencyJPY = "JPY"
)

// Currencies 按固定顺序返回支持的币种编码。
func Currencies() []string {
	return []string{CurrencyCNY, CurrencyUSD, CurrencyEUR, CurrencyJPY}
}

// IsCurrency 判断币种是否在支持目录内。
func IsCurrency(code string) bool {
	for _, item := range Currencies() {
		if item == code {
			return true
		}
	}
	return false
}

// Feature 是菜单模块化的功能。功能由代码维护，租户通过套餐开通后才能在角色上授权。
// Module 是能力所属的模块编码，对模块本身等于 Code；能力化后 /v1/system/features
// 返回能力，Module 用来把能力归到左侧菜单分组。
type Feature struct {
	Code        string
	Module      string
	Name        string
	Description string
	Permissions []string
}

// FeaturePackage 是平台级的功能套餐。Features 是套餐携带的能力编码。
// DurationDays 是租户使用该套餐时的默认授权时长，0 表示永久；
// Visibility 决定套餐出现在哪里（PUBLIC 可被用户创建租户时选择，BETA 只由平台分配）；
// MaxMembers 是租户可容纳的成员数上限，0 表示不限制；PriceCents 是价格，单位为人民币分。
type FeaturePackage struct {
	UID          string
	Name         string
	Remark       string
	IsBase       bool
	Status       string
	Visibility   string
	Features     []string
	DurationDays int32
	MaxMembers   int32
	PriceCents   int64
	Currency     string
	CreatedAt    time.Time
}

// NamespaceFeature 是租户的功能开通情况。
// PackageFeatures 是套餐自带的能力；ExtraFeatures 是租户额外增加的能力；
// Features 是两者并集（已配置状态，不因停用或到期清空）。
// ExpiresAt 是套餐到期时间，零值表示永久；Expired 表示当前是否已过期。
// DurationDays 是租户当前套餐的默认授权时长，0 表示永久。
type NamespaceFeature struct {
	NamespaceUID    string
	PackageUID      string
	PackageName     string
	PackageFeatures []string
	ExtraFeatures   []string
	Features        []string
	ExpiresAt       time.Time
	Expired         bool
	DurationDays    int32
}
