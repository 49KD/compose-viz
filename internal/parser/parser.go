package parser

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)


type ComposeFile struct {
	Version  string                    `yaml:"version"`
	Services map[string]ServiceConfig  `yaml:"services"`
	Volumes  map[string]any            `yaml:"volumes"`
	Networks map[string]any            `yaml:"networks"`
}

type ServiceConfig struct {
	Image         string            `yaml:"image"`
	Build         any               `yaml:"build"`
	Platform      string            `yaml:"platform"`
	ContainerName string            `yaml:"container_name"`
	Ports         []string          `yaml:"ports"`
	Expose        []string          `yaml:"expose"`
	Links         []string          `yaml:"links"`
	Command       []string          `yaml:"command"`
	Restart       string            `yaml:"restart"`
	Healthcheck   map[string]any    `yaml:"healthcheck"`
	DependsOn     any               `yaml:"depends_on"`
	Volumes       []string          `yaml:"volumes"`
	Networks      []string          `yaml:"networks"`
	Environment   map[string]string `yaml:"environment"`
}


func ParseFile(filename string) *ComposeFile {
	data, err := os.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	composeFile := ComposeFile{}
	err = yaml.Unmarshal(data, &composeFile)
	if err != nil {
		log.Fatalf("Parsing .yaml file %v error: %v", filename, err)
	}
	fmt.Printf("Compose file version: %v", composeFile.Version)
	fmt.Printf("Compose file DependsOn: %v", composeFile.Services["create-server"].DependsOn)
	return &composeFile
}
