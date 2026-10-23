//go:build !cgo || !unix

package tcmallocprom

func newCGOSource() source {
	return nopSource{}
}

type nopSource struct{}

func (nopSource) read() reading {
	return reading{}
}
