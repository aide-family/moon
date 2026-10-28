package data

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func testTime() metav1.Time {
	return metav1.NewTime(time.Unix(1700000000, 0))
}

func TestToKubeNode(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-a",
			CreationTimestamp: testTime(),
			Labels: map[string]string{
				"node-role.kubernetes.io/worker":        "",
				"node-role.kubernetes.io/control-plane": "",
				"kubernetes.io/hostname":                "node-a",
			},
		},
		Spec: corev1.NodeSpec{Unschedulable: true},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
			Addresses: []corev1.NodeAddress{
				{Type: corev1.NodeHostName, Address: "node-a"},
				{Type: corev1.NodeInternalIP, Address: "10.0.0.8"},
			},
			NodeInfo: corev1.NodeSystemInfo{KubeletVersion: "v1.32.3", OSImage: "Ubuntu 24.04"},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("8"),
				corev1.ResourceMemory: resource.MustParse("16384Mi"),
			},
		},
	}
	item := toKubeNode(node, map[string]int32{"node-a": 12})
	if item.Name != "node-a" || item.Status != "Ready" {
		t.Fatalf("name=%q status=%q", item.Name, item.Status)
	}
	if item.Schedulable {
		t.Fatal("cordoned node should not be schedulable")
	}
	if !reflect.DeepEqual(item.Roles, []string{"control-plane", "worker"}) {
		t.Fatalf("roles = %v", item.Roles)
	}
	if item.InternalIP != "10.0.0.8" || item.KubeletVersion != "v1.32.3" || item.OSImage != "Ubuntu 24.04" {
		t.Fatalf("node info = %+v", item)
	}
	if item.CPUCapacity != "8" || item.MemoryCapacity != "16Gi" {
		t.Fatalf("capacity cpu=%q memory=%q", item.CPUCapacity, item.MemoryCapacity)
	}
	if item.PodCount != 12 || item.CreatedAt != 1700000000 {
		t.Fatalf("pod count=%d created=%d", item.PodCount, item.CreatedAt)
	}
}

func TestNodeStatus(t *testing.T) {
	cases := []struct {
		name       string
		conditions []corev1.NodeCondition
		want       string
	}{
		{name: "ready", conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}}, want: "Ready"},
		{name: "not ready", conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionFalse}}, want: "NotReady"},
		{name: "unknown condition", conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionUnknown}}, want: "Unknown"},
		{name: "missing ready", conditions: []corev1.NodeCondition{{Type: corev1.NodeMemoryPressure, Status: corev1.ConditionTrue}}, want: "Unknown"},
		{name: "no conditions", want: "Unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := nodeStatus(corev1.NodeStatus{Conditions: tc.conditions}); got != tc.want {
				t.Fatalf("nodeStatus = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestToKubeWorkloads(t *testing.T) {
	replicas := int32(3)
	podSpec := corev1.PodSpec{
		InitContainers: []corev1.Container{{Name: "init", Image: "busybox:1.36"}},
		Containers: []corev1.Container{
			{Name: "app", Image: "nginx:1.27"},
			{Name: "sidecar", Image: "nginx:1.27"},
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web", CreationTimestamp: testTime()},
		Spec:       appsv1.DeploymentSpec{Replicas: &replicas, Template: corev1.PodTemplateSpec{Spec: podSpec}},
		Status:     appsv1.DeploymentStatus{ReadyReplicas: 2, UpdatedReplicas: 3, AvailableReplicas: 2},
	}
	got := toKubeDeployment(deployment)
	if got.Kind != "Deployment" || got.Name != "web" || got.Namespace != "default" {
		t.Fatalf("deployment = %+v", got)
	}
	if got.DesiredReplicas != 3 || got.ReadyReplicas != 2 || got.UpdatedReplicas != 3 || got.AvailableReplicas != 2 {
		t.Fatalf("replicas = %+v", got)
	}
	if !reflect.DeepEqual(got.Images, []string{"nginx:1.27", "busybox:1.36"}) {
		t.Fatalf("images = %v", got.Images)
	}
	if got.CreatedAt != 1700000000 {
		t.Fatalf("created = %d", got.CreatedAt)
	}

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: "db", Name: "mysql"},
		Spec:       appsv1.StatefulSetSpec{Replicas: &replicas, Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "mysql", Image: "mysql:8"}}}}},
		Status:     appsv1.StatefulSetStatus{ReadyReplicas: 1, UpdatedReplicas: 2, AvailableReplicas: 1},
	}
	gotStateful := toKubeStatefulSet(statefulSet)
	if gotStateful.Kind != "StatefulSet" || gotStateful.DesiredReplicas != 3 || gotStateful.ReadyReplicas != 1 || gotStateful.UpdatedReplicas != 2 || gotStateful.AvailableReplicas != 1 {
		t.Fatalf("statefulset = %+v", gotStateful)
	}

	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: "kube-system", Name: "agent"},
		Spec:       appsv1.DaemonSetSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "agent", Image: "agent:1"}}}}},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 5, NumberReady: 4, UpdatedNumberScheduled: 5, NumberAvailable: 4,
		},
	}
	gotDaemon := toKubeDaemonSet(daemonSet)
	if gotDaemon.Kind != "DaemonSet" || gotDaemon.DesiredReplicas != 5 || gotDaemon.ReadyReplicas != 4 || gotDaemon.UpdatedReplicas != 5 || gotDaemon.AvailableReplicas != 4 {
		t.Fatalf("daemonset = %+v", gotDaemon)
	}
}

func TestToKubePod(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "web-7d9f8c-abcde",
			CreationTimestamp: testTime(),
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: "web-7d9f8c"},
				{Kind: "Deployment", Name: "web"},
			},
		},
		Spec: corev1.PodSpec{
			NodeName: "node-a",
			Containers: []corev1.Container{
				{Name: "app", Image: "nginx:1.27"},
				{Name: "sidecar", Image: "envoy:1.31"},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "10.244.0.9",
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "app", Ready: true, RestartCount: 2},
				{Name: "sidecar", Ready: false, RestartCount: 1},
			},
		},
	}
	item := toKubePod(pod)
	if item.Namespace != "default" || item.Name != "web-7d9f8c-abcde" || item.Phase != "Running" {
		t.Fatalf("pod = %+v", item)
	}
	if item.Ready != "1/2" || item.Restarts != 3 {
		t.Fatalf("ready=%q restarts=%d", item.Ready, item.Restarts)
	}
	if item.NodeName != "node-a" || item.PodIP != "10.244.0.9" {
		t.Fatalf("node=%q ip=%q", item.NodeName, item.PodIP)
	}
	if item.OwnerKind != "ReplicaSet" || item.OwnerName != "web-7d9f8c" {
		t.Fatalf("owner = %s/%s", item.OwnerKind, item.OwnerName)
	}
	wantContainers := []bo.KubeContainer{
		{Name: "app", Ready: true, Restarts: 2, Image: "nginx:1.27"},
		{Name: "sidecar", Ready: false, Restarts: 1, Image: "envoy:1.31"},
	}
	if !reflect.DeepEqual(item.Containers, wantContainers) {
		t.Fatalf("containers = %+v", item.Containers)
	}
	if item.CreatedAt != 1700000000 {
		t.Fatalf("created = %d", item.CreatedAt)
	}
}

func TestToKubeIngress(t *testing.T) {
	className := "nginx"
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web", CreationTimestamp: testTime()},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &className,
			Rules: []networkingv1.IngressRule{
				{
					Host: "a.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{
								Name: "web", Port: networkingv1.ServiceBackendPort{Number: 80},
							}},
						}},
					}},
				},
				{
					Host: "b.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/api",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{
								Name: "api", Port: networkingv1.ServiceBackendPort{Name: "http"},
							}},
						}},
					}},
				},
			},
			TLS: []networkingv1.IngressTLS{{Hosts: []string{"a.example.com", "c.example.com"}, SecretName: "web-tls"}},
		},
	}
	item := toKubeIngress(ingress)
	if item.ClassName != "nginx" {
		t.Fatalf("class = %q", item.ClassName)
	}
	if !reflect.DeepEqual(item.Hosts, []string{"a.example.com", "b.example.com", "c.example.com"}) {
		t.Fatalf("hosts = %v", item.Hosts)
	}
	if len(item.Paths) != 2 {
		t.Fatalf("paths = %+v", item.Paths)
	}
	if item.Paths[0] != (bo.KubeIngressPath{Host: "a.example.com", Path: "/", PathType: "Prefix", ServiceName: "web", ServicePort: "80"}) {
		t.Fatalf("path[0] = %+v", item.Paths[0])
	}
	if item.Paths[1] != (bo.KubeIngressPath{Host: "b.example.com", Path: "/api", PathType: "Prefix", ServiceName: "api", ServicePort: "http"}) {
		t.Fatalf("path[1] = %+v", item.Paths[1])
	}
	if !reflect.DeepEqual(item.TLS, []string{"web-tls"}) {
		t.Fatalf("tls = %v", item.TLS)
	}
	if item.CreatedAt != 1700000000 {
		t.Fatalf("created = %d", item.CreatedAt)
	}

	// 没有 spec.ingressClassName 时回退到注解。
	fallback := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: "legacy", Annotations: map[string]string{"kubernetes.io/ingress.class": "traefik"}},
	}
	if class := ingressClassName(fallback); class != "traefik" {
		t.Fatalf("fallback class = %q", class)
	}
}

func TestMatchesName(t *testing.T) {
	if !matchesName("Web-Pod", "") || !matchesName("Web-Pod", "web") || !matchesName("web-pod", "POD") {
		t.Fatal("contains fold match")
	}
	if matchesName("web", "api") {
		t.Fatal("mismatch should not match")
	}
}

func TestMatchesPodOwner(t *testing.T) {
	pod := bo.KubePod{OwnerKind: "ReplicaSet", OwnerName: "web-7d9f8c"}
	cases := []struct {
		name      string
		ownerKind string
		ownerName string
		want      bool
	}{
		{name: "no filter", want: true},
		{name: "direct kind", ownerKind: "ReplicaSet", want: true},
		{name: "deployment kind maps to replicaset", ownerKind: "Deployment", want: true},
		{name: "other kind", ownerKind: "DaemonSet", want: false},
		{name: "workload name prefix", ownerName: "web", want: true},
		{name: "exact owner name", ownerName: "web-7d9f8c", want: true},
		{name: "other name", ownerName: "api", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := matchesPodOwner(pod, tc.ownerKind, tc.ownerName); got != tc.want {
				t.Fatalf("matchesPodOwner = %v, want %v", got, tc.want)
			}
		})
	}

	// 工作负载名前缀只能匹配 <name>-<ReplicaSet 哈希>，不能把 web 当成 web-other 的前缀。
	other := bo.KubePod{OwnerKind: "ReplicaSet", OwnerName: "web-other-7d9f8c"}
	if matchesPodOwner(other, "", "web") {
		t.Fatal("web should not match web-other-7d9f8c")
	}
	if !matchesPodOwner(other, "", "web-other") {
		t.Fatal("web-other should match web-other-7d9f8c")
	}
}

func TestToKubeService(t *testing.T) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web", CreationTimestamp: testTime()},
		Spec: corev1.ServiceSpec{
			Type:        corev1.ServiceTypeLoadBalancer,
			ClusterIP:   "10.96.0.10",
			ExternalIPs: []string{"203.0.113.5"},
			Selector:    map[string]string{"tier": "frontend", "app": "web"},
			Ports: []corev1.ServicePort{
				{Name: "http", Protocol: corev1.ProtocolTCP, Port: 80, TargetPort: intstr.FromInt32(8080), NodePort: 30080},
				{Name: "metrics", Protocol: corev1.ProtocolTCP, Port: 9090, TargetPort: intstr.FromString("metrics")},
				{Name: "unset", Protocol: corev1.ProtocolUDP, Port: 53},
			},
		},
		Status: corev1.ServiceStatus{LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{{IP: "198.51.100.7"}},
		}},
	}
	item := toKubeService(service)
	if item.Namespace != "default" || item.Name != "web" || item.Type != "LoadBalancer" {
		t.Fatalf("service = %+v", item)
	}
	if item.ClusterIP != "10.96.0.10" || item.LoadBalancerIP != "198.51.100.7" {
		t.Fatalf("ips = %+v", item)
	}
	if !reflect.DeepEqual(item.ExternalIPs, []string{"203.0.113.5"}) {
		t.Fatalf("external ips = %v", item.ExternalIPs)
	}
	// encoding/json 对 map 键排序，选择器输出稳定。
	if item.Selector != `{"app":"web","tier":"frontend"}` {
		t.Fatalf("selector = %q", item.Selector)
	}
	wantPorts := []bo.KubeServicePort{
		{Name: "http", Protocol: "TCP", Port: 80, TargetPort: "8080", NodePort: 30080},
		{Name: "metrics", Protocol: "TCP", Port: 9090, TargetPort: "metrics"},
		{Name: "unset", Protocol: "UDP", Port: 53, TargetPort: ""},
	}
	if !reflect.DeepEqual(item.Ports, wantPorts) {
		t.Fatalf("ports = %+v", item.Ports)
	}
	if item.CreatedAt != 1700000000 {
		t.Fatalf("created = %d", item.CreatedAt)
	}
	// 空选择器也输出合法的 JSON 对象。
	if got := serviceSelector(nil); got != "{}" {
		t.Fatalf("empty selector = %q", got)
	}
}

func TestToKubeConfigMap(t *testing.T) {
	immutable := true
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "app-config", CreationTimestamp: testTime()},
		Data:       map[string]string{"z": "plain-secret-value", "a": "another-value"},
		BinaryData: map[string][]byte{"bin": []byte("binary-secret-value")},
		Immutable:  &immutable,
	}
	item := toKubeConfigMap(configMap)
	if item.Namespace != "default" || item.Name != "app-config" || !item.Immutable {
		t.Fatalf("configmap = %+v", item)
	}
	if !reflect.DeepEqual(item.DataKeys, []string{"a", "z"}) {
		t.Fatalf("data keys = %v", item.DataKeys)
	}
	if !reflect.DeepEqual(item.BinaryDataKeys, []string{"bin"}) {
		t.Fatalf("binary data keys = %v", item.BinaryDataKeys)
	}
	// 业务对象只有键名，值不会跟着转出去。
	rendered := fmt.Sprintf("%+v", item)
	for _, secret := range []string{"plain-secret-value", "another-value", "binary-secret-value"} {
		if strings.Contains(rendered, secret) {
			t.Fatalf("configmap value %q leaked into bo: %s", secret, rendered)
		}
	}
}

func TestToKubeSecret(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web-tls", CreationTimestamp: testTime()},
		Type:       corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.key": []byte("key-secret-value"),
			"tls.crt": []byte("cert-secret-value"),
		},
	}
	item := toKubeSecret(secret)
	if item.Type != "kubernetes.io/tls" || item.Immutable {
		t.Fatalf("secret = %+v", item)
	}
	if !reflect.DeepEqual(item.DataKeys, []string{"tls.crt", "tls.key"}) {
		t.Fatalf("data keys = %v", item.DataKeys)
	}
	rendered := fmt.Sprintf("%+v", item)
	for _, leaked := range []string{"key-secret-value", "cert-secret-value"} {
		if strings.Contains(rendered, leaked) {
			t.Fatalf("secret value %q leaked into bo: %s", leaked, rendered)
		}
	}
}

func TestToKubePersistentVolume(t *testing.T) {
	volume := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{Name: "pv-data", CreationTimestamp: testTime()},
		Spec: corev1.PersistentVolumeSpec{
			Capacity:                      corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
			StorageClassName:              "fast",
			AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce, corev1.ReadOnlyMany},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			ClaimRef:                      &corev1.ObjectReference{Namespace: "default", Name: "data"},
		},
		Status: corev1.PersistentVolumeStatus{Phase: corev1.VolumeBound},
	}
	item := toKubePersistentVolume(volume)
	if item.Name != "pv-data" || item.Capacity != "10Gi" || item.StorageClass != "fast" {
		t.Fatalf("pv = %+v", item)
	}
	if !reflect.DeepEqual(item.AccessModes, []string{"ReadWriteOnce", "ReadOnlyMany"}) {
		t.Fatalf("access modes = %v", item.AccessModes)
	}
	if item.Claim != "default/data" || item.Phase != "Bound" || item.ReclaimPolicy != "Retain" {
		t.Fatalf("pv = %+v", item)
	}
	if item.CreatedAt != 1700000000 {
		t.Fatalf("created = %d", item.CreatedAt)
	}
	// 未绑定时 claim 为空。
	volume.Spec.ClaimRef = nil
	if item := toKubePersistentVolume(volume); item.Claim != "" {
		t.Fatalf("unbound claim = %q", item.Claim)
	}
}

func TestToKubePersistentVolumeClaim(t *testing.T) {
	storageClass := "fast"
	bound := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "data", CreationTimestamp: testTime()},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			VolumeName:       "pv-data",
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("5Gi")},
			},
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase:    corev1.ClaimBound,
			Capacity: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
		},
	}
	item := toKubePersistentVolumeClaim(bound)
	if item.Namespace != "default" || item.Name != "data" || item.Capacity != "10Gi" {
		t.Fatalf("pvc = %+v", item)
	}
	if item.StorageClass != "fast" || item.VolumeName != "pv-data" || item.Phase != "Bound" {
		t.Fatalf("pvc = %+v", item)
	}
	if !reflect.DeepEqual(item.AccessModes, []string{"ReadWriteOnce"}) || item.CreatedAt != 1700000000 {
		t.Fatalf("pvc = %+v", item)
	}

	// 未绑定时 status.capacity 为空，回退到 spec.resources.requests.storage。
	pending := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "pending"},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("5Gi")},
			},
		},
		Status: corev1.PersistentVolumeClaimStatus{Phase: corev1.ClaimPending},
	}
	pendingItem := toKubePersistentVolumeClaim(pending)
	if pendingItem.Capacity != "5Gi" || pendingItem.StorageClass != "" || pendingItem.Phase != "Pending" {
		t.Fatalf("pending pvc = %+v", pendingItem)
	}
}
