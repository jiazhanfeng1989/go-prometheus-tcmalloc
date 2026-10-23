//go:build !cgo || !unix

package tcmallocprom

import "testing"

func TestNewCollectorWithoutCGO(t *testing.T) {
	got := gather(t, NewCollector())
	if len(got) != 1 || got[presentMetric] != 0 {
		t.Fatalf("metrics = %#v, want only %s=0", got, presentMetric)
	}
}
