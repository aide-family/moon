package types_test

import (
	"testing"

	"github.com/aide-family/moon/pkg/util/types"
)

func TestNewPassword(t *testing.T) {
	pVal := "123456"
	enP := types.NewPassword(pVal)
	t.Log(enP.String())
	t.Log(enP.GetSalt())
}
