// Package oauthproviders 列出已实现授权码流程的第三方登录来源。
// 它是来源清单的唯一来源：登录页决定显示哪些按钮、后台按它校验与排序，
// 授权地址的拼接细节仍留在各个实现里。
package oauthproviders

// Provider 是一个第三方登录来源。
type Provider struct {
	// Key 是稳定标识，同时用于配置与接口参数，不随展示文案变化。
	Key string
	// Name 是界面展示名。
	Name string
}

// Catalog 按界面展示顺序列出全部来源。
var Catalog = []Provider{
	{Key: "GITHUB", Name: "GitHub"},
	{Key: "GITEE", Name: "Gitee"},
	{Key: "FEISHU", Name: "飞书"},
	{Key: "DINGTALK", Name: "钉钉"},
}

// Keys 返回全部来源标识，顺序与 Catalog 一致。
func Keys() []string {
	keys := make([]string, 0, len(Catalog))
	for _, item := range Catalog {
		keys = append(keys, item.Key)
	}
	return keys
}

// Lookup 按标识取来源；不存在时第二个返回值为 false。
func Lookup(key string) (Provider, bool) {
	for _, item := range Catalog {
		if item.Key == key {
			return item, true
		}
	}
	return Provider{}, false
}
