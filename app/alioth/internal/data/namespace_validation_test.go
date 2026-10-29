package data

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
)

// TestTextLimitsCountRunes 确认数据库写入前的长度校验按字符数而不是字节数。
// 停用原因、租户备注和角色备注都可能填中文或 emoji；用 MaxLen 按字节校验会让
// 500 字中文（约 1500 字节）在平台停用租户时报 "value is greater than the required length"。
func TestTextLimitsCountRunes(t *testing.T) {
	cases := []struct {
		name      string
		max       int
		validator func(string) error
	}{
		{name: "namespace disable reason", max: 500, validator: namespace.DisableReasonValidator},
		{name: "role remark", max: 5000, validator: role.RemarkValidator},
		{name: "feature package remark", max: 5000, validator: featurepackage.RemarkValidator},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.validator == nil {
				t.Fatal("validator is not initialized")
			}
			// 每个字符 3 字节，按字节校验会远超上限，按字符校验刚好等于上限。
			limit := strings.Repeat("停", tc.max)
			if err := tc.validator(limit); err != nil {
				t.Fatalf("validator rejected %d runes: %v", tc.max, err)
			}
			if err := tc.validator(limit + "停"); err == nil {
				t.Fatalf("validator accepted %d runes, want rejection", tc.max+1)
			}
		})
	}
}
