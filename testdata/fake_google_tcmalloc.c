#include <stdbool.h>
#include <stddef.h>
#include <string.h>

static bool name_eq(const char *name, size_t n, const char *want) {
	size_t w = strlen(want);
	return n == w && memcmp(name, want, n) == 0;
}

bool MallocExtension_Internal_GetNumericProperty(const char *name, size_t name_size, size_t *value) {
	if (name == NULL || value == NULL) {
		return false;
	}
	if (name_eq(name, name_size, "generic.current_allocated_bytes")) {
		*value = 1101;
		return true;
	}
	if (name_eq(name, name_size, "generic.heap_size")) {
		*value = 2202;
		return true;
	}
	if (name_eq(name, name_size, "generic.physical_memory_used")) {
		*value = 3303;
		return true;
	}
	if (name_eq(name, name_size, "tcmalloc.pageheap_free_bytes")) {
		*value = 4404;
		return true;
	}
	if (name_eq(name, name_size, "tcmalloc.pageheap_unmapped_bytes")) {
		*value = 7;
		return true;
	}
	if (name_eq(name, name_size, "tcmalloc.cpu_free")) {
		*value = 8;
		return true;
	}
	if (name_eq(name, name_size, "tcmalloc.central_cache_free")) {
		*value = 9;
		return true;
	}
	if (name_eq(name, name_size, "tcmalloc.thread_cache_free")) {
		*value = 10;
		return true;
	}
	return false;
}
