package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha/v2/click"
	"github.com/wenlng/go-captcha/v2/rotate"
	"github.com/wenlng/go-captcha/v2/slide"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/conf"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/util/captcha"
)

func NewCaptchaRepo(bc *conf.Bootstrap, d *data.Data, logger log.Logger) repository.Captcha {
	captchaConf := bc.GetAuth().GetCaptcha()

	return &captchaRepoImpl{
		captchaType: captchaConf.Type,
		expired:     captchaConf.GetExpire().AsDuration(),
		Data:        d,
		helper:      log.NewHelper(log.With(logger, "module", "data.repo.captcha")),
	}
}

type captchaRepoImpl struct {
	*data.Data
	expired     time.Duration
	captchaType config.Captcha
	cli         *redis.Client

	helper *log.Helper
}

func (c *captchaRepoImpl) Generate(ctx context.Context) (*captcha.GenResult, error) {

	genResult, err := c.GenerateCaptcha(ctx)
	if err != nil {
		return nil, err
	}
	captchaKey := uuid.New().String()
	genResult.CaptchaKey = captchaKey
	genResult.Expired = c.expired

	key := repository.CaptchaCacheKey.Key(captchaKey)
	_, err = c.GetCache().Client().SetNX(ctx, key, genResult, c.expired).Result()
	if err != nil {
		return nil, err
	}
	return genResult, nil
}

func (c *captchaRepoImpl) GenerateCaptcha(_ context.Context) (*captcha.GenResult, error) {
	switch c.captchaType {
	case config.Captcha_Click:
		clickCaptcha, err := captcha.GenerateClickCaptcha()
		if err != nil {
			return nil, err
		}
		clickCaptcha.CaptchaType = config.Captcha_Click
		return clickCaptcha, nil
	case config.Captcha_Slide:
		slideCaptcha, err := captcha.GenerateSlideCaptcha()
		if err != nil {
			return nil, err
		}
		slideCaptcha.CaptchaType = config.Captcha_Slide
		return slideCaptcha, nil
	case config.Captcha_Rotate:
		rotateCaptcha, err := captcha.GenerateRotateCaptcha()
		if err != nil {
			return nil, err
		}
		rotateCaptcha.CaptchaType = config.Captcha_Rotate
		return rotateCaptcha, nil
	default:
		return nil, fmt.Errorf("captcha type not support")
	}
}

func (c *captchaRepoImpl) Verify(ctx context.Context, req *bo.CaptchaVerify) bool {

	key := repository.CaptchaCacheKey.Key(req.CaptchaID)

	var genResult captcha.GenResult
	err := c.GetCache().Client().Get(ctx, key).Scan(&genResult)
	if err != nil {
		return false
	}

	switch genResult.CaptchaType {
	case config.Captcha_Click:
		src := strings.Split(req.Dots, ",")
		var dct map[int]*click.Dot
		if err = json.Unmarshal([]byte(genResult.DotData), &dct); err != nil {
			c.helper.WithContext(ctx).Errorf("captcha verify error: %v", err)
			return false
		}

		chkRet := false
		if (len(dct) * 2) == len(src) {
			for i := 0; i < len(dct); i++ {
				dot := dct[i]
				j := i * 2
				k := i*2 + 1
				sx, _ := strconv.Atoi(src[j])
				sy, _ := strconv.Atoi(src[k])

				chkRet = click.Validate(sx, sy, dot.X, dot.Y, dot.Width, dot.Height, 5)
				if !chkRet {
					break
				}
			}
		}
		return chkRet
	case config.Captcha_Slide:
		return slide.Validate(req.Sx, req.Sy, genResult.TileX, genResult.TileY, 5)
	case config.Captcha_Rotate:
		return rotate.Validate(req.Angle, genResult.Angle, 10)
	}
	return false
}
