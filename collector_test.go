package collector_test

import (
	"fmt"
	. "github.com/mrlyc/metrics-collector"
	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// CollectorSuite :
type CollectorSuite struct {
	suite.Suite
}

// TestCollect :
func (s *CollectorSuite) TestCollect() {
	cases := []struct {
		metric     interface{}
		metricType MetricType
	}{
		{metrics.NewCounter(), MetricCounterType},
		{metrics.NewGauge(), MetricGaugeType},
		{metrics.NewGaugeFloat64(), MetricGaugeType},
		{metrics.NewHistogram(metrics.NewExpDecaySample(1028, 0.015)), MetricHistogramType},
		{metrics.NewMeter(), MetricMeterType},
		{metrics.NewTimer(), MetricTimerType},
		{metrics.NewHealthcheck(func(healthcheck metrics.Healthcheck) {
			healthcheck.Unhealthy(fmt.Errorf("test"))
		}), MetricHealthCheckType},
		{metrics.NewHealthcheck(func(healthcheck metrics.Healthcheck) {
			healthcheck.Healthy()
		}), MetricHealthCheckType},
	}

	for _, c := range cases {
		registry := metrics.NewRegistry()
		s.NoError(registry.Register("x", c.metric))
		allMetrics := registry.GetAll()

		var collector *Collector
		collector = NewCollector(registry, time.Microsecond, func(metric *Metric) {
			s.Equal(c.metricType, metric.Type)
			xMetrics := allMetrics["x"]
			s.True(len(metric.Values) > 0)

			for key, value := range xMetrics {
				s.NotPanics(func() {
					s.Equal(value, metric.Values[key])
				})
			}

			collector.Stop()
		})
		collector.Start()
		collector.Wait()
	}
}

// TestCollectorSuite :
func TestCollectorSuite(t *testing.T) {
	suite.Run(t, new(CollectorSuite))
}
