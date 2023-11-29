package captcha

import (
	"context"
	"errors"
	"image/color"
	"time"

	"github.com/mojocn/base64Captcha"
)

type CaptchaType int8

const (
	// CaptchaTypeAudio 音频验证码
	CaptchaTypeAudio CaptchaType = iota + 1
	// CaptchaTypeString 字符验证码
	CaptchaTypeString
	// CaptchaTypeMath 算术验证码
	CaptchaTypeMath
	// CaptchaTypeChinese 汉字验证码
	CaptchaTypeChinese
	// CaptchaTypeDigit 数字验证码
	CaptchaTypeDigit
)

var CaptchaTypeMap = map[CaptchaType]string{
	CaptchaTypeAudio:   "audio",
	CaptchaTypeString:  "string",
	CaptchaTypeMath:    "math",
	CaptchaTypeChinese: "chinese",
	CaptchaTypeDigit:   "digit",
}

// var result = base64Captcha.DefaultMemStore
// 设置存储的验证码为 20240个，过期时间为 3分钟
var result = base64Captcha.NewMemoryStore(20240, 3*time.Minute)

func IsCaptchaType(captchaType CaptchaType) bool {
	_, ok := CaptchaTypeMap[captchaType]
	return ok
}

func getSizes(size ...int) (int, int) {
	height := 50
	width := 100
	switch len(size) {
	case 1:
		if size[0] > 0 {
			height, width = size[0], size[0]
		}
	case 2:
		if size[0] > 0 {
			height = size[0]
		}
		if size[1] > 0 {
			width = size[1]
		}
	}
	return height, width
}

// mathConfig 生成图形化算术验证码配置
func mathConfig(size ...int) *base64Captcha.DriverMath {
	height, width := getSizes(size...)
	mathType := &base64Captcha.DriverMath{
		Height:          height,
		Width:           width,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowHollowLine,
		BgColor: &color.RGBA{
			R: 40,
			G: 30,
			B: 89,
			A: 29,
		},
		Fonts: nil,
	}
	return mathType
}

// digitConfig 生成图形化数字验证码配置
func digitConfig(size ...int) *base64Captcha.DriverDigit {
	height, width := getSizes(size...)
	digitType := &base64Captcha.DriverDigit{
		Height:   height,
		Width:    width,
		Length:   5,
		MaxSkew:  0.45,
		DotCount: 80,
	}
	return digitType
}

// stringConfig 生成图形化字符串验证码配置
func stringConfig(size ...int) *base64Captcha.DriverString {
	height, width := getSizes(size...)
	stringType := &base64Captcha.DriverString{
		Height:          height,
		Width:           width,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowHollowLine | base64Captcha.OptionShowSlimeLine,
		Length:          5,
		Source:          "123456789qwertyuiopasdfghjklzxcvb",
		BgColor: &color.RGBA{
			R: 40,
			G: 30,
			B: 89,
			A: 29,
		},
		Fonts: nil,
	}
	return stringType
}

// chineseConfig 生成图形化汉字验证码配置
func chineseConfig(size ...int) *base64Captcha.DriverChinese {
	height, width := getSizes(size...)
	chineseType := &base64Captcha.DriverChinese{
		Height:          height,
		Width:           width,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowSlimeLine,
		Length:          2,
		Source:          "设想,你在,处理,消费者,的音,频输,出音,频可,能无,论什,么都,没有,任何,输出,或者,它可,能是,单声道,立体声,或是,环绕立,体声的,不想要,的值",
		BgColor: &color.RGBA{
			R: 40,
			G: 30,
			B: 89,
			A: 29,
		},
		Fonts: nil,
	}
	return chineseType
}

// autoConfig 生成图形化数字音频验证码配置
func autoConfig() *base64Captcha.DriverAudio {
	chineseType := &base64Captcha.DriverAudio{
		Length:   4,
		Language: "zh",
	}
	return chineseType
}

// CreateCode 生成验证码
//
//	id 	验证码id
//	bse64s 	图片base64编码
//	err 	错误
func CreateCode(_ context.Context, captchaType CaptchaType, size ...int) (string, string, error) {
	var driver base64Captcha.Driver
	switch captchaType {
	case CaptchaTypeAudio:
		driver = autoConfig()
	case CaptchaTypeString:
		driver = stringConfig(size...)
	case CaptchaTypeMath:
		driver = mathConfig(size...)
	case CaptchaTypeChinese:
		driver = chineseConfig(size...)
	case CaptchaTypeDigit:
		driver = digitConfig(size...)
	default:
		driver = digitConfig(size...)
	}
	if driver == nil {
		return "", "", errors.New("验证码类型错误")
	}
	// 创建验证码并传入创建的类型的配置，以及存储的对象
	c := base64Captcha.NewCaptcha(driver, result)
	id, b64s, err := c.Generate()
	return id, b64s, err
}

// VerifyCaptcha 验证验证码是否和答案一致
//
//	@Pram id 验证码id
//	@Pram VerifyValue 用户输入的答案
//	@Result true：正确，false：失败
func VerifyCaptcha(id, VerifyValue string) bool {
	// result 为步骤1 创建的图片验证码存储对象
	return result.Verify(id, VerifyValue, true)
}

// GetCodeAnswer 获取验证码答案
//
//	@Pram codeId 验证码id
//	@Result 验证码答案
func GetCodeAnswer(codeId string) string {
	// result 为步骤1 创建的图片验证码存储对象
	return result.Get(codeId, false)
}
