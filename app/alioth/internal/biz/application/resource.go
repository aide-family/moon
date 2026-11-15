package application

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
	"sigs.k8s.io/yaml"
)

const (
	maxResources     = 100
	maxResourceBytes = 512 * 1024
)

var blockedKinds = map[string]struct{}{
	"Namespace": {}, "Node": {}, "Event": {}, "APIService": {}, "ComponentStatus": {},
	"TokenReview": {}, "SubjectAccessReview": {},
}

var namespacedKinds = map[string]struct{}{
	"Deployment": {}, "StatefulSet": {}, "DaemonSet": {}, "Job": {}, "CronJob": {}, "Pod": {},
	"Service": {}, "Ingress": {}, "ConfigMap": {}, "Secret": {}, "PersistentVolumeClaim": {},
	"ServiceAccount": {}, "Role": {}, "RoleBinding": {}, "HorizontalPodAutoscaler": {},
	"PodDisruptionBudget": {}, "NetworkPolicy": {}, "ReplicaSet": {}, "Endpoints": {},
}

var clusterKinds = map[string]struct{}{
	"CustomResourceDefinition": {}, "ClusterRole": {}, "ClusterRoleBinding": {},
	"PersistentVolume": {}, "StorageClass": {}, "PriorityClass": {}, "IngressClass": {},
}

// ParseResources 解析发布包的资源清单。返回值里的 Manifest 已去掉 Secret 值与
// 服务器字段；secretValues 以 "resourceItem/key" 为键，交给数据层加密。
func ParseResources(inputs []bo.ApplicationResource) ([]bo.ApplicationResource, map[string]string, []string, []string, error) {
	if len(inputs) == 0 {
		return nil, nil, nil, nil, merr.Invalid("at least one resource is required")
	}
	if len(inputs) > maxResources {
		return nil, nil, nil, nil, merr.Invalid("too many resources")
	}
	resources := make([]bo.ApplicationResource, 0, len(inputs))
	secrets := map[string]string{}
	errors := []string{}
	warnings := []string{}
	seen := map[string]struct{}{}
	for index, input := range inputs {
		item := strings.TrimSpace(input.Item)
		if item == "" {
			item = fmt.Sprintf("r-%d", index+1)
		}
		if len(input.Manifest) > maxResourceBytes {
			errors = append(errors, fmt.Sprintf("%s: resource is too large", item))
			continue
		}
		var manifest map[string]any
		if err := yaml.Unmarshal([]byte(input.Manifest), &manifest); err != nil {
			errors = append(errors, fmt.Sprintf("%s: invalid yaml: %v", item, err))
			continue
		}
		apiVersion, _ := manifest["apiVersion"].(string)
		kind, _ := manifest["kind"].(string)
		if apiVersion == "" || kind == "" {
			errors = append(errors, fmt.Sprintf("%s: apiVersion and kind are required", item))
			continue
		}
		if _, blocked := blockedKinds[kind]; blocked {
			errors = append(errors, fmt.Sprintf("%s: kind %s is not allowed", item, kind))
			continue
		}
		metadata, _ := manifest["metadata"].(map[string]any)
		if metadata == nil {
			errors = append(errors, fmt.Sprintf("%s: metadata is required", item))
			continue
		}
		name, _ := metadata["name"].(string)
		if name == "" {
			errors = append(errors, fmt.Sprintf("%s: metadata.name is required", item))
			continue
		}
		if generateName, _ := metadata["generateName"].(string); generateName != "" {
			errors = append(errors, fmt.Sprintf("%s: generateName is not supported", item))
			continue
		}
		namespace, _ := metadata["namespace"].(string)
		scope := "UNKNOWN"
		if _, ok := namespacedKinds[kind]; ok {
			scope = "NAMESPACED"
		} else if _, ok := clusterKinds[kind]; ok {
			scope = "CLUSTER"
		}
		if scope == "CLUSTER" && namespace != "" {
			errors = append(errors, fmt.Sprintf("%s: cluster-scoped resource cannot set namespace", item))
			continue
		}
		key := fmt.Sprintf("%s|%s|%s|%s", apiVersion, kind, namespace, name)
		if _, ok := seen[key]; ok {
			errors = append(errors, fmt.Sprintf("%s: duplicate resource %s/%s", item, kind, name))
			continue
		}
		seen[key] = struct{}{}
		cleanServerFields(manifest)
		if kind == "Secret" {
			values := extractSecretValues(item, manifest)
			for key, value := range values {
				secrets[item+"/"+key] = value
			}
		}
		images := scanImages(manifest)
		resource := bo.ApplicationResource{
			Item: item, APIVersion: apiVersion, Kind: kind, Name: name, Namespace: namespace,
			Scope: scope, ClusterScoped: scope == "CLUSTER", Images: images,
			SecretKeys: secretKeys(item, secrets),
		}
		raw, err := yaml.Marshal(manifest)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: cannot serialize resource", item))
			continue
		}
		resource.Manifest = string(raw)
		resources = append(resources, resource)
	}
	if len(errors) > 0 {
		return resources, secrets, errors, warnings, nil
	}
	return resources, secrets, nil, warnings, nil
}

func cleanServerFields(manifest map[string]any) {
	delete(manifest, "status")
	metadata, _ := manifest["metadata"].(map[string]any)
	if metadata == nil {
		return
	}
	for _, key := range []string{"uid", "resourceVersion", "creationTimestamp", "generation", "managedFields", "selfLink"} {
		delete(metadata, key)
	}
}

func extractSecretValues(item string, manifest map[string]any) map[string]string {
	values := map[string]string{}
	if data, ok := manifest["data"].(map[string]any); ok {
		for key, raw := range data {
			text, ok := raw.(string)
			if !ok {
				continue
			}
			if decoded, err := base64.StdEncoding.DecodeString(text); err == nil {
				values[key] = string(decoded)
			}
		}
	}
	if data, ok := manifest["stringData"].(map[string]any); ok {
		for key, raw := range data {
			if text, ok := raw.(string); ok {
				values[key] = text
			}
		}
	}
	delete(manifest, "data")
	delete(manifest, "stringData")
	return values
}

func secretKeys(item string, secrets map[string]string) []string {
	prefix := item + "/"
	keys := make([]string, 0)
	for key := range secrets {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, strings.TrimPrefix(key, prefix))
		}
	}
	return keys
}

func scanImages(manifest map[string]any) []string {
	images := make([]string, 0)
	var walk func(value any)
	walk = func(value any) {
		switch typed := value.(type) {
		case map[string]any:
			if image, ok := typed["image"].(string); ok && strings.TrimSpace(image) != "" {
				images = append(images, strings.TrimSpace(image))
			}
			for _, child := range typed {
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(manifest)
	return unique(images)
}

func unique(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
