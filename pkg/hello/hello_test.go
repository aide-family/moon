package hello_test

import (
	"testing"

	"github.com/moon-monitor/moon/pkg/config"
	"github.com/moon-monitor/moon/pkg/hello"
)

func TestHello(t *testing.T) {
	hello.Hello()
	opts := []hello.Option{
		hello.WithMetadata(map[string]string{
			"summary": "test",
			"version": "v1.0.0",
		}),
		hello.WithEnv(config.Environment_TEST),
		hello.WithID("local.test"),
		hello.WithName("test"),
		hello.WithVersion("v1.0.0"),
	}
	hello.SetEnvWithOption(opts...)
	hello.Hello()
	t.Log()
}
