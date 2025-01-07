package label

import (
	"testing"
)

func TestLabels_String(t *testing.T) {
	label := NewLabels(map[string]string{
		"__moon__domain__":      "test",
		"__moon__domain_port__": "8080",
		"a":                     "a",
	})
	t.Log(label.String())
}

func TestLabels_Index(t *testing.T) {
	label := NewLabels(map[string]string{
		"__moon__domain__":      "test",
		"__moon__domain_port__": "8080",
		"a":                     "a",
	})
	t.Log(label.Index())
}

func TestLabels_Value(t *testing.T) {
	label := NewLabels(map[string]string{
		"__moon__domain__":      "test",
		"__moon__domain_port__": "8080",
		"a":                     "a",
	})
	v, _ := label.Value()
	t.Log(string(v.([]byte)))
}
