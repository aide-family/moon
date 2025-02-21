package logs

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// NewTestClient 创建一个用于测试的 Elasticsearch 客户端
func NewTestClient(t *testing.T) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{"https://localhost:9200"},
		Username:  "elastic",
		Password:  "x4qlYQdXKkjWjDn+16fR",
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	client, err := elasticsearch.NewClient(cfg)
	if types.IsNotNil(err) {
		return nil, err
	}

	return client, nil
}

// 测试 NewElasticsearch 函数
func TestNewElasticsearch(t *testing.T) {
	t.Run("TestNewElasticsearch", func(t *testing.T) {
		conf := &conf.Elasticsearch{
			Endpoint:    "https://localhost:9200",
			Username:    "elastic",
			Password:    "x4qlYQdXKkjWjDn+16fR",
			SearchIndex: "my_index",
		}
		es, err := NewElasticsearch(conf)
		if err != nil {
			t.Error(err)
		}
		if es.searchIndex != "my_index" {
			t.Error("searchIndex is not equal to test_index")
		}

		res, err := es.QueryLogs(context.TODO(), `{ "query": { "match_all": {} } }`, 1633072800000, 1633072800000)
		if err != nil {
			return
		}
		t.Log("res ", res)

	})
}

// TestElasticsearch_Create_Index 测试创建索引
func TestElasticsearch_Create_Index(t *testing.T) {
	client, err := NewTestClient(t)

	if types.IsNotNil(err) {
		t.Error("NewClient failed:", err)
		return
	}

	indexConfig := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   3,
			"number_of_replicas": 2,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
				},
				"description": map[string]interface{}{
					"type": "text",
				},
				"price": map[string]interface{}{
					"type": "float",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	// 将模板转换为 JSON
	templateJSON, err := json.Marshal(indexConfig)
	if err != nil {
		t.Fatalf("Error marshaling template to JSON: %s", err)
	}

	// 创建索引模板请求
	req := esapi.IndicesCreateRequest{
		Index: "my_test",                     // 模板名称
		Body:  bytes.NewReader(templateJSON), // 使用 bytes.NewReader 转换 JSON
	}
	// 发送请求
	res, err := req.Do(context.Background(), client)
	if err != nil {
		t.Fatalf("Error sending request to Elasticsearch: %s", err)
	}
	defer res.Body.Close()
	// 检查响应
	if res.IsError() {
		t.Fatalf("Error response from Elasticsearch: %s", res.String())
	}
}

func TestElasticsearch_Add_Document(t *testing.T) {
	client, err := NewTestClient(t)
	if types.IsNotNil(err) {
		t.Error("NewClient failed:", err)
	}

	// 定义要插入的文档数据
	doc := map[string]interface{}{
		"title":       "Elasticsearch Guide",
		"description": "A comprehensive guide to Elasticsearch",
		"price":       49.99,
		"created_at":  "2025-02-20",
	}

	// 将文档数据转换为 JSON
	docJSON, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Error marshaling document to JSON: %s", err)
	}

	// 创建索引文档请求
	req := esapi.IndexRequest{
		Index:      "my_test",                // 索引名称
		DocumentID: "1",                      // 文档 ID（可选，如果不指定，Elasticsearch 会自动生成）
		Body:       bytes.NewReader(docJSON), // 文档数据
		Refresh:    "true",                   // 是否立即刷新索引
	}

	// 发送请求
	res, err := req.Do(context.Background(), client)
	if err != nil {
		t.Fatalf("Error sending request to Elasticsearch: %s", err)
	}
	defer res.Body.Close()

	// 检查响应
	if res.IsError() {
		t.Fatalf("Error response from Elasticsearch: %s", res.String())
	}
}

func TestElasticsearch_Query_Document(t *testing.T) {
	client, err := NewTestClient(t)
	if types.IsNotNil(err) {
		t.Error("NewClient failed:", err)
	}

	query := `{ "query": { "match_all": {} } }`
	response, err := client.Search(
		client.Search.WithIndex("my_test"),
		client.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		return
	}

	var r map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&r); err != nil {
		fmt.Printf("Error parsing the response body: %s\n", err)
		return
	}

	searchJson, err := json.Marshal(r)
	if err != nil {
		return
	}
	t.Log("Query document response:", response.StatusCode)
	t.Logf("Query result: %s\n", string(searchJson))
	// {"_shards":{"failed":0,"skipped":0,"successful":3,"total":3},"hits":{"hits":[{"_id":"1","_index":"my_test","_score":1,"_source":{"created_at":"2025-02-20","description":"A comprehensive guide to Elasticsearch","price":49.99,"title":"Elasticsearch Guide"}}],"max_score":1,"total":{"relation":"eq","value":1}},"timed_out":false,"took":410}

}
