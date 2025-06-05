package bo

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/util/kv"
)

func TestScrapeTarget_Do(t *testing.T) {
	s := &ScrapeTarget{
		Target:      "localhost:9090",
		JobName:     "",
		Labels:      kv.StringMap{},
		Interval:    0,
		Timeout:     0,
		BasicAuth:   nil,
		TLS:         nil,
		Params:      nil,
		Headers:     nil,
		MetricsPath: "",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := s.Do(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(body))
}
