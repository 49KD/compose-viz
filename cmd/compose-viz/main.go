package main

import (
	"bytes"
	"flag"
	"log/slog"
	"os"
	"os/exec"

	"github.com/49KD/compose-viz/internal/parser"
	"github.com/49KD/compose-viz/internal/graph"
)

const (
	defaultOutFile         = "composeGraph"
	defaultNodeTemplate    = "html/template/default_node.html"
	defaultVolumeTemplate  = "html/template/default_volume.html"
)

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
	filePath := flag.String("f", "file", "Filepath for docker-compose.yml file to be processed")
	outPath := flag.String("o", defaultOutFile, "Filepath for generated dot file")
	graphTitle := flag.String("t", "defGraphTitle", "Title to be displayed on rendered graph")
	nodeTemplate := flag.String("n", defaultNodeTemplate, "HTML template to be used as node label")
	renderVolumes := flag.Bool("render-volumes", false, "Render volumes as separate nodes")
	format := flag.String("format", "dot", "Output format: dot or png")
	verbose := flag.Bool("v", false, "Enable verbose logging")

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

	switch *format {
	case "dot":
		filename := *outPath + ".dot"
		err := os.WriteFile(filename, []byte(dotString), 0644)
		if err != nil {
			panic(err)
		}
		slog.Info("DOT file written", "path", filename)

	case "png":
		cmd := exec.Command("dot", "-Tpng", "-o", *outPath+".png")
		cmd.Stdin = bytes.NewBufferString(dotString)
		if err := cmd.Run(); err != nil {
			slog.Error("dot command failed", "error", err)
			os.Exit(1)
		}
		slog.Info("PNG file rendered", "path", *outPath+".png")

	default:
		slog.Error("Unsupported format", "format", *format)
		os.Exit(1)
	}
}
