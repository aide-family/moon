package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/helper/middleware"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/captcha"
	"github.com/aide-family/moon/pkg/util/validate"
)

func LoginSignToUserBaseItem(b *middleware.JwtBaseInfo) *common.UserBaseItem {
	if b == nil {
		return nil
	}
	return &common.UserBaseItem{
		Username: b.Username,
		Nickname: b.Nickname,
		Avatar:   b.Avatar,
		Gender:   common.Gender(b.Gender.GetValue()),
		UserId:   b.UserID,
	}
}

func LoginReply(b *bo.LoginSign) *palace.LoginReply {
	return &palace.LoginReply{
		Token:          b.Token,
		ExpiredSeconds: int32(b.ExpiredSeconds),
		User:           LoginSignToUserBaseItem(b.Base),
	}
}

func ToTLS(tls *common.TLS) *do.TLS {
	if validate.IsNil(tls) {
		return nil
	}
	return &do.TLS{
		ServerName: tls.GetServerName(),
		ClientCert: tls.GetClientCert(),
		ClientKey:  tls.GetClientKey(),
	}
}

func ToTLSItem(tls *do.TLS) *common.TLS {
	if validate.IsNil(tls) {
		return nil
	}
	return &common.TLS{
		ServerName: tls.ServerName,
		ClientCert: tls.ClientCert,
		ClientKey:  tls.ClientKey,
	}
}

func ToBasicAuth(basicAuth *common.BasicAuth) *do.BasicAuth {
	if validate.IsNil(basicAuth) {
		return nil
	}
	return &do.BasicAuth{
		Username: basicAuth.GetUsername(),
		Password: basicAuth.GetPassword(),
	}
}

func ToBasicAuthItem(basicAuth *do.BasicAuth) *common.BasicAuth {
	if validate.IsNil(basicAuth) {
		return nil
	}
	return &common.BasicAuth{
		Username: basicAuth.Username,
		Password: basicAuth.Password,
	}
}

func ToGetCaptchaReply(res *captcha.GenResult) *palace.GetCaptchaReply {
	if validate.IsNil(res) {
		return nil
	}

	return &palace.GetCaptchaReply{
		CaptchaId:         res.CaptchaKey,
		MasterImageBase64: res.MasterImageBase64,
		ThumbImageBase64:  res.ThumbImageBase64,
		ThumbSize:         int32(res.ThumbSize),
		TileWidth:         int32(res.TileWidth),
		TileHeight:        int32(res.TileHeight),
	}
}

func ToCaptchaValidateRequest(req *palace.CaptchaValidateRequest) *bo.CaptchaVerify {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.CaptchaVerify{
		CaptchaID: req.GetCaptchaId(),
		Angle:     int(req.GetAngle()),
		Sx:        int(req.GetSx()),
		Sy:        int(req.GetSy()),
		Dots:      req.GetDots(),
	}
}
