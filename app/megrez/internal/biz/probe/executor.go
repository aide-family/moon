package probe

import (
	"context"
	"strconv"
	"time"

	"github.com/aide-family/moon/app/megrez/pkg/bo"
	probesdk "github.com/aide-family/moon/pkg/probe"
)

// execute 执行一次 HTTP 探测，并把结果转成回写结构。观测值按规则选择的指标取，
// 判定仍然发生在玉衡：执行器只负责「测」，不解析级别条件。
func execute(ctx context.Context, group bo.ProbeGroup, target bo.ProbeTarget, metric string) bo.ProbeResult {
	request := probesdk.Request{
		Method:          group.Request.Method,
		ExpectStatus:    group.Request.ExpectStatus,
		TimeoutMS:       group.Request.TimeoutMS,
		FollowRedirects: group.Request.FollowRedirects,
		SkipTLSVerify:   group.Request.SkipTLSVerify,
		VerifyCert:      group.Request.VerifyCert,
		Headers:         group.Request.Headers,
		BodyKeyword:     group.Request.BodyKeyword,
	}.Normalize()

	timeout := time.Duration(request.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	outcome := probesdk.Run(runCtx, target.Spec, request)

	value := outcome.Value(metric)
	return bo.ProbeResult{
		TargetUID:    target.UID,
		OK:           outcome.OK,
		Value:        strconv.FormatFloat(value, 'f', -1, 64),
		DurationMS:   int(outcome.DurationMS),
		StatusCode:   outcome.StatusCode,
		ErrorKind:    outcome.ErrorKind,
		ErrorMessage: outcome.ErrorMessage,
		DNSMS:        int(outcome.DNSMS),
		TCPMS:        int(outcome.TCPMS),
		TLSMS:        int(outcome.TLSMS),
		TTFBMS:       int(outcome.TTFBMS),
		CertDays:     outcome.CertDays,
		ResolvedIP:   outcome.ResolvedIP,
	}
}
