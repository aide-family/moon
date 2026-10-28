package data

import (
	"context"
	"strings"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aide-family/moon/pkg/merr"
)

// fakeYAMLObjectMeta 带上服务端记账字段和 kubectl 旧配置注解，用于验证它们会被去掉。
func fakeYAMLObjectMeta(namespace, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: namespace,
		Name:      name,
		Annotations: map[string]string{
			lastAppliedAnnotation: `{"data":{"stale":"stale-config-value"}}`,
			"example.com/keep":    "keep-me",
		},
		ManagedFields: []metav1.ManagedFieldsEntry{{
			Manager:    "kubectl",
			Operation:  metav1.ManagedFieldsOperationApply,
			APIVersion: "v1",
		}},
	}
}

func TestResourceYAMLNode(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: fakeYAMLObjectMeta("", "node-a"),
		Spec:       corev1.NodeSpec{Unschedulable: true},
	}
	repo := newFakeRepo(node)
	// 集群级资源忽略 namespace 参数。
	item, err := repo.ResourceYAML(context.Background(), testClusterUID, "Node", "ignored-namespace", "node-a")
	if err != nil {
		t.Fatalf("ResourceYAML: %v", err)
	}
	if item.Redacted {
		t.Fatal("node yaml should not be redacted")
	}
	for _, want := range []string{"apiVersion: v1", "kind: Node", "name: node-a", "unschedulable: true", "example.com/keep: keep-me"} {
		if !strings.Contains(item.YAML, want) {
			t.Fatalf("yaml missing %q:\n%s", want, item.YAML)
		}
	}
	for _, unwanted := range []string{"managedFields", "last-applied-configuration", "stale-config-value"} {
		if strings.Contains(item.YAML, unwanted) {
			t.Fatalf("yaml should not contain %q:\n%s", unwanted, item.YAML)
		}
	}
}

func TestResourceYAMLDeployment(t *testing.T) {
	replicas := int32(3)
	deployment := &appsv1.Deployment{
		ObjectMeta: fakeYAMLObjectMeta("default", "web"),
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "web"}},
		},
	}
	repo := newFakeRepo(deployment)
	item, err := repo.ResourceYAML(context.Background(), testClusterUID, "Deployment", "default", "web")
	if err != nil {
		t.Fatalf("ResourceYAML: %v", err)
	}
	if item.Redacted {
		t.Fatal("deployment yaml should not be redacted")
	}
	for _, want := range []string{"apiVersion: apps/v1", "kind: Deployment", "name: web", "replicas: 3"} {
		if !strings.Contains(item.YAML, want) {
			t.Fatalf("yaml missing %q:\n%s", want, item.YAML)
		}
	}
}

func TestResourceYAMLService(t *testing.T) {
	service := &corev1.Service{
		ObjectMeta: fakeYAMLObjectMeta("default", "web"),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"app": "web"},
			Ports:    []corev1.ServicePort{{Name: "http", Port: 80}},
		},
	}
	repo := newFakeRepo(service)
	item, err := repo.ResourceYAML(context.Background(), testClusterUID, "Service", "default", "web")
	if err != nil {
		t.Fatalf("ResourceYAML: %v", err)
	}
	for _, want := range []string{"apiVersion: v1", "kind: Service", "name: web"} {
		if !strings.Contains(item.YAML, want) {
			t.Fatalf("yaml missing %q:\n%s", want, item.YAML)
		}
	}
}

func TestResourceYAMLConfigMap(t *testing.T) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: fakeYAMLObjectMeta("default", "app-config"),
		Data: map[string]string{
			"password": "config-secret-value",
			"app":      "config-plain-value",
		},
		BinaryData: map[string][]byte{"bin": []byte("binary-secret-value")},
	}
	repo := newFakeRepo(configMap)
	item, err := repo.ResourceYAML(context.Background(), testClusterUID, "ConfigMap", "default", "app-config")
	if err != nil {
		t.Fatalf("ResourceYAML: %v", err)
	}
	if !item.Redacted {
		t.Fatal("configmap yaml should be redacted")
	}
	for _, want := range []string{"kind: ConfigMap", "name: app-config", "password: <redacted>", "app: <redacted>", "bin: <redacted>"} {
		if !strings.Contains(item.YAML, want) {
			t.Fatalf("yaml missing %q:\n%s", want, item.YAML)
		}
	}
	for _, leaked := range []string{"config-secret-value", "config-plain-value", "binary-secret-value"} {
		if strings.Contains(item.YAML, leaked) {
			t.Fatalf("configmap value %q leaked:\n%s", leaked, item.YAML)
		}
	}
}

func TestResourceYAMLSecret(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: fakeYAMLObjectMeta("default", "web-tls"),
		Type:       corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte("cert-secret-value"),
			"tls.key": []byte("key-secret-value"),
		},
		StringData: map[string]string{"plain": "stringdata-secret-value"},
	}
	repo := newFakeRepo(secret)
	item, err := repo.ResourceYAML(context.Background(), testClusterUID, "Secret", "default", "web-tls")
	if err != nil {
		t.Fatalf("ResourceYAML: %v", err)
	}
	if !item.Redacted {
		t.Fatal("secret yaml should be redacted")
	}
	for _, want := range []string{"kind: Secret", "name: web-tls", "type: kubernetes.io/tls", "tls.crt: <redacted>", "tls.key: <redacted>", "plain: <redacted>"} {
		if !strings.Contains(item.YAML, want) {
			t.Fatalf("yaml missing %q:\n%s", want, item.YAML)
		}
	}
	for _, leaked := range []string{"cert-secret-value", "key-secret-value", "stringdata-secret-value", "Y2VydC1zZWNyZXQtdmFsdWU="} {
		if strings.Contains(item.YAML, leaked) {
			t.Fatalf("secret value %q leaked:\n%s", leaked, item.YAML)
		}
	}
}

func TestResourceYAMLClusterScoped(t *testing.T) {
	volume := &corev1.PersistentVolume{
		ObjectMeta: fakeYAMLObjectMeta("", "pv-data"),
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
		},
	}
	claim := &corev1.PersistentVolumeClaim{
		ObjectMeta: fakeYAMLObjectMeta("default", "data"),
	}
	repo := newFakeRepo(volume, claim)
	// PersistentVolume 忽略命名空间。
	volumeItem, err := repo.ResourceYAML(context.Background(), testClusterUID, "PersistentVolume", "ignored", "pv-data")
	if err != nil {
		t.Fatalf("ResourceYAML PV: %v", err)
	}
	for _, want := range []string{"kind: PersistentVolume", "name: pv-data", "storage: 10Gi"} {
		if !strings.Contains(volumeItem.YAML, want) {
			t.Fatalf("pv yaml missing %q:\n%s", want, volumeItem.YAML)
		}
	}
	claimItem, err := repo.ResourceYAML(context.Background(), testClusterUID, "PersistentVolumeClaim", "default", "data")
	if err != nil {
		t.Fatalf("ResourceYAML PVC: %v", err)
	}
	for _, want := range []string{"kind: PersistentVolumeClaim", "name: data"} {
		if !strings.Contains(claimItem.YAML, want) {
			t.Fatalf("pvc yaml missing %q:\n%s", want, claimItem.YAML)
		}
	}
}

func TestResourceYAMLJobAndCronJob(t *testing.T) {
	job := &batchv1.Job{
		ObjectMeta: fakeYAMLObjectMeta("default", "backup"),
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "backup", Image: "backup:1"}},
			}},
		},
	}
	cronJob := &batchv1.CronJob{
		ObjectMeta: fakeYAMLObjectMeta("default", "nightly"),
		Spec: batchv1.CronJobSpec{
			Schedule: "0 2 * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "backup", Image: "backup:1"}},
					}},
				},
			},
		},
	}
	repo := newFakeRepo(job, cronJob)
	jobItem, err := repo.ResourceYAML(context.Background(), testClusterUID, "Job", "default", "backup")
	if err != nil {
		t.Fatalf("ResourceYAML Job: %v", err)
	}
	if jobItem.Redacted {
		t.Fatal("job yaml should not be redacted")
	}
	for _, want := range []string{"apiVersion: batch/v1", "kind: Job", "name: backup", "image: backup:1"} {
		if !strings.Contains(jobItem.YAML, want) {
			t.Fatalf("job yaml missing %q:\n%s", want, jobItem.YAML)
		}
	}
	cronJobItem, err := repo.ResourceYAML(context.Background(), testClusterUID, "CronJob", "default", "nightly")
	if err != nil {
		t.Fatalf("ResourceYAML CronJob: %v", err)
	}
	if cronJobItem.Redacted {
		t.Fatal("cronjob yaml should not be redacted")
	}
	for _, want := range []string{"apiVersion: batch/v1", "kind: CronJob", "name: nightly", "schedule:", "0 2 * * *", "image: backup:1"} {
		if !strings.Contains(cronJobItem.YAML, want) {
			t.Fatalf("cronjob yaml missing %q:\n%s", want, cronJobItem.YAML)
		}
	}
	for _, unwanted := range []string{"managedFields", "last-applied-configuration", "stale-config-value"} {
		if strings.Contains(jobItem.YAML, unwanted) || strings.Contains(cronJobItem.YAML, unwanted) {
			t.Fatalf("yaml should not contain %q", unwanted)
		}
	}
}

func TestResourceYAMLUnknownKind(t *testing.T) {
	repo := newFakeRepo()
	if _, err := repo.ResourceYAML(context.Background(), testClusterUID, "HorizontalPodAutoscaler", "default", "backup"); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func TestResourceYAMLNotFound(t *testing.T) {
	repo := newFakeRepo()
	if _, err := repo.ResourceYAML(context.Background(), testClusterUID, "Pod", "default", "missing"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestResourceYAMLTooLarge(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "node-a",
			Annotations: map[string]string{"example.com/big": strings.Repeat("a", 600*1024)},
		},
	}
	repo := newFakeRepo(node)
	_, err := repo.ResourceYAML(context.Background(), testClusterUID, "Node", "", "node-a")
	if err == nil {
		t.Fatal("want error for oversized yaml")
	}
	fromError := kerrors.FromError(err)
	if fromError.Reason != merr.ReasonClient || !strings.Contains(fromError.Message, "too large") {
		t.Fatalf("err = %v, want client too-large error", err)
	}
}
