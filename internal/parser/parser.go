package parser

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultImageName string = "N/A"

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
	Environment   any               `yaml:"environment"`
}

func (config *ServiceConfig) normalizeEnvironmentBlock () {
	normalisedEnv := make(map[string]string)
	switch env := config.Environment.(type) {
	case map[string]any:
		// fine
		for key, value := range env {
			if val, ok := value.(string); ok {
				normalisedEnv[key] = val
			}
		}
	case []any:
		// process VAR=val strings manually
		for _, value := range env {
			if val, ok := value.(string); ok {
				if parts := strings.SplitN(val, "=", 2); len(parts) == 2 {
					normalisedEnv[parts[0]] = parts[1]
				} else {
					normalisedEnv[parts[0]] = `{OS inherited}`
				}
			}
		}
	}
	config.Environment = normalisedEnv
}

func (config *ServiceConfig) normalizeImageName () {
	image := config.Image
	if image == "" {
		image = defaultImageName
	}
	config.Image = image
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

	for _, serviceConfig := range composeFile.Services {
		serviceConfig.normalizeEnvironmentBlock()
		serviceConfig.normalizeImageName()
	}

	fmt.Printf("Compose file version: %v", composeFile.Version)
	return &composeFile
}
