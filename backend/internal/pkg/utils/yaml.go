package utils

import (
	"encoding/base64"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/yaml"
)

var decoder runtime.Decoder

func init() {
	decoder = serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()
}

// ParseYAMLOrBase64 解析 YAML 或 Base64 编码的内容
func ParseYAMLOrBase64(yaml, base64Content string) ([]byte, error) {
	var content []byte

	if yaml != "" {
		content = []byte(yaml)
	} else if base64Content != "" {
		decoded, err := base64.StdEncoding.DecodeString(base64Content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 content: %w", err)
		}
		content = decoded
	} else {
		return nil, fmt.Errorf("either yaml or content field is required")
	}

	return content, nil
}

// DecodeYAML 将 YAML 内容解码为 Kubernetes 对象
func DecodeYAML(content []byte) (runtime.Object, error) {
	obj, _, err := decoder.Decode(content, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}
	return obj, nil
}

// EncodeToYAML 将 Kubernetes 对象编码为干净的 YAML 字符串
// apiVersion 和 kind 需要显式传入，因为从 API 获取的对象通常不包含这些字段
func EncodeToYAML(obj interface{}, apiVersion, kind string) (string, error) {
	// 先转为 map，以便操作
	jsonData, err := yaml.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal object: %w", err)
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(jsonData, &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// 清理系统自动添加的字段
	cleanSystemFields(data)

	// 设置 apiVersion 和 kind（从 API 获取的对象通常缺少这些字段）
	data["apiVersion"] = apiVersion
	data["kind"] = kind

	// 重新编码为 YAML，按照固定顺序输出
	result, err := marshalYAMLOrdered(data)
	if err != nil {
		return "", fmt.Errorf("failed to encode to YAML: %w", err)
	}
	return result, nil
}

// cleanSystemFields 清理 Kubernetes 系统自动添加的字段
func cleanSystemFields(data map[string]interface{}) {
	// 删除 status 字段
	delete(data, "status")

	// 清理 metadata 中的系统字段
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		delete(metadata, "creationTimestamp")
		delete(metadata, "resourceVersion")
		delete(metadata, "uid")
		delete(metadata, "generation")
		delete(metadata, "managedFields")
		delete(metadata, "selfLink")
		delete(metadata, "finalizers")
		delete(metadata, "ownerReferences")
	}
}

// marshalYAMLOrdered 按照 Kubernetes 惯例的顺序输出 YAML
func marshalYAMLOrdered(data map[string]interface{}) (string, error) {
	// 定义输出顺序
	orderedKeys := []string{"apiVersion", "kind", "metadata", "spec", "data"}

	var result string

	// 先按顺序输出已知字段
	for _, key := range orderedKeys {
		if val, ok := data[key]; ok {
			part, err := yaml.Marshal(map[string]interface{}{key: val})
			if err != nil {
				return "", err
			}
			result += string(part)
			delete(data, key)
		}
	}

	// 输出剩余字段
	if len(data) > 0 {
		part, err := yaml.Marshal(data)
		if err != nil {
			return "", err
		}
		result += string(part)
	}

	return result, nil
}
