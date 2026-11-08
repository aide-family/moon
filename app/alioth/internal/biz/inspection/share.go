package inspection

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// ShareReport 生成或轮换报告的公开链接。token 只在本次响应返回，库里只存哈希。
func (gateway *Gateway) ShareReport(ctx context.Context, uid string, days int32, password string) (*bo.InspectionReport, string, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionShare)
	if err != nil {
		return nil, "", err
	}
	if _, err := gateway.repo.GetInspectionReport(ctx, ns, uid); err != nil {
		return nil, "", gateway.access.Wrap(err)
	}
	if days != 0 && days != 7 && days != 30 {
		return nil, "", merr.Invalid("share expiry must be 0, 7 or 30 days")
	}
	token, err := newShareToken()
	if err != nil {
		return nil, "", merr.Server("generate share token failed")
	}
	var expiresAt *time.Time
	if days > 0 {
		value := time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
		expiresAt = &value
	}
	passwordHash := ""
	if password = strings.TrimSpace(password); password != "" {
		if len(password) > 64 {
			return nil, "", merr.Invalid("share password is too long")
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, "", merr.Server("hash share password failed")
		}
		passwordHash = string(hashed)
	}
	operator, _ := actor.UserID(ctx)
	item, err := gateway.repo.ShareInspectionReport(ctx, ns, uid, hashShareToken(token), expiresAt, passwordHash, operator)
	if err != nil {
		return nil, "", gateway.access.Wrap(err)
	}
	return item, token, nil
}

// RevokeReportShare 停用公开链接。
func (gateway *Gateway) RevokeReportShare(ctx context.Context, uid string) (*bo.InspectionReport, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionShare)
	if err != nil {
		return nil, err
	}
	item, err := gateway.repo.RevokeInspectionReportShare(ctx, ns, uid)
	return item, gateway.access.Wrap(err)
}

// PublicReport 按 token 读取公开报告。密码错误与 token 无效统一返回 NotFound，
// 不暴露报告是否存在；成功后记录一次访问。
func (gateway *Gateway) PublicReport(ctx context.Context, token, password string) (*bo.InspectionReport, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, merr.NotFound("report not found")
	}
	item, err := gateway.repo.GetSharedInspectionReport(ctx, hashShareToken(token), time.Now().UTC())
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if item.SharePasswordHash != "" {
		if bcrypt.CompareHashAndPassword([]byte(item.SharePasswordHash), []byte(password)) != nil {
			return nil, merr.NotFound("report not found")
		}
	}
	if err := gateway.repo.TouchInspectionReportShare(ctx, item.UID, time.Now().UTC()); err != nil {
		gateway.log.Error("touch inspection report share: " + err.Error())
	}
	return item, nil
}

func newShareToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func hashShareToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
