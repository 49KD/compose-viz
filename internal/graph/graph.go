package graph

import (
	"bytes"
	"text/template"
	"strings"
	"github.com/emicklei/dot"
	"github.com/49KD/compose-viz/internal/parser"
)


var serviceTemplate string = `
<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
 <tr> <td> <b>{{.ServiceName}}</b></td> </tr>
 <tr> <td align="left"><i>Image name: </i><br align="left"/>
{{.ImageName}}
 <br align="left"/></td></tr>
 <tr> <td align="left"><i>Container name: </i><br align="left"/>
{{.ContainerName}}
 <br align="left"/></td></tr>
 <tr> <td align="left"><i>Ports: </i><br align="left"/>{{.PortsString}}<br align="left"/></td></tr>
</table>
`

var portEntityTemplate string = `<br align="left"/>`

type ServiceToRender struct {
	ServiceName string
	ContainerName string
	ImageName string
	Ports []string
	PortsString string
	DependsOn any
}

type serviceNodePair struct {
	service ServiceToRender
	node *dot.Node
}

func (s *ServiceToRender) renderedPorts() string {
	var b strings.Builder
	for _, port := range s.Ports {
		b.WriteString(port + portEntityTemplate)
	}
	s.PortsString = b.String()
	return s.PortsString
}

var nodesServicesMap = make(map[string]serviceNodePair)


func setEdges(graph *dot.Graph, nodesMap *map[string]serviceNodePair){
	nMap := *nodesMap
	for _, nodeServicePair := range nMap {
		switch dependsOnBlock := nodeServicePair.service.DependsOn.(type) {
		case []interface{}:
			for _, dependency := range dependsOnBlock {
				if name, ok := dependency.(string); ok {
					toNode := *nMap[name].node
					graph.Edge(*nodeServicePair.node, toNode)
				}
			}
		case map[string]interface{}:
			for dependency := range dependsOnBlock {
				toNode := *nMap[dependency].node
				graph.Edge(*nodeServicePair.node, toNode)
			}
		}
	}
}


func RenderGraph(file *parser.ComposeFile) string {
	nodeStyle := map[string]string{
		"style": "filled",
		"pencolor": "#00000044",
		"fontname": "Helvetica,Arial,sans-serif",
		"shape": "plaintext",
	}
	tmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		panic(err)
	}
	graph := dot.NewGraph(dot.Directed)

	for attr, value := range nodeStyle {
		graph.Attr(attr, value)
	}

	for serviceName, serviceAttrs := range file.Services {
		image := serviceAttrs.Image
		if image == "" {
			image = "N/A"
		}
		service := ServiceToRender{
			serviceName,
			serviceAttrs.ContainerName,
			image,
			serviceAttrs.Ports,
			"",
			serviceAttrs.DependsOn,
		}
		service.renderedPorts()
		var buffer bytes.Buffer
		tmpl.Execute(&buffer, service)

		node := graph.Node(serviceName)
		node.Attr("label", dot.HTML(buffer.String()))
		node.Attr("style", "filled")
		node.Attr("shape", "plain")

		nodesServicesMap[serviceName] = serviceNodePair{service, &node}
	}
	setEdges(graph, &nodesServicesMap)
	return graph.String()
}
