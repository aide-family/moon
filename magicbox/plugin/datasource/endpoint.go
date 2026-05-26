package datasource

import "strings"

// NormalizeMetricEndpoint trims the endpoint and removes a trailing /api/v1 suffix.
// Prometheus-compatible clients append api/v1/... paths themselves; including /api/v1 in
// the base URL causes duplicated paths and HTTP 404.
func NormalizeMetricEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimRight(endpoint, "/")
	if strings.HasSuffix(strings.ToLower(endpoint), "/api/v1") {
		endpoint = endpoint[:len(endpoint)-len("/api/v1")]
		endpoint = strings.TrimRight(endpoint, "/")
	}
	return endpoint
}
