package data

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/aide-family/moon/pkg/merr"
)

// fakeBatchJob 构造一个带镜像的最小 Job，测试按需覆盖 spec 和 status。
func fakeBatchJob(namespace, name string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name, CreationTimestamp: testTime()},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "job", Image: "busybox:1"}},
			}},
		},
	}
}

// fakeCronJobObject 构造带模板、时区和并发策略的 CronJob。
func fakeCronJobObject(namespace, name, uid string) *batchv1.CronJob {
	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace, Name: name, UID: types.UID(uid), CreationTimestamp: testTime(),
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          "0 2 * * *",
			TimeZone:          stringPtr("Asia/Shanghai"),
			ConcurrencyPolicy: batchv1.ForbidConcurrent,
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "backup"}},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "backup", Image: "backup:1"}},
					}},
				},
			},
		},
	}
}

func TestListJobsAndStatus(t *testing.T) {
	suspend := true
	suspended := fakeBatchJob("default", "a-suspended")
	suspended.Spec.Suspend = &suspend
	suspended.Spec.Completions = int32Ptr(1)
	suspended.Spec.Parallelism = int32Ptr(1)

	backoffLimit := int32(2)
	failed := fakeBatchJob("default", "b-failed")
	failed.Spec.BackoffLimit = &backoffLimit
	failed.Status.Failed = 3

	complete := fakeBatchJob("default", "c-complete")
	complete.Spec.Completions = int32Ptr(3)
	complete.Status.Succeeded = 3

	single := fakeBatchJob("default", "d-single")
	single.Status.Succeeded = 1

	started := testTime()
	running := fakeBatchJob("default", "e-running")
	running.Status.Active = 1
	running.Status.StartTime = &started

	unknown := fakeBatchJob("default", "f-unknown")
	unknown.Spec.Completions = int32Ptr(2)
	unknown.Status.StartTime = &started

	pending := fakeBatchJob("default", "g-pending")

	owned := fakeBatchJob("other", "backup-123")
	owned.OwnerReferences = []metav1.OwnerReference{{APIVersion: "batch/v1", Kind: "CronJob", Name: "backup"}}
	owned.Status.Succeeded = 1
	owned.Status.StartTime = &started
	owned.Status.CompletionTime = &started

	repo := newFakeRepo(suspended, failed, complete, single, running, unknown, pending, owned)
	items, err := repo.ListJobs(context.Background(), testClusterUID, "", "")
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(items) != 8 {
		t.Fatalf("jobs = %+v", items)
	}
	// 排序口径是 namespace、name。
	if items[0].Namespace != "default" || items[0].Name != "a-suspended" || items[7].Namespace != "other" {
		t.Fatalf("sorted jobs = %+v", items)
	}
	wantStatus := map[string]string{
		"a-suspended": jobStatusSuspended,
		"b-failed":    jobStatusFailed,
		"c-complete":  jobStatusComplete,
		"d-single":    jobStatusComplete,
		"e-running":   jobStatusRunning,
		"f-unknown":   jobStatusUnknown,
		"g-pending":   jobStatusPending,
	}
	for _, item := range items[:7] {
		if item.Status != wantStatus[item.Name] {
			t.Fatalf("%s status = %q, want %q", item.Name, item.Status, wantStatus[item.Name])
		}
	}
	if items[7].Status != jobStatusComplete {
		t.Fatalf("owned job status = %q", items[7].Status)
	}
	if items[7].OwnerKind != "CronJob" || items[7].OwnerName != "backup" {
		t.Fatalf("owner = %s/%s", items[7].OwnerKind, items[7].OwnerName)
	}
	if items[7].StartTime == 0 || items[7].CompletionTime == 0 {
		t.Fatalf("job times = %d/%d", items[7].StartTime, items[7].CompletionTime)
	}
	if len(items[7].Images) != 1 || items[7].Images[0] != "busybox:1" {
		t.Fatalf("job images = %v", items[7].Images)
	}
	// 未设置 completions / parallelism 时为 0。
	if items[5].Completions != 2 || items[5].Parallelism != 0 {
		t.Fatalf("job counts = %+v", items[5])
	}
	filtered, err := repo.ListJobs(context.Background(), testClusterUID, "default", "FAIL")
	if err != nil {
		t.Fatalf("ListJobs filtered: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "b-failed" {
		t.Fatalf("filtered jobs = %+v", filtered)
	}
	otherNamespace, err := repo.ListJobs(context.Background(), testClusterUID, "other", "")
	if err != nil {
		t.Fatalf("ListJobs namespace: %v", err)
	}
	if len(otherNamespace) != 1 || otherNamespace[0].Namespace != "other" {
		t.Fatalf("other jobs = %+v", otherNamespace)
	}
}

func TestJobDetail(t *testing.T) {
	job := fakeBatchJob("default", "backup")
	job.Labels = map[string]string{"app": "backup"}
	other := fakeBatchJob("default", "other")
	repo := newFakeRepo(
		job,
		other,
		fakePod("default", "backup-2", "node-a", "Job", "backup"),
		fakePod("default", "backup-1", "node-b", "Job", "backup"),
		fakePod("default", "other-1", "node-a", "Job", "other"),
		fakePod("default", "standalone", "node-a", "", ""),
	)
	detail, err := repo.JobDetail(context.Background(), testClusterUID, "default", "backup")
	if err != nil {
		t.Fatalf("JobDetail: %v", err)
	}
	if detail.Job.Name != "backup" || len(detail.Job.Images) != 1 || detail.Job.Images[0] != "busybox:1" {
		t.Fatalf("job = %+v", detail.Job)
	}
	if len(detail.Pods) != 2 || detail.Pods[0].Name != "backup-1" || detail.Pods[1].Name != "backup-2" {
		t.Fatalf("pods = %+v", detail.Pods)
	}
	if _, err := repo.JobDetail(context.Background(), testClusterUID, "default", "missing"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
	if _, err := repo.JobDetail(context.Background(), testClusterUID, "default", " "); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func TestDeleteJob(t *testing.T) {
	clientset := fake.NewSimpleClientset(fakeBatchJob("default", "backup"))
	repo := newFakeRepoWithClientset(clientset)
	if err := repo.DeleteJob(context.Background(), testClusterUID, "default", "backup"); err != nil {
		t.Fatalf("DeleteJob: %v", err)
	}
	if _, err := clientset.BatchV1().Jobs("default").Get(context.Background(), "backup", metav1.GetOptions{}); !merr.IsNotFound(mustKubeError(err)) {
		t.Fatalf("job should be deleted, got %v", err)
	}
	if err := repo.DeleteJob(context.Background(), testClusterUID, "default", "missing"); !merr.IsNotFound(mustKubeError(err)) {
		t.Fatalf("err = %v, want not found", err)
	}
	if err := repo.DeleteJob(context.Background(), testClusterUID, "default", ""); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func TestListCronJobs(t *testing.T) {
	suspended := fakeCronJobObject("default", "backup", "backup-uid")
	suspend := true
	suspended.Spec.Suspend = &suspend
	lastSchedule := testTime()
	suspended.Status.LastScheduleTime = &lastSchedule
	suspended.Status.Active = []corev1.ObjectReference{
		{Kind: "Job", Namespace: "default", Name: "backup-1"},
		{Kind: "Job", Namespace: "default", Name: "backup-2"},
	}
	other := fakeCronJobObject("other", "report", "report-uid")
	other.Spec.Schedule = "*/5 * * * *"
	other.Spec.ConcurrencyPolicy = batchv1.ReplaceConcurrent
	other.Spec.TimeZone = nil

	repo := newFakeRepo(suspended, other)
	items, err := repo.ListCronJobs(context.Background(), testClusterUID, "", "")
	if err != nil {
		t.Fatalf("ListCronJobs: %v", err)
	}
	if len(items) != 2 || items[0].Namespace != "default" || items[1].Namespace != "other" {
		t.Fatalf("cronjobs = %+v", items)
	}
	backup := items[0]
	if backup.Name != "backup" || backup.Schedule != "0 2 * * *" || backup.TimeZone != "Asia/Shanghai" ||
		!backup.Suspend || backup.ConcurrencyPolicy != "Forbid" || backup.ActiveCount != 2 {
		t.Fatalf("cronjob = %+v", backup)
	}
	if backup.LastScheduleTime != lastSchedule.Unix() || backup.CreatedAt != testTime().Unix() {
		t.Fatalf("cronjob times = %+v", backup)
	}
	if strings.Join(backup.Images, ",") != "backup:1" {
		t.Fatalf("cronjob images = %v", backup.Images)
	}
	if items[1].TimeZone != "" || items[1].Suspend || items[1].ConcurrencyPolicy != "Replace" {
		t.Fatalf("cronjob = %+v", items[1])
	}
	filtered, err := repo.ListCronJobs(context.Background(), testClusterUID, "other", "REP")
	if err != nil {
		t.Fatalf("ListCronJobs filtered: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "report" {
		t.Fatalf("filtered cronjobs = %+v", filtered)
	}
}

func TestSetCronJobSuspend(t *testing.T) {
	clientset := fake.NewSimpleClientset(fakeCronJobObject("default", "backup", "backup-uid"))
	repo := newFakeRepoWithClientset(clientset)
	item, err := repo.SetCronJobSuspend(context.Background(), testClusterUID, "default", "backup", true)
	if err != nil {
		t.Fatalf("SetCronJobSuspend: %v", err)
	}
	if !item.Suspend {
		t.Fatalf("item = %+v", item)
	}
	stored, err := clientset.BatchV1().CronJobs("default").Get(context.Background(), "backup", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get cronjob: %v", err)
	}
	if stored.Spec.Suspend == nil || !*stored.Spec.Suspend {
		t.Fatalf("cronjob suspend = %+v", stored.Spec.Suspend)
	}
	item, err = repo.SetCronJobSuspend(context.Background(), testClusterUID, "default", "backup", false)
	if err != nil {
		t.Fatalf("SetCronJobSuspend resume: %v", err)
	}
	if item.Suspend {
		t.Fatalf("item = %+v", item)
	}
	if _, err := repo.SetCronJobSuspend(context.Background(), testClusterUID, "default", "missing", true); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestTriggerCronJob(t *testing.T) {
	clientset := fake.NewSimpleClientset(fakeCronJobObject("default", "backup", "backup-uid"))
	repo := newFakeRepoWithClientset(clientset)
	item, err := repo.TriggerCronJob(context.Background(), testClusterUID, "default", "backup")
	if err != nil {
		t.Fatalf("TriggerCronJob: %v", err)
	}
	if !strings.HasPrefix(item.Name, "backup-manual-") {
		t.Fatalf("job name = %q", item.Name)
	}
	if item.Namespace != "default" || len(item.Images) != 1 || item.Images[0] != "backup:1" {
		t.Fatalf("job = %+v", item)
	}
	created, err := clientset.BatchV1().Jobs("default").Get(context.Background(), item.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get job: %v", err)
	}
	// 复制 jobTemplate 标签并追加立即运行标记。
	if created.Labels["app"] != "backup" || created.Labels[manualJobLabelKey] != manualJobLabelValue {
		t.Fatalf("job labels = %+v", created.Labels)
	}
	// 不设置 CronJob ownerReference，避免被历史清理策略回收。
	if len(created.OwnerReferences) != 0 {
		t.Fatalf("job owner references = %+v", created.OwnerReferences)
	}
	if created.Spec.Template.Spec.Containers[0].Image != "backup:1" {
		t.Fatalf("job spec = %+v", created.Spec.Template.Spec)
	}
}

func TestTriggerCronJobTruncatesLongName(t *testing.T) {
	cronJob := fakeCronJobObject("default", strings.Repeat("a", 60), "long-uid")
	repo := newFakeRepo(cronJob)
	item, err := repo.TriggerCronJob(context.Background(), testClusterUID, "default", cronJob.Name)
	if err != nil {
		t.Fatalf("TriggerCronJob: %v", err)
	}
	if len(item.Name) != maxKubeNameLength {
		t.Fatalf("job name length = %d, name = %q", len(item.Name), item.Name)
	}
	// 保留后缀 <cronjob>-manual-<unix秒>，截断前缀。
	wantPrefix := strings.Repeat("a", maxKubeNameLength-len("-manual-")-10)
	if !strings.HasPrefix(item.Name, wantPrefix+"-manual-") {
		t.Fatalf("job name = %q, want prefix %q", item.Name, wantPrefix)
	}
	suffix := strings.TrimPrefix(item.Name, wantPrefix+"-manual-")
	if len(suffix) != 10 {
		t.Fatalf("job name suffix = %q", suffix)
	}
}

func TestCronJobDetail(t *testing.T) {
	cronJob := fakeCronJobObject("default", "backup", "backup-uid")
	objects := []runtime.Object{cronJob}
	for index := 0; index < 12; index++ {
		job := fakeBatchJob("default", "backup-"+twoDigit(index))
		job.OwnerReferences = []metav1.OwnerReference{{
			APIVersion: "batch/v1", Kind: "CronJob", Name: "backup", UID: types.UID("backup-uid"),
		}}
		job.CreationTimestamp = metav1.NewTime(time.Unix(testTime().Unix()+int64(index), 0))
		objects = append(objects, job)
	}
	unowned := fakeBatchJob("default", "backup-manual-1")
	unowned.Status.Succeeded = 1
	otherOwner := fakeBatchJob("default", "other-1")
	otherOwner.OwnerReferences = []metav1.OwnerReference{{Kind: "CronJob", Name: "other", UID: types.UID("other-uid")}}
	objects = append(objects, unowned, otherOwner)
	repo := newFakeRepo(objects...)
	detail, err := repo.CronJobDetail(context.Background(), testClusterUID, "default", "backup")
	if err != nil {
		t.Fatalf("CronJobDetail: %v", err)
	}
	if detail.CronJob.Name != "backup" || detail.CronJob.Schedule != "0 2 * * *" {
		t.Fatalf("cronjob = %+v", detail.CronJob)
	}
	// 最多 10 条，创建时间倒序；立即运行创建的独立 Job 不计入。
	if len(detail.Jobs) != maxCronJobDetailJobs {
		t.Fatalf("jobs = %+v", detail.Jobs)
	}
	if detail.Jobs[0].Name != "backup-11" || detail.Jobs[9].Name != "backup-02" {
		t.Fatalf("job order = %s ... %s", detail.Jobs[0].Name, detail.Jobs[9].Name)
	}
	if _, err := repo.CronJobDetail(context.Background(), testClusterUID, "default", "missing"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestDeleteCronJob(t *testing.T) {
	clientset := fake.NewSimpleClientset(fakeCronJobObject("default", "backup", "backup-uid"))
	repo := newFakeRepoWithClientset(clientset)
	if err := repo.DeleteCronJob(context.Background(), testClusterUID, "default", "backup"); err != nil {
		t.Fatalf("DeleteCronJob: %v", err)
	}
	if _, err := clientset.BatchV1().CronJobs("default").Get(context.Background(), "backup", metav1.GetOptions{}); !merr.IsNotFound(mustKubeError(err)) {
		t.Fatalf("cronjob should be deleted, got %v", err)
	}
	if err := repo.DeleteCronJob(context.Background(), testClusterUID, "default", "missing"); !merr.IsNotFound(mustKubeError(err)) {
		t.Fatalf("err = %v, want not found", err)
	}
}

// twoDigit 把 0 到 99 补成两位数字，便于断言排序后的 Job 名称。
func twoDigit(value int) string {
	return fmt.Sprintf("%02d", value)
}

func TestJobStatusWithZeroCompletions(t *testing.T) {
	finished := fakeBatchJob("default", "finished")
	finished.Spec.Completions = int32Ptr(0)
	finished.Status.Succeeded = 1
	running := fakeBatchJob("default", "running")
	running.Spec.Completions = int32Ptr(0)
	running.Status.Active = 1
	repo := newFakeRepo(finished, running)
	items, err := repo.ListJobs(context.Background(), testClusterUID, "default", "")
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(items) != 2 || items[0].Status != jobStatusComplete || items[1].Status != jobStatusRunning {
		t.Fatalf("jobs = %+v", items)
	}
	if items[0].Completions != 0 || items[0].Succeeded != 1 {
		t.Fatalf("job = %+v", items[0])
	}
}
