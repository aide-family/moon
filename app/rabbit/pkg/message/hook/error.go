// Package hook is a simple package that provides a hook.
package hook

import (
	"fmt"
	"io"
	"net/http"
)

func RequestAssert(resp *http.Response, unmarshalResponse func(body io.ReadCloser) error) error {
	if resp.StatusCode == http.StatusOK {
		return unmarshalResponse(resp.Body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(body))
}
