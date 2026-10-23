# prometheus-tcmalloc

Go module: [`github.com/jiazhanfeng1989/go-prometheus-tcmalloc`](https://github.com/jiazhanfeng1989/go-prometheus-tcmalloc)

Export **tcmalloc** allocator stats on a Prometheus `/metrics` endpoint.

[gperftools](https://github.com/gperftools/gperftools) and [google/tcmalloc](https://github.com/google/tcmalloc) are different allocators. This package talks to whichever one is already loaded in the process.

## When you need this

Use it when a Go process is **not** using the Go heap as its only allocator, and RSS does not match `go_memstats_*`.

Typical cases:

- The binary or a sidecar is started with `LD_PRELOAD=libtcmalloc.so.4` (gperftools).
- A C/C++ library inside the process was built with google/tcmalloc.
- Prometheus `process_resident_memory_bytes` is high, but Go runtime metrics look fine. You need to see free pageheap, unmapped pages, and allocator caches.

`github.com/prometheus/client_golang` only exposes Go runtime stats and process RSS. Neither tcmalloc ships Prometheus output. This package `dlsym`s the C stats API and does **not** link `libtcmalloc`, so the process still starts if tcmalloc is absent (`tcmalloc_present 0`).

Collection runs only when Prometheus scrapes `/metrics`. See [Performance and production](#performance-and-production).

## Install

```bash
go get github.com/jiazhanfeng1989/go-prometheus-tcmalloc
```

Building this module needs cgo on Linux or macOS (`libdl` / `libSystem`). You do not need tcmalloc headers. With `CGO_ENABLED=0` the package still builds and only exports `tcmalloc_present 0`.

## Use

If the process already serves `promhttp.Handler()`, a blank import registers the collector on the default registry:

```go
import _ "github.com/jiazhanfeng1989/go-prometheus-tcmalloc/autoload"
```

If you own the registry, skip the autoload import and register once:

```go
import tcmallocprom "github.com/jiazhanfeng1989/go-prometheus-tcmalloc"

prometheus.MustRegister(tcmallocprom.NewCollector())
```

Do not blank-import `autoload` and also call `MustRegister(NewCollector())`. That registers the collector twice and panics.

The allocator must already be loaded at runtime.

gperftools (common in images as `libtcmalloc.so.4`):

```text
LD_PRELOAD=/usr/lib/x86_64-linux-gnu/libtcmalloc.so.4
```

google/tcmalloc is usually linked into a C++ dependency rather than preloaded. If that library is loaded, this collector finds `MallocExtension_Internal_GetNumericProperty`.

`tcmalloc_info{impl="gperftools|google"} 1` tells you which one was detected. Property names differ; a name that this build does not support is omitted instead of exported as 0. `Describe` lists the full metric set; `Collect` only emits series that this allocator actually provided.

## Performance and production

This collector is meant for production. It does not sit on the request path, start a background goroutine, or link `libtcmalloc`. After the first scrape it reuses the `dlsym` function pointers.

A scrape is the only cost. Each scrape calls `GetNumericProperty` once per property (about 15–25 calls). gperftools takes the pageheap lock for those reads; google/tcmalloc walks allocator stats the same way. Other threads that are in `malloc`/`free` can stall for that lock. That is why the default Prometheus interval of 10–15s is the right setting.

It is a poor fit if `/metrics` is scraped every second, or if the process is already latency-sensitive around native allocations. If tcmalloc is not loaded, a scrape is a couple of failed `dlsym` lookups and `tcmalloc_present 0`.

## Metrics

| Metric | Meaning |
| --- | --- |
| `tcmalloc_present` | 1 if a tcmalloc stats symbol was found |
| `tcmalloc_info` | 1 labeled with `impl=gperftools` or `impl=google` |
| `tcmalloc_current_allocated_bytes` | Bytes still held by the application, excluding allocator caches |
| `tcmalloc_heap_bytes` | Heap tcmalloc reserved from the OS |
| `tcmalloc_physical_bytes` | Heap plus metadata, minus unmapped pages |
| `tcmalloc_virtual_memory_bytes` | Virtual memory google/tcmalloc has reserved |
| `tcmalloc_peak_memory_bytes` | Peak backed memory reported by google/tcmalloc |
| `tcmalloc_pageheap_free_bytes` | Free mapped pageheap bytes. These still count toward RSS |
| `tcmalloc_pageheap_unmapped_bytes` | Free bytes already `madvise`'d back to the OS |
| `tcmalloc_central_cache_free_bytes` | Free bytes in the central cache |
| `tcmalloc_transfer_cache_free_bytes` | Free bytes in transfer caches |
| `tcmalloc_thread_cache_free_bytes` | Free bytes in per-thread caches |
| `tcmalloc_cpu_free_bytes` | Free bytes in google/tcmalloc per-CPU caches |
| `tcmalloc_sharded_transfer_cache_free_bytes` | Free bytes in google/tcmalloc sharded transfer caches |
| `tcmalloc_local_bytes` | Local google/tcmalloc caches (CPU, thread, sharded transfer) |
| `tcmalloc_pageheap_committed_bytes` | Currently committed pageheap bytes (gperftools) |
| `tcmalloc_metadata_bytes` | google/tcmalloc internal metadata |
| `tcmalloc_slack_bytes` | Free plus unmapped pageheap bytes |
| `tcmalloc_external_fragmentation_bytes` | External fragmentation (google/tcmalloc) |
| `tcmalloc_per_cpu_caches_active` | 1 if google/tcmalloc per-CPU caches are on |
| `tcmalloc_max_total_thread_cache_bytes` | Thread-cache size cap |
| `tcmalloc_thread_cache_count` | Number of google/tcmalloc thread caches |
| `tcmalloc_pageheap_scavenge_count` | Scavenger runs. gperftools runs this on `free`, not a timer |
| `tcmalloc_pageheap_commit_count` | Commit operations (gperftools) |
| `tcmalloc_pageheap_decommit_count` | Decommit / release operations (gperftools) |
| `tcmalloc_pageheap_total_commit_bytes` | Cumulative committed bytes (gperftools) |
| `tcmalloc_pageheap_total_decommit_bytes` | Cumulative released bytes (gperftools) |
| `tcmalloc_memory_release_rate` | Release rate. `0` never returns free spans to the OS. gperftools default is `1` |

## Reading RSS that will not drop

- `pageheap_free_bytes` high and `pageheap_unmapped_bytes` near 0: free pages are still mapped, so RSS stays up.
- `current_allocated_bytes` high: C allocations are still live, not idle cache.
- gperftools `pageheap_scavenge_count` not increasing: nothing has called tcmalloc `free` since. Go GC does not call `free` for you.
- google/tcmalloc with `cpu_free_bytes` or `local_bytes` high: memory is sitting in per-CPU caches, not the pageheap.

## Test

```bash
go test ./...
golangci-lint run ./...
```

Tests build a fake `MallocExtension_GetNumericProperty` or `MallocExtension_Internal_GetNumericProperty` shared library, `dlopen` it, and check that `dlsym(RTLD_DEFAULT)` can read it. With no library loaded, only `tcmalloc_present 0` is exported.
