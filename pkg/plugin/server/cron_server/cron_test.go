package cron_server_test

import (
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/robfig/cron/v3"
)

func Test_CronSpecCustom(t *testing.T) {
	spec1 := cron_server.CronSpecCustom("*", "*", "*", "*", "*", "*")
	spec2 := cron_server.CronSpecCustom("0", "0", "*", "1", "*", "*")
	spec3 := cron_server.CronSpecCustom("0", "0", "*", "1", "1", "*")
	spec4 := cron_server.CronSpecCustom("0", "0", "*", "1", "1", "*")
	c := cron.New(cron.WithSeconds())
	_, err := c.AddFunc(string(spec1), func() {
		t.Logf("spec1: %s, ts: %v", spec1, timex.Now())
	})
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	_, err = c.AddFunc(string(spec2), func() {
		t.Logf("spec2: %s, ts: %v", spec2, timex.Now())
	})
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	_, err = c.AddFunc(string(spec3), func() {
		t.Logf("spec3: %s, ts: %v", spec3, timex.Now())
	})
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	_, err = c.AddFunc(string(spec4), func() {
		t.Logf("spec4: %s, ts: %v", spec4, timex.Now())
	})
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	c.Start()
	time.Sleep(time.Second * 10)
	c.Stop()
}
