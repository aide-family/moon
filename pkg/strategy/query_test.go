package strategy

import (
	"testing"
	"time"
)

func TestParseQuery(t *testing.T) {
	p := ParseQuery(map[string]any{
		"expr": "up == 1",
		"time": time.Now().Unix(),
	})

	t.Log(p)
}
