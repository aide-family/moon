package label_test

import (
	"testing"

	"github.com/moon-monitor/moon/pkg/util/kv/label"
)

func Test_NewAnnotation(t *testing.T) {
	annotation := label.NewAnnotation("summary", "description")
	summary := annotation.GetSummary()
	if summary != "summary" {
		t.Errorf("NewAnnotation() = %v, want %v", summary, "summary")
	}
	description := annotation.GetDescription()
	if description != "description" {
		t.Errorf("NewAnnotation() = %v, want %v", description, "description")
	}
	annotation.SetSummary("summary2")
	summary2 := annotation.GetSummary()
	if summary2 != "summary2" {
		t.Errorf("NewAnnotation() = %v, want %v", summary2, "summary2")
	}
	annotation.SetDescription("description2")
	description2 := annotation.GetDescription()
	if description2 != "description2" {
		t.Errorf("NewAnnotation() = %v, want %v", description2, "description2")
	}
	marshalBinary, err := annotation.MarshalBinary()
	if err != nil {
		t.Errorf("NewAnnotation() = %v, want %v", err, nil)
	}
	t.Logf("NewAnnotation() = %v", string(marshalBinary))
	if err := annotation.UnmarshalBinary([]byte(`{"summary":"summary3","description":"description3"}`)); err != nil {
		t.Errorf("NewAnnotation() = %v, want %v", err, nil)
	}
	summary3 := annotation.GetSummary()
	if summary3 != "summary3" {
		t.Errorf("NewAnnotation() = %v, want %v", summary3, "summary3")
	}
	description3 := annotation.GetDescription()
	if description3 != "description3" {
		t.Errorf("NewAnnotation() = %v, want %v", description3, "description3")
	}
	t.Logf("NewAnnotation() = %v", annotation)
}
