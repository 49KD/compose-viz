package main

import (
	"os"
	"flag"
	"github.com/49KD/compose-viz/internal/parser"
	"github.com/49KD/compose-viz/internal/graph"
)

func main() {
	var filePathFlag = flag.String("f", "foo", "help message for flag f")
	flag.Parse()
	composeFile := parser.ParseFile(*filePathFlag)
	dotString := graph.RenderGraph(composeFile)

	f, err := os.Create("cmp_graph")
	if err != nil{
		panic(err)
	}
	defer f.Close()

	f.WriteString(dotString)
}
