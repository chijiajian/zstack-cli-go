// Copyright 2025 zstack.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ResourceKind string

const (
	KindVM               ResourceKind = "VirtualMachine"
	KindVolume           ResourceKind = "Volume"
	KindImage            ResourceKind = "Image"
	KindInstanceOffering ResourceKind = "InstanceOffering"
	KindL3Network        ResourceKind = "L3Network"
)

type ResourceSpec struct {
	Kind       ResourceKind           `json:"kind" yaml:"kind"`
	APIVersion string                 `json:"apiVersion" yaml:"apiVersion"`
	Metadata   ResourceMetadata       `json:"metadata" yaml:"metadata"`
	Spec       map[string]interface{} `json:"spec" yaml:"spec"`
}

type ResourceMetadata struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	UUID        string            `json:"uuid" yaml:"uuid"`
	Tags        []string          `json:"tags" yaml:"tags"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
}

func ProcessFile(filePath string) error {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	var resource ResourceSpec
	ext := strings.ToLower(filepath.Ext(filePath))

	if ext == ".json" {
		if err := json.Unmarshal(data, &resource); err != nil {
			return fmt.Errorf("error parsing JSON file %s: %v", filePath, err)
		}
	} else if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &resource); err != nil {
			return fmt.Errorf("error parsing YAML file %s: %v", filePath, err)
		}
	} else {
		return fmt.Errorf("unsupported file format: %s (must be .yaml, .yml, or .json)", ext)
	}

	if resource.Kind == "" {
		return fmt.Errorf("missing 'kind' field in %s", filePath)
	}

	if resource.Metadata.Name == "" {
		return fmt.Errorf("missing 'metadata.name' field in %s", filePath)
	}

	return nil

}

func ProcessDirectory(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %v", dirPath, err)
	}

	var errors []string

	for _, file := range files {

		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())
		if err := ProcessFile(filePath); err != nil {
			errors = append(errors, fmt.Sprintf("error processing %s: %v", filePath, err))
		} else {
			fmt.Printf("Successfully processed %s\n", filePath)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors:\n%s", len(errors), strings.Join(errors, "\n"))
	}

	return nil
}

func ProcessFilePath(path string) error {

	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error accessing path %s: %v", path, err)
	}
	if fileInfo.IsDir() {
		return ProcessDirectory(path)
	}

	return ProcessFile(path)
}
