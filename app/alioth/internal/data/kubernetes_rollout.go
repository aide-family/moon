package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// deploymentRevisionAnnotation 是 Deployment 控制器写在自身和 ReplicaSet 上的版本号。
	deploymentRevisionAnnotation = "deployment.kubernetes.io/revision"
)

// historyObject 是工作负载的一层历史对象，可能是 ReplicaSet 或 ControllerRevision。
// 版本号、模板和创建时间都从真实对象读取，不落库。
type historyObject struct {
	revision  int32
	name      string
	template  corev1.PodTemplateSpec
	replicas  int32
	createdAt int64
}

// workloadRollout 汇总工作负载当前的 Pod 模板和历史对象。
// current 是当前版本号；ControllerRevision 没有副本数字段，replicas 取工作负载当前期望副本数。
type workloadRollout struct {
	template corev1.PodTemplateSpec
	current  int32
	replicas int32
	objects  []historyObject
}

// WorkloadImages 返回每个容器的当前镜像与历史版本。
// Deployment 的历史来自名下的 ReplicaSet，StatefulSet / DaemonSet 的历史来自名下的
// ControllerRevision；没有 ownerReference UID 命中的对象不算历史。
func (r *kubernetesRepo) WorkloadImages(ctx context.Context, clusterUID, namespace, kind, name string) ([]bo.KubeWorkloadImage, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace, kind, name = strings.TrimSpace(namespace), strings.TrimSpace(kind), strings.TrimSpace(name)
	if err := validateWorkloadTarget(namespace, kind, name); err != nil {
		return nil, err
	}
	rollout, err := r.workloadRollout(callCtx, client, namespace, kind, name)
	if err != nil {
		return nil, err
	}
	images := make([]bo.KubeWorkloadImage, 0, len(rollout.template.Spec.Containers)+len(rollout.template.Spec.InitContainers))
	collect := func(containers []corev1.Container, init bool) {
		for _, container := range containers {
			repository, tag := splitImage(container.Image)
			item := bo.KubeWorkloadImage{
				Container:  container.Name,
				Init:       init,
				Image:      container.Image,
				Repository: repository,
				Tag:        tag,
			}
			// 历史版本只保留与当前镜像不同的历史镜像，按版本从新到旧、按镜像去重。
			seen := make(map[string]struct{}, len(rollout.objects))
			for _, object := range rollout.objects {
				image := containerImage(object.template.Spec, container.Name, init)
				if image == "" || image == container.Image {
					continue
				}
				if _, ok := seen[image]; ok {
					continue
				}
				seen[image] = struct{}{}
				item.Versions = append(item.Versions, bo.KubeImageVersion{
					Image:     image,
					Revision:  object.revision,
					CreatedAt: object.createdAt,
				})
			}
			images = append(images, item)
		}
	}
	// 普通容器在前、初始化容器在后，和 Pod 模板的声明顺序一致。
	collect(rollout.template.Spec.Containers, false)
	collect(rollout.template.Spec.InitContainers, true)
	return images, nil
}

// SetWorkloadImage 按名称和 init 定位容器，只改镜像，其它字段不动。
func (r *kubernetesRepo) SetWorkloadImage(ctx context.Context, clusterUID, namespace, kind, name, container string, init bool, image string) (*bo.KubeWorkload, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace, kind, name = strings.TrimSpace(namespace), strings.TrimSpace(kind), strings.TrimSpace(name)
	container, image = strings.TrimSpace(container), strings.TrimSpace(image)
	if err := validateWorkloadTarget(namespace, kind, name); err != nil {
		return nil, err
	}
	if container == "" {
		return nil, merr.Invalid("container is required")
	}
	if image == "" {
		return nil, merr.Invalid("image is required")
	}
	switch kind {
	case workloadKindDeployment:
		deployment, err := client.AppsV1().Deployments(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		if !setContainerImage(&deployment.Spec.Template.Spec, container, init, image) {
			return nil, merr.Invalid("container not found")
		}
		updated, err := client.AppsV1().Deployments(namespace).Update(callCtx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		item := toKubeDeployment(updated)
		return &item, nil
	case workloadKindStatefulSet:
		statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		if !setContainerImage(&statefulSet.Spec.Template.Spec, container, init, image) {
			return nil, merr.Invalid("container not found")
		}
		updated, err := client.AppsV1().StatefulSets(namespace).Update(callCtx, statefulSet, metav1.UpdateOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		item := toKubeStatefulSet(updated)
		return &item, nil
	case workloadKindDaemonSet:
		daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		if !setContainerImage(&daemonSet.Spec.Template.Spec, container, init, image) {
			return nil, merr.Invalid("container not found")
		}
		updated, err := client.AppsV1().DaemonSets(namespace).Update(callCtx, daemonSet, metav1.UpdateOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		item := toKubeDaemonSet(updated)
		return &item, nil
	default:
		return nil, merr.Invalid("invalid resource kind")
	}
}

// WorkloadRevisions 返回历史版本，新版本在前。Current 只标记当前版本，
// 回滚到当前版本会被拒绝，因此前端需要这个标记。
func (r *kubernetesRepo) WorkloadRevisions(ctx context.Context, clusterUID, namespace, kind, name string) ([]bo.KubeWorkloadRevision, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace, kind, name = strings.TrimSpace(namespace), strings.TrimSpace(kind), strings.TrimSpace(name)
	if err := validateWorkloadTarget(namespace, kind, name); err != nil {
		return nil, err
	}
	rollout, err := r.workloadRollout(callCtx, client, namespace, kind, name)
	if err != nil {
		return nil, err
	}
	revisions := make([]bo.KubeWorkloadRevision, 0, len(rollout.objects))
	for _, object := range rollout.objects {
		replicas := object.replicas
		if kind != workloadKindDeployment {
			// ControllerRevision 不记录副本数，展示工作负载当前期望副本数。
			replicas = rollout.replicas
		}
		revisions = append(revisions, bo.KubeWorkloadRevision{
			Revision:  object.revision,
			Name:      object.name,
			Current:   rollout.current > 0 && object.revision == rollout.current,
			Replicas:  replicas,
			Images:    workloadImages(object.template.Spec),
			CreatedAt: object.createdAt,
		})
	}
	return revisions, nil
}

// RollbackWorkload 把目标历史对象的模板写回工作负载。
// 目标版本不存在或等于当前版本时返回客户端错误，此时工作负载保持不变。
func (r *kubernetesRepo) RollbackWorkload(ctx context.Context, clusterUID, namespace, kind, name string, revision int32) (*bo.KubeWorkload, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace, kind, name = strings.TrimSpace(namespace), strings.TrimSpace(kind), strings.TrimSpace(name)
	if err := validateWorkloadTarget(namespace, kind, name); err != nil {
		return nil, err
	}
	if revision < 1 {
		return nil, merr.Invalid("revision is required")
	}
	rollout, err := r.workloadRollout(callCtx, client, namespace, kind, name)
	if err != nil {
		return nil, err
	}
	if rollout.current > 0 && revision == rollout.current {
		return nil, merr.Invalid("cannot roll back to the current revision")
	}
	target, ok := rollout.objectFor(revision)
	if !ok {
		return nil, merr.Invalid("workload revision not found")
	}
	switch kind {
	case workloadKindDeployment:
		deployment, err := client.AppsV1().Deployments(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		deployment.Spec.Template = *target.template.DeepCopy()
		updated, err := client.AppsV1().Deployments(namespace).Update(callCtx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		item := toKubeDeployment(updated)
		return &item, nil
	case workloadKindStatefulSet:
		statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		statefulSet.Spec.Template = *target.template.DeepCopy()
		updated, err := client.AppsV1().StatefulSets(namespace).Update(callCtx, statefulSet, metav1.UpdateOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		item := toKubeStatefulSet(updated)
		return &item, nil
	case workloadKindDaemonSet:
		daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		daemonSet.Spec.Template = *target.template.DeepCopy()
		updated, err := client.AppsV1().DaemonSets(namespace).Update(callCtx, daemonSet, metav1.UpdateOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		item := toKubeDaemonSet(updated)
		return &item, nil
	default:
		return nil, merr.Invalid("invalid resource kind")
	}
}

// workloadRollout 读取工作负载当前的 Pod 模板、当前版本号和历史对象。
func (r *kubernetesRepo) workloadRollout(ctx context.Context, client kubernetes.Interface, namespace, kind, name string) (*workloadRollout, error) {
	switch kind {
	case workloadKindDeployment:
		deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		objects, err := r.deploymentHistory(ctx, client, deployment)
		if err != nil {
			return nil, err
		}
		rollout := &workloadRollout{template: deployment.Spec.Template, objects: objects}
		// Deployment 的当前版本号在自身注解上；缺失时无法判定「当前版本」。
		current, ok := revisionOf(deployment.Annotations)
		if ok {
			rollout.current = current
		}
		if deployment.Spec.Replicas != nil {
			rollout.replicas = *deployment.Spec.Replicas
		}
		return rollout, nil
	case workloadKindStatefulSet:
		statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		objects, err := r.controllerRevisionHistory(ctx, client, namespace, statefulSet.UID)
		if err != nil {
			return nil, err
		}
		rollout := &workloadRollout{template: statefulSet.Spec.Template, objects: objects}
		rollout.current = controllerRevisionNumber(objects, statefulSet.Status.CurrentRevision, statefulSet.Status.UpdateRevision)
		if statefulSet.Spec.Replicas != nil {
			rollout.replicas = *statefulSet.Spec.Replicas
		}
		return rollout, nil
	case workloadKindDaemonSet:
		daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		objects, err := r.controllerRevisionHistory(ctx, client, namespace, daemonSet.UID)
		if err != nil {
			return nil, err
		}
		rollout := &workloadRollout{template: daemonSet.Spec.Template, objects: objects}
		// DaemonSet 的 status 不记录当前版本名，最高的 ControllerRevision 就是已应用的版本：
		// 每次模板变更（含回滚）都会生成一个更高的版本号。
		if len(objects) > 0 {
			rollout.current = objects[0].revision
		}
		rollout.replicas = daemonSet.Status.DesiredNumberScheduled
		return rollout, nil
	default:
		return nil, merr.Invalid("invalid resource kind")
	}
}

// deploymentHistory 用 ownerReferences UID 找出属于该 Deployment 的 ReplicaSet。
func (r *kubernetesRepo) deploymentHistory(ctx context.Context, client kubernetes.Interface, deployment *appsv1.Deployment) ([]historyObject, error) {
	replicaSets, err := client.AppsV1().ReplicaSets(deployment.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	objects := make([]historyObject, 0, len(replicaSets.Items))
	for index := range replicaSets.Items {
		replicaSet := &replicaSets.Items[index]
		if !ownedByUID(replicaSet.OwnerReferences, deployment.UID) {
			continue
		}
		// 没有版本号注解的 ReplicaSet 不是一次 rollout 的版本，不参与历史。
		revision, ok := revisionOf(replicaSet.Annotations)
		if !ok {
			continue
		}
		replicas := int32(0)
		if replicaSet.Spec.Replicas != nil {
			replicas = *replicaSet.Spec.Replicas
		}
		objects = append(objects, historyObject{
			revision:  revision,
			name:      replicaSet.Name,
			template:  replicaSet.Spec.Template,
			replicas:  replicas,
			createdAt: unixTime(replicaSet.CreationTimestamp),
		})
	}
	return sortedHistoryObjects(objects), nil
}

// controllerRevisionHistory 用 ownerReferences UID 找出属于该工作负载的 ControllerRevision，
// 模板从 data.raw 反序列化。
func (r *kubernetesRepo) controllerRevisionHistory(ctx context.Context, client kubernetes.Interface, namespace string, ownerUID types.UID) ([]historyObject, error) {
	revisions, err := client.AppsV1().ControllerRevisions(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	objects := make([]historyObject, 0, len(revisions.Items))
	for index := range revisions.Items {
		revision := &revisions.Items[index]
		if !ownedByUID(revision.OwnerReferences, ownerUID) || revision.Revision < 1 {
			continue
		}
		number := int32(revision.Revision)
		var template corev1.PodTemplateSpec
		if len(revision.Data.Raw) > 0 {
			if err := json.Unmarshal(revision.Data.Raw, &template); err != nil {
				// 单个损坏的历史对象不该让整个版本列表失败。
				r.log.Error(fmt.Sprintf("controller revision template: %v", err))
				continue
			}
		}
		objects = append(objects, historyObject{
			revision:  number,
			name:      revision.Name,
			template:  template,
			createdAt: unixTime(revision.CreationTimestamp),
		})
	}
	return sortedHistoryObjects(objects), nil
}

// objectFor 按版本号取历史对象。
func (rollout *workloadRollout) objectFor(revision int32) (historyObject, bool) {
	for _, object := range rollout.objects {
		if object.revision == revision {
			return object, true
		}
	}
	return historyObject{}, false
}

// sortedHistoryObjects 按版本号从新到旧排序，同一版本号只保留最新的历史对象。
func sortedHistoryObjects(objects []historyObject) []historyObject {
	sort.SliceStable(objects, func(i, j int) bool {
		if objects[i].revision != objects[j].revision {
			return objects[i].revision > objects[j].revision
		}
		return objects[i].createdAt > objects[j].createdAt
	})
	unique := objects[:0]
	var previous int32
	for index, object := range objects {
		if index > 0 && object.revision == previous {
			continue
		}
		previous = object.revision
		unique = append(unique, object)
	}
	return unique
}

// controllerRevisionNumber 按 ControllerRevision 名称找当前版本号。
// 优先取 CurrentRevision，没有命中时回退到 UpdateRevision。
func controllerRevisionNumber(objects []historyObject, names ...string) int32 {
	for _, name := range names {
		if name == "" {
			continue
		}
		for _, object := range objects {
			if object.name == name {
				return object.revision
			}
		}
	}
	return 0
}

// validateWorkloadTarget 校验工作负载定位参数。Repo 方法也会被直接调用，
// 不能只依赖 proto 校验。
func validateWorkloadTarget(namespace, kind, name string) error {
	if namespace == "" || name == "" || !validWorkloadKind(kind) || kind == "" {
		return merr.Invalid("invalid request")
	}
	return nil
}

// setContainerImage 按容器名和 init 定位容器，命中时返回 true。
func setContainerImage(spec *corev1.PodSpec, name string, init bool, image string) bool {
	containers := spec.Containers
	if init {
		containers = spec.InitContainers
	}
	for index := range containers {
		if containers[index].Name == name {
			containers[index].Image = image
			return true
		}
	}
	return false
}

// containerImage 取 Pod 模板中指定容器的镜像，容器不存在时返回空串。
func containerImage(spec corev1.PodSpec, name string, init bool) string {
	containers := spec.Containers
	if init {
		containers = spec.InitContainers
	}
	for _, container := range containers {
		if container.Name == name {
			return container.Image
		}
	}
	return ""
}

// splitImage 把镜像拆成仓库和 tag。registry:5000/app:v1 拆成 registry:5000/app 和 v1；
// 没有 tag 时 tag 为空；带 digest 时把 digest 当作 tag 返回。
func splitImage(image string) (string, string) {
	image = strings.TrimSpace(image)
	if image == "" {
		return "", ""
	}
	lastSlash := strings.LastIndex(image, "/")
	if at := strings.LastIndex(image, "@"); at > lastSlash {
		return image[:at], image[at+1:]
	}
	if colon := strings.LastIndex(image, ":"); colon > lastSlash {
		return image[:colon], image[colon+1:]
	}
	return image, ""
}

// revisionOf 从注解里读出版本号，缺失或非法时返回 false。
func revisionOf(annotations map[string]string) (int32, bool) {
	value, ok := annotations[deploymentRevisionAnnotation]
	if !ok {
		return 0, false
	}
	revision, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || revision < 1 {
		return 0, false
	}
	return int32(revision), true
}

// ownedByUID 判断对象的直接属主是否是给定 UID 的对象。同名对象可能存于不同命名空间或
// 重建过多次，UID 才能确认历史对象真的属于当前工作负载。
func ownedByUID(references []metav1.OwnerReference, uid types.UID) bool {
	if uid == "" {
		return false
	}
	for _, reference := range references {
		if reference.UID == uid {
			return true
		}
	}
	return false
}
