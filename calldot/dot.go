package calldot

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"golang.org/x/tools/go/buildutil"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	minlen    uint
	nodesep   float64
	nodeshape string
	nodestyle string
	rankdir   string
)

func init() {
	flag.Var((*buildutil.TagsFlag)(&build.Default.BuildTags), "tags", buildutil.TagsFlagDoc)
	// Graphviz options
	flag.UintVar(&minlen, "minlen", 2, "Minimum edge length (for wider output).")
	flag.Float64Var(&nodesep, "nodesep", 0.35, "Minimum space between two adjacent nodes in the same rank (for taller output).")
	flag.StringVar(&nodeshape, "nodeshape", "box", "graph node shape (see graphvis manpage for valid values)")
	flag.StringVar(&nodestyle, "nodestyle", "filled,rounded", "graph node style (see graphvis manpage for valid values)")
	flag.StringVar(&rankdir, "rankdir", "LR", "Direction of graph layout [LR | RL | TB | BT]")
}

const tmplCluster = `{{define "cluster" -}}
    {{printf "subgraph %q {" .}}
        {{printf "%s" .Attrs.Lines}}
        {{range .Nodes}}
        {{template "node" .}}
        {{- end}}
        {{range .Clusters}}
        {{template "cluster" .}}
        {{- end}}
    {{println "}" }}
{{- end}}`

const tmplNode = `{{define "edge" -}}
    {{printf "%q -> %q [ %s ]" .From .To .Attrs}}
{{- end}}`

const tmplEdge = `{{define "node" -}}
    {{printf "%q [ %s ]" .ID .Attrs}}
{{- end}}`

const tmplGraph = `digraph gocallvis {
    label="{{.Title}}";
    labeljust="l";
    fontname="Arial";
    fontsize="14";
    rankdir="{{.Options.rankdir}}";
    bgcolor="lightgray";
    style="solid";
    penwidth="0.5";
    pad="0.0";
    nodesep="{{.Options.nodesep}}";

    node [shape="{{.Options.nodeshape}}" style="{{.Options.nodestyle}}" fillcolor="honeydew" fontname="Verdana" penwidth="1.0" margin="0.05,0.0"];
    edge [minlen="{{.Options.minlen}}"]

    {{template "cluster" .Cluster}}

    {{- range .Edges}}
    {{template "edge" .}}
    {{- end}}
}
`

//==[ type def/func: dotCluster ]===============================================
type dotCluster struct {
	ID       string
	Clusters map[string]*dotCluster
	Nodes    []*DotNode
	Attrs    dotAttrs
}

func NewDotCluster(id string) *dotCluster {
	return &dotCluster{
		ID:       id,
		Clusters: make(map[string]*dotCluster),
		Attrs:    make(dotAttrs),
	}
}

func (c *dotCluster) SetAttrs(attrs map[string]string) {
	c.Attrs = attrs
}

func (c *dotCluster) String() string {
	return fmt.Sprintf("cluster_%s", c.ID)
}

//==[ type def/func: DotNode    ]===============================================
type DotNode struct {
	ID    string
	Attrs dotAttrs
}

func (n *DotNode) String() string {
	return n.ID
}

//==[ type def/func: DotEdge    ]===============================================
type DotEdge struct {
	From  *DotNode
	To    *DotNode
	Attrs dotAttrs
}

//==[ type def/func: dotAttrs   ]===============================================
type dotAttrs map[string]string

func (p dotAttrs) List() []string {
	l := []string{}
	for k, v := range p {
		l = append(l, fmt.Sprintf("%s=%q", k, v))
	}
	return l
}

func (p dotAttrs) String() string {
	return strings.Join(p.List(), " ")
}

func (p dotAttrs) Lines() string {
	return fmt.Sprintf("%s;", strings.Join(p.List(), ";\n"))
}

//==[ type def/func: dotGraph   ]===============================================
type dotGraph struct {
	Title   string
	Minlen  uint
	Attrs   dotAttrs
	Cluster *dotCluster
	Nodes   []*DotNode
	Edges   []*DotEdge
	Options map[string]string
}

func NewDotGraph(title string, c *dotCluster, nodes []*DotNode, edges []*DotEdge) *dotGraph {
	return &dotGraph{
		Title:   title,
		Cluster: c,
		Nodes:   nodes,
		Edges:   edges,
		Minlen:  minlen,
		Options: map[string]string{
			"minlen":    fmt.Sprint(minlen),
			"nodesep":   fmt.Sprint(nodesep),
			"nodeshape": fmt.Sprint(nodeshape),
			"nodestyle": fmt.Sprint(nodestyle),
			"rankdir":   fmt.Sprint(rankdir),
		},
	}
}

func (g *dotGraph) WriteDot(w io.Writer) error {
	t := template.New("dot")
	for _, s := range []string{tmplCluster, tmplNode, tmplEdge, tmplGraph} {
		if _, err := t.Parse(s); err != nil {
			return err
		}
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, g); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}

func DotToImage(outfname string, format string, dot []byte) (string, error) {
	return runDotToImage(outfname, format, dot)
}

// location of dot executable for converting from .dot to .svg
// it's usually at: /usr/bin/dot
var dotSystemBinary string

// runDotToImageCallSystemGraphviz generates a SVG using the 'dot' utility, returning the filepath
func runDotToImageCallSystemGraphviz(outfname string, format string, dot []byte) (string, error) {
	if dotSystemBinary == "" {
		dot, err := exec.LookPath("dot")
		if err != nil {
			log.Fatalln("unable to find program 'dot', please install it or check your PATH")
		}
		dotSystemBinary = dot
	}

	var img string
	if outfname == "" {
		img = filepath.Join(os.TempDir(), fmt.Sprintf("go-callvis_export.%s", format))
	} else {
		img = fmt.Sprintf("%s.%s", outfname, format)
	}
	cmd := exec.Command(dotSystemBinary, fmt.Sprintf("-T%s", format), "-o", img)
	cmd.Stdin = bytes.NewReader(dot)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command '%v': %v\n%v", cmd, err, stderr.String())
	}
	return img, nil
}
