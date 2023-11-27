package helper

import (
	"testing"
)

func TestIssueToken(t *testing.T) {
	t.Log(IssueToken(1, "admin"))
}
