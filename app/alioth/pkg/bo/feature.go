package bo

import "time"

// Feature 是菜单模块化的功能。功能由代码维护，租户通过套餐开通后才能在角色上授权。
type Feature struct {
	Code        string
	Name        string
	Description string
	Permissions []string
}

// FeaturePackage 是平台级的功能套餐。Features 是套餐携带的功能编码。
type FeaturePackage struct {
	UID       string
	Name      string
	Remark    string
	IsBase    bool
	Status    string
	Features  []string
	CreatedAt time.Time
}

// NamespaceFeature 是租户的功能开通情况。
// PackageFeatures 是套餐自带的功能；Features 是叠加租户差异后的有效功能。
type NamespaceFeature struct {
	NamespaceUID    string
	PackageUID      string
	PackageName     string
	PackageFeatures []string
	Features        []string
}
