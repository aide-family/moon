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

// ShareReport 生成一条新的公开链接。token 只在本次响应返回，库里只存哈希。
// 每次调用新增一条分享记录，不会自动停用旧链接；旧链接由用户逐条关闭。
func (gateway *Gateway) ShareReport(ctx context.Context, uid string, days int32, password string) (*bo.InspectionReport, *bo.InspectionReportShare, string, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionShare)
	if err != nil {
		return nil, nil, "", err
	}
	if _, err := gateway.repo.GetInspectionReport(ctx, ns, uid); err != nil {
		return nil, nil, "", gateway.access.Wrap(err)
	}
	if days != 0 && days != 7 && days != 30 {
		return nil, nil, "", merr.Invalid("share expiry must be 0, 7 or 30 days")
	}
	token, err := newShareToken()
	if err != nil {
		return nil, nil, "", merr.Server("generate share token failed")
	}
	var expiresAt *time.Time
	if days > 0 {
		value := time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
		expiresAt = &value
	}
	passwordHash := ""
	if password = strings.TrimSpace(password); password != "" {
		if len(password) > 64 {
			return nil, nil, "", merr.Invalid("share password is too long")
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, nil, "", merr.Server("hash share password failed")
		}
		passwordHash = string(hashed)
	}
	operator, _ := actor.UserID(ctx)
	share, err := gateway.repo.CreateInspectionReportShare(ctx, bo.InspectionReportShare{
		NamespaceUID: ns, ReportUID: uid, TokenHash: hashShareToken(token),
		PasswordHash: passwordHash, ExpiresAt: expiresAt, CreatedBy: operator,
	})
	if err != nil {
		return nil, nil, "", gateway.access.Wrap(err)
	}
	// 旧字段保留给历史公开链接兜底读取，始终指向最新一条分享。
	report, err := gateway.repo.ShareInspectionReport(ctx, ns, uid, share.TokenHash, expiresAt, passwordHash, operator)
	if err != nil {
		return nil, nil, "", gateway.access.Wrap(err)
	}
	share.HasPassword = passwordHash != ""
	return report, share, token, nil
}

// ListReportShares 列出报告的分享记录。旧版单链接数据会转成一条只读记录返回。
func (gateway *Gateway) ListReportShares(ctx context.Context, uid string) ([]bo.InspectionReportShare, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionShare)
	if err != nil {
		return nil, err
	}
	report, err := gateway.repo.GetInspectionReport(ctx, ns, uid)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	items, err := gateway.repo.ListInspectionReportShares(ctx, ns, uid)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if len(items) == 0 && report.ShareEnabled {
		createdAt := report.CreatedAt
		if report.ShareCreatedAt != nil {
			createdAt = *report.ShareCreatedAt
		}
		item := bo.InspectionReportShare{
			NamespaceUID: ns, ReportUID: uid, ExpiresAt: report.ShareExpiresAt,
			CreatedBy: report.ShareCreatedBy, CreatedAt: createdAt,
			ViewCount: report.ShareViewCount, LastViewedAt: report.ShareLastViewedAt,
			HasPassword: report.SharePasswordHash != "",
		}
		items = append(items, item)
	}
	for index := range items {
		items[index].HasPassword = items[index].PasswordHash != ""
		items[index].TokenHash = ""
		items[index].PasswordHash = ""
	}
	return items, nil
}

// RevokeReportShareLink 停用一条分享记录，并把报告旧字段同步到剩余的最新链接。
func (gateway *Gateway) RevokeReportShareLink(ctx context.Context, uid, shareUID string) (*bo.InspectionReportShare, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionShare)
	if err != nil {
		return nil, err
	}
	share, err := gateway.repo.RevokeInspectionReportShareByUID(ctx, ns, uid, shareUID, time.Now().UTC())
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if err := gateway.syncReportShareFields(ctx, ns, uid); err != nil {
		gateway.log.Error("sync inspection report share: " + err.Error())
	}
	share.HasPassword = share.PasswordHash != ""
	share.TokenHash = ""
	share.PasswordHash = ""
	return share, nil
}

// RevokeReportShare 停用报告的全部公开链接，兼容旧接口。
func (gateway *Gateway) RevokeReportShare(ctx context.Context, uid string) (*bo.InspectionReport, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionShare)
	if err != nil {
		return nil, err
	}
	if err := gateway.repo.RevokeAllInspectionReportShares(ctx, ns, uid, time.Now().UTC()); err != nil {
		return nil, gateway.access.Wrap(err)
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
	hash := hashShareToken(token)
	share, item, err := gateway.repo.GetInspectionReportShareByToken(ctx, hash, time.Now().UTC())
	if err != nil {
		// 兼容升级前生成的单链接。
		legacy, legacyErr := gateway.repo.GetSharedInspectionReport(ctx, hash, time.Now().UTC())
		if legacyErr != nil {
			return nil, gateway.access.Wrap(err)
		}
		if legacy.SharePasswordHash != "" {
			if bcrypt.CompareHashAndPassword([]byte(legacy.SharePasswordHash), []byte(password)) != nil {
				return nil, merr.NotFound("report not found")
			}
		}
		if touchErr := gateway.repo.TouchInspectionReportShare(ctx, legacy.UID, time.Now().UTC()); touchErr != nil {
			gateway.log.Error("touch inspection report share: " + touchErr.Error())
		}
		return legacy, nil
	}
	if share.PasswordHash != "" {
		if bcrypt.CompareHashAndPassword([]byte(share.PasswordHash), []byte(password)) != nil {
			return nil, merr.NotFound("report not found")
		}
	}
	if err := gateway.repo.TouchInspectionReportShareByUID(ctx, share.UID, time.Now().UTC()); err != nil {
		gateway.log.Error("touch inspection report share: " + err.Error())
	}
	return item, nil
}

func (gateway *Gateway) syncReportShareFields(ctx context.Context, namespaceUID, reportUID string) error {
	items, err := gateway.repo.ListInspectionReportShares(ctx, namespaceUID, reportUID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	var latest *bo.InspectionReportShare
	for index := range items {
		item := items[index]
		if item.RevokedAt != nil {
			continue
		}
		if item.ExpiresAt != nil && !item.ExpiresAt.After(now) {
			continue
		}
		latest = &item
		break
	}
	if latest == nil {
		_, err := gateway.repo.RevokeInspectionReportShare(ctx, namespaceUID, reportUID)
		return err
	}
	_, err = gateway.repo.ShareInspectionReport(
		ctx, namespaceUID, reportUID, latest.TokenHash, latest.ExpiresAt, latest.PasswordHash, latest.CreatedBy,
	)
	return err
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
