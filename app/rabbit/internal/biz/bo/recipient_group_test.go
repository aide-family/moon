package bo

import (
	"testing"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"
)

func TestRecipientGroupItemBo_ToAPIV1RecipientGroupItem_timestamps(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 3, 4, 5, 6, 0, time.UTC)
	item := &RecipientGroupItemBo{
		UID:       snowflake.ID(1),
		Name:      "ops",
		Status:    enum.GlobalStatus_ENABLED,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	got := item.ToAPIV1RecipientGroupItem()
	if got.GetCreatedAt() != timex.FormatTime(&createdAt) {
		t.Fatalf("createdAt = %q, want %q", got.GetCreatedAt(), timex.FormatTime(&createdAt))
	}
	if got.GetUpdatedAt() != timex.FormatTime(&updatedAt) {
		t.Fatalf("updatedAt = %q, want %q", got.GetUpdatedAt(), timex.FormatTime(&updatedAt))
	}
}
