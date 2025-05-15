package httpx

import (
	"fmt"
	"net/url"
)

func ParseQuery(qr map[string]any) url.Values {
	if len(qr) == 0 {
		return nil
	}
	query := url.Values{}
	for k, v := range qr {
		query.Add(k, fmt.Sprintf("%v", v))
	}
	return query
}
