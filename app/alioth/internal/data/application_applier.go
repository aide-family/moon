package data

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/yaml"
)

var _ repository.ApplicationApplier = (*applicationApplier)(nil)

const (
	appStoreFieldManager = "alioth-appstore"
	appStoreReadyTimeout = 5 * time.Minute
)

var applyOrder = map[string]int{
	"CustomResourceDefinition": 0,
	"ServiceAccount":           1, "ClusterRole": 2, "Role": 3, "ClusterRoleBinding": 4, "RoleBinding": 5,
	"ConfigMap": 6, "Secret": 7, "PersistentVolumeClaim": 8, "Service": 9,
	"Deployment": 10, "StatefulSet": 11, "DaemonSet": 12, "Job": 13, "CronJob": 14, "Pod": 15,
	"HorizontalPodAutoscaler": 16, "PodDisruptionBudget": 17, "NetworkPolicy": 18, "Ingress": 19,
}

var workloadKinds = map[string]struct{}{"Deployment": {}, "StatefulSet": {}, "DaemonSet": {}}

func NewApplicationApplier(deps *Dependencies, logger *slog.Logger) repository.ApplicationApplier {
	return &applicationApplier{repoBase: newBase(deps, logger), clients: deps.clients}
}

type applicationApplier struct {
	*repoBase
	clients *clusterClients
}

func (applier *applicationApplier) NamespaceExists(ctx context.Context, clusterUID, namespace string) (bool, error) {
	client, err := applier.clients.client(ctx, clusterUID)
	if err != nil {
		return false, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	_, err = client.CoreV1().Namespaces().Get(callCtx, namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, appKubeError(err)
	}
	return true, nil
}

func (applier *applicationApplier) Preflight(ctx context.Context, clusterUID, namespace string, resources []bo.ApplicationResource) ([]bo.ResourceRef, error) {
	if ok, err := applier.NamespaceExists(ctx, clusterUID, namespace); err != nil {
		return nil, err
	} else if !ok {
		return nil, merr.Invalid("target namespace does not exist")
	}
	mapper, err := applier.mapper(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	hasCRD := false
	for _, resource := range resources {
		if resource.Kind == "CustomResourceDefinition" {
			hasCRD = true
			break
		}
	}
	client, err := applier.clients.dynamicClient(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	refs := make([]bo.ResourceRef, 0, len(resources))
	for _, resource := range resources {
		mapping, err := applier.mapping(mapper, resource.APIVersion, resource.Kind)
		if err != nil {
			if hasCRD {
				// 同包 CRD 建立后 CR 才能被 discovery 识别；预检先按集群级保守处理。
				refs = append(refs, bo.ResourceRef{APIVersion: resource.APIVersion, Kind: resource.Kind, Name: resource.Name})
				continue
			}
			return nil, err
		}
		ref := bo.ResourceRef{APIVersion: resource.APIVersion, Kind: resource.Kind, Name: resource.Name}
		if mapping.Scope.Name() == "namespace" {
			if resource.Namespace != "" && resource.Namespace != namespace {
				return nil, merr.Invalid("resource namespace must match the target namespace")
			}
			ref.Namespace = namespace
		} else {
			ref.Namespace = ""
		}
		if resource.Kind != "CustomResourceDefinition" {
			if err := applier.dryRunOne(ctx, client, mapping, namespace, resource); err != nil {
				return nil, err
			}
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

func (applier *applicationApplier) Apply(ctx context.Context, clusterUID, namespace, deploymentUID, releaseUID, name string, resources []bo.ApplicationResource, secrets map[string]string) ([]bo.ResourceRef, error) {
	client, err := applier.clients.dynamicClient(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	sorted := append([]bo.ApplicationResource(nil), resources...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return resourceOrder(sorted[i].Kind) < resourceOrder(sorted[j].Kind)
	})
	mapper, err := applier.mapper(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	// 先应用 CRD，再重建 mapper，使同包 CR 能找到新的 GVK。
	crdIndex := 0
	for _, resource := range sorted {
		if resource.Kind != "CustomResourceDefinition" {
			continue
		}
		if err := applier.applyOne(ctx, client, mapper, namespace, deploymentUID, releaseUID, name, resource, secrets); err != nil {
			return nil, err
		}
		crdIndex++
	}
	if crdIndex > 0 {
		time.Sleep(2 * time.Second)
		mapper, _ = applier.mapper(ctx, clusterUID)
	}
	if crdIndex > 0 {
		if err := applier.waitCRDs(ctx, client, sorted); err != nil {
			return nil, err
		}
		mapper, _ = applier.mapper(ctx, clusterUID)
	}
	refs := make([]bo.ResourceRef, 0, len(sorted))
	for _, resource := range sorted {
		if resource.Kind == "CustomResourceDefinition" {
			refs = append(refs, bo.ResourceRef{APIVersion: resource.APIVersion, Kind: resource.Kind, Name: resource.Name})
			continue
		}
		if err := applier.applyOne(ctx, client, mapper, namespace, deploymentUID, releaseUID, name, resource, secrets); err != nil {
			return refs, err
		}
		ref := bo.ResourceRef{APIVersion: resource.APIVersion, Kind: resource.Kind, Name: resource.Name}
		if mapping, mappingErr := applier.mapping(mapper, resource.APIVersion, resource.Kind); mappingErr == nil && mapping.Scope.Name() == "namespace" {
			ref.Namespace = namespace
		}
		refs = append(refs, ref)
	}
	if err := applier.waitReady(ctx, clusterUID, namespace, sorted); err != nil {
		return refs, err
	}
	return refs, nil
}

func (applier *applicationApplier) Delete(ctx context.Context, clusterUID, releaseUID string, refs []bo.ResourceRef, keepPVC, deleteClusterScoped bool) error {
	client, err := applier.clients.dynamicClient(ctx, clusterUID)
	if err != nil {
		return err
	}
	mapper, err := applier.mapper(ctx, clusterUID)
	if err != nil {
		return err
	}
	for index := len(refs) - 1; index >= 0; index-- {
		ref := refs[index]
		if keepPVC && ref.Kind == "PersistentVolumeClaim" {
			continue
		}
		mapping, err := applier.mapping(mapper, ref.APIVersion, ref.Kind)
		if err != nil {
			continue
		}
		if !deleteClusterScoped && mapping.Scope.Name() != "namespace" {
			continue
		}
		resource := client.Resource(mapping.Resource)
		var target dynamic.ResourceInterface = resource
		if mapping.Scope.Name() == "namespace" {
			target = resource.Namespace(ref.Namespace)
		}
		object, err := target.Get(ctx, ref.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return appKubeError(err)
		}
		if object.GetLabels()["alioth.io/release-uid"] != releaseUID {
			continue
		}
		if err := target.Delete(ctx, ref.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return appKubeError(err)
		}
	}
	return nil
}

func (applier *applicationApplier) dryRunOne(ctx context.Context, client dynamic.Interface, mapping *meta.RESTMapping, namespace string, resource bo.ApplicationResource) error {
	var manifest map[string]any
	if err := yaml.Unmarshal([]byte(resource.Manifest), &manifest); err != nil {
		return merr.Invalid("invalid resource yaml")
	}
	metadata, _ := manifest["metadata"].(map[string]any)
	if metadata == nil {
		metadata = map[string]any{}
		manifest["metadata"] = metadata
	}
	if mapping.Scope.Name() == "namespace" {
		metadata["namespace"] = namespace
	} else {
		delete(metadata, "namespace")
	}
	body, err := json.Marshal(manifest)
	if err != nil {
		return merr.Invalid("invalid resource")
	}
	resourceClient := client.Resource(mapping.Resource)
	var target dynamic.ResourceInterface = resourceClient
	if mapping.Scope.Name() == "namespace" {
		target = resourceClient.Namespace(namespace)
	}
	force := true
	_, err = target.Patch(ctx, resource.Name, types.ApplyPatchType, body, metav1.PatchOptions{
		FieldManager: appStoreFieldManager, Force: &force, DryRun: []string{"All"},
	})
	if err != nil {
		return appKubeError(err)
	}
	return nil
}

func (applier *applicationApplier) applyOne(ctx context.Context, client dynamic.Interface, mapper meta.RESTMapper, namespace, deploymentUID, releaseUID, name string, resource bo.ApplicationResource, secrets map[string]string) error {
	mapping, err := applier.mapping(mapper, resource.APIVersion, resource.Kind)
	if err != nil {
		return err
	}
	var manifest map[string]any
	if err := yaml.Unmarshal([]byte(resource.Manifest), &manifest); err != nil {
		return merr.Invalid("invalid resource yaml")
	}
	metadata, _ := manifest["metadata"].(map[string]any)
	if metadata == nil {
		metadata = map[string]any{}
		manifest["metadata"] = metadata
	}
	labels, _ := metadata["labels"].(map[string]any)
	if labels == nil {
		labels = map[string]any{}
	}
	labels["app.kubernetes.io/managed-by"] = "alioth"
	labels["alioth.io/deployment-uid"] = deploymentUID
	labels["alioth.io/release-uid"] = releaseUID
	labels["alioth.io/release-name"] = name
	metadata["labels"] = labels
	if resource.Kind == "Secret" {
		data := map[string]any{}
		prefix := resource.Item + "/"
		for key, value := range secrets {
			if strings.HasPrefix(key, prefix) {
				data[strings.TrimPrefix(key, prefix)] = base64.StdEncoding.EncodeToString([]byte(value))
			}
		}
		if len(data) > 0 {
			manifest["data"] = data
		}
	}
	if mapping.Scope.Name() == "namespace" {
		metadata["namespace"] = namespace
	} else {
		delete(metadata, "namespace")
	}
	body, err := json.Marshal(manifest)
	if err != nil {
		return merr.Invalid("invalid resource")
	}
	resourceClient := client.Resource(mapping.Resource)
	var target dynamic.ResourceInterface = resourceClient
	if mapping.Scope.Name() == "namespace" {
		target = resourceClient.Namespace(namespace)
	}
	existing, err := target.Get(ctx, resource.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return appKubeError(err)
	}
	if err == nil && existing.GetLabels()["alioth.io/release-uid"] != releaseUID {
		return merr.Conflict("resource already exists and is not owned by this release")
	}
	force := true
	_, err = target.Patch(ctx, resource.Name, types.ApplyPatchType, body, metav1.PatchOptions{FieldManager: appStoreFieldManager, Force: &force})
	if err != nil {
		return appKubeError(err)
	}
	return nil
}

// waitCRDs 等待同包 CRD 的 Established 条件，最多 30 秒。
func (applier *applicationApplier) waitCRDs(ctx context.Context, client dynamic.Interface, resources []bo.ApplicationResource) error {
	gvr := schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}
	for _, resource := range resources {
		if resource.Kind != "CustomResourceDefinition" {
			continue
		}
		established := false
		for attempt := 0; attempt < 30; attempt++ {
			object, err := client.Resource(gvr).Get(ctx, resource.Name, metav1.GetOptions{})
			if err == nil {
				status, _ := object.Object["status"].(map[string]any)
				conditions, _ := status["conditions"].([]any)
				for _, item := range conditions {
					condition, _ := item.(map[string]any)
					if condition["type"] == "Established" && condition["status"] == "True" {
						established = true
						break
					}
				}
			}
			if established {
				break
			}
			select {
			case <-ctx.Done():
				return appKubeError(ctx.Err())
			case <-time.After(time.Second):
			}
		}
		if !established {
			return merr.Invalid("CRD did not become established before timeout")
		}
	}
	return nil
}

func (applier *applicationApplier) waitReady(ctx context.Context, clusterUID, namespace string, resources []bo.ApplicationResource) error {
	client, err := applier.clients.client(ctx, clusterUID)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(appStoreReadyTimeout)
	for {
		pending := 0
		for _, resource := range resources {
			if _, ok := workloadKinds[resource.Kind]; !ok {
				continue
			}
			ready, err := workloadReady(ctx, client, namespace, resource.Kind, resource.Name)
			if err != nil {
				return err
			}
			if !ready {
				pending++
			}
		}
		if pending == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return merr.Invalid("workload did not become ready before timeout")
		}
		select {
		case <-ctx.Done():
			return merr.Invalid("deployment canceled")
		case <-time.After(3 * time.Second):
		}
	}
}

func workloadReady(ctx context.Context, client kubernetes.Interface, namespace, kind, name string) (bool, error) {
	desired := int32(1)
	switch kind {
	case "Deployment":
		item, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, appKubeError(err)
		}
		if item.Spec.Replicas != nil {
			desired = *item.Spec.Replicas
		}
		return item.Status.ObservedGeneration >= item.Generation && item.Status.ReadyReplicas >= desired, nil
	case "StatefulSet":
		item, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, appKubeError(err)
		}
		if item.Spec.Replicas != nil {
			desired = *item.Spec.Replicas
		}
		return item.Status.ObservedGeneration >= item.Generation && item.Status.ReadyReplicas >= desired, nil
	case "DaemonSet":
		item, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, appKubeError(err)
		}
		desired = item.Status.DesiredNumberScheduled
		return item.Status.ObservedGeneration >= item.Generation && item.Status.NumberReady >= desired, nil
	}
	return true, nil
}

func (applier *applicationApplier) mapper(ctx context.Context, clusterUID string) (meta.RESTMapper, error) {
	discoveryClient, err := applier.clients.discoveryClient(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	groups, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, appKubeError(err)
	}
	return restmapper.NewDiscoveryRESTMapper(groups), nil
}

func (applier *applicationApplier) mapping(mapper meta.RESTMapper, apiVersion, kind string) (*meta.RESTMapping, error) {
	if mapper == nil {
		return nil, merr.Invalid("resource mapping is not ready")
	}
	groupVersion, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, merr.Invalid("invalid apiVersion")
	}
	mapping, err := mapper.RESTMapping(groupVersion.WithKind(kind).GroupKind(), groupVersion.Version)
	if err != nil {
		return nil, merr.Invalid(fmt.Sprintf("resource type %s %s is not installed", apiVersion, kind))
	}
	return mapping, nil
}

func resourceOrder(kind string) int {
	if order, ok := applyOrder[kind]; ok {
		return order
	}
	return 100
}

func appKubeError(err error) error {
	if apierrors.IsForbidden(err) {
		return merr.Forbidden("cluster credential is not allowed to perform this operation")
	}
	if apierrors.IsUnauthorized(err) {
		return merr.Invalid("cluster credential was rejected by the cluster")
	}
	if apierrors.IsAlreadyExists(err) || apierrors.IsConflict(err) {
		return merr.Conflict("cluster resource conflict")
	}
	return merr.Server("cluster request failed")
}
