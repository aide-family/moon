package impl

import (
	"context"
	_ "embed"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/conf"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/data/impl/build"
	"github.com/aide-family/moon/cmd/palace/internal/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/hash"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/template"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewCacheRepo(bc *conf.Bootstrap, d *data.Data, logger log.Logger) repository.Cache {
	return &cacheReoImpl{
		bc:      bc,
		signKey: bc.GetAuth().GetJwt().GetSignKey(),
		Data:    d,
		helper:  log.NewHelper(log.With(logger, "module", "data.repo.cache")),
	}
}

type cacheReoImpl struct {
	bc      *conf.Bootstrap
	signKey string
	*data.Data

	helper *log.Helper
}

func (c *cacheReoImpl) CacheTeams(ctx context.Context, teams ...do.Team) error {
	key := repository.UserCacheKey.Key()
	teamsMap := make(map[string]any)
	for _, team := range teams {
		teamItem := build.ToTeam(ctx, team)
		if validate.IsNil(teamItem) {
			continue
		}
		teamsMap[teamItem.UniqueKey()] = teamItem
	}
	return c.GetCache().Client().HSet(ctx, key, teamsMap).Err()
}

func (c *cacheReoImpl) GetTeam(ctx context.Context, teamID uint32) (do.Team, error) {
	key := repository.TeamCacheKey.Key()
	teamKey := strconv.Itoa(int(teamID))
	exist, err := c.GetCache().Client().HExists(ctx, key, teamKey).Result()
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, merr.ErrorNotFound("team not found")
	}
	var team system.Team
	if err = c.GetCache().Client().HGet(ctx, key, teamKey).Scan(&team); err != nil {
		return nil, err
	}
	return &team, nil
}

func (c *cacheReoImpl) GetTeams(ctx context.Context, ids ...uint32) ([]do.Team, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	key := repository.TeamCacheKey.Key()
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exist == 0 {
		return nil, nil
	}
	teamKeys := slices.Map(ids, func(id uint32) string { return strconv.Itoa(int(id)) })
	teamMap, err := c.GetCache().Client().HMGet(ctx, key, teamKeys...).Result()
	if err != nil {
		return nil, err
	}
	teams := make([]do.Team, 0, len(teamMap))
	for _, v := range teamMap {
		var team system.Team
		if err := team.UnmarshalBinary([]byte(v.(string))); err != nil {
			continue
		}
		teams = append(teams, &team)
	}
	return teams, nil
}

func (c *cacheReoImpl) CacheTeamMembers(ctx context.Context, members ...do.TeamMember) error {
	if len(members) == 0 {
		return nil
	}
	key := repository.TeamMemberCacheKey.Key()
	membersMap := make(map[string]any)
	for _, member := range members {
		memberItem := build.ToTeamMember(ctx, member)
		if validate.IsNil(memberItem) {
			continue
		}
		membersMap[memberItem.UniqueKey()] = memberItem
	}
	return c.GetCache().Client().HSet(ctx, key, membersMap).Err()
}

func (c *cacheReoImpl) GetTeamMember(ctx context.Context, memberID uint32) (do.TeamMember, error) {
	key := repository.TeamMemberCacheKey.Key()
	memberKey := strconv.Itoa(int(memberID))
	exist, err := c.GetCache().Client().HExists(ctx, key, memberKey).Result()
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, merr.ErrorNotFound("team member not found")
	}
	var member system.TeamMember
	if err = c.GetCache().Client().HGet(ctx, key, memberKey).Scan(&member); err != nil {
		return nil, err
	}
	return &member, nil
}

func (c *cacheReoImpl) GetTeamMembers(ctx context.Context, ids ...uint32) ([]do.TeamMember, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	key := repository.TeamMemberCacheKey.Key()
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exist == 0 {
		return nil, nil
	}
	memberKeys := slices.Map(ids, func(id uint32) string { return strconv.Itoa(int(id)) })
	memberMap, err := c.GetCache().Client().HMGet(ctx, key, memberKeys...).Result()
	if err != nil {
		return nil, err
	}
	members := make([]do.TeamMember, 0, len(memberMap))
	for _, v := range memberMap {
		var member system.TeamMember
		if err := member.UnmarshalBinary([]byte(v.(string))); err != nil {
			continue
		}
		members = append(members, &member)
	}
	return members, nil
}

func (c *cacheReoImpl) CacheUsers(ctx context.Context, users ...do.User) error {
	key := repository.UserCacheKey.Key()
	usersMap := make(map[string]any)
	for _, user := range users {
		userItem := build.ToUser(ctx, user)
		if validate.IsNil(userItem) {
			continue
		}
		usersMap[userItem.UniqueKey()] = userItem
	}
	return c.GetCache().Client().HSet(ctx, key, usersMap).Err()
}

func (c *cacheReoImpl) GetUser(ctx context.Context, userID uint32) (do.User, error) {
	key := repository.UserCacheKey.Key()
	userKey := strconv.Itoa(int(userID))
	exist, err := c.GetCache().Client().HExists(ctx, key, userKey).Result()
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, merr.ErrorUserNotFound("user not found")
	}
	var user system.User
	if err = c.GetCache().Client().HGet(ctx, key, userKey).Scan(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *cacheReoImpl) GetUsers(ctx context.Context, ids ...uint32) ([]do.User, error) {
	key := repository.UserCacheKey.Key()
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exist == 0 {
		return nil, nil
	}
	userKeys := slices.Map(ids, func(id uint32) string { return strconv.Itoa(int(id)) })
	userMap, err := c.GetCache().Client().HMGet(ctx, key, userKeys...).Result()
	if err != nil {
		return nil, err
	}
	users := make([]do.User, 0, len(userMap))
	for _, v := range userMap {
		var user system.User
		if err := user.UnmarshalBinary([]byte(v.(string))); err != nil {
			continue
		}
		users = append(users, &user)
	}
	return users, nil
}

func (c *cacheReoImpl) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.GetCache().Client().SetNX(ctx, key, 1, expiration).Result()
}

func (c *cacheReoImpl) Unlock(ctx context.Context, key string) error {
	return c.GetCache().Client().Del(ctx, key).Err()
}

func (c *cacheReoImpl) BanToken(ctx context.Context, token string) error {
	jwtClaims, err := middleware.ParseJwtClaimsFromToken(token, c.signKey)
	if err != nil {
		return err
	}
	expiration := jwtClaims.ExpiresAt.Sub(timex.Now())
	if expiration <= 0 {
		return merr.ErrorInvalidToken("token is invalid")
	}
	return c.GetCache().Client().Set(ctx, repository.BankTokenKey.Key(hash.MD5(token)), 1, expiration).Err()
}

func (c *cacheReoImpl) VerifyToken(ctx context.Context, token string) error {
	exist, err := c.GetCache().Client().Exists(ctx, repository.BankTokenKey.Key(hash.MD5(token))).Result()
	if err != nil {
		return err
	}
	if exist > 0 {
		return merr.ErrorInvalidToken("token is ban")
	}
	return nil
}

func (c *cacheReoImpl) VerifyOAuthToken(ctx context.Context, oauthParams *bo.OAuthLoginParams) error {
	key := repository.OAuthTokenKey.Key(oauthParams.APP, oauthParams.OpenID, oauthParams.Token)
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		return merr.ErrorInternalServerError("cache err").WithCause(err)
	}
	if exist == 0 {
		return merr.ErrorUnauthorized("oauth unauthorized").WithMetadata(map[string]string{
			"exist": "false",
		})
	}
	return c.GetCache().Client().Del(ctx, key).Err()
}

func (c *cacheReoImpl) CacheVerifyOAuthToken(ctx context.Context, oauthParams *bo.OAuthLoginParams) error {
	key := repository.OAuthTokenKey.Key(oauthParams.APP, oauthParams.OpenID, oauthParams.Token)
	return c.GetCache().Client().Set(ctx, key, "##code##", 10*time.Minute).Err()
}

func (c *cacheReoImpl) VerifyEmailCode(ctx context.Context, params *bo.VerifyEmailCodeParams) error {
	key := repository.EmailCodeKey.Key(params.Email)
	cacheCode, err := c.GetCache().Client().Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return merr.ErrorCaptchaError("captcha is expire").WithMetadata(map[string]string{
				"code": "captcha is expire",
			})
		}
		return merr.ErrorInternalServerError("cache err").WithCause(err)
	}
	defer c.GetCache().Client().Del(ctx, key).Val()
	if strings.EqualFold(cacheCode, params.Code) {
		return nil
	}
	return merr.ErrorCaptchaError("captcha err").WithMetadata(map[string]string{
		"code": "The verification code is incorrect. Please retrieve a new one and try again.",
	})
}

//go:embed template/verify_email.html
var verifyEmailTemplate string

func (c *cacheReoImpl) SendVerifyEmailCode(ctx context.Context, params *bo.VerifyEmailParams) error {
	if err := validate.CheckEmail(params.Email); err != nil {
		return err
	}
	code := strings.ToUpper(hash.MD5(timex.Now().String())[:6])
	err := c.GetCache().Client().Set(ctx, repository.EmailCodeKey.Key(params.Email), code, 5*time.Minute).Err()
	if err != nil {
		return err
	}

	bodyParams := map[string]string{
		"Email":       params.Email,
		"Code":        code,
		"RedirectURI": c.bc.GetAuth().GetOauth2().GetRedirectUri(),
	}
	emailBody, err := template.HtmlFormatter(verifyEmailTemplate, bodyParams)
	if err != nil {
		return err
	}
	sendEmailParams := &bo.SendEmailParams{
		Email:       params.Email,
		Body:        emailBody,
		Subject:     "Email verification code.",
		ContentType: "text/html",
		RequestID:   uuid.New().String(),
	}
	return params.SendEmailFun(ctx, sendEmailParams)
}
