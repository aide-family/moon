package logs

import (
	"strings"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	logElasticsearch  = "elasticsearch"
	logLoki           = "loki"
	logAliAliCloudSLS = "aliYunSls"
)

// NewLogQuery creates a new log query based on the configuration.
func NewLogQuery(c *conf.LogQuery) (datasource.LogDatasource, error) {
	switch strings.ToLower(c.GetType()) {
	case logElasticsearch:
		es := c.GetEs()
		return NewElasticsearch(es, WithEsEndpoint(es.GetEndpoint()))
	case logLoki:
		loki := c.GetLoki()
		return NewLokiDatasource(WithLokiEndpoint(loki.GetEndpoint()),
			WithLokiBasicAuth(loki.GetUsername(), loki.GetPassword()),
			WithLokiLimit(loki.GetLimit())), nil
	case logAliAliCloudSLS:
		aliYun := c.GetAliYun()
		return NewAliYunLog(aliYun, WithAliYunEndpoint(aliYun.GetEndpoint())), nil
	default:
		return nil, merr.ErrorNotificationSystemError("不支持的日志查询类型:%s", c.GetType())
	}
}
