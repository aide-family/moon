package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/plugin/cache"
)

type Cache interface {
	Lock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
	BanToken(ctx context.Context, token string) error
	VerifyToken(ctx context.Context, token string) error

	VerifyOAuthToken(ctx context.Context, oauthParams *bo.OAuthLoginParams) error
	CacheVerifyOAuthToken(ctx context.Context, oauthParams *bo.OAuthLoginParams) error
	SendVerifyEmailCode(ctx context.Context, params *bo.VerifyEmailParams) error
	VerifyEmailCode(ctx context.Context, params *bo.VerifyEmailCodeParams) error

	CacheUsers(ctx context.Context, users ...do.User) error
	GetUser(ctx context.Context, userID uint32) (do.User, error)
	GetUsers(ctx context.Context, ids ...uint32) ([]do.User, error)

	CacheTeams(ctx context.Context, teams ...do.Team) error
	GetTeam(ctx context.Context, teamID uint32) (do.Team, error)
	GetTeams(ctx context.Context, ids ...uint32) ([]do.Team, error)

	CacheTeamMembers(ctx context.Context, members ...do.TeamMember) error
	GetTeamMember(ctx context.Context, memberID uint32) (do.TeamMember, error)
	GetTeamMembers(ctx context.Context, ids ...uint32) ([]do.TeamMember, error)

	CacheMenus(ctx context.Context, menus ...do.Menu) error
	GetMenu(ctx context.Context, operation string) (do.Menu, error)
	GetMenus(ctx context.Context, operations ...string) ([]do.Menu, error)
}

const (
	EmailCodeKey                        cache.K = "palace:verify:email:code"
	BankTokenKey                        cache.K = "palace:token:ban"
	OAuthTokenKey                       cache.K = "palace:token:oauth"
	UserCacheKey                        cache.K = "palace:user:cache"
	TeamCacheKey                        cache.K = "palace:team:cache"
	TeamMemberCacheKey                  cache.K = "palace:team:member:cache"
	TeamDatasourceMetricMetadataSyncKey cache.K = "palace:team:datasource:metric:metadata:sync"
	MenuCacheKey                        cache.K = "palace:menu:cache"
)
