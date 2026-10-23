//go:build !cgo || !unix

package dlopen

import "fmt"

// Open reports that this helper needs cgo on unix.
func Open(path string) error {
	return fmt.Errorf("dlopen %s: requires cgo on unix", path)
}
