package data

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// maxResourceYAMLBytes 是单次 YAML 预览允许返回的字节上限。
	maxResourceYAMLBytes = 512 * 1024
	// redactedValue 替换 Secret / ConfigMap 中的敏感值。
	redactedValue = "<redacted>"
	// lastAppliedAnnotation 记录 kubectl apply 的完整旧配置，预览时去掉，避免把旧值带出去。
	lastAppliedAnnotation = "kubectl.kubernetes.io/last-applied-configuration"
)

// YAML 预览支持的单资源类型，和 proto 的 kind 枚举一致。
const (
	yamlKindNode                  = "Node"
	yamlKindPod                   = "Pod"
	yamlKindIngress               = "Ingress"
	yamlKindService               = "Service"
	yamlKindConfigMap             = "ConfigMap"
	yamlKindSecret                = "Secret"
	yamlKindPersistentVolume      = "PersistentVolume"
	yamlKindPersistentVolumeClaim = "PersistentVolumeClaim"
)

// ResourceYAML 读取单个资源并转成 YAML。
// Secret / ConfigMap 的值在序列化前替换成占位符，因此响应里永远没有明文。
func (r *kubernetesRepo) ResourceYAML(ctx context.Context, clusterUID, kind, namespace, name string) (*bo.KubeResourceYAML, error) {
	object, apiVersion, resourceKind, err := r.getResourceObject(ctx, clusterUID, kind, namespace, name)
	if err != nil {
		return nil, err
	}
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		r.log.Error(fmt.Sprintf("resource yaml: %v", err))
		return nil, merr.Server("request failed")
	}
	// typed client 拿到的对象不一定带 TypeMeta，显式补齐后 YAML 才能直接 kubectl apply。
	content["apiVersion"] = apiVersion
	content["kind"] = resourceKind
	redacted := redactSensitiveValues(resourceKind, content)
	stripResourceMetadata(content)
	encoded, err := yaml.Marshal(content)
	if err != nil {
		r.log.Error(fmt.Sprintf("resource yaml: %v", err))
		return nil, merr.Server("request failed")
	}
	if len(encoded) > maxResourceYAMLBytes {
		// 错误文案固定，不回传资源内容。
		return nil, merr.Client("resource yaml is too large to preview")
	}
	return &bo.KubeResourceYAML{YAML: string(encoded), Redacted: redacted}, nil
}

// getResourceObject 按资源类型取单个对象，并返回该类型对应的 apiVersion。
// Node 和 PersistentVolume 是集群级资源，不使用 namespace。
func (r *kubernetesRepo) getResourceObject(ctx context.Context, clusterUID, kind, namespace, name string) (runtime.Object, string, string, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, "", "", err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	switch kind {
	case yamlKindNode:
		object, err := client.CoreV1().Nodes().Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", kubeError(r.log, err)
		}
		return object, "v1", yamlKindNode, nil
	case yamlKindPod:
		object, err := client.CoreV1().Pods(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", kubeError(r.log, err)
		}
		return object, "v1", yamlKindPod, nil
	case workloadKindDeployment:
		object, err := client.AppsV1().Deployments(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", kubeError(r.log, err)
		}
		return object, "apps/v1", workloadKindDeployment, nil
	case workloadKindStatefulSet:
		object, err := client.AppsV1().StatefulSets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", kubeError(r.log, err)
		}
		return object, "apps/v1", workloadKindStatefulSet, nil
	case workloadKindDaemonSet:
		object, err := client.AppsV1().DaemonSets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", kubeError(r.log, err)
		}
		return object, "apps/v1", workloadKindDaemonSet, nil
	case yamlKindIngress:
		object, err := client.NetworkingV1().Ingresses(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", kubeError(r.log, err)
		}
		return object, "networking.k8s.io/v1", yamlKindIngress, nil
	case yamlKindService:
		object, err := client.CoreV1().Services(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", kubeError(r.log, err)
		}
		return object, "v1", yamlKindService, nil
	case yamlKindConfigMap:
		object, err := client.CoreV1().ConfigMaps(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", kubeError(r.log, err)
		}
		return object, "v1", yamlKindConfigMap, nil
	case yamlKindSecret:
		object, err := client.CoreV1().Secrets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", kubeError(r.log, err)
		}
		return object, "v1", yamlKindSecret, nil
	case yamlKindPersistentVolume:
		object, err := client.CoreV1().PersistentVolumes().Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", kubeError(r.log, err)
		}
		return object, "v1", yamlKindPersistentVolume, nil
	case yamlKindPersistentVolumeClaim:
		object, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", kubeError(r.log, err)
		}
		return object, "v1", yamlKindPersistentVolumeClaim, nil
	default:
		return nil, "", "", merr.Invalid("invalid resource kind")
	}
}

// redactSensitiveValues 把 Secret / ConfigMap 的值替换成占位符，返回是否替换过任何值。
// 键名保留，方便用户确认配置项是否齐全；值不进入响应。
func redactSensitiveValues(kind string, content map[string]interface{}) bool {
	switch kind {
	case yamlKindSecret:
		redacted := redactFieldValues(content, "data")
		// stringData 只用于写入，API Server 读取时不会返回；这里兜底防御。
		if redactFieldValues(content, "stringData") {
			redacted = true
		}
		return redacted
	case yamlKindConfigMap:
		redacted := redactFieldValues(content, "data")
		if redactFieldValues(content, "binaryData") {
			redacted = true
		}
		return redacted
	default:
		return false
	}
}

// redactFieldValues 把 object[field] 下所有值替换成占位符。字段不存在或为空时返回 false。
func redactFieldValues(content map[string]interface{}, field string) bool {
	values, ok := content[field].(map[string]interface{})
	if !ok || len(values) == 0 {
		return false
	}
	for key := range values {
		values[key] = redactedValue
	}
	return true
}

// stripResourceMetadata 去掉 managedFields 和 kubectl apply 的旧配置注解。
// 前者是服务端记账信息，后者可能含上一版配置值，都不属于资源预览的内容。
func stripResourceMetadata(content map[string]interface{}) {
	metadata, ok := content["metadata"].(map[string]interface{})
	if !ok {
		return
	}
	delete(metadata, "managedFields")
	annotations, ok := metadata["annotations"].(map[string]interface{})
	if !ok {
		return
	}
	delete(annotations, lastAppliedAnnotation)
	if len(annotations) == 0 {
		// 合并后的空 map 在 YAML 里会显示 annotations: {}，一并去掉。
		delete(metadata, "annotations")
	}
}
