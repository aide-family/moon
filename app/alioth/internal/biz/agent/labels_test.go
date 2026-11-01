package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/biz/agenthub"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// TestNormalizeAgentLabels 覆盖空值、规范化、非法 JSON、非字符串值与键值限制。
func TestNormalizeAgentLabels(t *testing.T) {
	longKey := make([]byte, maxAgentLabelKeyLength+1)
	for i := range longKey {
		longKey[i] = 'a'
	}
	longValue := make([]byte, maxAgentLabelValueLength+1)
	for i := range longValue {
		longValue[i] = 'a'
	}
	// Go 的 json.Marshal 对 map 按键排序，因此输出是稳定的规范形式。
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr string
	}{
		{name: "empty", raw: "", want: "{}"},
		{name: "spaces", raw: "  \n\t ", want: "{}"},
		{name: "empty object", raw: "{}", want: "{}"},
		{name: "null object", raw: "null", want: "{}"},
		{name: "orders keys", raw: `{"role":"web","env":"prod"}`, want: `{"env":"prod","role":"web"}`},
		{name: "allows empty value", raw: `{"env":""}`, want: `{"env":""}`},
		{name: "allows dot underscore dash", raw: `{"k8s.io-node_role.1":"a.b_c-d"}`, want: `{"k8s.io-node_role.1":"a.b_c-d"}`},
		{name: "invalid json", raw: `not-json`, wantErr: "invalid labels"},
		{name: "json array", raw: `["env"]`, wantErr: "invalid labels"},
		{name: "non string value", raw: `{"env":1}`, wantErr: "invalid labels"},
		{name: "empty key", raw: `{"":"prod"}`, wantErr: "invalid labels"},
		{name: "key has space", raw: `{"my env":"prod"}`, wantErr: "invalid labels"},
		{name: "value has slash", raw: `{"env":"prod/us"}`, wantErr: "invalid labels"},
		{name: "key too long", raw: `{"` + string(longKey) + `":"prod"}`, wantErr: "invalid labels"},
		{name: "value too long", raw: `{"env":"` + string(longValue) + `"}`, wantErr: "invalid labels"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeAgentLabels(tt.raw)
			if tt.wantErr != "" {
				if message := kerrors.FromError(err).Message; message != tt.wantErr {
					t.Fatalf("NormalizeAgentLabels(%q) error = %v, want message %q", tt.raw, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeAgentLabels(%q): %v", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeAgentLabels(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

// TestNormalizeAgentLabelsTooMany 单独确认超过 32 对时返回 too many labels。
func TestNormalizeAgentLabelsTooMany(t *testing.T) {
	pairs := make([]string, 0, maxAgentLabelPairs+1)
	for i := 0; i <= maxAgentLabelPairs; i++ {
		pairs = append(pairs, fmt.Sprintf(`"key-%d":"v"`, i))
	}
	_, err := NormalizeAgentLabels("{" + strings.Join(pairs, ",") + "}")
	if message := kerrors.FromError(err).Message; message != "too many labels" {
		t.Fatalf("error = %v, want too many labels", err)
	}
}

// TestCreateNormalizesLabels 确认创建时把原始 labels 规范化后交给仓储。
func TestCreateNormalizesLabels(t *testing.T) {
	repo := &stubAgentRepo{}
	registry := NewRegistry(repo, agenthub.NewHub(), testChecker(perm.AgentWrite, perm.MenuPermission("agent")))
	item, err := registry.Create(tenantContext(), bo.Agent{Name: "web-1", Labels: `{"role":"web","env":"prod"}`})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if repo.created.Labels != `{"env":"prod","role":"web"}` {
		t.Fatalf("persisted labels = %q", repo.created.Labels)
	}
	if item.Labels != repo.created.Labels {
		t.Fatalf("returned labels = %q", item.Labels)
	}
	if _, err := registry.Create(tenantContext(), bo.Agent{Name: "web-2", Labels: `not-json`}); !merr.IsInvalid(err) {
		t.Fatalf("invalid labels error = %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("create calls = %d, want 1", repo.createCalls)
	}
}

// TestUpdateNormalizesLabels 确认更新时同样校验并规范化 labels。
func TestUpdateNormalizesLabels(t *testing.T) {
	repo := &stubAgentRepo{}
	registry := NewRegistry(repo, agenthub.NewHub(), testChecker(perm.AgentWrite, perm.MenuPermission("agent")))
	if _, err := registry.Update(tenantContext(), bo.Agent{UID: "agent-1", Labels: ""}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if repo.updated.Labels != "{}" {
		t.Fatalf("update labels = %q, want {}", repo.updated.Labels)
	}
}

// TestMatchRequiresSSHExec 确认按标签匹配先校验 ssh:exec，不碰仓储。
func TestMatchRequiresSSHExec(t *testing.T) {
	repo := &stubAgentRepo{matched: []bo.Agent{{UID: "agent-1"}}}
	registry := NewRegistry(repo, agenthub.NewHub(), testChecker(perm.SSHRead, perm.MenuPermission("ssh_command")))
	_, err := registry.Match(tenantContext(), map[string]string{"env": "prod"})
	if kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("reason = %q, want %q (err = %v)", kerrors.Reason(err), merr.ReasonForbidden, err)
	}
	if repo.matchCalls != 0 {
		t.Fatalf("match calls = %d, want 0", repo.matchCalls)
	}
}

// TestMatchValidatesLabels 确认空标签是客户端错误，不是「匹配全部」。
func TestMatchValidatesLabels(t *testing.T) {
	repo := &stubAgentRepo{matched: []bo.Agent{{UID: "agent-1"}}}
	registry := NewRegistry(repo, agenthub.NewHub(), testChecker(perm.SSHExec, perm.MenuPermission("web_shell")))
	for _, labels := range []map[string]string{nil, {}} {
		_, err := registry.Match(tenantContext(), labels)
		if message := kerrors.FromError(err).Message; message != "labels are required" {
			t.Fatalf("match %v error = %v, want labels are required", labels, err)
		}
	}
	if repo.matchCalls != 0 {
		t.Fatalf("match calls = %d, want 0", repo.matchCalls)
	}
}

// TestMatchReturnsOnlineAgents 确认把租户、标签传给仓储，并用 Hub 覆盖在线状态。
func TestMatchReturnsOnlineAgents(t *testing.T) {
	hub := agenthub.NewHub()
	hub.Put(agenthub.NewSession("agent-1", "namespace-1"))
	repo := &stubAgentRepo{matched: []bo.Agent{{UID: "agent-1"}, {UID: "agent-2"}}}
	registry := NewRegistry(repo, hub, testChecker(perm.SSHExec, perm.MenuPermission("web_shell")))
	items, err := registry.Match(tenantContext(), map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if repo.matchNamespace != "namespace-1" || repo.matchLabels["env"] != "prod" {
		t.Fatalf("match args = %q, %v", repo.matchNamespace, repo.matchLabels)
	}
	if len(items) != 2 || !items[0].Online || items[1].Online {
		t.Fatalf("items = %+v", items)
	}
}

// stubAgentRepo 只实现本文件用到的仓储方法，其余方法由嵌入接口兜底。
type stubAgentRepo struct {
	repository.AgentRepo
	created        bo.Agent
	updated        bo.Agent
	createCalls    int
	updateCalls    int
	matched        []bo.Agent
	matchCalls     int
	matchNamespace string
	matchLabels    map[string]string
}

func (repo *stubAgentRepo) CreateAgent(_ context.Context, in bo.Agent, _ string) (*bo.Agent, error) {
	repo.createCalls++
	repo.created = in
	return &in, nil
}

func (repo *stubAgentRepo) UpdateAgent(_ context.Context, in bo.Agent) (*bo.Agent, error) {
	repo.updateCalls++
	repo.updated = in
	return &in, nil
}

func (repo *stubAgentRepo) MatchAgents(_ context.Context, namespaceUID string, labels map[string]string) ([]bo.Agent, error) {
	repo.matchCalls++
	repo.matchNamespace = namespaceUID
	repo.matchLabels = labels
	return repo.matched, nil
}

// 测试替身只实现权限检查会用到的方法，漏调权限检查而走到别的仓储方法时会 panic。
type stubUserRepo struct{ repository.UserRepo }

func (stubUserRepo) UserByUID(context.Context, string) (*bo.Account, error) {
	return &bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleUser}, nil
}

type stubNamespaceRepo struct{ repository.NamespaceRepo }

func (stubNamespaceRepo) GetNamespace(_ context.Context, uid string) (*bo.Namespace, error) {
	return &bo.Namespace{UID: uid, Status: "ACTIVE"}, nil
}

type stubRoleRepo struct {
	repository.RoleRepo
	granted map[string]struct{}
}

func (repo stubRoleRepo) Permissions(context.Context, string, string) (map[string]struct{}, error) {
	return repo.granted, nil
}

type stubFeatureRepo struct {
	repository.FeatureRepo
	capabilities []string
	modules      []string
}

func (repo stubFeatureRepo) EffectiveCapabilities(context.Context, string) ([]string, error) {
	return repo.capabilities, nil
}

func (repo stubFeatureRepo) EffectiveModules(context.Context, string) ([]string, error) {
	return repo.modules, nil
}

// testChecker 返回已开通全部能力、只持有指定权限的租户成员。
func testChecker(permissions ...string) access.Checker {
	granted := make(map[string]struct{}, len(permissions))
	for _, name := range permissions {
		granted[name] = struct{}{}
	}
	capabilities := make([]string, 0, len(perm.Capabilities()))
	for _, item := range perm.Capabilities() {
		capabilities = append(capabilities, item.Code)
	}
	modules := make([]string, 0, len(perm.Features()))
	for _, item := range perm.Features() {
		modules = append(modules, item.Code)
	}
	return access.Checker{
		Users:      stubUserRepo{},
		Namespaces: stubNamespaceRepo{},
		Roles:      stubRoleRepo{granted: granted},
		Features:   stubFeatureRepo{capabilities: capabilities, modules: modules},
		Log:        slog.New(slog.DiscardHandler),
	}
}

func tenantContext() context.Context {
	ctx := actor.WithUser(context.Background(), "user-1")
	return actor.WithNamespace(ctx, "namespace-1")
}
