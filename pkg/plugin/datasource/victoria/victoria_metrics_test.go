package victoria_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	transporthttp "github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/plugin/datasource"
	"github.com/aide-family/moon/pkg/plugin/datasource/victoria"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/timex"
)

type config struct {
	endpoint       string
	basicAuth      datasource.BasicAuth
	tls            datasource.TLS
	scrapeInterval time.Duration
}

func (c *config) GetEndpoint() string {
	return c.endpoint
}

func (c *config) GetHeaders() []*kv.KV {
	return nil
}

func (c *config) GetMethod() common.DatasourceQueryMethod {
	return common.DatasourceQueryMethod_GET
}

func (c *config) GetBasicAuth() datasource.BasicAuth {
	return nil
}

func (c *config) GetTLS() datasource.TLS {
	return nil
}

func (c *config) GetCA() string {
	return ""
}

func (c *config) GetScrapeInterval() time.Duration {
	return 15 * time.Second
}

type Query struct {
	Target string `json:"target,omitempty"`
}

func TestVictoria_Proxy(t *testing.T) {
	c := &config{
		endpoint: "http://127.0.0.1:8428",
	}
	instance := victoria.New(c, log.GetLogger())
	opts := []transporthttp.ServerOption{
		transporthttp.Address(":8080"),
	}
	srv := transporthttp.NewServer(opts...)
	r := srv.Route("/")
	r.GET("/proxy/{target:[^/]+(?:/[^?]*)}", func(c transporthttp.Context) error {
		var in Query
		if err := c.BindVars(&in); err != nil {
			return err
		}
		log.Infof("proxy target: %s", in.Target)
		return instance.Proxy(c, in.Target)
	})

	app := kratos.New(kratos.Server(srv))
	if err := app.Run(); err != nil {
		return
	}
}

func TestVictoria_MetadataNames(t *testing.T) {
	c := &config{
		endpoint: "http://127.0.0.1:8428",
	}
	instance := victoria.New(c, log.GetLogger())
	names, err := instance.Metadata(context.Background())
	if err != nil {
		return
	}
	for metricMetadata := range names {
		t.Logf("%s", metricMetadata)
	}
}

func TestVictoria_QueryRange(t *testing.T) {
	c := &config{
		endpoint: "http://127.0.0.1:8428",
	}
	instance := victoria.New(c, log.GetLogger())
	query := &datasource.MetricQueryRequest{
		Expr: "flag",
	}
	query.Time = timex.Now().Unix()
	query.Step = 14
	query.StartTime = timex.Now().Add(-time.Hour * 12).Unix()
	query.EndTime = timex.Now().Unix()
	response, err := instance.Query(context.Background(), query)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	t.Logf("%s", response.String())
}

func TestVictoria_Query(t *testing.T) {
	c := &config{
		endpoint: "http://127.0.0.1:8428",
	}
	instance := victoria.New(c, log.GetLogger())
	query := &datasource.MetricQueryRequest{
		Expr: "flag",
	}
	query.Time = timex.Now().Unix()
	response, err := instance.Query(context.Background(), query)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	t.Logf("%s", response.String())
}
