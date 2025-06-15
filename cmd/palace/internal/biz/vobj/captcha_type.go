package vobj

//go:generate stringer -type=CaptchaType -linecomment -output=captcha_type.string.go
type CaptchaType int8

const (
	CaptchaTypeUnknown CaptchaType = iota // Unknown
	CaptchaTypeClick                      // 点击验证码
	CaptchaTypeSlide                      // 滑动验证码
	CaptchaTypeRotate                     // 选择验证码
)
