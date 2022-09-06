[![codecov](https://codecov.io/gh/xing393939/gotools/branch/main/graph/badge.svg)](https://codecov.io/gh/xing393939/gotools)

# Go Tools

### goplantuml

```
go install github.com/xing393939/gotools/cmd/goplantuml
goplantuml path/to/dir > diagram_file_name.puml

Usage of goplantuml:
  -aggregate-private-members
        Show aggregations for private members. Ignored if -show-aggregations is not used.
  -hide-connections
        hides all connections in the diagram
  -hide-fields
        hides fields
  -hide-methods
        hides methods
  -ignore string
        comma separated list of folders to ignore
  -notes string
        Comma separated list of notes to be added to the diagram
  -build-tags string
        Comma separated list of tags for build constraint
  -output string
        output file path. If omitted, then this will default to standard output
  -recursive
        walk all directories recursively
  -show-aggregations
        renders public aggregations even when -hide-connections is used (do not render by default)
  -show-aliases
        Shows aliases even when -hide-connections is used
  -show-compositions
        Shows compositions even when -hide-connections is used
  -show-connection-labels
        Shows labels in the connections to identify the connections types (e.g. extends, implements, aggregates, alias of
  -show-implementations
        Shows implementations even when -hide-connections is used
  -show-options-as-note
        Show a note in the diagram with the none evident options ran with this CLI
  -title string
        Title of the generated diagram
  -hide-private-members
        Hides all private members (fields and methods)
```

### gocallvis

```
go install github.com/xing393939/gotools/cmd/gocallvis
gocallvis [flags] package

Usage of gocallvis:
  -algo string
    	The algorithm used to construct the call graph. Possible values inlcude: "static", "cha", "rta", "pointer" (default "pointer")
  -cacheDir string
    	Enable caching to avoid unnecessary re-rendering, you can force rendering by adding 'refresh=true' to the URL query or emptying the cache directory
  -debug
    	Enable verbose log.
  -file string
    	output filename - omit to use server mode
  -focus string
    	Focus specific package using name or import path. (default "main")
  -format string
    	output file format [svg | png | jpg | ...] (default "svg")
  -graphviz
    	Use Graphviz's dot program to render images.
  -group string
    	Grouping functions by packages and/or types [pkg, type] (separated by comma) (default "pkg")
  -http string
    	HTTP service address. (default ":7878")
  -ignore string
    	Ignore package paths containing given prefixes (separated by comma)
  -include string
    	Include package paths with given prefixes (separated by comma)
  -limit string
    	Limit package paths to given prefixes (separated by comma)
  -minlen uint
    	Minimum edge length (for wider output). (default 2)
  -nodesep float
    	Minimum space between two adjacent nodes in the same rank (for taller output). (default 0.35)
  -nodeshape string
    	graph node shape (see graphvis manpage for valid values) (default "box")
  -nodestyle string
    	graph node style (see graphvis manpage for valid values) (default "filled,rounded")
  -nointer
    	Omit calls to unexported functions.
  -nostd
    	Omit calls to/from packages in standard library.
  -rankdir string
    	Direction of graph layout [LR | RL | TB | BT] (default "LR")
  -tags build tags
    	a list of build tags to consider satisfied during the build. For more information about build tags, see the description of build constraints in the documentation for the go/build package
  -tests
    	Include test code.
  -version
    	Show version and exit.
```
