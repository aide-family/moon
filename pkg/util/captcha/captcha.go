package captcha

import (
	"context"
	"errors"
	"image/color"
	"time"

	"github.com/mojocn/base64Captcha"
)

// Type 验证码类型
type Type int8

const (
	// TypeAudio 音频验证码
	TypeAudio Type = iota + 1
	// TypeString 字符验证码
	TypeString
	// TypeMath 算术验证码
	TypeMath
	// TypeChinese 汉字验证码
	TypeChinese
	// TypeDigit 数字验证码
	TypeDigit
)

// Theme 验证码主题
type Theme string

const (
	_ Theme = "dark"
	// LightTheme light
	LightTheme Theme = "light"
)

// var result = base64Captcha.DefaultMemStore
// 设置存储的验证码为 20240个，过期时间为 3分钟
var result = base64Captcha.NewMemoryStore(20240, 3*time.Minute)

func getSizes(size ...int) (int, int) {
	height := 50
	width := 100
	switch len(size) {
	case 1:
		if size[0] > 50 {
			height, width = size[0], size[0]
		}
	case 2:
		if size[0] > 50 {
			height = size[0]
		}
		if size[1] > 50 {
			width = size[1]
		}
	}
	return height, width
}

// mathConfig 生成图形化算术验证码配置
func mathConfig(theme Theme, size ...int) *base64Captcha.DriverMath {
	height, width := getSizes(size...)
	mathType := &base64Captcha.DriverMath{
		Height:          height,
		Width:           width,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowHollowLine,
		BgColor:         getBgColor(theme),
		Fonts:           nil,
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

// getBgColor 生成图形化背景色
func getBgColor(theme Theme) *color.RGBA {
	switch theme {
	default:
		return &color.RGBA{
			R: 0xff,
			G: 0xff,
			B: 0xff,
			A: 0xff,
		}
	case LightTheme:
		return &color.RGBA{
			R: 40,
			G: 30,
			B: 89,
			A: 29,
		}
	}
}

// stringConfig 生成图形化字符串验证码配置
func stringConfig(theme Theme, size ...int) *base64Captcha.DriverString {
	height, width := getSizes(size...)
	stringType := &base64Captcha.DriverString{
		Height:          height,
		Width:           width,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowHollowLine | base64Captcha.OptionShowSlimeLine,
		Length:          5,
		Source:          "123456789qwertyuiopasdfghjklzxcvb",
		BgColor:         getBgColor(theme),
		Fonts:           nil,
	}
	return stringType
}

// chineseConfig 生成图形化汉字验证码配置
func chineseConfig(theme Theme, size ...int) *base64Captcha.DriverChinese {
	height, width := getSizes(size...)
	chineseType := &base64Captcha.DriverChinese{
		Height:          height,
		Width:           width,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowSlimeLine,
		Length:          2,
		Source:          "设想,你在,处理,消费者,的音,频输,出音,频可,能无,论什,么都,没有,任何,输出,或者,它可,能是,单声道,立体声,或是,环绕立,体声的,不想要,的值",
		BgColor:         getBgColor(theme),
		Fonts:           nil,
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
func CreateCode(_ context.Context, captchaType Type, theme Theme, size ...int) (string, string, error) {
	var driver base64Captcha.Driver
	switch captchaType {
	case TypeAudio:
		driver = autoConfig()
	case TypeString:
		driver = stringConfig(theme, size...)
	case TypeMath:
		driver = mathConfig(theme, size...)
	case TypeChinese:
		driver = chineseConfig(theme, size...)
	case TypeDigit:
		driver = digitConfig(size...)
	default:
		driver = digitConfig(size...)
	}
	if driver == nil {
		return "", "", errors.New("验证码类型错误")
	}
	// 创建验证码并传入创建的类型的配置，以及存储的对象
	c := base64Captcha.NewCaptcha(driver, result)
	id, b64s, _, err := c.Generate()
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
func GetCodeAnswer(codeID string) string {
	// result 为步骤1 创建的图片验证码存储对象
	return result.Get(codeID, false)
}
