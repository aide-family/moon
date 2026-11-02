package data

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8stesting "k8s.io/client-go/testing"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func detailNode(name, uid string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name, UID: types.UID(uid)},
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{{Key: "dedicated", Value: "gpu", Effect: corev1.TaintEffectNoSchedule}},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{{
				Type: corev1.NodeReady, Status: corev1.ConditionTrue,
				Reason: "KubeletReady", Message: "kubelet is posting ready status",
				LastTransitionTime: metav1.Unix(1700000000, 0),
			}},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
				"nvidia.com/gpu":      resource.MustParse("2"),
			},
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("3800m"),
				corev1.ResourceMemory: resource.MustParse("7Gi"),
			},
		},
	}
}

func detailPod() *corev1.Pod {
	pod := fakePod("default", "web-1", "node-a", "ReplicaSet", "web-abc")
	pod.UID = types.UID("pod-uid")
	pod.Spec.Containers = []corev1.Container{{
		Name: "web", Image: "nginx:1.27",
		Ports: []corev1.ContainerPort{{ContainerPort: 8080, Protocol: corev1.ProtocolTCP}},
	}}
	pod.Spec.InitContainers = []corev1.Container{{Name: "init", Image: "busybox:1.36"}}
	pod.Spec.Volumes = []corev1.Volume{
		{Name: "data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "data", ReadOnly: true}}},
		{Name: "secret", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "db-secret"}}},
		{Name: "config", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: "app-config"}}}},
		{Name: "cache", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
	}
	pod.Status.Conditions = []corev1.PodCondition{{
		Type: corev1.PodReady, Status: corev1.ConditionTrue,
		LastTransitionTime: metav1.Unix(1700000100, 0),
	}}
	pod.Status.ContainerStatuses = []corev1.ContainerStatus{{
		Name: "web", Image: "nginx:1.27", Ready: true, RestartCount: 2,
		State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.Unix(1700000200, 0)}},
	}}
	pod.Status.InitContainerStatuses = []corev1.ContainerStatus{{
		Name: "init", Image: "busybox:1.36",
		State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "Completed"}},
	}}
	controller := true
	pod.OwnerReferences = []metav1.OwnerReference{{
		Kind: "ReplicaSet", Name: "web-abc", UID: types.UID("rs-uid"), Controller: &controller,
	}}
	return pod
}

func TestNodeDetail(t *testing.T) {
	repo := newFakeRepo(detailNode("node-a", "node-uid"), fakePod("default", "web-1", "node-a", "ReplicaSet", "web-abc"))
	detail, err := repo.NodeDetail(context.Background(), testClusterUID, "node-a")
	if err != nil {
		t.Fatalf("NodeDetail: %v", err)
	}
	if detail.Node.Name != "node-a" || detail.Node.UID != "node-uid" || detail.Node.PodCount != 1 {
		t.Fatalf("node = %+v", detail.Node)
	}
	if len(detail.Conditions) != 1 || detail.Conditions[0].Type != "Ready" || detail.Conditions[0].Reason != "KubeletReady" {
		t.Fatalf("conditions = %+v", detail.Conditions)
	}
	if len(detail.Taints) != 1 || detail.Taints[0].Effect != "NoSchedule" || detail.Taints[0].Value != "gpu" {
		t.Fatalf("taints = %+v", detail.Taints)
	}
	if detail.Capacity["cpu"] != "4" || detail.Capacity["nvidia.com/gpu"] != "2" {
		t.Fatalf("capacity = %+v", detail.Capacity)
	}
	if detail.Allocatable["cpu"] != "3800m" || detail.Allocatable["memory"] != "7Gi" {
		t.Fatalf("allocatable = %+v", detail.Allocatable)
	}
}

// TestNodeDetailWithoutPodPermission 确认全量 Pod 列表被拒绝时节点详情仍然成功，
// Pod 数按 0 计：Pod 数只是辅助字段，不能让受限凭据的节点页整体失败。
func TestNodeDetailWithoutPodPermission(t *testing.T) {
	clientset := fake.NewSimpleClientset(detailNode("node-a", "node-uid"))
	clientset.PrependReactor("list", "pods", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "pods"}, "", nil)
	})
	repo := newFakeRepoWithClientset(clientset)
	detail, err := repo.NodeDetail(context.Background(), testClusterUID, "node-a")
	if err != nil {
		t.Fatalf("NodeDetail: %v", err)
	}
	if detail.Node.PodCount != 0 {
		t.Fatalf("pod count = %d, want 0", detail.Node.PodCount)
	}
}

func TestNodeDetailNotFound(t *testing.T) {
	repo := newFakeRepo()
	if _, err := repo.NodeDetail(context.Background(), testClusterUID, "missing"); err == nil {
		t.Fatal("want not found error")
	}
}

func TestPodDetail(t *testing.T) {
	repo := newFakeRepo(detailPod())
	detail, err := repo.PodDetail(context.Background(), testClusterUID, "default", "web-1")
	if err != nil {
		t.Fatalf("PodDetail: %v", err)
	}
	if detail.Pod.Name != "web-1" || detail.Pod.UID != "pod-uid" {
		t.Fatalf("pod = %+v", detail.Pod)
	}
	if len(detail.Containers) != 1 {
		t.Fatalf("containers = %+v", detail.Containers)
	}
	container := detail.Containers[0]
	if container.State != "RUNNING" || container.StartedAt != 1700000200 || len(container.Ports) != 1 || container.Ports[0] != "8080/TCP" {
		t.Fatalf("container = %+v", container)
	}
	if len(detail.InitContainers) != 1 || detail.InitContainers[0].State != "TERMINATED" || detail.InitContainers[0].StateReason != "Completed" {
		t.Fatalf("init containers = %+v", detail.InitContainers)
	}
	if len(detail.Volumes) != 4 {
		t.Fatalf("volumes = %+v", detail.Volumes)
	}
	byName := make(map[string]bo.KubeVolume, len(detail.Volumes))
	for _, volume := range detail.Volumes {
		byName[volume.Name] = volume
	}
	if byName["data"].Type != "PersistentVolumeClaim" || !byName["data"].ReadOnly || byName["data"].Summary != "data" {
		t.Fatalf("pvc volume = %+v", byName["data"])
	}
	if byName["secret"].Type != "Secret" || byName["secret"].Summary != "db-secret" {
		t.Fatalf("secret volume = %+v", byName["secret"])
	}
	if byName["config"].Type != "ConfigMap" || byName["config"].Summary != "app-config" {
		t.Fatalf("configmap volume = %+v", byName["config"])
	}
	if byName["cache"].Type != "EmptyDir" {
		t.Fatalf("emptyDir volume = %+v", byName["cache"])
	}
	if len(detail.OwnerReferences) != 1 || !detail.OwnerReferences[0].Controller || detail.OwnerReferences[0].Kind != "ReplicaSet" {
		t.Fatalf("owners = %+v", detail.OwnerReferences)
	}
	if len(detail.Conditions) != 1 || detail.Conditions[0].Type != "Ready" {
		t.Fatalf("conditions = %+v", detail.Conditions)
	}
}

func TestWorkloadDetailConditionsAndReplicaSets(t *testing.T) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default", Name: "web",
			Annotations: map[string]string{"deployment.kubernetes.io/revision": "3"},
		},
		Spec: appsv1.DeploymentSpec{Replicas: int32Ptr(3)},
		Status: appsv1.DeploymentStatus{
			Replicas: 3, ReadyReplicas: 2, UpdatedReplicas: 3, AvailableReplicas: 2, UnavailableReplicas: 1,
			Conditions: []appsv1.DeploymentCondition{{
				Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue,
				Reason: "MinimumReplicasAvailable", LastTransitionTime: metav1.Unix(1700000300, 0),
			}},
		},
	}
	oldReplicaSet := fakeReplicaSet("default", "web-old", "web", "nginx:1.26", 1, 1, 1)
	oldReplicaSet.UID = types.UID("rs-old")
	oldReplicaSet.Annotations = map[string]string{"deployment.kubernetes.io/revision": "2"}
	newReplicaSet := fakeReplicaSet("default", "web-new", "web", "nginx:1.27", 3, 2, 3)
	newReplicaSet.UID = types.UID("rs-new")
	newReplicaSet.Annotations = map[string]string{"deployment.kubernetes.io/revision": "3"}
	repo := newFakeRepo(deployment, oldReplicaSet, newReplicaSet)
	detail, err := repo.WorkloadDetail(context.Background(), testClusterUID, "Deployment", "default", "web")
	if err != nil {
		t.Fatalf("WorkloadDetail: %v", err)
	}
	if detail.ReplicaStatus.Desired != 3 || detail.ReplicaStatus.Current != 3 {
		t.Fatalf("replica status = %+v", detail.ReplicaStatus)
	}
	if detail.ReplicaStatus.Available != 2 || detail.ReplicaStatus.Unavailable != 1 {
		t.Fatalf("replica available = %+v", detail.ReplicaStatus)
	}
	if len(detail.Conditions) != 1 || detail.Conditions[0].Reason != "MinimumReplicasAvailable" {
		t.Fatalf("conditions = %+v", detail.Conditions)
	}
	if len(detail.ReplicaSets) != 2 {
		t.Fatalf("replicasets = %+v", detail.ReplicaSets)
	}
	// 版本号倒序：新版本在前，并且只有它标记为当前版本。
	if detail.ReplicaSets[0].Name != "web-new" || detail.ReplicaSets[0].Revision != 3 || !detail.ReplicaSets[0].Current {
		t.Fatalf("first replicaset = %+v", detail.ReplicaSets[0])
	}
	if detail.ReplicaSets[1].Revision != 2 || detail.ReplicaSets[1].Current {
		t.Fatalf("second replicaset = %+v", detail.ReplicaSets[1])
	}
}

func TestListEvents(t *testing.T) {
	nodeEvent := &corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Namespace: "default", Name: "node-a.1"},
		InvolvedObject: corev1.ObjectReference{Kind: "Node", Name: "node-a", UID: types.UID("node-uid")},
		Type:           "Warning", Reason: "NodeNotReady", Message: "node is not ready",
		Count: 3, LastTimestamp: metav1.Unix(1700000200, 0), FirstTimestamp: metav1.Unix(1700000100, 0),
		Source: corev1.EventSource{Component: "kubelet", Host: "node-a"},
	}
	staleEvent := &corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Namespace: "default", Name: "node-a.old"},
		InvolvedObject: corev1.ObjectReference{Kind: "Node", Name: "node-a", UID: types.UID("old-node-uid")},
		Type:           "Normal", Reason: "NodeReady", Message: "old node",
		LastTimestamp: metav1.Unix(1700000300, 0),
	}
	otherPodEvent := &corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Namespace: "default", Name: "web-1.1"},
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "web-1", UID: types.UID("pod-uid")},
		Reason:         "Pulled", LastTimestamp: metav1.Unix(1700000400, 0),
	}
	repo := newFakeRepo(nodeEvent, staleEvent, otherPodEvent)

	events, err := repo.ListEvents(context.Background(), testClusterUID, "default", "Node", "node-a", "node-uid", 20)
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	// UID 过滤丢弃同名旧对象的事件。
	if len(events) != 1 || events[0].Reason != "NodeNotReady" {
		t.Fatalf("events = %+v", events)
	}
	if events[0].Count != 3 || events[0].Source != "kubelet/node-a" || events[0].LastSeenAt != 1700000200 {
		t.Fatalf("event = %+v", events[0])
	}

	// 没有 UID 时退化为 kind + name 过滤，并包含两个同名对象的事件，按时间倒序。
	events, err = repo.ListEvents(context.Background(), testClusterUID, "default", "Node", "node-a", "", 20)
	if err != nil {
		t.Fatalf("ListEvents without uid: %v", err)
	}
	if len(events) != 2 || events[0].LastSeenAt != 1700000300 {
		t.Fatalf("events = %+v", events)
	}

	// limit 截断。
	events, err = repo.ListEvents(context.Background(), testClusterUID, "default", "Node", "node-a", "", 1)
	if err != nil {
		t.Fatalf("ListEvents with limit: %v", err)
	}
	if len(events) != 1 || events[0].LastSeenAt != 1700000300 {
		t.Fatalf("limited events = %+v", events)
	}
}

func TestServerVersion(t *testing.T) {
	repo := newFakeRepo()
	version, err := repo.ServerVersion(context.Background(), testClusterUID)
	if err != nil {
		t.Fatalf("ServerVersion: %v", err)
	}
	if version == "" {
		t.Fatal("version is empty")
	}
}
