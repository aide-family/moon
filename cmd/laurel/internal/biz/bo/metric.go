package bo

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/plugin/cache"
)

type MetricVec interface {
	cache.Object
	GetMetricName() string
	Type() vobj.MetricType
}

var _ MetricVec = (*CounterMetricVec)(nil)

type CounterMetricVec struct {
	Namespace string   `json:"namespace"`
	SubSystem string   `json:"subSystem"`
	Name      string   `json:"name"`
	Labels    []string `json:"labels"`
	Help      string   `json:"help"`
}

// Type implements MetricVec.
func (c *CounterMetricVec) Type() vobj.MetricType {
	return vobj.MetricTypeCounter
}

// MarshalBinary implements cache.Object.
func (c *CounterMetricVec) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// UniqueKey implements cache.Object.
func (c *CounterMetricVec) UniqueKey() string {
	return c.GetMetricName()
}

// UnmarshalBinary implements cache.Object.
func (c *CounterMetricVec) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *CounterMetricVec) GetMetricName() string {
	return strings.Join([]string{c.Namespace, c.SubSystem, c.Name, vobj.MetricTypeCounter.String()}, "_")
}

func (c *CounterMetricVec) New() *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: c.Namespace,
		Subsystem: c.SubSystem,
		Name:      strings.Join([]string{c.Name, vobj.MetricTypeCounter.String()}, "_"),
		Help:      c.Help,
	}, c.Labels)
}

var _ MetricVec = (*GaugeMetricVec)(nil)

type GaugeMetricVec struct {
	Namespace string   `json:"namespace"`
	SubSystem string   `json:"subSystem"`
	Name      string   `json:"name"`
	Labels    []string `json:"labels"`
	Help      string   `json:"help"`
}

// Type implements MetricVec.
func (g *GaugeMetricVec) Type() vobj.MetricType {
	return vobj.MetricTypeGauge
}

// MarshalBinary implements cache.Object.
func (g *GaugeMetricVec) MarshalBinary() (data []byte, err error) {
	return json.Marshal(g)
}

// UniqueKey implements cache.Object.
func (g *GaugeMetricVec) UniqueKey() string {
	return g.GetMetricName()
}

// UnmarshalBinary implements cache.Object.
func (g *GaugeMetricVec) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, g)
}

func (g *GaugeMetricVec) GetMetricName() string {
	return strings.Join([]string{g.Namespace, g.SubSystem, g.Name, vobj.MetricTypeGauge.String()}, "_")
}

func (g *GaugeMetricVec) New() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: g.Namespace,
		Subsystem: g.SubSystem,
		Name:      strings.Join([]string{g.Name, vobj.MetricTypeGauge.String()}, "_"),
		Help:      g.Help,
	}, g.Labels)
}

var _ MetricVec = (*HistogramMetricVec)(nil)

type HistogramMetricVec struct {
	Namespace                       string    `json:"namespace"`
	SubSystem                       string    `json:"subSystem"`
	Name                            string    `json:"name"`
	Labels                          []string  `json:"labels"`
	Help                            string    `json:"help"`
	Buckets                         []float64 `json:"buckets"`
	NativeHistogramBucketFactor     float64   `json:"nativeHistogramBucketFactor"`
	NativeHistogramZeroThreshold    float64   `json:"nativeHistogramZeroThreshold"`
	NativeHistogramMaxBucketNumber  uint32    `json:"nativeHistogramMaxBucketNumber"`
	NativeHistogramMinResetDuration int64     `json:"nativeHistogramMinResetDuration"`
	NativeHistogramMaxZeroThreshold float64   `json:"nativeHistogramMaxZeroThreshold"`
	NativeHistogramMaxExemplars     int64     `json:"nativeHistogramMaxExemplars"`
	NativeHistogramExemplarTTL      int64     `json:"nativeHistogramExemplarTTL"`
}

// Type implements MetricVec.
func (h *HistogramMetricVec) Type() vobj.MetricType {
	return vobj.MetricTypeHistogram
}

// MarshalBinary implements cache.Object.
func (h *HistogramMetricVec) MarshalBinary() (data []byte, err error) {
	return json.Marshal(h)
}

// UniqueKey implements cache.Object.
func (h *HistogramMetricVec) UniqueKey() string {
	return h.GetMetricName()
}

// UnmarshalBinary implements cache.Object.
func (h *HistogramMetricVec) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, h)
}

func (h *HistogramMetricVec) GetMetricName() string {
	return strings.Join([]string{h.Namespace, h.SubSystem, h.Name, vobj.MetricTypeHistogram.String()}, "_")
}

func (h *HistogramMetricVec) New() *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:                       h.Namespace,
		Subsystem:                       h.SubSystem,
		Name:                            strings.Join([]string{h.Name, vobj.MetricTypeHistogram.String()}, "_"),
		Help:                            h.Help,
		Buckets:                         h.Buckets,
		NativeHistogramBucketFactor:     h.NativeHistogramBucketFactor,
		NativeHistogramZeroThreshold:    h.NativeHistogramZeroThreshold,
		NativeHistogramMaxBucketNumber:  h.NativeHistogramMaxBucketNumber,
		NativeHistogramMinResetDuration: time.Duration(h.NativeHistogramMinResetDuration),
		NativeHistogramMaxZeroThreshold: h.NativeHistogramMaxZeroThreshold,
		NativeHistogramMaxExemplars:     int(h.NativeHistogramMaxExemplars),
		NativeHistogramExemplarTTL:      time.Duration(h.NativeHistogramExemplarTTL),
	}, h.Labels)
}

var _ MetricVec = (*SummaryMetricVec)(nil)

type SummaryMetricVec struct {
	Namespace  string              `json:"namespace"`
	SubSystem  string              `json:"subSystem"`
	Name       string              `json:"name"`
	Labels     []string            `json:"labels"`
	Help       string              `json:"help"`
	Objectives map[float64]float64 `json:"objectives"`
	MaxAge     int64               `json:"maxAge"`
	AgeBuckets uint32              `json:"ageBuckets"`
	BufCap     uint32              `json:"bufCap"`
}

// Type implements MetricVec.
func (s *SummaryMetricVec) Type() vobj.MetricType {
	return vobj.MetricTypeSummary
}

// MarshalBinary implements cache.Object.
func (s *SummaryMetricVec) MarshalBinary() (data []byte, err error) {
	return json.Marshal(s)
}

// UniqueKey implements cache.Object.
func (s *SummaryMetricVec) UniqueKey() string {
	return s.GetMetricName()
}

// UnmarshalBinary implements cache.Object.
func (s *SummaryMetricVec) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *SummaryMetricVec) GetMetricName() string {
	return strings.Join([]string{s.Namespace, s.SubSystem, s.Name, vobj.MetricTypeSummary.String()}, "_")
}

func (s *SummaryMetricVec) New() *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  s.Namespace,
		Subsystem:  s.SubSystem,
		Name:       strings.Join([]string{s.Name, vobj.MetricTypeSummary.String()}, "_"),
		Help:       s.Help,
		Objectives: s.Objectives,
		MaxAge:     time.Duration(s.MaxAge),
		AgeBuckets: s.AgeBuckets,
		BufCap:     s.BufCap,
	}, s.Labels)
}

type MetricData struct {
	MetricType vobj.MetricType   `json:"metricType"`
	Namespace  string            `json:"namespace"`
	SubSystem  string            `json:"subSystem"`
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	Value      float64           `json:"value"`
}

func (m *MetricData) GetMetricName() string {
	return strings.Join([]string{m.Namespace, m.SubSystem, m.Name, m.MetricType.String()}, "_")
}
