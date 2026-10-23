//go:build cgo && unix

package tcmallocprom

/*
#cgo linux LDFLAGS: -ldl

#include <dlfcn.h>
#include <stddef.h>
#include <stdint.h>

typedef int (*gperf_get_numeric_property_fn)(const char *name, size_t *value);
typedef _Bool (*google_get_numeric_property_fn)(const char *name, size_t name_size, size_t *value);
typedef double (*get_release_rate_fn)(void);

static gperf_get_numeric_property_fn lookup_gperf_numeric_property(void) {
	dlerror();
	return (gperf_get_numeric_property_fn)dlsym(RTLD_DEFAULT, "MallocExtension_GetNumericProperty");
}

static google_get_numeric_property_fn lookup_google_numeric_property(void) {
	dlerror();
	return (google_get_numeric_property_fn)dlsym(RTLD_DEFAULT, "MallocExtension_Internal_GetNumericProperty");
}

static get_release_rate_fn lookup_release_rate(void) {
	dlerror();
	return (get_release_rate_fn)dlsym(RTLD_DEFAULT, "MallocExtension_GetMemoryReleaseRate");
}

static int read_gperf_numeric_property(gperf_get_numeric_property_fn fn, const char *name, unsigned long long *out) {
	size_t value = 0;
	if (fn == NULL || name == NULL || out == NULL) {
		return 0;
	}
	if (!fn(name, &value)) {
		return 0;
	}
	*out = (unsigned long long)value;
	return 1;
}

static int read_google_numeric_property(google_get_numeric_property_fn fn, const char *name, size_t name_size, unsigned long long *out) {
	size_t value = 0;
	if (fn == NULL || name == NULL || out == NULL) {
		return 0;
	}
	if (!fn(name, name_size, &value)) {
		return 0;
	}
	*out = (unsigned long long)value;
	return 1;
}

static int read_release_rate(get_release_rate_fn fn, double *out) {
	if (fn == NULL || out == NULL) {
		return 0;
	}
	*out = fn();
	return 1;
}
*/
import "C"
import "sync"

type cName struct {
	ptr  *C.char
	size C.size_t
}

// cgoSource resolves tcmalloc on the first successful scrape and keeps the
// function pointers. A miss is not cached, so a library opened later is picked up.
type cgoSource struct {
	mu     sync.Mutex
	gperf  C.gperf_get_numeric_property_fn
	google C.google_get_numeric_property_fn
	rateFn C.get_release_rate_fn
	impl   string

	// names live for the process. Prometheus scrapes must not malloc C strings.
	gperfNames  [][]cName
	googleNames [][]cName
}

func newCGOSource() source {
	gperfNames, googleNames := propertyCNames()
	return &cgoSource{gperfNames: gperfNames, googleNames: googleNames}
}

var (
	propNamesOnce sync.Once
	gperfNames    [][]cName
	googleNames   [][]cName
)

func propertyCNames() ([][]cName, [][]cName) {
	propNamesOnce.Do(func() {
		gperfNames = cNamesFor(implGperftools)
		googleNames = cNamesFor(implGoogle)
	})
	return gperfNames, googleNames
}

func cNamesFor(impl string) [][]cName {
	out := make([][]cName, len(properties))
	for i, p := range properties {
		names := p.namesFor(impl)
		aliases := make([]cName, len(names))
		for j, name := range names {
			aliases[j] = cName{ptr: C.CString(name), size: C.size_t(len(name))}
		}
		out[i] = aliases
	}
	return out
}

func (s *cgoSource) read() reading {
	gperf, google, rateFn, impl := s.fns()
	if gperf == nil && google == nil {
		return reading{}
	}
	out := reading{
		present: true,
		impl:    impl,
		values:  make([]uint64, len(properties)),
		ok:      make([]bool, len(properties)),
	}
	for i := range properties {
		value, ok := s.readProperty(gperf, google, impl, i)
		if !ok {
			continue
		}
		out.values[i] = value
		out.ok[i] = true
	}
	if rateFn != nil {
		var rate C.double
		if C.read_release_rate(rateFn, &rate) == 1 {
			out.rate = float64(rate)
			out.hasRate = true
		}
	}
	return out
}

func (s *cgoSource) readProperty(gperf C.gperf_get_numeric_property_fn, google C.google_get_numeric_property_fn, impl string, i int) (uint64, bool) {
	names := s.gperfNames[i]
	if impl == implGoogle {
		names = s.googleNames[i]
	}
	for _, name := range names {
		var value C.ulonglong
		if gperf != nil {
			if C.read_gperf_numeric_property(gperf, name.ptr, &value) == 1 {
				return uint64(value), true
			}
			continue
		}
		if google != nil && C.read_google_numeric_property(google, name.ptr, name.size, &value) == 1 {
			return uint64(value), true
		}
	}
	return 0, false
}

func (s *cgoSource) fns() (C.gperf_get_numeric_property_fn, C.google_get_numeric_property_fn, C.get_release_rate_fn, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.gperf == nil && s.google == nil {
		if gperf := C.lookup_gperf_numeric_property(); gperf != nil {
			s.gperf = gperf
			s.rateFn = C.lookup_release_rate()
			s.impl = implGperftools
		} else if google := C.lookup_google_numeric_property(); google != nil {
			s.google = google
			s.impl = implGoogle
		}
	}
	return s.gperf, s.google, s.rateFn, s.impl
}
