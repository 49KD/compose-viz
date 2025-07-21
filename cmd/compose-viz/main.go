package main

import (
	"os"
	"flag"
	"log/slog"
	"github.com/49KD/compose-viz/internal/parser"
	"github.com/49KD/compose-viz/internal/graph"
)

const defaultOutFile string = "composeGraph.dot"
const defaultNodeTemplate string = "html/template/default_node.html"
const defaultVolumeTemplate string = "html/template/default_volume.html"

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
	var graphTitle = flag.String("t", "defGraphTitle", "Title to be displayed on rendered graph")
	var nodeTemplate = flag.String("n", defaultNodeTemplate, "HTML template to be used as node label")
	var renderVolumes = flag.Bool("render-volumes", true, "Render volumes as separate nodes")

	var verbose = flag.Bool("v", false, "Enable verbose logging")

	flag.Parse()

	setLogging(*verbose)

	slog.Debug("Trying to parse a file", "filename", *filePath)
	composeFile := parser.ParseFile(*filePath)

	slog.Debug("Rendering file into dot-graph file")
	opts := graph.RenderOptions{
		RenderVolumes: *renderVolumes,
		GraphTitle: *graphTitle,
		NodeTemplatePath: *nodeTemplate,
		VolumeTemplatePath: defaultVolumeTemplate,
	}
	dotString := graph.RenderGraph(composeFile, opts)

	f, err := os.Create(*outFilePath)
	if err != nil{
		panic(err)
	}
	defer f.Close()

	slog.Debug("Writing graph into file", "filename", *outFilePath)
	f.WriteString(dotString)
}
