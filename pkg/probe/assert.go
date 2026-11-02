package probe

import "regexp"

// compileFullMatch 编译「完整匹配」正则：把用户正则整体包进锚点，只有整段
// 文本命中才算匹配。请求校验与执行时共用同一个包装，避免两边口径漂移。
func compileFullMatch(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile("^(?:" + pattern + ")$")
}
