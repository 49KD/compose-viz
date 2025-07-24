package graph

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/49KD/compose-viz/internal/parser"
	"github.com/emicklei/dot"
)

const portEntityTemplate string = `<br align="left"/>`
const volumeEntityTemplate string = `<br align="left"/>`

type RenderOptions struct {
	RenderVolumes      bool
	GraphTitle         string
	NodeTemplatePath   string
	VolumeTemplatePath string
}

type ServiceToRender struct {
	ServiceName   string
	ContainerName string
	ImageName     string
	Ports         []string
	PortsString   string
	DependsOn     any
	VolumesString string
}

type VolumeToRender struct {
	VolumeName  string
	Mountpoints []string
}

type serviceNodePair struct {
	service ServiceToRender
	node    *dot.Node
}

func (s *ServiceToRender) renderPorts() string {
	var b strings.Builder
	for _, port := range s.Ports {
		b.WriteString(port + portEntityTemplate)
	}
	s.PortsString = b.String()
	return s.PortsString
}

func (s *ServiceToRender) renderVolumes(
	service *parser.ServiceConfig,
	namedVolumes map[string]struct{},
) string {
	var b strings.Builder
	for _, volume := range service.Volumes {
		parts := strings.SplitN(volume, ":", 2)
		if len(parts) < 2 {
			continue // possibly anonymous volume
		}
		source := parts[0]
		if _, ok := namedVolumes[source]; !ok {
			b.WriteString(volume + volumeEntityTemplate)
		}
	}
	s.VolumesString = b.String()
	return s.VolumesString
}

func renderNamedVolumes(
	volumeNodeTmpl string,
	graph *dot.Graph,
	composeFile *parser.ComposeFile,
	volumesUsage map[string][]string,
	servicesNodes map[string]serviceNodePair,
) {
	tmpl := template.Must(template.ParseFiles(volumeNodeTmpl))
	volGraph := graph.Subgraph("volumes_cluster", dot.ClusterOption{})
	volGraph.Attr("rank", "min")
	volGraph.Attr("rankdir", "TB")
	volGraph.Attr("label", "")
	for volumeName := range composeFile.Volumes {
		v := VolumeToRender{
			VolumeName: volumeName,
			Mountpoints: []string{},
		}
		for _, service := range volumesUsage[volumeName] {
			v.Mountpoints = append(v.Mountpoints, service)
		}
		var buffer bytes.Buffer
		tmpl.Execute(&buffer, v)

		node := volGraph.Node(volumeName)
		node.Attr("label", dot.HTML(buffer.String()))
		node.Attr("shape", "plane")
		for _, mp := range v.Mountpoints {
			toNode := *servicesNodes[mp].node
			edge := graph.Edge(node, toNode)
			edge.Attr("arrowhead", "vee")
			if len(v.Mountpoints) == 1 {
				edge.Attr("constraint", "true")
			} else {
				edge.Attr("constraint", "false")
			}
		}
	}
}

var nodesServicesMap = make(map[string]serviceNodePair)

func setEdges(graph *dot.Graph, nMap map[string]serviceNodePair){
	for _, nodeServicePair := range nMap {
		switch dependsOnBlock := nodeServicePair.service.DependsOn.(type) {
		case []any:
			for _, dependency := range dependsOnBlock {
				if name, ok := dependency.(string); ok {
					toNode := *nMap[name].node
					graph.Edge(*nodeServicePair.node, toNode)

				}
			}
		case map[string]any:
			for dependency, conditionBlock := range dependsOnBlock {
				condLabel := "?"
				if cbMap, ok := conditionBlock.(map[string]any); ok {
					if rawCond, exists := cbMap["condition"]; exists {
						if condStr, ok := rawCond.(string); ok {
							condLabel = condStr
						}
					}
				}
				toNode := *nMap[dependency].node
				edge := graph.Edge(*nodeServicePair.node, toNode)
				switch condLabel {
				case "service_healthy":
					edge.Dashed()
				case "service_completed_successfully":
					edge.Dotted()
				}
			}
		}
	}
}

func extractNamedVolumes(
	serviceName string,
	service *parser.ServiceConfig,
	namedVolumes map[string]struct{},
	volumesUsage map[string][]string,
) {
	for _, v := range service.Volumes {
		parts := strings.SplitN(v, ":", 2)
		if len(parts) == 2 && !strings.HasPrefix(parts[0], "/") && !strings.HasPrefix(parts[0], ".") {
			namedVolumes[v] = struct{}{}
			volumesUsage[parts[0]] = append(volumesUsage[parts[0]], serviceName)
		}
	}
}

func setGraphAttrs(graph *dot.Graph, title string) {
	nodeStyle := map[string]string{
		"nodesep":  "1",
		"style":    "filled",
		"pencolor": "#00000044",
		"fontname": "Helvetica,Arial,sans-serif",
		"shape":    "plaintext",
		"rankdir":  "TB",
	}
	for attr, value := range nodeStyle {
		graph.Attr(attr, value)
	}
	if title != "defGraphTitle" {
		graph.Attr("label", title)
	}
}


func RenderGraph(file *parser.ComposeFile, opts RenderOptions) string {
	tmpl := template.Must(template.ParseFiles(opts.NodeTemplatePath))

	mainGraph := dot.NewGraph(dot.Directed)

	setGraphAttrs(mainGraph, opts.GraphTitle)

	networkClusters := make(map[string]*dot.Graph)

	namedVolumes := map[string]struct{}{}
	volumesUsage := make(map[string][]string)

	for serviceName, serviceAttrs := range file.Services {
		image := serviceAttrs.Image
		if image == "" {
			image = "N/A"
		}
		service := ServiceToRender{
			ServiceName: serviceName,
			ContainerName: serviceAttrs.ContainerName,
			ImageName: image,
			Ports: serviceAttrs.Ports,
			PortsString: "",
			DependsOn: serviceAttrs.DependsOn,
			VolumesString: "",
		}
		service.renderPorts()
		if opts.RenderVolumes {
			extractNamedVolumes(
				serviceName,
				&serviceAttrs,
				namedVolumes,
				volumesUsage,
			)
		}
		service.renderVolumes(&serviceAttrs, namedVolumes)
		var buffer bytes.Buffer
		tmpl.Execute(&buffer, service)

		var parentGraph *dot.Graph
		networks := serviceAttrs.Networks
		switch len(networks){
		case 1:
			network := networks[0]
			if _, exists := networkClusters[network]; !exists {
				cluster := mainGraph.Subgraph("cluster_" + network, dot.ClusterOption{})
				cluster.Attr("label", network)
				cluster.Attr("style", "dashed")
				networkClusters[network] = cluster
			}
			parentGraph = networkClusters[network]
		default:
			parentGraph = mainGraph
		}

		node := parentGraph.Node(serviceName)
		node.Attr("label", dot.HTML(buffer.String()))
		node.Attr("style", "filled")
		node.Attr("shape", "plain")

		nodesServicesMap[serviceName] = serviceNodePair{service, &node}
	}
	setEdges(mainGraph, nodesServicesMap)
	if opts.RenderVolumes {
		renderNamedVolumes(
			opts.VolumeTemplatePath,
			mainGraph,
			file,
			volumesUsage,
			nodesServicesMap,
		)
	}
	return mainGraph.String()
}
