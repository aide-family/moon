package datasource

import "context"

type (
	// EsResponse is the response of Elasticsearch query API.
	// {"_shards":{"failed":0,"skipped":0,"successful":3,"total":3},"hits":{"hits":[{"_id":"1","_index":"my_test","_score":1,"_source":{"created_at":"2025-02-20","description":"A comprehensive guide to Elasticsearch","price":49.99,"title":"Elasticsearch Guide"}}],"max_score":1,"total":{"relation":"eq","value":1}},"timed_out":false,"took":410}
	EsResponse struct {
		Hits EsHits `json:"hits"`
	}

	// EsHits is the hits of Elasticsearch query response.
	EsHits struct {
		Total EsHitTotal `json:"total"`
		Hits  []EsHit    `json:"hits"`
	}

	// EsHitTotal is the total of Elasticsearch query response.
	EsHitTotal struct {
		Value    int    `json:"value"`
		Relation string `json:"relation"`
	}

	// EsHit is the hit of Elasticsearch query response.
	EsHit struct {
		ID     string         `json:"_id"`
		Source map[string]any `json:"_source"`
	}

	// LogResponse is the response of log query API.
	LogResponse struct {
		Values        []string `json:"values"`
		DatasourceUrl string   `json:"datasourceUrl"`
		Timestamp     int64    `json:"timestamp"`
	}

	// LokiQueryResponse is the response of Loki query API.
	LokiQueryResponse struct {
		Status string    `json:"status"`
		Data   *LokiData `json:"data"`
	}
	// LokiData is the data of Loki query response.
	LokiData struct {
		ResultType string        `json:"status"`
		Result     []*LokiResult `json:"result"`
	}
	// LokiResult is the result of Loki query response.
	LokiResult struct {
		Stream map[string]interface{} `json:"stream"`
		Values []interface{}          `json:"values"`
	}

	// LogDatasource is the interface of log datasource.
	LogDatasource interface {
		// QueryLogs queries logs from datasource.
		QueryLogs(ctx context.Context, expr string, start, end int64) (*LogResponse, error)
		// Check checks
		Check(ctx context.Context) error
	}
)
