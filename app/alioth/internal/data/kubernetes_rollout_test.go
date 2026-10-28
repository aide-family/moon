package data

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/aide-family/moon/pkg/merr"
)

// fakeRolloutDeployment 构造带 UID 和版本号注解的 Deployment，历史对象按 UID 归属。
func fakeRolloutDeployment(namespace, name, uid, revision string, replicas int32, podSpec corev1.PodSpec) *appsv1.Deployment {
	annotations := map[string]string{}
	if revision != "" {
		annotations[deploymentRevisionAnnotation] = revision
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace, Name: name, UID: types.UID(uid), Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(replicas),
			Template: corev1.PodTemplateSpec{Spec: podSpec},
		},
	}
}

// fakeOwnedReplicaSet 构造归属于指定 UID 的 ReplicaSet，revision 为 0 时不写注解。
func fakeOwnedReplicaSet(namespace, name, ownerName string, ownerUID types.UID, revision int32, replicas int32, podSpec corev1.PodSpec) *appsv1.ReplicaSet {
	var annotations map[string]string
	if revision > 0 {
		annotations = map[string]string{deploymentRevisionAnnotation: strconv.Itoa(int(revision))}
	}
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         namespace,
			Name:              name,
			CreationTimestamp: testTime(),
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "apps/v1", Kind: "Deployment", Name: ownerName, UID: ownerUID,
			}},
			Annotations: annotations,
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: int32Ptr(replicas),
			Template: corev1.PodTemplateSpec{Spec: podSpec},
		},
	}
}

// fakeControllerRevisionObject 构造归属于指定 UID 的 ControllerRevision，模板放在 data.raw 里。
func fakeControllerRevisionObject(t *testing.T, namespace, name, ownerName string, ownerUID types.UID, revision int64, podSpec corev1.PodSpec) *appsv1.ControllerRevision {
	t.Helper()
	raw, err := json.Marshal(corev1.PodTemplateSpec{Spec: podSpec})
	if err != nil {
		t.Fatalf("marshal pod template: %v", err)
	}
	return &appsv1.ControllerRevision{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         namespace,
			Name:              name,
			CreationTimestamp: testTime(),
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "apps/v1", Kind: "StatefulSet", Name: ownerName, UID: ownerUID,
			}},
		},
		Revision: revision,
		Data:     runtime.RawExtension{Raw: raw},
	}
}

// podSpecOf 按给定顺序构造 Pod 模板，镜像列表的顺序由调用方决定。
func podSpecOf(containers []corev1.Container, initContainers []corev1.Container) corev1.PodSpec {
	return corev1.PodSpec{Containers: containers, InitContainers: initContainers}
}

// containerOf 构造一个只有名称和镜像的容器。
func containerOf(name, image string) corev1.Container {
	return corev1.Container{Name: name, Image: image}
}

func TestWorkloadImagesDeployment(t *testing.T) {
	deployment := fakeRolloutDeployment("default", "web", "web-uid", "4", 2, podSpecOf(
		[]corev1.Container{containerOf("app", "nginx:4"), containerOf("sidecar", "busybox")},
		[]corev1.Container{containerOf("init", "alpine:4")},
	))
	// app 的历史在 rev2 与 rev1 是同一个镜像，按镜像去重后只保留 rev2。
	replicaSets := []runtime.Object{
		fakeOwnedReplicaSet("default", "web-1", "web", "web-uid", 1, 1, podSpecOf(
			[]corev1.Container{containerOf("app", "nginx:2"), containerOf("sidecar", "busybox")}, []corev1.Container{containerOf("init", "alpine:1")})),
		fakeOwnedReplicaSet("default", "web-2", "web", "web-uid", 2, 0, podSpecOf(
			[]corev1.Container{containerOf("app", "nginx:2"), containerOf("sidecar", "busybox")}, []corev1.Container{containerOf("init", "alpine:2")})),
		fakeOwnedReplicaSet("default", "web-3", "web", "web-uid", 3, 0, podSpecOf(
			[]corev1.Container{containerOf("app", "nginx:3"), containerOf("sidecar", "busybox")}, []corev1.Container{containerOf("init", "alpine:3")})),
		// 别的 Deployment 的历史对象不能算进来。
		fakeOwnedReplicaSet("default", "other-9", "other", "other-uid", 9, 1, podSpecOf(
			[]corev1.Container{containerOf("app", "nginx:9")}, nil)),
		// 没有 UID 归属的 ReplicaSet 也不能算进来。
		fakeOwnedReplicaSet("default", "web-stale", "web", "", 5, 1, podSpecOf(
			[]corev1.Container{containerOf("app", "nginx:5")}, nil)),
	}
	objects := append([]runtime.Object{deployment}, replicaSets...)
	repo := newFakeRepo(objects...)
	images, err := repo.WorkloadImages(context.Background(), testClusterUID, "default", "Deployment", "web")
	if err != nil {
		t.Fatalf("WorkloadImages: %v", err)
	}
	if len(images) != 3 {
		t.Fatalf("images = %+v", images)
	}
	app := images[0]
	if app.Container != "app" || app.Init || app.Image != "nginx:4" || app.Repository != "nginx" || app.Tag != "4" {
		t.Fatalf("app image = %+v", app)
	}
	wantAppVersions := []struct {
		image    string
		revision int32
	}{{"nginx:3", 3}, {"nginx:2", 2}}
	if len(app.Versions) != len(wantAppVersions) {
		t.Fatalf("app versions = %+v", app.Versions)
	}
	for index, want := range wantAppVersions {
		if app.Versions[index].Image != want.image || app.Versions[index].Revision != want.revision {
			t.Fatalf("app version[%d] = %+v, want %s rev %d", index, app.Versions[index], want.image, want.revision)
		}
		if app.Versions[index].CreatedAt == 0 {
			t.Fatalf("app version[%d] should keep created time", index)
		}
	}
	// sidecar 每个版本的镜像都和当前一致，没有历史版本。
	if images[1].Container != "sidecar" || images[1].Tag != "" || len(images[1].Versions) != 0 {
		t.Fatalf("sidecar image = %+v", images[1])
	}
	init := images[2]
	if init.Container != "init" || !init.Init || init.Image != "alpine:4" {
		t.Fatalf("init image = %+v", init)
	}
	if len(init.Versions) != 3 || init.Versions[0].Image != "alpine:3" || init.Versions[2].Image != "alpine:1" {
		t.Fatalf("init versions = %+v", init.Versions)
	}
}

func TestWorkloadImagesStatefulSet(t *testing.T) {
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "db", UID: "db-uid"},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(2),
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "app", Image: "mysql:3"}},
			}},
		},
	}
	revisionOne := fakeControllerRevisionObject(t, "default", "db-1", "db", "db-uid", 1, corev1.PodSpec{
		Containers: []corev1.Container{{Name: "app", Image: "mysql:1"}},
	})
	revisionTwo := fakeControllerRevisionObject(t, "default", "db-2", "db", "db-uid", 2, corev1.PodSpec{
		Containers: []corev1.Container{{Name: "app", Image: "mysql:2"}},
	})
	other := fakeControllerRevisionObject(t, "default", "other-9", "other", "other-uid", 9, corev1.PodSpec{
		Containers: []corev1.Container{{Name: "app", Image: "mysql:9"}},
	})
	repo := newFakeRepo(statefulSet, revisionOne, revisionTwo, other)
	images, err := repo.WorkloadImages(context.Background(), testClusterUID, "default", "StatefulSet", "db")
	if err != nil {
		t.Fatalf("WorkloadImages: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("images = %+v", images)
	}
	app := images[0]
	if app.Image != "mysql:3" || len(app.Versions) != 2 {
		t.Fatalf("app image = %+v", app)
	}
	if app.Versions[0].Image != "mysql:2" || app.Versions[0].Revision != 2 || app.Versions[1].Revision != 1 {
		t.Fatalf("versions = %+v", app.Versions)
	}
}

func TestSetWorkloadImage(t *testing.T) {
	clientset := fake.NewSimpleClientset(fakeDeployment("default", "web", 2))
	repo := newFakeRepoWithClientset(clientset)
	item, err := repo.SetWorkloadImage(context.Background(), testClusterUID, "default", "Deployment", "web", "app", false, "nginx:1.28")
	if err != nil {
		t.Fatalf("SetWorkloadImage: %v", err)
	}
	if item == nil || item.Name != "web" || len(item.Images) != 1 || item.Images[0] != "nginx:1.28" {
		t.Fatalf("item = %+v", item)
	}
	deployment, err := clientset.AppsV1().Deployments("default").Get(context.Background(), "web", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if deployment.Spec.Template.Spec.Containers[0].Image != "nginx:1.28" {
		t.Fatalf("deployment image = %+v", deployment.Spec.Template.Spec.Containers)
	}
	// 只改目标容器，其它字段保持不变。
	if *deployment.Spec.Replicas != 2 {
		t.Fatalf("replicas = %d", *deployment.Spec.Replicas)
	}
	if _, err := repo.SetWorkloadImage(context.Background(), testClusterUID, "default", "Deployment", "web", "missing", false, "nginx:1.28"); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	// init 为 true 时必须找初始化容器，普通容器 app 不算命中。
	if _, err := repo.SetWorkloadImage(context.Background(), testClusterUID, "default", "Deployment", "web", "app", true, "nginx:1.28"); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	if _, err := repo.SetWorkloadImage(context.Background(), testClusterUID, "default", "Deployment", "web", "app", false, "  "); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	if _, err := repo.SetWorkloadImage(context.Background(), testClusterUID, "default", "Deployment", "missing", "app", false, "nginx:1.28"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
	if _, err := repo.SetWorkloadImage(context.Background(), testClusterUID, "default", "Job", "web", "app", false, "nginx:1.28"); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func TestSetWorkloadImageInitContainer(t *testing.T) {
	deployment := fakeDeploymentWithTemplate("default", "web", corev1.PodSpec{
		Containers:     []corev1.Container{{Name: "app", Image: "nginx:1"}},
		InitContainers: []corev1.Container{{Name: "init", Image: "alpine:1"}},
	})
	clientset := fake.NewSimpleClientset(deployment)
	repo := newFakeRepoWithClientset(clientset)
	if _, err := repo.SetWorkloadImage(context.Background(), testClusterUID, "default", "Deployment", "web", "init", true, "alpine:2"); err != nil {
		t.Fatalf("SetWorkloadImage init: %v", err)
	}
	updated, err := clientset.AppsV1().Deployments("default").Get(context.Background(), "web", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if updated.Spec.Template.Spec.InitContainers[0].Image != "alpine:2" {
		t.Fatalf("init image = %+v", updated.Spec.Template.Spec.InitContainers)
	}
	if updated.Spec.Template.Spec.Containers[0].Image != "nginx:1" {
		t.Fatalf("normal container should not change: %+v", updated.Spec.Template.Spec.Containers)
	}
}

func TestWorkloadRevisionsDeployment(t *testing.T) {
	deployment := fakeRolloutDeployment("default", "web", "web-uid", "2", 3, podSpecOf(
		[]corev1.Container{containerOf("app", "nginx:2")}, nil,
	))
	repo := newFakeRepo(
		deployment,
		fakeOwnedReplicaSet("default", "web-2", "web", "web-uid", 2, 3, podSpecOf([]corev1.Container{containerOf("app", "nginx:2")}, nil)),
		fakeOwnedReplicaSet("default", "web-1", "web", "web-uid", 1, 2, podSpecOf([]corev1.Container{containerOf("app", "nginx:1")}, nil)),
		fakeOwnedReplicaSet("default", "web-no-revision", "web", "web-uid", 0, 1, podSpecOf([]corev1.Container{containerOf("app", "nginx:0")}, nil)),
		fakeOwnedReplicaSet("default", "other-9", "other", "other-uid", 9, 1, podSpecOf([]corev1.Container{containerOf("app", "nginx:9")}, nil)),
	)
	revisions, err := repo.WorkloadRevisions(context.Background(), testClusterUID, "default", "Deployment", "web")
	if err != nil {
		t.Fatalf("WorkloadRevisions: %v", err)
	}
	if len(revisions) != 2 {
		t.Fatalf("revisions = %+v", revisions)
	}
	current := revisions[0]
	if current.Revision != 2 || current.Name != "web-2" || !current.Current || current.Replicas != 3 {
		t.Fatalf("current revision = %+v", current)
	}
	if strings.Join(current.Images, ",") != "nginx:2" {
		t.Fatalf("current images = %v", current.Images)
	}
	previous := revisions[1]
	if previous.Revision != 1 || previous.Name != "web-1" || previous.Current || previous.Replicas != 2 {
		t.Fatalf("previous revision = %+v", previous)
	}
}

func TestWorkloadRevisionsStatefulSet(t *testing.T) {
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "db", UID: "db-uid"},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(3),
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "app", Image: "mysql:2"}},
			}},
		},
		Status: appsv1.StatefulSetStatus{CurrentRevision: "db-2"},
	}
	repo := newFakeRepo(
		statefulSet,
		fakeControllerRevisionObject(t, "default", "db-1", "db", "db-uid", 1, corev1.PodSpec{
			Containers: []corev1.Container{{Name: "app", Image: "mysql:1"}},
		}),
		fakeControllerRevisionObject(t, "default", "db-2", "db", "db-uid", 2, corev1.PodSpec{
			Containers: []corev1.Container{{Name: "app", Image: "mysql:2"}},
		}),
	)
	revisions, err := repo.WorkloadRevisions(context.Background(), testClusterUID, "default", "StatefulSet", "db")
	if err != nil {
		t.Fatalf("WorkloadRevisions: %v", err)
	}
	if len(revisions) != 2 {
		t.Fatalf("revisions = %+v", revisions)
	}
	if !revisions[0].Current || revisions[0].Revision != 2 || revisions[0].Replicas != 3 {
		t.Fatalf("current revision = %+v", revisions[0])
	}
	if revisions[1].Current || revisions[1].Revision != 1 {
		t.Fatalf("previous revision = %+v", revisions[1])
	}
}

func TestRollbackWorkloadDeployment(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		fakeRolloutDeployment("default", "web", "web-uid", "3", 2, podSpecOf([]corev1.Container{containerOf("app", "nginx:3")}, nil)),
		fakeOwnedReplicaSet("default", "web-2", "web", "web-uid", 2, 2, corev1.PodSpec{
			Containers: []corev1.Container{{Name: "app", Image: "nginx:2"}},
		}),
		fakeOwnedReplicaSet("default", "web-3", "web", "web-uid", 3, 2, corev1.PodSpec{
			Containers: []corev1.Container{{Name: "app", Image: "nginx:3"}},
		}),
	)
	repo := newFakeRepoWithClientset(clientset)
	item, err := repo.RollbackWorkload(context.Background(), testClusterUID, "default", "Deployment", "web", 2)
	if err != nil {
		t.Fatalf("RollbackWorkload: %v", err)
	}
	if item == nil || item.Name != "web" || strings.Join(item.Images, ",") != "nginx:2" {
		t.Fatalf("item = %+v", item)
	}
	deployment, err := clientset.AppsV1().Deployments("default").Get(context.Background(), "web", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if deployment.Spec.Template.Spec.Containers[0].Image != "nginx:2" {
		t.Fatalf("deployment image = %+v", deployment.Spec.Template.Spec.Containers)
	}
	// 回滚到当前版本（3）和不存在版本（9）都返回客户端错误，模板保持回滚后的状态。
	if _, err := repo.RollbackWorkload(context.Background(), testClusterUID, "default", "Deployment", "web", 3); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	if _, err := repo.RollbackWorkload(context.Background(), testClusterUID, "default", "Deployment", "web", 9); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	if _, err := repo.RollbackWorkload(context.Background(), testClusterUID, "default", "Deployment", "web", 0); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	deployment, err = clientset.AppsV1().Deployments("default").Get(context.Background(), "web", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if deployment.Spec.Template.Spec.Containers[0].Image != "nginx:2" {
		t.Fatalf("failed rollback should keep template, got %+v", deployment.Spec.Template.Spec.Containers)
	}
}

func TestRollbackWorkloadStatefulSet(t *testing.T) {
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "db", UID: "db-uid"},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(3),
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "app", Image: "mysql:2"}},
			}},
		},
		Status: appsv1.StatefulSetStatus{CurrentRevision: "db-2"},
	}
	clientset := fake.NewSimpleClientset(
		statefulSet,
		fakeControllerRevisionObject(t, "default", "db-1", "db", "db-uid", 1, corev1.PodSpec{
			Containers: []corev1.Container{{Name: "app", Image: "mysql:1"}},
		}),
		fakeControllerRevisionObject(t, "default", "db-2", "db", "db-uid", 2, corev1.PodSpec{
			Containers: []corev1.Container{{Name: "app", Image: "mysql:2"}},
		}),
	)
	repo := newFakeRepoWithClientset(clientset)
	item, err := repo.RollbackWorkload(context.Background(), testClusterUID, "default", "StatefulSet", "db", 1)
	if err != nil {
		t.Fatalf("RollbackWorkload: %v", err)
	}
	if item == nil || item.Kind != "StatefulSet" || strings.Join(item.Images, ",") != "mysql:1" {
		t.Fatalf("item = %+v", item)
	}
	updated, err := clientset.AppsV1().StatefulSets("default").Get(context.Background(), "db", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get statefulset: %v", err)
	}
	if updated.Spec.Template.Spec.Containers[0].Image != "mysql:1" {
		t.Fatalf("statefulset image = %+v", updated.Spec.Template.Spec.Containers)
	}
	if _, err := repo.RollbackWorkload(context.Background(), testClusterUID, "default", "StatefulSet", "db", 2); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	if _, err := repo.RollbackWorkload(context.Background(), testClusterUID, "default", "StatefulSet", "db", 7); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func TestWorkloadRolloutRejectsInvalidTargets(t *testing.T) {
	repo := newFakeRepo()
	if _, err := repo.WorkloadImages(context.Background(), testClusterUID, "default", "Job", "backup"); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	if _, err := repo.WorkloadImages(context.Background(), testClusterUID, "", "Deployment", "web"); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	if _, err := repo.WorkloadRevisions(context.Background(), testClusterUID, "default", "Deployment", ""); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	if _, err := repo.WorkloadRevisions(context.Background(), testClusterUID, "default", "Deployment", "missing"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}
