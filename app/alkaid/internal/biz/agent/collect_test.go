package agent

import (
	"os"
	"testing"
)

func TestFirstNTPLine(t *testing.T) {
	path := t.TempDir() + "/timesyncd.conf"
	body := "#NTP=\nNTP=ntp.aliyun.com ntp.tencent.com\nFallbackNTP=time.cloudflare.com\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := firstNTPLine(path, "NTP", "FallbackNTP"); got != "ntp.aliyun.com" {
		t.Fatalf("ntp = %q", got)
	}
}

func TestFormatBytes(t *testing.T) {
	if got := formatBytes(1536); got != "1.5KiB" {
		t.Fatalf("bytes = %q", got)
	}
}
