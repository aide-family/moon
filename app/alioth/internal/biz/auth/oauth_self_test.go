package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

const (
	selfOAuthUID  = "user-self"
	otherOAuthUID = "user-other"
)

// oauthStubLink 把桩里的绑定映射成 bo.OAuthLink。UID 用 provider + open_id 派生，
// 保持稳定，测试据此定位要解绑的那一条。绑定时间按插入顺序递增。
func oauthStubLink(link oauthLink, index int) bo.OAuthLink {
	return bo.OAuthLink{
		UID:      "link-" + link.provider + "-" + link.openID,
		Provider: link.provider,
		OpenID:   link.openID,
		Email:    link.email,
		BoundAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(index) * time.Minute),
	}
}

// ListOAuthByUser 只返回指定账号的绑定，顺序与真实仓储一致：最近绑定的在前。
func (stub *oauthRepoStub) ListOAuthByUser(_ context.Context, userUID string) ([]bo.OAuthLink, error) {
	links := make([]bo.OAuthLink, 0, len(stub.links))
	for i := len(stub.links) - 1; i >= 0; i-- {
		if stub.links[i].userUID != userUID {
			continue
		}
		links = append(links, oauthStubLink(stub.links[i], i))
	}
	return links, nil
}

// OAuthByUserProvider 按账号和来源返回最近的一条绑定。
func (stub *oauthRepoStub) OAuthByUserProvider(_ context.Context, userUID, provider string) (*bo.OAuthLink, error) {
	for i := len(stub.links) - 1; i >= 0; i-- {
		if stub.links[i].userUID == userUID && stub.links[i].provider == provider {
			item := oauthStubLink(stub.links[i], i)
			return &item, nil
		}
	}
	return nil, merr.NotFound("not found")
}

// UnlinkOAuth 只在 uid 和 user_uid 同时匹配时删除，模拟真实仓储的删除条件。
func (stub *oauthRepoStub) UnlinkOAuth(_ context.Context, userUID, uid string) error {
	for i := range stub.links {
		if stub.links[i].userUID != userUID {
			continue
		}
		if oauthStubLink(stub.links[i], i).UID != uid {
			continue
		}
		stub.links = append(stub.links[:i], stub.links[i+1:]...)
		return nil
	}
	return merr.NotFound("not found")
}

// selfOAuthContext 模拟已登录的当前账号。个人资料页的自助接口只从上下文取 uid。
func selfOAuthContext() context.Context {
	return actor.WithUser(context.Background(), selfOAuthUID)
}

// dingtalkProfileStub 让钉钉授权码换 profile 的链路指向 httptest，不访问真实网络。
// 钉钉的 open_id 和邮箱是绑定用例唯一关心的字段。
func dingtalkProfileStub(t *testing.T, openID, email string) {
	t.Helper()
	newDingtalkStub(t, dingtalkReply{
		tokenResponse:    `{"accessToken":"u-dingtalk-token"}`,
		userInfoResponse: fmt.Sprintf(`{"nick":"张三","openId":%q,"email":%q}`, openID, email),
	})
}

// newSelfOAuthFixture 准备一个当前账号和一个可用的钉钉客户端。
func newSelfOAuthFixture(t *testing.T) *authFixture {
	t.Helper()
	fixture := newAuthFixture(t)
	configureClients(fixture, "DINGTALK")
	fixture.addAccount(selfOAuthUID, "self@example.com")
	return fixture
}

// assertOAuthError 同时断言 HTTP 语义（code/reason）和前端依赖的原文。
func assertOAuthError(t *testing.T, err error, code int32, reason, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("want %s error %q, got nil", reason, message)
	}
	info := kerrors.FromError(err)
	if info.Code != code || info.Reason != reason {
		t.Fatalf("code=%d reason=%s, want code=%d reason=%s", info.Code, info.Reason, code, reason)
	}
	if info.Message != message {
		t.Fatalf("message = %q, want %q", info.Message, message)
	}
}

// TestBindSelfOAuthRejectsOpenIDBoundToAnotherAccount 覆盖把他人已绑定的第三方身份
// 绑到自己的场景：必须 409，且不能调用 MoveOAuth 把绑定抢过来。
func TestBindSelfOAuthRejectsOpenIDBoundToAnotherAccount(t *testing.T) {
	dingtalkProfileStub(t, "open-shared", "shared@example.com")
	fixture := newSelfOAuthFixture(t)
	fixture.addAccount(otherOAuthUID, "other@example.com")
	fixture.oauth.addLink(otherOAuthUID, "DINGTALK", "open-shared", "old@example.com")

	_, err := fixture.authenticator.BindSelfOAuth(selfOAuthContext(), "DINGTALK", "auth-code", oauthRedirectURI)
	assertOAuthError(t, err, 409, merr.ReasonConflict, "oauth account is already bound")

	if len(fixture.oauth.moves) != 0 {
		t.Fatalf("绑定他人已绑定的身份时不能调用 MoveOAuth: %+v", fixture.oauth.moves)
	}
	if len(fixture.oauth.links) != 1 || fixture.oauth.links[0].userUID != otherOAuthUID {
		t.Fatalf("冲突后绑定归属不应变化: %+v", fixture.oauth.links)
	}
}

// TestBindSelfOAuthRepeatedBindOnSameAccountIsIdempotent 覆盖重复绑定到本账号：
// 第二次返回同一条绑定，不新增行，也不走改绑。
func TestBindSelfOAuthRepeatedBindOnSameAccountIsIdempotent(t *testing.T) {
	dingtalkProfileStub(t, "open-self", "self-oauth@example.com")
	fixture := newSelfOAuthFixture(t)

	first, err := fixture.authenticator.BindSelfOAuth(selfOAuthContext(), "DINGTALK", "auth-code", oauthRedirectURI)
	if err != nil {
		t.Fatalf("BindSelfOAuth: %v", err)
	}
	if first.Provider != "DINGTALK" || first.OpenID != "open-self" || first.Email != "self-oauth@example.com" {
		t.Fatalf("绑定结果 = %+v", first)
	}
	if first.UID == "" || first.BoundAt.IsZero() {
		t.Fatalf("绑定结果缺少 uid 或绑定时间: %+v", first)
	}
	if len(fixture.oauth.links) != 1 {
		t.Fatalf("首次绑定后绑定条数 = %d, want 1", len(fixture.oauth.links))
	}

	second, err := fixture.authenticator.BindSelfOAuth(selfOAuthContext(), "DINGTALK", "auth-code", oauthRedirectURI)
	if err != nil {
		t.Fatalf("重复 BindSelfOAuth: %v", err)
	}
	if second.UID != first.UID {
		t.Fatalf("重复绑定应返回同一条绑定: %q != %q", second.UID, first.UID)
	}
	if len(fixture.oauth.links) != 1 {
		t.Fatalf("重复绑定不应新增行，绑定条数 = %d", len(fixture.oauth.links))
	}
	if len(fixture.oauth.moves) != 0 {
		t.Fatalf("重复绑定不能调用 MoveOAuth: %+v", fixture.oauth.moves)
	}
}

// TestUnbindSelfOAuthRejectsOtherUsersLink 覆盖解绑他人的 uid：按不存在处理（404），
// 不泄露该绑定是否存在，也不删除。
func TestUnbindSelfOAuthRejectsOtherUsersLink(t *testing.T) {
	fixture := newSelfOAuthFixture(t)
	fixture.addAccount(otherOAuthUID, "other@example.com")
	fixture.oauth.addLink(otherOAuthUID, "DINGTALK", "open-other", "other-oauth@example.com")

	uid := oauthStubLink(fixture.oauth.links[0], 0).UID
	err := fixture.authenticator.UnbindSelfOAuth(selfOAuthContext(), uid)
	assertOAuthError(t, err, 404, merr.ReasonNotFound, "oauth link not found")

	if len(fixture.oauth.links) != 1 || fixture.oauth.links[0].userUID != otherOAuthUID {
		t.Fatalf("解绑他人 uid 不应删除绑定: %+v", fixture.oauth.links)
	}
}

// TestUnbindSelfOAuthRejectsLastLoginMethod 覆盖最后一个登录方式：空邮箱和
// @users.alioth.local 占位邮箱都算不可用邮箱，必须 409 且不删除。
func TestUnbindSelfOAuthRejectsLastLoginMethod(t *testing.T) {
	cases := []struct {
		name  string
		email string
	}{
		{name: "empty email", email: ""},
		{name: "placeholder email", email: "user-self@users.alioth.local"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newAuthFixture(t)
			fixture.addAccount(selfOAuthUID, tc.email)
			fixture.oauth.addLink(selfOAuthUID, "DINGTALK", "open-self", "self-oauth@example.com")

			uid := oauthStubLink(fixture.oauth.links[0], 0).UID
			err := fixture.authenticator.UnbindSelfOAuth(selfOAuthContext(), uid)
			assertOAuthError(t, err, 409, merr.ReasonConflict, "cannot unlink the last login method")

			if len(fixture.oauth.links) != 1 {
				t.Fatalf("最后一个登录方式不能删除，绑定条数 = %d", len(fixture.oauth.links))
			}
		})
	}
}

// TestUnbindSelfOAuthAllowsWhenAnotherLoginMethodRemains 确认约束不会过宽：
// 有可用邮箱，或还有别的 OAuth 绑定时允许解绑。
func TestUnbindSelfOAuthAllowsWhenAnotherLoginMethodRemains(t *testing.T) {
	t.Run("usable email", func(t *testing.T) {
		fixture := newSelfOAuthFixture(t)
		fixture.oauth.addLink(selfOAuthUID, "DINGTALK", "open-self", "self-oauth@example.com")

		uid := oauthStubLink(fixture.oauth.links[0], 0).UID
		if err := fixture.authenticator.UnbindSelfOAuth(selfOAuthContext(), uid); err != nil {
			t.Fatalf("账号还有可用邮箱时应允许解绑: %v", err)
		}
		if len(fixture.oauth.links) != 0 {
			t.Fatalf("解绑后绑定条数 = %d, want 0", len(fixture.oauth.links))
		}
	})

	t.Run("another oauth link", func(t *testing.T) {
		fixture := newAuthFixture(t)
		fixture.addAccount(selfOAuthUID, "user-self@users.alioth.local")
		fixture.oauth.addLink(selfOAuthUID, "DINGTALK", "open-dingtalk", "dingtalk@example.com")
		fixture.oauth.addLink(selfOAuthUID, "GITHUB", "open-github", "github@example.com")

		uid := oauthStubLink(fixture.oauth.links[0], 0).UID
		if err := fixture.authenticator.UnbindSelfOAuth(selfOAuthContext(), uid); err != nil {
			t.Fatalf("还有别的 OAuth 绑定时应允许解绑: %v", err)
		}
		if len(fixture.oauth.links) != 1 || fixture.oauth.links[0].openID != "open-github" {
			t.Fatalf("只应删除目标绑定: %+v", fixture.oauth.links)
		}
	})
}

// TestSelfOAuthLinksListsOnlyCurrentUser 确认列表只返回当前账号的绑定，
// 不会因为上下文里没有租户而混入其他账号的数据。
func TestSelfOAuthLinksListsOnlyCurrentUser(t *testing.T) {
	fixture := newSelfOAuthFixture(t)
	fixture.addAccount(otherOAuthUID, "other@example.com")
	fixture.oauth.addLink(selfOAuthUID, "DINGTALK", "open-self", "self-oauth@example.com")
	fixture.oauth.addLink(otherOAuthUID, "GITHUB", "open-other", "other@example.com")

	links, err := fixture.authenticator.SelfOAuthLinks(selfOAuthContext())
	if err != nil {
		t.Fatalf("SelfOAuthLinks: %v", err)
	}
	if len(links) != 1 || links[0].OpenID != "open-self" {
		t.Fatalf("只应返回当前账号的绑定: %+v", links)
	}
}

// TestSelfOAuthOperationsRequireOnlyLogin 确认三个自助接口只要求登录态：
// 不带 X-Namespace 也能通过认证，未登录返回 401，而不是会触发整页跳转的 403。
func TestSelfOAuthOperationsRequireOnlyLogin(t *testing.T) {
	fixture := newSelfOAuthFixture(t)
	token, err := fixture.authenticator.sign(selfOAuthUID, "access", time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	operations := []string{
		v1.OperationSelfServiceListSelfOAuthLinks,
		v1.OperationSelfServiceBindSelfOAuth,
		v1.OperationSelfServiceUnbindSelfOAuth,
	}
	for _, operation := range operations {
		ctx, err := fixture.authenticator.Authenticate(context.Background(), operation, "Bearer "+token, "")
		if err != nil {
			t.Fatalf("%s 只要求登录态，未选租户也应通过认证: %v", operation, err)
		}
		if uid, ok := actor.UserID(ctx); !ok || uid != selfOAuthUID {
			t.Fatalf("%s 认证后当前账号 = %q, want %q", operation, uid, selfOAuthUID)
		}
		_, err = fixture.authenticator.Authenticate(context.Background(), operation, "", "")
		if kerrors.Reason(err) != merr.ReasonUnauthenticated {
			t.Fatalf("%s 未登录时应返回 401，实际 %v", operation, err)
		}
	}
}
