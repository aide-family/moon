package probe

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/aide-family/moon/app/megrez/pkg/bo"
	probesdk "github.com/aide-family/moon/pkg/probe"
)

// runProbe 是执行一次探测的入口，默认走 pkg/probe。单测把它替换成假实现，
// 才能覆盖「执行器缺权限」分支，而不用要求 root 或 raw socket。
var runProbe = probesdk.Run

// execute 按规则驱动执行一次探测（HTTP / TCP / ICMP / DNS / TLS），并把结果转成
// 回写结构。观测值按规则选择的指标取，判定仍然发生在玉衡：执行器只负责「测」，
// 不解析级别条件。
//
// 第二个返回值表示结果是否可以回写。ICMP 需要非特权 ping socket 或 NET_RAW，
// 缺权限属于执行器自身故障：只记日志、不回写，避免把权限问题误判成目标不可用
// 而触发误告警。
func execute(ctx context.Context, logger *slog.Logger, driver string, group bo.ProbeGroup, target bo.ProbeTarget, metric string, maxTimeoutMS int) (bo.ProbeResult, bool) {
	request := probesdk.Request{
		Method:          group.Request.Method,
		ExpectStatus:    group.Request.ExpectStatus,
		TimeoutMS:       group.Request.TimeoutMS,
		FollowRedirects: group.Request.FollowRedirects,
		SkipTLSVerify:   group.Request.SkipTLSVerify,
		VerifyCert:      group.Request.VerifyCert,
		Headers:         group.Request.Headers,
		BodyKeyword:     group.Request.BodyKeyword,
		ExpectHeaders:   group.Request.ExpectHeaders,
		BodyRegex:       group.Request.BodyRegex,
		ICMPCount:       group.Request.ICMPCount,
		ICMPPacketSize:  group.Request.ICMPPacketSize,
		DNSRecordType:   group.Request.DNSRecordType,
		DNSExpectValues: group.Request.DNSExpectValues,
		DNSServer:       group.Request.DNSServer,
		TLSServerName:   group.Request.TLSServerName,
	}.Normalize()

	timeout := time.Duration(request.TimeoutMS) * time.Millisecond
	if maxTimeoutMS > 0 && timeout > time.Duration(maxTimeoutMS)*time.Millisecond {
		timeout = time.Duration(maxTimeoutMS) * time.Millisecond
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	outcome := runProbe(runCtx, driver, target.Spec, request)

	// 权限不足是执行器环境问题（容器没给 NET_RAW 或内核没开非特权 ping），
	// 不是目标故障，回写会制造假失败。
	if outcome.ErrorKind == probesdk.ErrorPermission {
		if logger != nil {
			logger.Warn("probe permission denied, result is not reported; ICMP needs an unprivileged ping socket (net.ipv4.ping_group_range) or CAP_NET_RAW",
				"driver", driver, "target_uid", target.UID, "target", target.Spec, "error", outcome.ErrorMessage)
		}
		return bo.ProbeResult{}, false
	}

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
		PacketLoss:   outcome.PacketLoss,
		JitterMS:     int(outcome.JitterMS),
		MinRTTMS:     int(outcome.MinRTTMS),
		MaxRTTMS:     int(outcome.MaxRTTMS),
		AnswerCount:  outcome.AnswerCount,
		DNSAnswers:   outcome.DNSAnswers,
		TLSVersion:   outcome.TLSVersion,
		TLSCipher:    outcome.TLSCipher,
		CertSubject:  outcome.CertSubject,
		CertIssuer:   outcome.CertIssuer,
		CertSANs:     outcome.CertSANs,
	}, true
}
