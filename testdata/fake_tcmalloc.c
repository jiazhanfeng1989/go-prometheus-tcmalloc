#include <stddef.h>
#include <string.h>

int MallocExtension_GetNumericProperty(const char *name, size_t *value) {
	if (name == NULL || value == NULL) {
		return 0;
	}
	if (strcmp(name, "generic.current_allocated_bytes") == 0) {
		*value = 1001;
		return 1;
	}
	if (strcmp(name, "generic.heap_size") == 0) {
		*value = 2002;
		return 1;
	}
	if (strcmp(name, "generic.total_physical_bytes") == 0) {
		*value = 3003;
		return 1;
	}
	if (strcmp(name, "tcmalloc.central_cache_free_bytes") == 0) {
		*value = 11;
		return 1;
	}
	if (strcmp(name, "tcmalloc.transfer_cache_free_bytes") == 0) {
		*value = 12;
		return 1;
	}
	if (strcmp(name, "tcmalloc.thread_cache_free_bytes") == 0) {
		*value = 13;
		return 1;
	}
	if (strcmp(name, "tcmalloc.pageheap_free_bytes") == 0) {
		*value = 4004;
		return 1;
	}
	if (strcmp(name, "tcmalloc.pageheap_unmapped_bytes") == 0) {
		*value = 5;
		return 1;
	}
	if (strcmp(name, "tcmalloc.pageheap_committed_bytes") == 0) {
		*value = 14;
		return 1;
	}
	if (strcmp(name, "tcmalloc.slack_bytes") == 0) {
		*value = 15;
		return 1;
	}
	if (strcmp(name, "tcmalloc.max_total_thread_cache_bytes") == 0) {
		*value = 16;
		return 1;
	}
	if (strcmp(name, "tcmalloc.pageheap_scavenge_count") == 0) {
		*value = 6;
		return 1;
	}
	if (strcmp(name, "tcmalloc.pageheap_commit_count") == 0) {
		*value = 17;
		return 1;
	}
	if (strcmp(name, "tcmalloc.pageheap_decommit_count") == 0) {
		*value = 18;
		return 1;
	}
	if (strcmp(name, "tcmalloc.pageheap_total_commit_bytes") == 0) {
		*value = 19;
		return 1;
	}
	return 0;
}

double MallocExtension_GetMemoryReleaseRate(void) {
	return 1.25;
}
