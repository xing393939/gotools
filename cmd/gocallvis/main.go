package main

import (
	"flag"
	"fmt"
	"github.com/xing393939/gotools/callvis"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const Usage = "gocallvis [flags] package: visualize call graph of a Go program.\n\n"

var (
	focusFlag     = flag.String("focus", "main", "Focus specific package using name or import path.")
	groupFlag     = flag.String("group", "pkg", "Grouping functions by packages and/or types [pkg, type] (separated by comma)")
	limitFlag     = flag.String("limit", "", "Limit package paths to given prefixes (separated by comma)")
	ignoreFlag    = flag.String("ignore", "", "Ignore package paths containing given prefixes (separated by comma)")
	sweepFlag     = flag.String("sweep", "", "Sweep edges containing given prefixes (separated by comma)")
	includeFlag   = flag.String("include", "", "Include package paths with given prefixes (separated by comma)")
	nostdFlag     = flag.Bool("nostd", false, "Omit calls to/from packages in standard library.")
	nointerFlag   = flag.Bool("nointer", false, "Omit calls to unexported functions.")
	testFlag      = flag.Bool("tests", false, "Include test code.")
	httpFlag      = flag.String("http", ":7878", "HTTP service address.")
	outputFile    = flag.String("file", "", "output filename - omit to use server mode")
	outputFormat  = flag.String("format", "svg", "output file format [svg | png | jpg | ...]")
	cacheDir      = flag.String("cacheDir", "", "Enable caching to avoid unnecessary re-rendering, you can force rendering by adding 'refresh=true' to the URL query or emptying the cache directory")
	callgraphAlgo = flag.String("algo", callvis.CallGraphTypePointer, fmt.Sprintf("The algorithm used to construct the call graph. Possible values inlcude: %q, %q, %q, %q",
		callvis.CallGraphTypeStatic, callvis.CallGraphTypeCha, callvis.CallGraphTypeRta, callvis.CallGraphTypePointer))
	versionFlag = flag.Bool("version", false, "Show version and exit.")
)

func parseHTTPAddr(addr string) string {
	host, port, _ := net.SplitHostPort(addr)
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "80"
	}
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", host, port),
	}
	return u.String()
}

//noinspection GoUnhandledErrorResult
func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Fprintln(os.Stderr, callvis.Version())
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		fmt.Fprint(os.Stderr, Usage)
		flag.PrintDefaults()
		os.Exit(2)
	}

	args := flag.Args()
	analysisObj := new(callvis.Analysis)
	analysisObj.OptsSetup(*cacheDir, *focusFlag, *groupFlag, *ignoreFlag, *sweepFlag, *includeFlag, *limitFlag, *nointerFlag, *nostdFlag)
	if err := analysisObj.DoAnalysis(callvis.CallGraphType(*callgraphAlgo), "", *testFlag, args); err != nil {
		log.Fatal(err)
	}

	httpAddr := *httpFlag
	urlAddr := parseHTTPAddr(httpAddr)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && !strings.HasSuffix(r.URL.Path, ".svg") {
			http.NotFound(w, r)
			return
		}
		analysisObj.OptsSetup(*cacheDir, *focusFlag, *groupFlag, *ignoreFlag, *sweepFlag, *includeFlag, *limitFlag, *nointerFlag, *nostdFlag)
		analysisObj.OverrideByHTTP(r)
		img, err := analysisObj.OutputDot("output", *outputFormat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("serving file:", img)
		http.ServeFile(w, r, img)
	})

	if *outputFile == "" {
		log.Printf("http serving at %s", urlAddr)
		if err := http.ListenAndServe(httpAddr, nil); err != nil {
			log.Fatal(err)
		}
	} else {
		analysisObj.OutputDot(*outputFile, *outputFormat)
	}
}
