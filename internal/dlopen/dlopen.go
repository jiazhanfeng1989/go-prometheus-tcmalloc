//go:build cgo && unix

// Package dlopen is a test helper that loads a shared library into the process.
// It is not part of the collector API.
package dlopen

/*
#cgo linux LDFLAGS: -ldl

#include <dlfcn.h>
#include <stdlib.h>
#include <string.h>

static char *open_global(const char *path) {
	dlerror();
	if (dlopen(path, RTLD_NOW | RTLD_GLOBAL) != NULL) {
		return NULL;
	}
	const char *err = dlerror();
	if (err == NULL) {
		err = "dlopen failed";
	}
	return strdup(err);
}
*/
import "C"
import (
	"os"
	"unsafe"
)

// Open loads path with RTLD_GLOBAL so a later dlsym(RTLD_DEFAULT) can see its symbols.
func Open(path string) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	cerr := C.open_global(cpath)
	if cerr == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(cerr))
	return &os.PathError{Op: "dlopen", Path: path, Err: stringError(C.GoString(cerr))}
}

type stringError string

func (e stringError) Error() string { return string(e) }
