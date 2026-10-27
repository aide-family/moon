package data

import (
	"context"
	"fmt"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

func int32Ptr(value int32) *int32 {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func fakeDeploymentWithTemplate(namespace, name string, podSpec corev1.PodSpec) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(1), Template: corev1.PodTemplateSpec{Spec: podSpec}},
	}
}

func fakeStatefulSetWithTemplate(namespace, name string, podSpec corev1.PodSpec) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec:       appsv1.StatefulSetSpec{Replicas: int32Ptr(1), Template: corev1.PodTemplateSpec{Spec: podSpec}},
	}
}

func fakeDaemonSetWithTemplate(namespace, name string, podSpec corev1.PodSpec) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec:       appsv1.DaemonSetSpec{Template: corev1.PodTemplateSpec{Spec: podSpec}},
	}
}

func fakeReplicaSet(namespace, name, deploymentName, image string, desired, ready, current int32) *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "apps/v1", Kind: "Deployment", Name: deploymentName,
			}},
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: int32Ptr(desired),
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "app", Image: image}},
			}},
		},
		Status: appsv1.ReplicaSetStatus{Replicas: current, ReadyReplicas: ready},
	}
}

func TestWorkloadDetailDeployment(t *testing.T) {
	repo := newFakeRepo(
		fakeDeployment("default", "web", 2),
		fakeReplicaSet("default", "web-abc", "web", "nginx:1.27", 2, 2, 2),
		fakeReplicaSet("default", "other-xyz", "other", "nginx:1.25", 1, 1, 1),
		fakePod("default", "web-abc-1", "node-a", "ReplicaSet", "web-abc"),
		fakePod("default", "web-abc-2", "node-b", "ReplicaSet", "web-abc"),
		fakePod("default", "other-xyz-1", "node-a", "ReplicaSet", "other-xyz"),
		fakePod("default", "standalone", "node-a", "", ""),
	)
	detail, err := repo.WorkloadDetail(context.Background(), testClusterUID, "Deployment", "default", "web")
	if err != nil {
		t.Fatalf("WorkloadDetail: %v", err)
	}
	if detail.Workload.Kind != "Deployment" || detail.Workload.Name != "web" || detail.Workload.DesiredReplicas != 2 {
		t.Fatalf("workload = %+v", detail.Workload)
	}
	if len(detail.ReplicaSets) != 1 || detail.ReplicaSets[0].Name != "web-abc" {
		t.Fatalf("replicasets = %+v", detail.ReplicaSets)
	}
	rs := detail.ReplicaSets[0]
	if rs.DesiredReplicas != 2 || rs.ReadyReplicas != 2 || rs.CurrentReplicas != 2 || len(rs.Images) != 1 || rs.Images[0] != "nginx:1.27" {
		t.Fatalf("replicaset = %+v", rs)
	}
	if len(detail.Pods) != 2 || detail.Pods[0].Name != "web-abc-1" || detail.Pods[1].Name != "web-abc-2" {
		t.Fatalf("pods = %+v", detail.Pods)
	}
}

func TestWorkloadDetailStatefulSet(t *testing.T) {
	repo := newFakeRepo(
		fakeStatefulSet("default", "db", 3),
		fakePod("default", "db-0", "node-a", "StatefulSet", "db"),
		fakePod("default", "web-abc-1", "node-a", "ReplicaSet", "web-abc"),
	)
	detail, err := repo.WorkloadDetail(context.Background(), testClusterUID, "StatefulSet", "default", "db")
	if err != nil {
		t.Fatalf("WorkloadDetail: %v", err)
	}
	if detail.Workload.Kind != "StatefulSet" || detail.Workload.Name != "db" {
		t.Fatalf("workload = %+v", detail.Workload)
	}
	if len(detail.ReplicaSets) != 0 {
		t.Fatalf("statefulset should have no replicasets: %+v", detail.ReplicaSets)
	}
	if len(detail.Pods) != 1 || detail.Pods[0].Name != "db-0" {
		t.Fatalf("pods = %+v", detail.Pods)
	}
}

func TestWorkloadDetailInvalidKind(t *testing.T) {
	repo := newFakeRepo()
	if _, err := repo.WorkloadDetail(context.Background(), testClusterUID, "Job", "default", "backup"); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func TestWorkloadDetailNotFound(t *testing.T) {
	repo := newFakeRepo()
	if _, err := repo.WorkloadDetail(context.Background(), testClusterUID, "Deployment", "default", "missing"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestIngressDetail(t *testing.T) {
	className := "nginx"
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web"},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &className,
			DefaultBackend: &networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{
				Name: "fallback", Port: networkingv1.ServiceBackendPort{Number: 80},
			}},
			Rules: []networkingv1.IngressRule{{
				Host: "app.example.com",
				IngressRuleValue: networkingv1.IngressRuleValue{HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{Path: "/", PathType: pathTypePtr(networkingv1.PathTypePrefix), Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: "web", Port: networkingv1.ServiceBackendPort{Number: 80}}}},
						{Path: "/api", PathType: pathTypePtr(networkingv1.PathTypePrefix), Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: "api", Port: networkingv1.ServiceBackendPort{Number: 8080}}}},
					},
				}},
			}},
		},
	}
	repo := newFakeRepo(
		ingress,
		fakeService("default", "web", map[string]string{"app": "web"}, 80, intstr.FromInt32(8080), 0),
		fakeService("default", "api", map[string]string{"app": "api"}, 8080, intstr.FromInt32(8080), 0),
	)
	detail, err := repo.IngressDetail(context.Background(), testClusterUID, "default", "web")
	if err != nil {
		t.Fatalf("IngressDetail: %v", err)
	}
	if detail.Ingress.Name != "web" || detail.Ingress.ClassName != "nginx" {
		t.Fatalf("ingress = %+v", detail.Ingress)
	}
	if len(detail.Services) != 2 || detail.Services[0].Name != "api" || detail.Services[1].Name != "web" {
		t.Fatalf("services = %+v", detail.Services)
	}
	for _, service := range detail.Services {
		if service.Name == "fallback" {
			t.Fatalf("missing service should be skipped: %+v", detail.Services)
		}
	}
}

func TestServiceDetail(t *testing.T) {
	service := fakeService("default", "web", map[string]string{"app": "web"}, 80, intstr.FromInt32(8080), 30080)
	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web"},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{
				IP:        "10.0.0.1",
				Hostname:  "web-1",
				NodeName:  stringPtr("node-a"),
				TargetRef: &corev1.ObjectReference{Kind: "Pod", Name: "web-1"},
			}},
			NotReadyAddresses: []corev1.EndpointAddress{{
				IP:        "10.0.0.2",
				TargetRef: &corev1.ObjectReference{Kind: "Pod", Name: "web-2"},
			}},
			Ports: []corev1.EndpointPort{{Name: "http", Port: 8080, Protocol: corev1.ProtocolTCP}},
		}},
	}
	repo := newFakeRepo(service, endpoints)
	detail, err := repo.ServiceDetail(context.Background(), testClusterUID, "default", "web")
	if err != nil {
		t.Fatalf("ServiceDetail: %v", err)
	}
	if detail.Service.Name != "web" || detail.Service.Type != "NodePort" {
		t.Fatalf("service = %+v", detail.Service)
	}
	if len(detail.Endpoints) != 2 {
		t.Fatalf("endpoints = %+v", detail.Endpoints)
	}
	ready := detail.Endpoints[0]
	if ready.Address != "10.0.0.1" || !ready.Ready || ready.NodeName != "node-a" ||
		ready.TargetKind != "Pod" || ready.TargetName != "web-1" {
		t.Fatalf("ready endpoint = %+v", ready)
	}
	if len(ready.Ports) != 1 || ready.Ports[0].Name != "http" || ready.Ports[0].Port != 8080 || ready.Ports[0].Protocol != "TCP" {
		t.Fatalf("endpoint ports = %+v", ready.Ports)
	}
	notReady := detail.Endpoints[1]
	if notReady.Address != "10.0.0.2" || notReady.Ready || notReady.TargetName != "web-2" {
		t.Fatalf("not ready endpoint = %+v", notReady)
	}
}

func TestServiceDetailWithoutEndpoints(t *testing.T) {
	repo := newFakeRepo(fakeService("default", "web", nil, 80, intstr.FromInt32(80), 0))
	detail, err := repo.ServiceDetail(context.Background(), testClusterUID, "default", "web")
	if err != nil {
		t.Fatalf("ServiceDetail: %v", err)
	}
	if len(detail.Endpoints) != 0 {
		t.Fatalf("endpoints = %+v", detail.Endpoints)
	}
}

func TestServiceDetailNotFound(t *testing.T) {
	repo := newFakeRepo()
	if _, err := repo.ServiceDetail(context.Background(), testClusterUID, "default", "missing"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestConfigMapDetailRelations(t *testing.T) {
	configMap := fakeConfigMap("default", "app-config", map[string]string{"app": "config-value"}, nil, false)
	// 通过卷引用。
	volumeDeployment := fakeDeploymentWithTemplate("default", "cm-volume", corev1.PodSpec{
		Volumes: []corev1.Volume{{
			Name: "config",
			VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "app-config"},
			}},
		}},
	})
	// 通过 envFrom 引用。
	envStatefulSet := fakeStatefulSetWithTemplate("default", "cm-env", corev1.PodSpec{
		Containers: []corev1.Container{{
			Name:    "app",
			EnvFrom: []corev1.EnvFromSource{{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "app-config"}}}},
		}},
	})
	// 通过 env.valueFrom 引用。
	valueDeployment := fakeDeploymentWithTemplate("default", "cm-value", corev1.PodSpec{
		Containers: []corev1.Container{{
			Name: "app",
			Env: []corev1.EnvVar{{
				Name: "APP",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "app-config"}, Key: "app"},
				},
			}},
		}},
	})
	// 引用别的 ConfigMap，不能算进来。
	unrelatedDaemonSet := fakeDaemonSetWithTemplate("default", "other", corev1.PodSpec{
		Volumes: []corev1.Volume{{
			Name: "config",
			VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "other-config"},
			}},
		}},
	})
	// 别的命名空间不参与扫描。
	otherNamespace := fakeDeploymentWithTemplate("other", "cm-volume", corev1.PodSpec{
		Volumes: []corev1.Volume{{
			Name: "config",
			VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "app-config"},
			}},
		}},
	})
	repo := newFakeRepo(configMap, volumeDeployment, envStatefulSet, valueDeployment, unrelatedDaemonSet, otherNamespace)
	detail, err := repo.ConfigMapDetail(context.Background(), testClusterUID, "default", "app-config")
	if err != nil {
		t.Fatalf("ConfigMapDetail: %v", err)
	}
	if detail.ConfigMap.Name != "app-config" || strings.Join(detail.ConfigMap.DataKeys, ",") != "app" {
		t.Fatalf("configmap = %+v", detail.ConfigMap)
	}
	got := workloadRefKeys(detail.RelatedWorkloads)
	want := []string{"Deployment/default/cm-value", "Deployment/default/cm-volume", "StatefulSet/default/cm-env"}
	if got != strings.Join(want, ",") {
		t.Fatalf("related = %s, want %s", got, strings.Join(want, ","))
	}
}

func TestSecretDetailRelations(t *testing.T) {
	secret := fakeSecret("default", "db-secret", corev1.SecretTypeOpaque, map[string][]byte{"password": []byte("super-secret-value")}, false)
	envDeployment := fakeDeploymentWithTemplate("default", "secret-env", corev1.PodSpec{
		Containers: []corev1.Container{{
			Name: "app",
			Env: []corev1.EnvVar{{
				Name: "PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "db-secret"}, Key: "password"},
				},
			}},
		}},
	})
	pullDaemonSet := fakeDaemonSetWithTemplate("default", "secret-pull", corev1.PodSpec{
		ImagePullSecrets: []corev1.LocalObjectReference{{Name: "db-secret"}},
		Containers:       []corev1.Container{{Name: "app", Image: "nginx"}},
	})
	projectedStatefulSet := fakeStatefulSetWithTemplate("default", "secret-projected", corev1.PodSpec{
		Volumes: []corev1.Volume{{
			Name: "projected",
			VolumeSource: corev1.VolumeSource{Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{{Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{Name: "db-secret"},
				}}},
			}},
		}},
	})
	unrelatedDeployment := fakeDeploymentWithTemplate("default", "other", corev1.PodSpec{
		Containers: []corev1.Container{{
			Name:    "app",
			EnvFrom: []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "other-secret"}}}},
		}},
	})
	repo := newFakeRepo(secret, envDeployment, pullDaemonSet, projectedStatefulSet, unrelatedDeployment)
	detail, err := repo.SecretDetail(context.Background(), testClusterUID, "default", "db-secret")
	if err != nil {
		t.Fatalf("SecretDetail: %v", err)
	}
	if detail.Secret.Name != "db-secret" || detail.Secret.Type != "Opaque" || strings.Join(detail.Secret.DataKeys, ",") != "password" {
		t.Fatalf("secret = %+v", detail.Secret)
	}
	got := workloadRefKeys(detail.RelatedWorkloads)
	want := "DaemonSet/default/secret-pull,Deployment/default/secret-env,StatefulSet/default/secret-projected"
	if got != want {
		t.Fatalf("related = %s, want %s", got, want)
	}
	// 详情对象里只有键名，Secret 值不会出现在任何输出中。
	rendered := fmt.Sprintf("%+v", detail)
	if strings.Contains(rendered, "super-secret-value") {
		t.Fatalf("secret value leaked: %s", rendered)
	}
}

func TestPersistentVolumeClaimDetailRelations(t *testing.T) {
	claim := fakePersistentVolumeClaim("default", "data", "10Gi", "5Gi")
	mountDeployment := fakeDeploymentWithTemplate("default", "data-mount", corev1.PodSpec{
		Volumes: []corev1.Volume{{
			Name:         "data",
			VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "data"}},
		}},
	})
	unrelatedStatefulSet := fakeStatefulSetWithTemplate("default", "other", corev1.PodSpec{
		Volumes: []corev1.Volume{{
			Name:         "data",
			VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "other-data"}},
		}},
	})
	repo := newFakeRepo(claim, mountDeployment, unrelatedStatefulSet)
	detail, err := repo.PersistentVolumeClaimDetail(context.Background(), testClusterUID, "default", "data")
	if err != nil {
		t.Fatalf("PersistentVolumeClaimDetail: %v", err)
	}
	if detail.Claim.Name != "data" || detail.Claim.Capacity != "10Gi" {
		t.Fatalf("claim = %+v", detail.Claim)
	}
	got := workloadRefKeys(detail.RelatedWorkloads)
	if got != "Deployment/default/data-mount" {
		t.Fatalf("related = %s", got)
	}
}

func TestConfigMapDetailNotFound(t *testing.T) {
	repo := newFakeRepo()
	if _, err := repo.ConfigMapDetail(context.Background(), testClusterUID, "default", "missing"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestSecretDetailNotFound(t *testing.T) {
	repo := newFakeRepo()
	if _, err := repo.SecretDetail(context.Background(), testClusterUID, "default", "missing"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestPersistentVolumeClaimDetailNotFound(t *testing.T) {
	repo := newFakeRepo()
	if _, err := repo.PersistentVolumeClaimDetail(context.Background(), testClusterUID, "default", "missing"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestIngressDetailNotFound(t *testing.T) {
	repo := newFakeRepo()
	if _, err := repo.IngressDetail(context.Background(), testClusterUID, "default", "missing"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func workloadRefKeys(refs []bo.KubeWorkloadRef) string {
	keys := make([]string, 0, len(refs))
	for _, ref := range refs {
		keys = append(keys, ref.Kind+"/"+ref.Namespace+"/"+ref.Name)
	}
	return strings.Join(keys, ",")
}
