package tcmallocprom

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestOmitsUnsupportedProperties(t *testing.T) {
	values := make([]uint64, len(properties))
	ok := make([]bool, len(properties))
	for i, p := range properties {
		if p.metric == "tcmalloc_pageheap_free_bytes" {
			values[i] = 9
			ok[i] = true
		}
	}
	got := gather(t, newCollector(fixedSource{reading: reading{
		present: true,
		values:  values,
		ok:      ok,
		hasRate: false,
	}}))

	if got[presentMetric] != 1 {
		t.Fatalf("present = %v, want 1", got[presentMetric])
	}
	if got["tcmalloc_pageheap_free_bytes"] != 9 {
		t.Fatalf("pageheap_free = %v, want 9", got["tcmalloc_pageheap_free_bytes"])
	}
	if _, exists := got["tcmalloc_heap_bytes"]; exists {
		t.Fatal("unsupported property was exported as a zero gauge")
	}
	if _, exists := got[releaseMetric]; exists {
		t.Fatal("missing release-rate symbol was exported")
	}
}

func TestAbsentSourceExportsOnlyPresence(t *testing.T) {
	got := gather(t, newCollector(fixedSource{}))
	if len(got) != 1 || got[presentMetric] != 0 {
		t.Fatalf("metrics = %#v, want only %s=0", got, presentMetric)
	}
}

func TestExportsImplLabel(t *testing.T) {
	got := gatherFamilies(t, newCollector(fixedSource{reading: reading{
		present: true,
		impl:    implGoogle,
	}}))
	var impl *dto.MetricFamily
	for _, family := range got {
		if family.GetName() == infoMetric {
			impl = family
			break
		}
	}
	if impl == nil {
		t.Fatal("tcmalloc_info was not exported")
	}
	if impl.GetType() != dto.MetricType_GAUGE {
		t.Fatalf("type = %s", impl.GetType())
	}
	if len(impl.Metric) != 1 {
		t.Fatalf("series = %d", len(impl.Metric))
	}
	labels := impl.Metric[0].GetLabel()
	if len(labels) != 1 || labels[0].GetName() != "impl" || labels[0].GetValue() != implGoogle {
		t.Fatalf("labels = %v, want impl=%s", labels, implGoogle)
	}
	if impl.Metric[0].GetGauge().GetValue() != 1 {
		t.Fatalf("value = %v, want 1", impl.Metric[0].GetGauge().GetValue())
	}
}

func TestMetricNamesAreUnique(t *testing.T) {
	seen := map[string]struct{}{presentMetric: {}, infoMetric: {}, releaseMetric: {}}
	for _, p := range properties {
		if _, dup := seen[p.metric]; dup {
			t.Fatalf("duplicate metric %s", p.metric)
		}
		seen[p.metric] = struct{}{}
	}
}

type fixedSource struct {
	reading reading
}

func (f fixedSource) read() reading { return f.reading }

func gather(t *testing.T, c prometheus.Collector) map[string]float64 {
	t.Helper()
	families := gatherFamilies(t, c)
	out := make(map[string]float64, len(families))
	for _, family := range families {
		if len(family.Metric) != 1 {
			t.Fatalf("%s has %d series", family.GetName(), len(family.Metric))
		}
		out[family.GetName()] = metricValue(t, family)
	}
	return out
}

func gatherFamilies(t *testing.T, c prometheus.Collector) []*dto.MetricFamily {
	t.Helper()
	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("register: %v", err)
	}
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	return families
}

func assertFamilyType(t *testing.T, c prometheus.Collector, name string, want dto.MetricType) {
	t.Helper()
	for _, family := range gatherFamilies(t, c) {
		if family.GetName() != name {
			continue
		}
		if family.GetType() != want {
			t.Fatalf("%s type = %s, want %s", name, family.GetType(), want)
		}
		return
	}
	t.Fatalf("metric %s was not exported", name)
}

func metricValue(t *testing.T, family *dto.MetricFamily) float64 {
	t.Helper()
	metric := family.Metric[0]
	switch family.GetType() {
	case dto.MetricType_GAUGE:
		return metric.GetGauge().GetValue()
	case dto.MetricType_COUNTER:
		return metric.GetCounter().GetValue()
	default:
		t.Fatalf("%s has type %s", family.GetName(), family.GetType())
		return 0
	}
}
