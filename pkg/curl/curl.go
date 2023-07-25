package curl

import (
	"context"
	"go.opentelemetry.io/otel"
	"os/exec"
)

func Curl(ctx context.Context, path string) (string, error) {
	_, span := otel.Tracer("pkg/curl").Start(ctx, "Curl")
	defer span.End()

	if len(path) == 0 {
		return "", nil
	}

	cmd := exec.Command("curl", "-X", "POST", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
