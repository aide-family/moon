package strategy

import (
	"testing"
	"time"
)

func TestParseQuery(t *testing.T) {
	p, err := ParseQuery(map[string]any{
		"expr": "up == 1",
		"time": time.Now().Unix(),
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(p)
}
