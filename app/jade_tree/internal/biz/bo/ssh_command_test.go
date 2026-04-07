package bo

import (
	"testing"

	"github.com/aide-family/magicbox/enum"
)

func TestNewListSSHCommandsBoDefaults(t *testing.T) {
	b := NewListSSHCommandsBo(0, 0, "k")
	if b.Page != 1 || b.PageSize != 20 {
		t.Fatalf("unexpected page defaults: page=%d size=%d", b.Page, b.PageSize)
	}
	if b.Keyword != "k" {
		t.Fatalf("keyword not preserved")
	}
}

func TestNewListSSHCommandAuditsBoDefaults(t *testing.T) {
	b := NewListSSHCommandAuditsBo(
		0,
		0,
		enum.SSHCommandAuditStatus_SSHCommandAuditStatus_PENDING,
		"  key  ",
		enum.SSHCommandAuditKind_SSHCommandAuditKind_UPDATE,
	)
	if b.Page != 1 || b.PageSize != 20 {
		t.Fatalf("unexpected page defaults: page=%d size=%d", b.Page, b.PageSize)
	}
	if b.StatusFilter != enum.SSHCommandAuditStatus_SSHCommandAuditStatus_PENDING {
		t.Fatalf("status filter not preserved")
	}
	if b.Keyword != "key" {
		t.Fatalf("keyword not normalized")
	}
	if b.Kind != enum.SSHCommandAuditKind_SSHCommandAuditKind_UPDATE {
		t.Fatalf("kind not preserved")
	}
}
