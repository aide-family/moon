package datasource_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/vobj"
)

func TestEndpointDuration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	url := "http://www.baidu.com"
	res := datasource.EndpointDuration(ctx, url, vobj.HTTPMethodGet, nil, "", 5)
	for _, point := range res {
		bs, _ := json.Marshal(point)
		t.Log(string(bs))
	}

	url = "https://www.baidu.com"
	res = datasource.EndpointDuration(ctx, url, vobj.HTTPMethodGet, nil, "", 5)
	for _, point := range res {
		bs, _ := json.Marshal(point)
		t.Log(string(bs))
	}

	url = "https://www.baidu.com"
	res = datasource.EndpointDuration(ctx, url, vobj.HTTPMethodPost, nil, "", 5)
	for _, point := range res {
		bs, _ := json.Marshal(point)
		t.Log(string(bs))
	}
}
