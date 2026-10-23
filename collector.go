// Package tcmallocprom exports tcmalloc stats to Prometheus.
//
// It supports gperftools tcmalloc and google/tcmalloc. The process must
// already have one of them loaded, usually through LD_PRELOAD or a C++
// dependency. This package looks the C stats API up with dlsym and does not
// link libtcmalloc, so a binary without tcmalloc still starts.
// Collection runs when Prometheus scrapes /metrics. It is not on the request path.
package tcmallocprom

import "github.com/prometheus/client_golang/prometheus"

const (
	presentMetric = "tcmalloc_present"
	infoMetric    = "tcmalloc_info"
	releaseMetric = "tcmalloc_memory_release_rate"

	implGperftools = "gperftools"
	implGoogle     = "google"
)

// Collector reads tcmalloc numeric properties on each scrape.
type Collector struct {
	src     source
	present *prometheus.Desc
	impl    *prometheus.Desc
	release *prometheus.Desc
	descs   []*prometheus.Desc
}

type reading struct {
	present bool
	impl    string
	values  []uint64
	ok      []bool
	rate    float64
	hasRate bool
}

type source interface {
	read() reading
}

// NewCollector returns a collector for the default Prometheus registry.
// Register it with prometheus.MustRegister, or blank-import the autoload package.
func NewCollector() *Collector {
	return newCollector(newCGOSource())
}

func newCollector(src source) *Collector {
	c := &Collector{
		src:     src,
		present: prometheus.NewDesc(presentMetric, "1 if a tcmalloc stats API is loaded in this process, otherwise 0.", nil, nil),
		impl:    prometheus.NewDesc(infoMetric, "Information about the detected tcmalloc implementation.", []string{"impl"}, nil),
		release: prometheus.NewDesc(releaseMetric, "tcmalloc memory release rate. 0 disables releasing free spans back to the OS. The gperftools default is 1.", nil, nil),
		descs:   make([]*prometheus.Desc, len(properties)),
	}
	for i, p := range properties {
		c.descs[i] = prometheus.NewDesc(p.metric, p.help, nil, nil)
	}
	return c
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.present
	ch <- c.impl
	ch <- c.release
	for _, desc := range c.descs {
		ch <- desc
	}
}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	sample := c.src.read()
	ch <- prometheus.MustNewConstMetric(c.present, prometheus.GaugeValue, boolValue(sample.present))
	if !sample.present {
		return
	}
	if sample.impl != "" {
		ch <- prometheus.MustNewConstMetric(c.impl, prometheus.GaugeValue, 1, sample.impl)
	}
	for i, p := range properties {
		if i >= len(sample.ok) || !sample.ok[i] {
			continue
		}
		ch <- prometheus.MustNewConstMetric(c.descs[i], p.valueType(), float64(sample.values[i]))
	}
	if sample.hasRate {
		ch <- prometheus.MustNewConstMetric(c.release, prometheus.GaugeValue, sample.rate)
	}
}

func boolValue(v bool) float64 {
	if v {
		return 1
	}
	return 0
}
