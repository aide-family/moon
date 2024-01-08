package bo

import (
	"testing"
)

func TestBuildApiDuration(t *testing.T) {
	t.Log(buildApiDuration("1m"))
	t.Log(buildApiDuration("1h"))
	t.Log(buildApiDuration("1d"))
	t.Log(buildApiDuration("d"))
	t.Log(buildApiDuration("1"))
}
