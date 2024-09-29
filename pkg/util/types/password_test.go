package types_test

import (
	"testing"

	"github.com/aide-family/moon/pkg/util/types"
)

func TestNewPassword(t *testing.T) {
	s := types.MD5("123456" + "3c4d9a0a5a703938dd1d2d46e1c924f9")
	t.Log(s)
	pVal := s
	enP := types.NewPassword(pVal)
	t.Log(enP.String())
	t.Log(enP.GetSalt())
}
