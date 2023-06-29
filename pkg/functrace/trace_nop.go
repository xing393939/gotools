//go:build !trace
// +build !trace

package functrace

func Trace() func() {
	return func() {

	}
}
