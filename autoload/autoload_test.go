package autoload

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestBlankImportRegistersPresence(t *testing.T) {
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		switch family.GetName() {
		case "tcmalloc_present":
			if family.GetType() != dto.MetricType_GAUGE {
				t.Fatalf("type = %s", family.GetType())
			}
			if got := family.Metric[0].GetGauge().GetValue(); got != 0 {
				t.Fatalf("tcmalloc_present = %v, want 0 without libtcmalloc", got)
			}
			return
		case "tcmalloc_pageheap_free_bytes":
			t.Fatal("pageheap metrics must not be exported when tcmalloc is absent")
		}
	}
	t.Fatal("tcmalloc_present was not registered")
}
