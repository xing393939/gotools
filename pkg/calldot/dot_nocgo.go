//go:build !cgo
// +build !cgo

package calldot

func runDotToImage(outfname string, format string, dot []byte) (string, error) {
	return runDotToImageCallSystemGraphviz(outfname, format, dot)
}
