package tcmallocprom

import "github.com/prometheus/client_golang/prometheus"

type metricKind int

const (
	kindGauge metricKind = iota
	kindCounter
)

type property struct {
	gperftools []string
	google     []string
	metric     string
	help       string
	kind       metricKind
}

func (p property) valueType() prometheus.ValueType {
	if p.kind == kindCounter {
		return prometheus.CounterValue
	}
	return prometheus.GaugeValue
}

func (p property) namesFor(impl string) []string {
	switch impl {
	case implGperftools:
		return p.gperftools
	case implGoogle:
		return p.google
	default:
		return nil
	}
}

func n(names ...string) []string { return names }

func gauge(metric, help string, gperftools, google []string) property {
	return property{gperftools: gperftools, google: google, metric: metric, help: help, kind: kindGauge}
}

func counter(metric, help string, gperftools []string) property {
	return property{gperftools: gperftools, metric: metric, help: help, kind: kindCounter}
}

// Numeric properties from gperftools and google/tcmalloc. Each impl is queried
// with its own names. A miss is omitted rather than exported as 0.
var properties = []property{
	gauge(
		"tcmalloc_current_allocated_bytes",
		"Bytes currently allocated by the application through tcmalloc, excluding allocator caches and fragmentation.",
		n("generic.current_allocated_bytes"),
		n("generic.current_allocated_bytes", "generic.bytes_in_use_by_app"),
	),
	gauge(
		"tcmalloc_heap_bytes",
		"Bytes of system memory reserved by tcmalloc, including free and unmapped pageheap bytes.",
		n("generic.heap_size"),
		n("generic.heap_size"),
	),
	gauge(
		"tcmalloc_physical_bytes",
		"Bytes tcmalloc counts as physical: heap plus metadata, minus pageheap bytes already unmapped.",
		n("generic.total_physical_bytes"),
		n("generic.physical_memory_used"),
	),
	gauge(
		"tcmalloc_virtual_memory_bytes",
		"Virtual memory tcmalloc has reserved, including unmapped pages.",
		nil,
		n("generic.virtual_memory_used"),
	),
	gauge(
		"tcmalloc_peak_memory_bytes",
		"Peak backed memory observed by google/tcmalloc.",
		nil,
		n("generic.peak_memory_usage"),
	),
	gauge(
		"tcmalloc_central_cache_free_bytes",
		"Free bytes held in the tcmalloc central cache.",
		n("tcmalloc.central_cache_free_bytes"),
		n("tcmalloc.central_cache_free"),
	),
	gauge(
		"tcmalloc_transfer_cache_free_bytes",
		"Free bytes held in tcmalloc transfer caches.",
		n("tcmalloc.transfer_cache_free_bytes"),
		n("tcmalloc.transfer_cache_free"),
	),
	gauge(
		"tcmalloc_thread_cache_free_bytes",
		"Free bytes held in tcmalloc per-thread caches.",
		n("tcmalloc.thread_cache_free_bytes"),
		n("tcmalloc.thread_cache_free", "tcmalloc.current_total_thread_cache_bytes"),
	),
	gauge(
		"tcmalloc_cpu_free_bytes",
		"Free bytes held in google/tcmalloc per-CPU caches.",
		nil,
		n("tcmalloc.cpu_free"),
	),
	gauge(
		"tcmalloc_sharded_transfer_cache_free_bytes",
		"Free bytes held in google/tcmalloc sharded transfer caches.",
		nil,
		n("tcmalloc.sharded_transfer_cache_free"),
	),
	gauge(
		"tcmalloc_local_bytes",
		"Free bytes in google/tcmalloc local caches (CPU, thread, and sharded transfer).",
		nil,
		n("tcmalloc.local_bytes"),
	),
	gauge(
		"tcmalloc_pageheap_free_bytes",
		"Free bytes in mapped pageheap spans. These count toward RSS until tcmalloc unmaps them.",
		n("tcmalloc.pageheap_free_bytes"),
		n("tcmalloc.pageheap_free_bytes"),
	),
	gauge(
		"tcmalloc_pageheap_unmapped_bytes",
		"Free bytes tcmalloc has released to the OS with madvise. Usually not charged as RSS.",
		n("tcmalloc.pageheap_unmapped_bytes"),
		n("tcmalloc.pageheap_unmapped_bytes"),
	),
	gauge(
		"tcmalloc_pageheap_committed_bytes",
		"Bytes the pageheap currently has committed.",
		n("tcmalloc.pageheap_committed_bytes"),
		nil,
	),
	gauge(
		"tcmalloc_metadata_bytes",
		"Bytes used by google/tcmalloc internal metadata.",
		nil,
		n("tcmalloc.metadata_bytes"),
	),
	gauge(
		"tcmalloc_slack_bytes",
		"Free plus unmapped pageheap bytes.",
		n("tcmalloc.slack_bytes"),
		n("tcmalloc.slack_bytes"),
	),
	gauge(
		"tcmalloc_external_fragmentation_bytes",
		"External fragmentation reported by google/tcmalloc.",
		nil,
		n("tcmalloc.external_fragmentation_bytes"),
	),
	gauge(
		"tcmalloc_per_cpu_caches_active",
		"1 if google/tcmalloc is using per-CPU caches, otherwise 0.",
		nil,
		n("tcmalloc.per_cpu_caches_active"),
	),
	gauge(
		"tcmalloc_max_total_thread_cache_bytes",
		"Upper limit on bytes stored across all tcmalloc thread caches.",
		n("tcmalloc.max_total_thread_cache_bytes"),
		n("tcmalloc.max_total_thread_cache_bytes"),
	),
	gauge(
		"tcmalloc_thread_cache_count",
		"Number of google/tcmalloc thread caches in use.",
		nil,
		n("tcmalloc.thread_cache_count"),
	),
	counter(
		"tcmalloc_pageheap_scavenge_count",
		"Times the pageheap scavenger has run. tcmalloc runs it from the free path, not on a timer.",
		n("tcmalloc.pageheap_scavenge_count"),
	),
	counter(
		"tcmalloc_pageheap_commit_count",
		"Times the pageheap has committed memory.",
		n("tcmalloc.pageheap_commit_count"),
	),
	counter(
		"tcmalloc_pageheap_decommit_count",
		"Times the pageheap has decommitted or released memory.",
		n("tcmalloc.pageheap_decommit_count"),
	),
	counter(
		"tcmalloc_pageheap_total_commit_bytes",
		"Cumulative bytes the pageheap has committed.",
		n("tcmalloc.pageheap_total_commit_bytes"),
	),
	counter(
		"tcmalloc_pageheap_total_decommit_bytes",
		"Cumulative bytes the pageheap has decommitted or released.",
		n("tcmalloc.pageheap_total_decommit_bytes"),
	),
}
