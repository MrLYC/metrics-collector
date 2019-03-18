package collector

import (
	"context"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"sync"
	"time"
)

// MetricType :
type MetricType string

//
const (
	MetricCounterType     MetricType = "counter"
	MetricGaugeType                  = "gauge"
	MetricHistogramType              = "histogram"
	MetricMeterType                  = "meter"
	MetricTimerType                  = "timer"
	MetricHealthCheckType            = "health_check"
)

// Metric :
type Metric struct {
	Metric interface{}
	ID     string
	Name   string
	Type   MetricType
	Values map[string]interface{}
}

func NewMetric() *Metric {
	return &Metric{
		Values: make(map[string]interface{}),
	}
}

type CollectFn func(*Metric)

// Collector :
type Collector struct {
	registry metrics.Registry
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	Run CollectFn
}

// NewCollector :
func NewCollector(registry metrics.Registry, interval time.Duration, run CollectFn) *Collector {
	ctx, cancel := context.WithCancel(context.Background())
	return &Collector{
		registry: registry,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		Run:      run,
	}
}

// Stop :
func (c *Collector) Stop() {
	c.cancel()
}

// Wait :
func (c *Collector) Wait() {
	c.wg.Wait()
}

// Start :
func (c *Collector) Start() {
	go func() {
		ticker := time.NewTicker(c.interval)
	loop:
		for {
			select {
			case <-c.ctx.Done():
				break loop
			case <-ticker.C:
				c.registry.Each(func(name string, i interface{}) {
					metric := NewMetric()
					metric.Name = name
					metric.ID = fmt.Sprintf("%p", i)
					metric.Metric = i
					switch m := i.(type) {
					case metrics.Counter:
						metric.Type = MetricCounterType
						metric.Values["count"] = m.Count()
					case metrics.Gauge:
						metric.Type = MetricGaugeType
						metric.Values["value"] = m.Value()
					case metrics.GaugeFloat64:
						metric.Type = MetricGaugeType
						metric.Values["value"] = m.Value()
					case metrics.Healthcheck:
						metric.Type = MetricHealthCheckType
						metric.Values["error"] = nil
						m.Check()
						if err := m.Error(); nil != err {
							metric.Values["error"] = m.Error().Error()
						}
					case metrics.Histogram:
						metric.Type = MetricHistogramType
						h := m.Snapshot()
						ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
						metric.Values["count"] = h.Count()
						metric.Values["min"] = h.Min()
						metric.Values["max"] = h.Max()
						metric.Values["mean"] = h.Mean()
						metric.Values["stddev"] = h.StdDev()
						metric.Values["median"] = ps[0]
						metric.Values["75%"] = ps[1]
						metric.Values["95%"] = ps[2]
						metric.Values["99%"] = ps[3]
						metric.Values["99.9%"] = ps[4]
					case metrics.Meter:
						metric.Type = MetricMeterType
						s := m.Snapshot()
						metric.Values["count"] = s.Count()
						metric.Values["1m.rate"] = s.Rate1()
						metric.Values["5m.rate"] = s.Rate5()
						metric.Values["15m.rate"] = s.Rate15()
						metric.Values["mean.rate"] = s.RateMean()
					case metrics.Timer:
						metric.Type = MetricTimerType
						t := m.Snapshot()
						ps := t.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
						metric.Values["count"] = t.Count()
						metric.Values["min"] = t.Min()
						metric.Values["max"] = t.Max()
						metric.Values["mean"] = t.Mean()
						metric.Values["stddev"] = t.StdDev()
						metric.Values["median"] = ps[0]
						metric.Values["75%"] = ps[1]
						metric.Values["95%"] = ps[2]
						metric.Values["99%"] = ps[3]
						metric.Values["99.9%"] = ps[4]
						metric.Values["1m.rate"] = t.Rate1()
						metric.Values["5m.rate"] = t.Rate5()
						metric.Values["15m.rate"] = t.Rate15()
						metric.Values["mean.rate"] = t.RateMean()
					}
					c.Run(metric)
				})
			}
		}
		ticker.Stop()
		c.wg.Done()
	}()
}
