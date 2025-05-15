package do

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/kv"
)

type Realtime interface {
	GetID() uint32
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetTeamID() uint32
	GetStatus() vobj.AlertStatus
	GetFingerprint() string
	GetLabels() kv.StringMap
	GetSummary() string
	GetDescription() string
	GetValue() string
	GetGeneratorURL() string
	GetStartsAt() time.Time
	GetEndsAt() time.Time
}
