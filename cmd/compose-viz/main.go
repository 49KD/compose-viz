package main

import (
	"os"
	"flag"
	"log/slog"
	"github.com/49KD/compose-viz/internal/parser"
	"github.com/49KD/compose-viz/internal/graph"
)

const defaultOutFile string = "composeGraph.dot"

func setLogging(verbose bool){
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}

func main() {
	var filePath = flag.String("f", "file", "Filepath for docker-compose.yml file to be processed")
	var outFilePath = flag.String("o", defaultOutFile, "Filepath for generated dot file")
	var verbose = flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()

	setLogging(*verbose)

	slog.Debug("Trying to parse a file", "filename", *filePath)
	composeFile := parser.ParseFile(*filePath)
	slog.Debug("Rendering file into dot-graph file")
	dotString := graph.RenderGraph(composeFile)

	f, err := os.Create(*outFilePath)
	if err != nil{
		panic(err)
	}
	defer f.Close()

	slog.Debug("Writing graph into file", "filename", *outFilePath)
	f.WriteString(dotString)
}
