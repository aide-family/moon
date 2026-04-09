// Package collector provides business collectors for metrics exposition.
package collector

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/internal/conf"
)

const (
	metricNamespace = "probe"

	probeTypeTCP  = "tcp"
	probeTypePort = "port"
	probeTypeHTTP = "http"
	probeTypeCert = "cert"
	probeTypePing = "ping"
)

var (
	tcpSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypeTCP, "success"),
		"Whether the last TCP probe to the target succeeded (1) or not (0).",
		[]string{"target", "initiator"}, nil,
	)
	tcpDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypeTCP, "duration_seconds"),
		"Duration of the last TCP probe in seconds.",
		[]string{"target", "initiator"}, nil,
	)
	portOpenDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypePort, "open"),
		"Whether the target port is open (1) or closed/unreachable (0).",
		[]string{"target", "initiator"}, nil,
	)
	portDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypePort, "duration_seconds"),
		"Duration of the last port probe in seconds.",
		[]string{"target", "initiator"}, nil,
	)
	httpSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypeHTTP, "success"),
		"Whether the last HTTP probe to the target succeeded (1) or not (0).",
		[]string{"target", "initiator"}, nil,
	)
	httpDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypeHTTP, "duration_seconds"),
		"Duration of the last HTTP probe in seconds.",
		[]string{"target", "initiator"}, nil,
	)
	httpStatusCodeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypeHTTP, "status_code"),
		"HTTP response status code of the last probe. 0 means request failed before receiving response.",
		[]string{"target", "initiator"}, nil,
	)
	certSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypeCert, "success"),
		"Whether the last TLS certificate probe succeeded (1) or not (0).",
		[]string{"target", "initiator"}, nil,
	)
	certNotAfterDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypeCert, "not_after_timestamp_seconds"),
		"TLS certificate expiration timestamp (Unix seconds).",
		[]string{"target", "initiator"}, nil,
	)
	certRemainingDaysDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypeCert, "remaining_days"),
		"Remaining days until TLS certificate expiration.",
		[]string{"target", "initiator"}, nil,
	)
	pingSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypePing, "success"),
		"Whether the last ping probe to the target succeeded (1) or not (0).",
		[]string{"target", "initiator"}, nil,
	)
	pingDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricNamespace, probeTypePing, "duration_seconds"),
		"Duration of the last ping probe in seconds.",
		[]string{"target", "initiator"}, nil,
	)
)

type target struct {
	Type    string
	Host    string
	Port    string
	URL     string
	Name    string
	Timeout time.Duration
}

type ProbeCollector struct {
	repo           repository.ProbeTask
	defaultTimeout time.Duration
	maxConcurrent  int
	initiator      string
	configTargets  []target
}

func NewProbeCollector(cfg *conf.Bootstrap, repo repository.ProbeTask) *ProbeCollector {
	probe := cfg.GetProbe()
	timeout := 5 * time.Second
	if probe != nil && probe.GetTimeout() != nil && probe.GetTimeout().AsDuration() > 0 {
		timeout = probe.GetTimeout().AsDuration()
	}
	concurrency := 32
	if probe != nil && probe.GetConcurrency() > 0 {
		concurrency = int(probe.GetConcurrency())
	}
	host := ""
	ip := ""
	if probe != nil && probe.GetInitiator() != nil {
		host = strings.TrimSpace(probe.GetInitiator().GetHostname())
		ip = strings.TrimSpace(probe.GetInitiator().GetIp())
	}
	if host == "" {
		h, _ := os.Hostname()
		host = h
	}
	if ip == "" {
		ip = discoverLocalIP()
	}
	initiator := host + "/" + ip
	return &ProbeCollector{
		repo:           repo,
		defaultTimeout: timeout,
		maxConcurrent:  concurrency,
		initiator:      initiator,
		configTargets:  normalizeConfigTargets(probe, timeout),
	}
}

func (c *ProbeCollector) Enabled(cfg *conf.Bootstrap) bool {
	return cfg.GetProbe() != nil && strings.EqualFold(cfg.GetProbe().GetEnabled(), "true")
}

func (c *ProbeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- tcpSuccessDesc
	ch <- tcpDurationDesc
	ch <- portOpenDesc
	ch <- portDurationDesc
	ch <- httpSuccessDesc
	ch <- httpDurationDesc
	ch <- httpStatusCodeDesc
	ch <- certSuccessDesc
	ch <- certNotAfterDesc
	ch <- certRemainingDaysDesc
	ch <- pingSuccessDesc
	ch <- pingDurationDesc
}

func (c *ProbeCollector) Collect(ch chan<- prometheus.Metric) {
	targets := c.loadTargets(context.Background())
	if len(targets) == 0 {
		return
	}
	sem := make(chan struct{}, c.maxConcurrent)
	var wg sync.WaitGroup
	for _, t := range targets {
		target := t
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			c.collectTarget(ch, target)
		})
	}
	wg.Wait()
}

func (c *ProbeCollector) collectTarget(ch chan<- prometheus.Metric, target target) {
	switch target.Type {
	case probeTypeTCP:
		c.collectTCP(ch, target)
	case probeTypePort:
		c.collectPort(ch, target)
	case probeTypeHTTP:
		c.collectHTTP(ch, target)
	case probeTypeCert:
		c.collectCert(ch, target)
	case probeTypePing:
		c.collectPing(ch, target)
	}
}

func (c *ProbeCollector) collectTCP(ch chan<- prometheus.Metric, target target) {
	start := time.Now()
	ok := tcpProbe(target.Host, target.Port, target.Timeout)
	success := 0.0
	if ok {
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(tcpSuccessDesc, prometheus.GaugeValue, success, target.Name, c.initiator)
	ch <- prometheus.MustNewConstMetric(tcpDurationDesc, prometheus.GaugeValue, time.Since(start).Seconds(), target.Name, c.initiator)
}

func (c *ProbeCollector) collectPort(ch chan<- prometheus.Metric, target target) {
	start := time.Now()
	ok := tcpProbe(target.Host, target.Port, target.Timeout)
	open := 0.0
	if ok {
		open = 1
	}
	ch <- prometheus.MustNewConstMetric(portOpenDesc, prometheus.GaugeValue, open, target.Name, c.initiator)
	ch <- prometheus.MustNewConstMetric(portDurationDesc, prometheus.GaugeValue, time.Since(start).Seconds(), target.Name, c.initiator)
}

func (c *ProbeCollector) collectHTTP(ch chan<- prometheus.Metric, target target) {
	start := time.Now()
	code, err := httpProbe(target.URL, target.Timeout)
	success := 0.0
	if err == nil && code >= 200 && code < 400 {
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(httpSuccessDesc, prometheus.GaugeValue, success, target.Name, c.initiator)
	ch <- prometheus.MustNewConstMetric(httpDurationDesc, prometheus.GaugeValue, time.Since(start).Seconds(), target.Name, c.initiator)
	ch <- prometheus.MustNewConstMetric(httpStatusCodeDesc, prometheus.GaugeValue, float64(code), target.Name, c.initiator)
}

func (c *ProbeCollector) collectCert(ch chan<- prometheus.Metric, target target) {
	notAfter, ok := certProbe(target.Host, target.Port, target.Timeout)
	success := 0.0
	notAfterTS := 0.0
	remainingDays := 0.0
	if ok {
		success = 1
		notAfterTS = float64(notAfter.Unix())
		remainingDays = time.Until(notAfter).Hours() / 24
	}
	ch <- prometheus.MustNewConstMetric(certSuccessDesc, prometheus.GaugeValue, success, target.Name, c.initiator)
	ch <- prometheus.MustNewConstMetric(certNotAfterDesc, prometheus.GaugeValue, notAfterTS, target.Name, c.initiator)
	ch <- prometheus.MustNewConstMetric(certRemainingDaysDesc, prometheus.GaugeValue, remainingDays, target.Name, c.initiator)
}

func (c *ProbeCollector) collectPing(ch chan<- prometheus.Metric, target target) {
	start := time.Now()
	ok := pingProbe(target.Host, target.Timeout)
	success := 0.0
	if ok {
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(pingSuccessDesc, prometheus.GaugeValue, success, target.Name, c.initiator)
	ch <- prometheus.MustNewConstMetric(pingDurationDesc, prometheus.GaugeValue, time.Since(start).Seconds(), target.Name, c.initiator)
}

func (c *ProbeCollector) loadTargets(ctx context.Context) []target {
	merged := append([]target{}, c.configTargets...)
	items, err := c.repo.ListEnabled(ctx)
	if err != nil {
		return merged
	}
	for _, it := range items {
		if it == nil {
			continue
		}
		timeout := c.defaultTimeout
		if it.TimeoutSeconds > 0 {
			timeout = time.Duration(it.TimeoutSeconds) * time.Second
		}
		t := target{
			Type:    strings.ToLower(strings.TrimSpace(it.Type)),
			Host:    strings.TrimSpace(it.Host),
			Port:    strings.TrimSpace(it.Port),
			URL:     strings.TrimSpace(it.URL),
			Name:    strings.TrimSpace(it.Name),
			Timeout: timeout,
		}
		merged = append(merged, t)
	}
	return merged
}

func normalizeConfigTargets(probe *conf.Probe, defaultTimeout time.Duration) []target {
	if probe == nil {
		return nil
	}
	out := make([]target, 0, len(probe.GetTargets()))
	for _, it := range probe.GetTargets() {
		if it == nil {
			continue
		}
		taskType := strings.ToLower(strings.TrimSpace(it.GetType()))
		host := strings.TrimSpace(it.GetHost())
		port := strings.TrimSpace(it.GetPort())
		rawURL := strings.TrimSpace(it.GetUrl())
		name := strings.TrimSpace(it.GetName())
		timeout := defaultTimeout
		if it.GetTimeoutSeconds() > 0 {
			timeout = time.Duration(it.GetTimeoutSeconds()) * time.Second
		}
		if taskType == "" {
			if rawURL != "" {
				taskType = probeTypeHTTP
			} else {
				taskType = probeTypeTCP
			}
		}
		switch taskType {
		case probeTypeTCP:
			if host == "" || port == "" {
				continue
			}
			if name == "" {
				name = net.JoinHostPort(host, port)
			}
		case probeTypePort:
			if port == "" {
				continue
			}
			if host == "" {
				host = "127.0.0.1"
			}
			if name == "" {
				name = net.JoinHostPort(host, port)
			}
		case probeTypeHTTP:
			if rawURL == "" {
				continue
			}
			if name == "" {
				name = rawURL
			}
		case probeTypeCert:
			if host == "" {
				continue
			}
			if port == "" {
				port = "443"
			}
			if name == "" {
				name = net.JoinHostPort(host, port)
			}
		case probeTypePing:
			if host == "" {
				continue
			}
			if name == "" {
				name = host
			}
		default:
			continue
		}
		out = append(out, target{
			Type:    taskType,
			Host:    host,
			Port:    port,
			URL:     rawURL,
			Name:    name,
			Timeout: timeout,
		})
	}
	return out
}

func discoverLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer func() { _ = conn.Close() }()
	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || addr.IP == nil {
		return "127.0.0.1"
	}
	return addr.IP.String()
}

func tcpProbe(host, port string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func httpProbe(targetURL string, timeout time.Duration) (int, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(targetURL)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode, nil
}

func certProbe(host, port string, timeout time.Duration) (time.Time, bool) {
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(host, port), &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return time.Time{}, false
	}
	defer func() { _ = conn.Close() }()
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return time.Time{}, false
	}
	return state.PeerCertificates[0].NotAfter, true
}

func pingProbe(host string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	args := []string{"-c", "1", host}
	if runtime.GOOS == "linux" {
		seconds := int(timeout / time.Second)
		if seconds <= 0 {
			seconds = 1
		}
		args = []string{"-c", "1", "-W", strconv.Itoa(seconds), host}
	}
	cmd := exec.CommandContext(ctx, "ping", args...)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
