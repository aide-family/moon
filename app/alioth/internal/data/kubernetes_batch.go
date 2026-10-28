package data

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// jobStatusSuspended 等是 Job 展示状态的固定取值，判定顺序见 m16 文档。
	jobStatusSuspended = "Suspended"
	jobStatusFailed    = "Failed"
	jobStatusComplete  = "Complete"
	jobStatusRunning   = "Running"
	jobStatusUnknown   = "Unknown"
	jobStatusPending   = "Pending"
	// defaultJobBackoffLimit 是 Job 未设置 backoffLimit 时 Kubernetes 的默认值。
	defaultJobBackoffLimit = 6
	// manualJobLabelKey 标记立即运行创建的 Job，避免控制器和历史清理策略把它当成调度产物。
	manualJobLabelKey = "cronjob.kubernetes.io/instantiate"
	// manualJobLabelValue 是立即运行 Job 的标签值。
	manualJobLabelValue = "manual"
	// maxKubeNameLength 是 Job 名称允许的最大长度。
	maxKubeNameLength = 63
	// maxCronJobDetailJobs 是 CronJob 详情返回的 Job 条数上限。
	maxCronJobDetailJobs = 10
)

// ListJobs 返回全部匹配的 Job，按 namespace、name 升序，分页由调用方切片。
func (r *kubernetesRepo) ListJobs(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubeJob, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	list, err := client.BatchV1().Jobs(strings.TrimSpace(namespace)).List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	items := make([]bo.KubeJob, 0, len(list.Items))
	for index := range list.Items {
		job := &list.Items[index]
		if !matchesName(job.Name, name) {
			continue
		}
		items = append(items, toKubeJob(job))
	}
	sortKubeJobs(items)
	return items, nil
}

// JobDetail 返回 Job 和它的 Pod。Pod 按名称升序，其它控制器的 Pod 不计入。
func (r *kubernetesRepo) JobDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeJobDetail, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, merr.Invalid("namespace and name are required")
	}
	job, err := client.BatchV1().Jobs(namespace).Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	detail := &bo.KubeJobDetail{Job: toKubeJob(job)}
	pods, err := client.CoreV1().Pods(namespace).List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	for index := range pods.Items {
		pod := &pods.Items[index]
		if len(pod.OwnerReferences) == 0 {
			continue
		}
		if pod.OwnerReferences[0].Kind != "Job" || pod.OwnerReferences[0].Name != name {
			continue
		}
		detail.Pods = append(detail.Pods, toKubePod(pod))
	}
	sortKubePods(detail.Pods)
	return detail, nil
}

// DeleteJob 删除 Job，使用默认的后台级联删除，Pod 由控制器清理。
func (r *kubernetesRepo) DeleteJob(ctx context.Context, clusterUID, namespace, name string) error {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return merr.Invalid("namespace and name are required")
	}
	if err := client.BatchV1().Jobs(namespace).Delete(callCtx, name, metav1.DeleteOptions{}); err != nil {
		return kubeError(r.log, err)
	}
	return nil
}

// ListCronJobs 返回全部匹配的 CronJob，按 namespace、name 升序，分页由调用方切片。
func (r *kubernetesRepo) ListCronJobs(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubeCronJob, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	list, err := client.BatchV1().CronJobs(strings.TrimSpace(namespace)).List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	items := make([]bo.KubeCronJob, 0, len(list.Items))
	for index := range list.Items {
		cronJob := &list.Items[index]
		if !matchesName(cronJob.Name, name) {
			continue
		}
		items = append(items, toKubeCronJob(cronJob))
	}
	sortKubeCronJobs(items)
	return items, nil
}

// CronJobDetail 返回 CronJob 和它最近创建的 Job，最多 10 条，创建时间倒序。
// 只认 ownerReference UID 指向该 CronJob 的 Job，立即运行创建的独立 Job 不计入。
func (r *kubernetesRepo) CronJobDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeCronJobDetail, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, merr.Invalid("namespace and name are required")
	}
	cronJob, err := client.BatchV1().CronJobs(namespace).Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	detail := &bo.KubeCronJobDetail{CronJob: toKubeCronJob(cronJob)}
	jobs, err := client.BatchV1().Jobs(namespace).List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	for index := range jobs.Items {
		job := &jobs.Items[index]
		if !ownedByUID(job.OwnerReferences, cronJob.UID) {
			continue
		}
		detail.Jobs = append(detail.Jobs, toKubeJob(job))
	}
	sort.SliceStable(detail.Jobs, func(i, j int) bool {
		if detail.Jobs[i].CreatedAt != detail.Jobs[j].CreatedAt {
			return detail.Jobs[i].CreatedAt > detail.Jobs[j].CreatedAt
		}
		return detail.Jobs[i].Name > detail.Jobs[j].Name
	})
	if len(detail.Jobs) > maxCronJobDetailJobs {
		detail.Jobs = detail.Jobs[:maxCronJobDetailJobs]
	}
	return detail, nil
}

// SetCronJobSuspend 设置 spec.suspend 并返回更新后的 CronJob。
func (r *kubernetesRepo) SetCronJobSuspend(ctx context.Context, clusterUID, namespace, name string, suspend bool) (*bo.KubeCronJob, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, merr.Invalid("namespace and name are required")
	}
	cronJob, err := client.BatchV1().CronJobs(namespace).Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	cronJob.Spec.Suspend = &suspend
	updated, err := client.BatchV1().CronJobs(namespace).Update(callCtx, cronJob, metav1.UpdateOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	item := toKubeCronJob(updated)
	return &item, nil
}

// TriggerCronJob 用 jobTemplate 立即创建一个独立 Job。
// 名称是 <cronjob>-manual-<unix秒>，不设置 CronJob ownerReference，避免被历史清理策略回收。
func (r *kubernetesRepo) TriggerCronJob(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeJob, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, merr.Invalid("namespace and name are required")
	}
	cronJob, err := client.BatchV1().CronJobs(namespace).Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	labels := make(map[string]string, len(cronJob.Spec.JobTemplate.Labels)+1)
	for key, value := range cronJob.Spec.JobTemplate.Labels {
		labels[key] = value
	}
	labels[manualJobLabelKey] = manualJobLabelValue
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      manualJobName(cronJob.Name, time.Now()),
			Labels:    labels,
		},
		Spec: *cronJob.Spec.JobTemplate.Spec.DeepCopy(),
	}
	created, err := client.BatchV1().Jobs(namespace).Create(callCtx, job, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			// 名称以秒为单位，同一秒内重复触发会撞名；让调用方稍后重试。
			return nil, merr.Conflict("a manual job with the same name already exists")
		}
		return nil, kubeError(r.log, err)
	}
	item := toKubeJob(created)
	return &item, nil
}

// DeleteCronJob 删除 CronJob。默认的删除策略会级联清理它创建的 Job。
func (r *kubernetesRepo) DeleteCronJob(ctx context.Context, clusterUID, namespace, name string) error {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return merr.Invalid("namespace and name are required")
	}
	if err := client.BatchV1().CronJobs(namespace).Delete(callCtx, name, metav1.DeleteOptions{}); err != nil {
		return kubeError(r.log, err)
	}
	return nil
}

// toKubeJob 把 Job 转成业务对象。状态按 m16 文档的顺序判定。
func toKubeJob(job *batchv1.Job) bo.KubeJob {
	item := bo.KubeJob{
		Namespace: job.Namespace,
		Name:      job.Name,
		Status:    jobStatus(job),
		Active:    job.Status.Active,
		Succeeded: job.Status.Succeeded,
		Failed:    job.Status.Failed,
		Images:    workloadImages(job.Spec.Template.Spec),
		CreatedAt: unixTime(job.CreationTimestamp),
	}
	if job.Spec.Completions != nil {
		item.Completions = *job.Spec.Completions
	}
	if job.Spec.Parallelism != nil {
		item.Parallelism = *job.Spec.Parallelism
	}
	if len(job.OwnerReferences) > 0 {
		item.OwnerKind = job.OwnerReferences[0].Kind
		item.OwnerName = job.OwnerReferences[0].Name
	}
	if job.Status.StartTime != nil {
		item.StartTime = job.Status.StartTime.Unix()
	}
	if job.Status.CompletionTime != nil {
		item.CompletionTime = job.Status.CompletionTime.Unix()
	}
	return item
}

// jobStatus 按固定顺序判定状态：暂停 > 失败 > 完成 > 运行中 > 有开始时间 > 等待中。
func jobStatus(job *batchv1.Job) string {
	if job.Spec.Suspend != nil && *job.Spec.Suspend {
		return jobStatusSuspended
	}
	backoffLimit := int32(defaultJobBackoffLimit)
	if job.Spec.BackoffLimit != nil {
		backoffLimit = *job.Spec.BackoffLimit
	}
	if job.Status.Failed > backoffLimit {
		return jobStatusFailed
	}
	completions := int32(0)
	if job.Spec.Completions != nil {
		completions = *job.Spec.Completions
	}
	if completions > 0 {
		if job.Status.Succeeded >= completions {
			return jobStatusComplete
		}
	} else if job.Status.Succeeded > 0 {
		// completions 为 0（未设置或显式 0）时，成功过一次即视为完成。
		return jobStatusComplete
	}
	if job.Status.Active > 0 {
		return jobStatusRunning
	}
	if job.Status.StartTime != nil {
		return jobStatusUnknown
	}
	return jobStatusPending
}

// toKubeCronJob 把 CronJob 转成业务对象，字段全部取自真实对象。
func toKubeCronJob(cronJob *batchv1.CronJob) bo.KubeCronJob {
	item := bo.KubeCronJob{
		Namespace:         cronJob.Namespace,
		Name:              cronJob.Name,
		Schedule:          cronJob.Spec.Schedule,
		Suspend:           cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend,
		ConcurrencyPolicy: string(cronJob.Spec.ConcurrencyPolicy),
		ActiveCount:       int32(len(cronJob.Status.Active)),
		Images:            workloadImages(cronJob.Spec.JobTemplate.Spec.Template.Spec),
		CreatedAt:         unixTime(cronJob.CreationTimestamp),
	}
	if cronJob.Spec.TimeZone != nil {
		item.TimeZone = *cronJob.Spec.TimeZone
	}
	if cronJob.Status.LastScheduleTime != nil {
		item.LastScheduleTime = cronJob.Status.LastScheduleTime.Unix()
	}
	return item
}

// manualJobName 生成 <cronjob>-manual-<unix秒>。名称超过 63 个字符时截断前缀，保留可读后缀。
func manualJobName(cronJobName string, now time.Time) string {
	suffix := "-manual-" + strconv.FormatInt(now.Unix(), 10)
	limit := maxKubeNameLength - len(suffix)
	if limit < 1 {
		limit = 1
	}
	prefix := cronJobName
	if len(prefix) > limit {
		prefix = prefix[:limit]
	}
	return prefix + suffix
}

func sortKubeJobs(items []bo.KubeJob) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace != items[j].Namespace {
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].Name < items[j].Name
	})
}

func sortKubeCronJobs(items []bo.KubeCronJob) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace != items[j].Namespace {
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].Name < items[j].Name
	})
}
