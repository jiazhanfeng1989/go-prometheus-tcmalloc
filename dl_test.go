//go:build cgo && unix

package tcmallocprom

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	dto "github.com/prometheus/client_model/go"

	"github.com/jiazhanfeng1989/go-prometheus-tcmalloc/internal/dlopen"
)

func TestDLSymLoadsPreloadedAllocator(t *testing.T) {
	if os.Getenv("TCMALLOC_FAKE") != "gperftools" {
		runFakeAllocatorChild(t, "gperftools")
		return
	}

	before := gather(t, NewCollector())
	if len(before) != 1 || before[presentMetric] != 0 {
		t.Fatalf("before dlopen metrics = %#v, want only %s=0", before, presentMetric)
	}

	lib := buildFakeAllocator(t, "testdata/fake_tcmalloc.c")
	if err := dlopen.Open(lib); err != nil {
		t.Fatal(err)
	}

	got := gather(t, NewCollector())
	if implLabel(t, NewCollector()) != implGperftools {
		t.Fatalf("impl = %q, want %s", implLabel(t, NewCollector()), implGperftools)
	}
	want := map[string]float64{
		presentMetric:                           1,
		infoMetric:                              1,
		"tcmalloc_current_allocated_bytes":      1001,
		"tcmalloc_heap_bytes":                   2002,
		"tcmalloc_physical_bytes":               3003,
		"tcmalloc_central_cache_free_bytes":     11,
		"tcmalloc_transfer_cache_free_bytes":    12,
		"tcmalloc_thread_cache_free_bytes":      13,
		"tcmalloc_pageheap_free_bytes":          4004,
		"tcmalloc_pageheap_unmapped_bytes":      5,
		"tcmalloc_pageheap_committed_bytes":     14,
		"tcmalloc_slack_bytes":                  15,
		"tcmalloc_max_total_thread_cache_bytes": 16,
		"tcmalloc_pageheap_scavenge_count":      6,
		"tcmalloc_pageheap_commit_count":        17,
		"tcmalloc_pageheap_decommit_count":      18,
		"tcmalloc_pageheap_total_commit_bytes":  19,
		releaseMetric:                           1.25,
	}
	assertExactMetrics(t, got, want)
	assertFamilyType(t, NewCollector(), "tcmalloc_pageheap_scavenge_count", dto.MetricType_COUNTER)
	assertFamilyType(t, NewCollector(), "tcmalloc_pageheap_free_bytes", dto.MetricType_GAUGE)
}

func TestDLSymLoadsGoogleTcmalloc(t *testing.T) {
	if os.Getenv("TCMALLOC_FAKE") != "google" {
		runFakeAllocatorChild(t, "google")
		return
	}

	lib := buildFakeAllocator(t, "testdata/fake_google_tcmalloc.c")
	if err := dlopen.Open(lib); err != nil {
		t.Fatal(err)
	}

	c := NewCollector()
	got := gather(t, c)
	if implLabel(t, c) != implGoogle {
		t.Fatalf("impl = %q, want %s", implLabel(t, c), implGoogle)
	}
	want := map[string]float64{
		presentMetric:                       1,
		infoMetric:                          1,
		"tcmalloc_current_allocated_bytes":  1101,
		"tcmalloc_heap_bytes":               2202,
		"tcmalloc_physical_bytes":           3303,
		"tcmalloc_pageheap_free_bytes":      4404,
		"tcmalloc_pageheap_unmapped_bytes":  7,
		"tcmalloc_cpu_free_bytes":           8,
		"tcmalloc_central_cache_free_bytes": 9,
		"tcmalloc_thread_cache_free_bytes":  10,
	}
	assertExactMetrics(t, got, want)
}

func TestSameCollectorPicksUpLaterDLOpen(t *testing.T) {
	if os.Getenv("TCMALLOC_FAKE") != "gperftools-retry" {
		runFakeAllocatorChild(t, "gperftools-retry")
		return
	}

	c := NewCollector()
	before := gather(t, c)
	if len(before) != 1 || before[presentMetric] != 0 {
		t.Fatalf("before dlopen metrics = %#v, want only %s=0", before, presentMetric)
	}

	lib := buildFakeAllocator(t, "testdata/fake_tcmalloc.c")
	if err := dlopen.Open(lib); err != nil {
		t.Fatal(err)
	}

	got := gather(t, c)
	if got[presentMetric] != 1 {
		t.Fatalf("present = %v, want 1", got[presentMetric])
	}
	if implLabel(t, c) != implGperftools {
		t.Fatalf("impl = %q, want %s", implLabel(t, c), implGperftools)
	}
	if got["tcmalloc_current_allocated_bytes"] != 1001 {
		t.Fatalf("current_allocated = %v, want 1001", got["tcmalloc_current_allocated_bytes"])
	}
}

func runFakeAllocatorChild(t *testing.T, fake string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v", "-test.count=1")
	cmd.Env = append(os.Environ(), "TCMALLOC_FAKE="+fake)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child %s: %v\n%s", fake, err, output)
	}
}

func buildFakeAllocator(t *testing.T, src string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "libfake_tcmalloc.so")
	cmd := exec.Command("cc", "-shared", "-fPIC", "-o", out, src)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build %s: %v\n%s", src, err, output)
	}
	return out
}

func implLabel(t *testing.T, c *Collector) string {
	t.Helper()
	for _, family := range gatherFamilies(t, c) {
		if family.GetName() != infoMetric {
			continue
		}
		if len(family.Metric) != 1 || len(family.Metric[0].GetLabel()) != 1 {
			t.Fatalf("tcmalloc_info series = %#v", family.Metric)
		}
		return family.Metric[0].GetLabel()[0].GetValue()
	}
	t.Fatal("tcmalloc_info was not exported")
	return ""
}

func assertExactMetrics(t *testing.T, got, want map[string]float64) {
	t.Helper()
	for name, v := range want {
		if got[name] != v {
			t.Errorf("%s = %v, want %v", name, got[name], v)
		}
	}
	for name, v := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("unexpected metric %s = %v", name, v)
		}
	}
}
