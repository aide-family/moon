package data

import (
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

// TestLegacyRoleK8sPermissions 确认自定义角色迁移按旧 cluster 权限补齐：
// cluster:read 补全部 k8s 读和 Pod 日志，cluster:write 补全部 k8s 写；
// 已经有任何 k8s:* 权限的角色不再补，保持幂等。
func TestLegacyRoleK8sPermissions(t *testing.T) {
	cases := []struct {
		name  string
		names []string
		want  []string
	}{
		{name: "no legacy cluster permissions"},
		{name: "unrelated permissions", names: []string{perm.MemberRead, perm.RoleRead}},
		{name: "cluster read", names: []string{perm.ClusterRead}, want: perm.K8sReadPermissions()},
		{name: "cluster write", names: []string{perm.ClusterWrite}, want: perm.K8sWritePermissions()},
		{
			name:  "cluster read and write",
			names: []string{perm.ClusterRead, perm.ClusterWrite},
			want:  append(perm.K8sReadPermissions(), perm.K8sWritePermissions()...),
		},
		{name: "already migrated", names: []string{perm.ClusterRead, perm.ClusterWrite, perm.K8sPodRead}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := legacyRoleK8sPermissions(permissionSet(tc.names))
			if len(got) != len(tc.want) {
				t.Fatalf("permissions = %v, want %v", got, tc.want)
			}
			want := permissionSet(tc.want)
			for _, code := range got {
				if _, ok := want[code]; !ok {
					t.Fatalf("permission %s not expected, want %v", code, tc.want)
				}
			}
		})
	}
}

func permissionSet(names []string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, name := range names {
		out[name] = struct{}{}
	}
	return out
}

// TestCapabilityCodesOfExpandsLegacyChannelCodes 确认租户套餐快照和旧数据里的
// 历史渠道编码在读取时展开成当前能力：旧 Webhook 渠道覆盖四个应用能力，
// 旧短信渠道覆盖两个提供商能力，目录外的编码继续被忽略。
func TestCapabilityCodesOfExpandsLegacyChannelCodes(t *testing.T) {
	got := capabilityCodesOf([]string{
		perm.CapabilityNotifyChannelSMTP,
		"notify_channel_webhook",
		"notify_channel_sms",
		perm.CapabilityMember,
		"unknown",
	})
	want := []string{
		perm.CapabilityNotifyChannelSMTP,
		perm.CapabilityNotifyChannelWebhookOther,
		perm.CapabilityNotifyChannelWebhookDingTalk,
		perm.CapabilityNotifyChannelWebhookWeChat,
		perm.CapabilityNotifyChannelWebhookFeishu,
		perm.CapabilityNotifyChannelSMSAliCloud,
		perm.CapabilityNotifyChannelSMSTencent,
		perm.CapabilityMember,
	}
	if len(got) != len(want) {
		t.Fatalf("capability codes = %v, want %v", got, want)
	}
	wantSet := permissionSet(want)
	for _, code := range got {
		if _, ok := wantSet[code]; !ok {
			t.Fatalf("unexpected capability %s, want %v", code, want)
		}
	}
	base := permissionSet(perm.BaseCapabilities())
	for _, code := range got {
		if code == perm.CapabilityMember {
			continue
		}
		if _, ok := base[code]; ok {
			continue
		}
		t.Fatalf("base package is missing channel capability %s", code)
	}
}

// TestCapabilityCodesOf 确认旧的模块级编码会展开成该模块的全部能力，
// 能力编码去重并按目录顺序返回，目录外的编码被忽略。
func TestCapabilityCodesOf(t *testing.T) {
	got := capabilityCodesOf([]string{
		perm.FeatureCluster,
		perm.CapabilityAlertRealtime,
		perm.CapabilityAlertRealtime,
		"unknown",
		perm.FeatureMember,
	})
	want := append(
		[]string{perm.CapabilityAlertRealtime, perm.CapabilityMember},
		perm.CapabilityCodesOfModule(perm.FeatureCluster)...,
	)
	if len(got) != len(want) {
		t.Fatalf("capability codes = %v, want %v", got, want)
	}
	wantSet := permissionSet(want)
	for _, code := range got {
		if _, ok := wantSet[code]; !ok {
			t.Fatalf("unexpected capability %s, want %v", code, want)
		}
	}
}
