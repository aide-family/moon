package bo

import (
	"testing"
)

func TestBuildApiDuration(t *testing.T) {
	t.Log(BuildApiDuration("1m"))
	t.Log(BuildApiDuration("1h"))
	t.Log(BuildApiDuration("1d"))
	t.Log(BuildApiDuration("d"))
	t.Log(BuildApiDuration("1"))
}
