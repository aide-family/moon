package hello_test

import (
	"testing"

	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/hello"
)

func TestHello(t *testing.T) {
	hello.Hello()
	opts := []hello.Option{
		hello.WithMetadata(map[string]string{
			"summary": "test",
			"version": "v3.0.0",
		}),
		hello.WithEnv(config.Environment_TEST),
		hello.WithID("local.test"),
		hello.WithName("test"),
		hello.WithVersion("v3.0.0"),
	}
	hello.SetEnvWithOption(opts...)
	hello.Hello()
	t.Log()
}
