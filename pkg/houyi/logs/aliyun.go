package logs

import (
	"context"
	"fmt"
	"time"

	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/mlog"
	"github.com/aide-family/moon/pkg/util/types"

	sls "github.com/aliyun/aliyun-log-go-sdk"
)

type (
	AliYunLog struct {
		c mlog.AliyunLogConfig
		// endpoint 数据源地址
		endpoint string
	}

	// AliYunLogOption is a functional option for AliYunLogOption.
	AliYunLogOption func(l *AliYunLog)
)

func (l *AliYunLog) Check(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

// NewAliYunLog 创建阿里云日志对象
func NewAliYunLog(c mlog.AliyunLogConfig, opt ...AliYunLogOption) *AliYunLog {
	a := &AliYunLog{
		c: c,
	}

	for _, o := range opt {
		o(a)
	}

	return a
}

// WithAliYunEndpoint sets the aliYun endpoint.
func WithAliYunEndpoint(endpoint string) AliYunLogOption {
	return func(l *AliYunLog) {
		l.endpoint = endpoint
	}
}

func (l *AliYunLog) QueryLogs(_ context.Context, expr string, start, end int64) (*datasource.LogResponse, error) {

	opts := l.c
	provider := sls.NewStaticCredentialsProvider(opts.GetAccessKey(), opts.GetAccessSecret(), opts.GetSecurityToken())
	client := sls.CreateNormalInterfaceV2(opts.GetEndpoint(), provider)

	// Client 是否成功
	_, err := client.ListProject()
	if err != nil {
		return nil, merr.ErrorNotificationSystemError(" aliYun log client init failed: %v", err.Error())
	}

	// 每次查询的最大日志条数
	maxLineNum := int64(100)
	// 查询的偏移量
	offset := int64(0)

	logsResponse, err := client.GetLogs(opts.GetProject(), opts.GetStore(), "", start, end, expr, maxLineNum, offset, true)
	if types.IsNotNil(err) {
		return nil, err
	}

	result := make([]string, 0, len(logsResponse.Logs))

	for _, logGroup := range logsResponse.Logs {
		for key, value := range logGroup {
			logValue := fmt.Sprintf("%s:%s", key, value)
			result = append(result, logValue)
		}
	}

	return &datasource.LogResponse{
		Values:        result,
		DatasourceUrl: l.endpoint,
		Timestamp:     time.Now().Unix(),
	}, nil
}
