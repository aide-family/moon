package data

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/aide-family/moon/pkg/merr"
)

const testClusterUID = "cluster-uid"

func newFakeRepo(objects ...runtime.Object) *kubernetesRepo {
	return newFakeRepoWithClientset(fake.NewSimpleClientset(objects...))
}

func newFakeRepoWithClientset(clientset *fake.Clientset) *kubernetesRepo {
	pool := &clusterClients{cached: map[string]kubernetes.Interface{testClusterUID: clientset}}
	return &kubernetesRepo{clients: pool, log: testLogger()}
}

func fakeNode(name, status string, unschedulable bool) *corev1.Node {
	ready := corev1.ConditionTrue
	if status == "NotReady" {
		ready = corev1.ConditionFalse
	}
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       corev1.NodeSpec{Unschedulable: unschedulable},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: ready}},
			Addresses:  []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.0.1"}},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
		},
	}
}

func fakePod(namespace, name, nodeName, ownerKind, ownerName string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: corev1.PodSpec{
			NodeName:   nodeName,
			Containers: []corev1.Container{{Name: "app", Image: "nginx"}},
		},
	}
	if ownerKind != "" {
		pod.OwnerReferences = []metav1.OwnerReference{{Kind: ownerKind, Name: ownerName}}
	}
	return pod
}

func fakeDeployment(namespace, name string, replicas int32) *appsv1.Deployment {
	desired := replicas
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: appsv1.DeploymentSpec{
			Replicas: &desired,
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "nginx:1.27"}}}},
		},
		Status: appsv1.DeploymentStatus{ReadyReplicas: replicas, UpdatedReplicas: replicas, AvailableReplicas: replicas},
	}
}

func fakeStatefulSet(namespace, name string, replicas int32) *appsv1.StatefulSet {
	desired := replicas
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &desired,
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "mysql:8"}}}},
		},
	}
}

func fakeDaemonSet(namespace, name string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "agent", Image: "agent:1"}}}},
		},
		Status: appsv1.DaemonSetStatus{DesiredNumberScheduled: 3, NumberReady: 2},
	}
}

func TestListNodesWithFake(t *testing.T) {
	repo := newFakeRepo(
		fakeNode("node-a", "Ready", false),
		fakeNode("node-b", "NotReady", true),
		fakePod("default", "web-1", "node-a", "ReplicaSet", "web-7d9f"),
		fakePod("default", "web-2", "node-a", "ReplicaSet", "web-7d9f"),
		fakePod("kube-system", "agent-1", "node-b", "DaemonSet", "agent"),
		fakePod("default", "pending", "", "", ""),
	)
	items, err := repo.ListNodes(context.Background(), testClusterUID, "")
	if err != nil {
		t.Fatalf("ListNodes: %v", err)
	}
	if len(items) != 2 || items[0].Name != "node-a" || items[1].Name != "node-b" {
		t.Fatalf("nodes = %+v", items)
	}
	if items[0].PodCount != 2 || items[1].PodCount != 1 {
		t.Fatalf("pod counts = %d/%d", items[0].PodCount, items[1].PodCount)
	}
	if items[1].Schedulable || items[1].Status != "NotReady" {
		t.Fatalf("node-b = %+v", items[1])
	}
	filtered, err := repo.ListNodes(context.Background(), testClusterUID, "B")
	if err != nil {
		t.Fatalf("ListNodes filtered: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "node-b" {
		t.Fatalf("filtered nodes = %+v", filtered)
	}
}

func TestListNodesIgnoresTerminalPods(t *testing.T) {
	succeeded := fakePod("default", "done", "node-a", "", "")
	succeeded.Status.Phase = corev1.PodSucceeded
	failed := fakePod("default", "broken", "node-a", "", "")
	failed.Status.Phase = corev1.PodFailed
	repo := newFakeRepo(fakeNode("node-a", "Ready", false), succeeded, failed, fakePod("default", "live", "node-a", "", ""))
	items, err := repo.ListNodes(context.Background(), testClusterUID, "")
	if err != nil {
		t.Fatalf("ListNodes: %v", err)
	}
	if len(items) != 1 || items[0].PodCount != 1 {
		t.Fatalf("pod count = %+v", items)
	}
}

func TestListNodesToleratesForbiddenPodList(t *testing.T) {
	clientset := fake.NewSimpleClientset(fakeNode("node-a", "Ready", false))
	clientset.PrependReactor("list", "pods", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "pods"}, "", errors.New("forbidden"))
	})
	items, err := newFakeRepoWithClientset(clientset).ListNodes(context.Background(), testClusterUID, "")
	if err != nil {
		t.Fatalf("ListNodes: %v", err)
	}
	if len(items) != 1 || items[0].PodCount != 0 {
		t.Fatalf("nodes = %+v", items)
	}
}

func TestListWorkloadsWithFake(t *testing.T) {
	repo := newFakeRepo(
		fakeDeployment("default", "web", 3),
		fakeDeployment("default", "api", 2),
		fakeStatefulSet("db", "mysql", 1),
		fakeDaemonSet("kube-system", "agent"),
	)
	items, err := repo.ListWorkloads(context.Background(), testClusterUID, "", "", "")
	if err != nil {
		t.Fatalf("ListWorkloads: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("workloads = %+v", items)
	}
	// 排序口径是 namespace、kind、name。
	wantOrder := []string{"db/StatefulSet/mysql", "default/Deployment/api", "default/Deployment/web", "kube-system/DaemonSet/agent"}
	for index, want := range wantOrder {
		got := items[index].Namespace + "/" + items[index].Kind + "/" + items[index].Name
		if got != want {
			t.Fatalf("items[%d] = %s, want %s", index, got, want)
		}
	}
	deployments, err := repo.ListWorkloads(context.Background(), testClusterUID, "", "Deployment", "")
	if err != nil {
		t.Fatalf("ListWorkloads kind: %v", err)
	}
	if len(deployments) != 2 {
		t.Fatalf("deployments = %+v", deployments)
	}
	daemonSets, err := repo.ListWorkloads(context.Background(), testClusterUID, "kube-system", "DaemonSet", "agent")
	if err != nil {
		t.Fatalf("ListWorkloads daemonset: %v", err)
	}
	if len(daemonSets) != 1 || daemonSets[0].DesiredReplicas != 3 || daemonSets[0].ReadyReplicas != 2 {
		t.Fatalf("daemonsets = %+v", daemonSets)
	}
	if _, err := repo.ListWorkloads(context.Background(), testClusterUID, "", "Job", ""); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func TestListPodsWithFake(t *testing.T) {
	repo := newFakeRepo(
		fakePod("default", "web-7d9f8c-abcde", "node-a", "ReplicaSet", "web-7d9f8c"),
		fakePod("default", "web-7d9f8c-fghij", "node-a", "ReplicaSet", "web-7d9f8c"),
		fakePod("default", "agent-xyz", "node-b", "DaemonSet", "agent"),
		fakePod("default", "standalone", "node-b", "", ""),
		fakePod("other", "job-pod", "node-a", "Job", "backup"),
	)
	all, err := repo.ListPods(context.Background(), testClusterUID, "", "", "", "")
	if err != nil {
		t.Fatalf("ListPods: %v", err)
	}
	if len(all) != 5 {
		t.Fatalf("pods = %+v", all)
	}
	if all[0].Namespace != "default" || all[0].Name != "agent-xyz" {
		t.Fatalf("sorted first = %s/%s", all[0].Namespace, all[0].Name)
	}
	deploymentPods, err := repo.ListPods(context.Background(), testClusterUID, "default", "", "Deployment", "web")
	if err != nil {
		t.Fatalf("ListPods deployment: %v", err)
	}
	if len(deploymentPods) != 2 {
		t.Fatalf("deployment pods = %+v", deploymentPods)
	}
	jobPods, err := repo.ListPods(context.Background(), testClusterUID, "other", "job-pod", "Job", "backup")
	if err != nil {
		t.Fatalf("ListPods job: %v", err)
	}
	if len(jobPods) != 1 || jobPods[0].OwnerKind != "Job" {
		t.Fatalf("job pods = %+v", jobPods)
	}
}

func TestListIngressesWithFake(t *testing.T) {
	className := "nginx"
	build := func(namespace, name string) *networkingv1.Ingress {
		return &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
			Spec: networkingv1.IngressSpec{
				IngressClassName: &className,
				Rules: []networkingv1.IngressRule{{
					Host: "app.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: pathTypePtr(networkingv1.PathTypePrefix),
							Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{
								Name: "web", Port: networkingv1.ServiceBackendPort{Number: 80},
							}},
						}},
					}},
				}},
			},
		}
	}
	repo := newFakeRepo(build("default", "web"), build("other", "api"))
	items, err := repo.ListIngresses(context.Background(), testClusterUID, "", "")
	if err != nil {
		t.Fatalf("ListIngresses: %v", err)
	}
	if len(items) != 2 || items[0].Namespace != "default" || items[1].Namespace != "other" {
		t.Fatalf("ingresses = %+v", items)
	}
	if items[0].ClassName != "nginx" || len(items[0].Paths) != 1 || items[0].Paths[0].ServicePort != "80" {
		t.Fatalf("ingress = %+v", items[0])
	}
	filtered, err := repo.ListIngresses(context.Background(), testClusterUID, "other", "api")
	if err != nil {
		t.Fatalf("ListIngresses filtered: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "api" {
		t.Fatalf("filtered ingresses = %+v", filtered)
	}
}

func TestSetNodeUnschedulableWithFake(t *testing.T) {
	clientset := fake.NewSimpleClientset(fakeNode("node-a", "Ready", false))
	repo := newFakeRepoWithClientset(clientset)
	if err := repo.SetNodeUnschedulable(context.Background(), testClusterUID, "node-a", true); err != nil {
		t.Fatalf("cordoning: %v", err)
	}
	node, err := clientset.CoreV1().Nodes().Get(context.Background(), "node-a", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get node: %v", err)
	}
	if !node.Spec.Unschedulable {
		t.Fatal("node should be cordoned")
	}
	if err := repo.SetNodeUnschedulable(context.Background(), testClusterUID, "node-a", false); err != nil {
		t.Fatalf("uncordoning: %v", err)
	}
	node, err = clientset.CoreV1().Nodes().Get(context.Background(), "node-a", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get node: %v", err)
	}
	if node.Spec.Unschedulable {
		t.Fatal("node should be uncordoned")
	}
	if err := repo.SetNodeUnschedulable(context.Background(), testClusterUID, "missing", true); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestScaleWorkloadWithFake(t *testing.T) {
	clientset := fake.NewSimpleClientset(fakeDeployment("default", "web", 1))
	var scaledTo int32 = -1
	// 假客户端的 Scale 子资源不会写回 Deployment，这里拦截并记录目标副本数。
	clientset.PrependReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "scale" {
			return false, nil, nil
		}
		return true, &autoscalingv1.Scale{
			ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "default"},
			Spec:       autoscalingv1.ScaleSpec{Replicas: 1},
		}, nil
	})
	clientset.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "scale" {
			return false, nil, nil
		}
		scale := action.(k8stesting.UpdateAction).GetObject().(*autoscalingv1.Scale)
		scaledTo = scale.Spec.Replicas
		return true, scale, nil
	})
	repo := newFakeRepoWithClientset(clientset)
	item, err := repo.ScaleWorkload(context.Background(), testClusterUID, "default", "Deployment", "web", 3)
	if err != nil {
		t.Fatalf("ScaleWorkload: %v", err)
	}
	if scaledTo != 3 {
		t.Fatalf("scaled to %d, want 3", scaledTo)
	}
	if item == nil || item.Name != "web" || item.Kind != "Deployment" {
		t.Fatalf("item = %+v", item)
	}
	if _, err := repo.ScaleWorkload(context.Background(), testClusterUID, "default", "DaemonSet", "agent", 3); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	if _, err := repo.ScaleWorkload(context.Background(), testClusterUID, "default", "Deployment", "missing", 3); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestRestartWorkloadWithFake(t *testing.T) {
	clientset := fake.NewSimpleClientset(fakeDeployment("default", "web", 2))
	repo := newFakeRepoWithClientset(clientset)
	item, err := repo.RestartWorkload(context.Background(), testClusterUID, "default", "Deployment", "web")
	if err != nil {
		t.Fatalf("RestartWorkload: %v", err)
	}
	if item == nil || item.Name != "web" {
		t.Fatalf("item = %+v", item)
	}
	deployment, err := clientset.AppsV1().Deployments("default").Get(context.Background(), "web", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] == "" {
		t.Fatalf("template annotations = %v", deployment.Spec.Template.Annotations)
	}
	if _, err := repo.RestartWorkload(context.Background(), testClusterUID, "default", "Job", "backup"); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func TestDeletePodWithFake(t *testing.T) {
	clientset := fake.NewSimpleClientset(fakePod("default", "web-1", "node-a", "ReplicaSet", "web"))
	repo := newFakeRepoWithClientset(clientset)
	if err := repo.DeletePod(context.Background(), testClusterUID, "default", "web-1"); err != nil {
		t.Fatalf("DeletePod: %v", err)
	}
	if _, err := clientset.CoreV1().Pods("default").Get(context.Background(), "web-1", metav1.GetOptions{}); !merr.IsNotFound(mustKubeError(err)) {
		t.Fatalf("pod should be deleted, got %v", err)
	}
	if err := repo.DeletePod(context.Background(), testClusterUID, "default", "missing"); !merr.IsNotFound(mustKubeError(err)) {
		t.Fatalf("err = %v, want not found", err)
	}
}

// mustKubeError 把 client-go 的 NotFound 转成仓储错误，便于断言。
func mustKubeError(err error) error {
	return kubeError(testLogger(), err)
}

func pathTypePtr(pathType networkingv1.PathType) *networkingv1.PathType {
	return &pathType
}

func fakeService(namespace, name string, selector map[string]string, port int32, targetPort intstr.IntOrString, nodePort int32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeNodePort,
			ClusterIP: "10.96.0.10",
			Selector:  selector,
			Ports: []corev1.ServicePort{{
				Name: "http", Protocol: corev1.ProtocolTCP, Port: port, TargetPort: targetPort, NodePort: nodePort,
			}},
		},
	}
}

func fakeConfigMap(namespace, name string, data map[string]string, binaryData map[string][]byte, immutable bool) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Data:       data,
		BinaryData: binaryData,
		Immutable:  &immutable,
	}
}

func fakeSecret(namespace, name string, secretType corev1.SecretType, data map[string][]byte, immutable bool) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Type:       secretType,
		Data:       data,
		Immutable:  &immutable,
	}
}

func fakePersistentVolume(name, capacity string, claimRef *corev1.ObjectReference) *corev1.PersistentVolume {
	return &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PersistentVolumeSpec{
			Capacity:                      corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(capacity)},
			StorageClassName:              "fast",
			AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			ClaimRef:                      claimRef,
		},
		Status: corev1.PersistentVolumeStatus{Phase: corev1.VolumeBound},
	}
}

func fakePersistentVolumeClaim(namespace, name, statusCapacity, requestCapacity string) *corev1.PersistentVolumeClaim {
	storageClass := "fast"
	claim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			VolumeName:       "pv-" + name,
		},
		Status: corev1.PersistentVolumeClaimStatus{Phase: corev1.ClaimBound},
	}
	if requestCapacity != "" {
		claim.Spec.Resources = corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(requestCapacity)},
		}
	}
	if statusCapacity != "" {
		claim.Status.Capacity = corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(statusCapacity)}
	}
	return claim
}

func TestListServicesWithFake(t *testing.T) {
	repo := newFakeRepo(
		fakeService("default", "web", map[string]string{"app": "web"}, 80, intstr.FromInt32(8080), 30080),
		fakeService("other", "api", nil, 9090, intstr.FromString("http"), 0),
	)
	items, err := repo.ListServices(context.Background(), testClusterUID, "", "")
	if err != nil {
		t.Fatalf("ListServices: %v", err)
	}
	if len(items) != 2 || items[0].Namespace != "default" || items[1].Namespace != "other" {
		t.Fatalf("services = %+v", items)
	}
	if items[0].Selector != `{"app":"web"}` || items[0].Ports[0].TargetPort != "8080" || items[0].Ports[0].NodePort != 30080 {
		t.Fatalf("service = %+v", items[0])
	}
	// 空选择器始终是合法 JSON 对象，nodePort 未设置时为 0。
	if items[1].Selector != "{}" || items[1].Ports[0].NodePort != 0 || items[1].Ports[0].TargetPort != "http" {
		t.Fatalf("service = %+v", items[1])
	}
	filtered, err := repo.ListServices(context.Background(), testClusterUID, "default", "WEB")
	if err != nil {
		t.Fatalf("ListServices filtered: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "web" {
		t.Fatalf("filtered services = %+v", filtered)
	}
}

func TestListConfigMapsWithFake(t *testing.T) {
	repo := newFakeRepo(
		fakeConfigMap("default", "app-config", map[string]string{"b": "config-value-one", "a": "config-value-two"}, map[string][]byte{"bin": []byte("config-binary-value")}, true),
		fakeConfigMap("other", "extra-config", map[string]string{"k": "config-value-three"}, nil, false),
	)
	items, err := repo.ListConfigMaps(context.Background(), testClusterUID, "", "")
	if err != nil {
		t.Fatalf("ListConfigMaps: %v", err)
	}
	if len(items) != 2 || items[0].Name != "app-config" || items[1].Name != "extra-config" {
		t.Fatalf("configmaps = %+v", items)
	}
	if strings.Join(items[0].DataKeys, ",") != "a,b" || strings.Join(items[0].BinaryDataKeys, ",") != "bin" {
		t.Fatalf("keys = %v / %v", items[0].DataKeys, items[0].BinaryDataKeys)
	}
	if !items[0].Immutable || items[1].Immutable {
		t.Fatalf("immutable = %+v", items)
	}
	rendered := fmt.Sprintf("%+v", items)
	for _, leaked := range []string{"config-value-one", "config-value-two", "config-value-three", "config-binary-value"} {
		if strings.Contains(rendered, leaked) {
			t.Fatalf("configmap value %q leaked into bo: %s", leaked, rendered)
		}
	}
	filtered, err := repo.ListConfigMaps(context.Background(), testClusterUID, "other", "extra")
	if err != nil {
		t.Fatalf("ListConfigMaps filtered: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Namespace != "other" {
		t.Fatalf("filtered configmaps = %+v", filtered)
	}
}

func TestListSecretsWithFake(t *testing.T) {
	repo := newFakeRepo(
		fakeSecret("default", "web-tls", corev1.SecretTypeTLS, map[string][]byte{
			"tls.key": []byte("secret-key-value"), "tls.crt": []byte("secret-cert-value"),
		}, false),
		fakeSecret("default", "opaque", corev1.SecretTypeOpaque, map[string][]byte{"password": []byte("secret-password-value")}, true),
	)
	items, err := repo.ListSecrets(context.Background(), testClusterUID, "default", "")
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if len(items) != 2 || items[0].Name != "opaque" || items[1].Name != "web-tls" {
		t.Fatalf("secrets = %+v", items)
	}
	if items[0].Type != "Opaque" || !items[0].Immutable || strings.Join(items[0].DataKeys, ",") != "password" {
		t.Fatalf("secret = %+v", items[0])
	}
	if items[1].Type != "kubernetes.io/tls" || strings.Join(items[1].DataKeys, ",") != "tls.crt,tls.key" {
		t.Fatalf("secret = %+v", items[1])
	}
	rendered := fmt.Sprintf("%+v", items)
	for _, leaked := range []string{"secret-key-value", "secret-cert-value", "secret-password-value"} {
		if strings.Contains(rendered, leaked) {
			t.Fatalf("secret value %q leaked into bo: %s", leaked, rendered)
		}
	}
}

func TestListPersistentVolumesWithFake(t *testing.T) {
	repo := newFakeRepo(
		fakePersistentVolume("pv-b", "20Gi", &corev1.ObjectReference{Namespace: "default", Name: "data-b"}),
		fakePersistentVolume("pv-a", "10Gi", nil),
	)
	// PV 是集群级资源：没有命名空间参数，按名称升序。
	items, err := repo.ListPersistentVolumes(context.Background(), testClusterUID, "")
	if err != nil {
		t.Fatalf("ListPersistentVolumes: %v", err)
	}
	if len(items) != 2 || items[0].Name != "pv-a" || items[1].Name != "pv-b" {
		t.Fatalf("pvs = %+v", items)
	}
	if items[0].Claim != "" || items[0].Capacity != "10Gi" {
		t.Fatalf("unbound pv = %+v", items[0])
	}
	if items[1].Claim != "default/data-b" || items[1].StorageClass != "fast" || items[1].ReclaimPolicy != "Delete" {
		t.Fatalf("pv = %+v", items[1])
	}
	filtered, err := repo.ListPersistentVolumes(context.Background(), testClusterUID, "PV-B")
	if err != nil {
		t.Fatalf("ListPersistentVolumes filtered: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "pv-b" {
		t.Fatalf("filtered pvs = %+v", filtered)
	}
}

func TestListPersistentVolumeClaimsWithFake(t *testing.T) {
	repo := newFakeRepo(
		fakePersistentVolumeClaim("default", "data-b", "20Gi", "10Gi"),
		fakePersistentVolumeClaim("default", "data-a", "", "5Gi"),
		fakePersistentVolumeClaim("other", "logs", "1Gi", "1Gi"),
	)
	items, err := repo.ListPersistentVolumeClaims(context.Background(), testClusterUID, "default", "")
	if err != nil {
		t.Fatalf("ListPersistentVolumeClaims: %v", err)
	}
	if len(items) != 2 || items[0].Name != "data-a" || items[1].Name != "data-b" {
		t.Fatalf("pvcs = %+v", items)
	}
	// 已绑定时用 status.capacity，未绑定时回退到 requests。
	if items[1].Capacity != "20Gi" || items[1].VolumeName != "pv-data-b" {
		t.Fatalf("pvc = %+v", items[1])
	}
	if items[0].Capacity != "5Gi" {
		t.Fatalf("pending pvc = %+v", items[0])
	}
	all, err := repo.ListPersistentVolumeClaims(context.Background(), testClusterUID, "", "LOG")
	if err != nil {
		t.Fatalf("ListPersistentVolumeClaims filtered: %v", err)
	}
	if len(all) != 1 || all[0].Namespace != "other" {
		t.Fatalf("filtered pvcs = %+v", all)
	}
}

func TestListServicesForbidden(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	clientset.PrependReactor("list", "services", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "services"}, "", errors.New("forbidden"))
	})
	repo := newFakeRepoWithClientset(clientset)
	_, err := repo.ListServices(context.Background(), testClusterUID, "", "")
	// 集群侧 Forbidden 是凭据能力问题，不是平台权限问题，映射成可读的调用方错误。
	if err == nil || kerrors.FromError(err).Reason != merr.ReasonClient ||
		!strings.Contains(kerrors.FromError(err).Message, "not allowed") {
		t.Fatalf("err = %v, want cluster credential client error", err)
	}
}
