package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ResourceKind 定义了支持的资源类型
type ResourceKind string

const (
	KindVM               ResourceKind = "VirtualMachine"
	KindVolume           ResourceKind = "Volume"
	KindImage            ResourceKind = "Image"
	KindInstanceOffering ResourceKind = "InstanceOffering"
	KindL3Network        ResourceKind = "L3Network"
	// 添加更多资源类型...
)

// ResourceSpec 表示通用资源规范
type ResourceSpec struct {
	Kind       ResourceKind           `json:"kind" yaml:"kind"`
	APIVersion string                 `json:"apiVersion" yaml:"apiVersion"`
	Metadata   ResourceMetadata       `json:"metadata" yaml:"metadata"`
	Spec       map[string]interface{} `json:"spec" yaml:"spec"`
}

// ResourceMetadata 包含资源元数据
type ResourceMetadata struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	UUID        string            `json:"uuid" yaml:"uuid"`
	Tags        []string          `json:"tags" yaml:"tags"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
}

// ProcessFile 处理单个文件并创建相应的资源
func ProcessFile(filePath string) error {
	// 读取文件内容
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	// 检测文件类型
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

	// 验证必要字段
	if resource.Kind == "" {
		return fmt.Errorf("missing 'kind' field in %s", filePath)
	}

	if resource.Metadata.Name == "" {
		return fmt.Errorf("missing 'metadata.name' field in %s", filePath)
	}

	// 根据资源类型创建相应的资源
	return nil
	//CreateResourceFromSpec(resource)
}

// ProcessDirectory 处理目录中的所有配置文件
func ProcessDirectory(dirPath string) error {
	// 获取目录中的所有文件
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %v", dirPath, err)
	}

	// 记录错误，但继续处理其他文件
	var errors []string

	// 处理每个文件
	for _, file := range files {
		// 忽略目录和隐藏文件
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		// 只处理 YAML 和 JSON 文件
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

	// 如果有错误，返回组合错误信息
	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors:\n%s", len(errors), strings.Join(errors, "\n"))
	}

	return nil
}

// ProcessFilePath 处理文件路径，可以是单个文件或目录
func ProcessFilePath(path string) error {
	// 检查路径是否存在
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error accessing path %s: %v", path, err)
	}

	// 如果是目录，处理目录中的所有文件
	if fileInfo.IsDir() {
		return ProcessDirectory(path)
	}

	// 处理单个文件
	return ProcessFile(path)
}
