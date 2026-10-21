package notify

import (
	"context"
	"testing"

	"github.com/aide-family/Alioth/app/merak/pkg/bo"
)

func TestDeliverReturnsRenderedContentForAudit(t *testing.T) {
	result := Deliver(context.Background(), bo.TaskSnapshot{
		ChannelType:     "UNSUPPORTED",
		ChannelApp:      "OTHER",
		TemplateType:    "EMAIL",
		TemplateSubject: "Alert: {{.name}}",
		TemplateBody:    "Host {{.name}} is {{.state}}",
		Params:          `{"name":"node-1","state":"down"}`,
	})
	if result.OK {
		t.Fatal("unsupported channel unexpectedly delivered")
	}
	if result.Subject != "Alert: node-1" || result.Content != "Host node-1 is down" {
		t.Fatalf("rendered delivery content = subject %q, content %q", result.Subject, result.Content)
	}
	if result.Detail == "" {
		t.Fatal("failed delivery has no detail")
	}
}
