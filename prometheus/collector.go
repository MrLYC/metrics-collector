package prometheus

import (
	"crypto/sha256"
	"fmt"
	"github.com/mrlyc/metrics-collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
	"time"
)

// Provider :
type Provider struct {
	*collector.Collector
}

func makeMetricName(key string) string {
	name := make([]rune, len(key))
	for i, k := range key {
		switch k {
		case ' ', '.', '-', '=':
			k = '_'
		}
		name[i] = k
	}
	return string(name)
}

func labelToID(labels map[string]string) string {
	h := sha256.New()
	for key, value := range labels {
		h.Write([]byte(fmt.Sprintf("%s:%s\n", key, value)))
	}
	return fmt.Sprintf("%s", h.Sum(nil))
}

// NewProvider :
func NewProvider(namespace string, subSystem string, registry metrics.Registry, registerer prometheus.Registerer, interval time.Duration, labels map[string]string) *Provider {
	registered := make(map[string]*prometheus.GaugeVec)
	return &Provider{
		Collector: collector.NewCollector(registry, interval, func(metric *collector.Metric) {
			gauges, ok := registered[metric.ID]
			if !ok {
				gauges = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name:        makeMetricName(metric.Name),
					Namespace:   namespace,
					Subsystem:   subSystem,
					ConstLabels: labels,
				}, []string{"key"})
				err := registerer.Register(gauges)
				if err != nil {
					return
				}

				registered[metric.ID] = gauges
			}

			for key, value := range metric.Values {
				switch v := value.(type) {
				case float64:
					gauges.With(prometheus.Labels{
						"key": key,
					}).Set(v)
				case int64:
					gauges.With(prometheus.Labels{
						"key": key,
					}).Set(float64(v))
				}
			}
		}),
	}
}
