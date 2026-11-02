package data

import (
	"context"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// maxDetailEvents 是节点 / Pod 详情返回的最近事件条数上限。
// 事件只是详情的辅助信息，不做分页和检索。
const maxDetailEvents = 20

// ListEvents 读取与某个资源相关的事件，按最后发生时间倒序。
// 优先按 involvedObject.uid 过滤：同名对象重建后旧事件不应挂到新对象上。
// 没有 UID（例如只拿到了名称）时退化为 kind + name 过滤。
func (r *kubernetesRepo) ListEvents(ctx context.Context, clusterUID, namespace, kind, name, uid string, limit int32) ([]bo.KubeEvent, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}
	list, err := client.CoreV1().Events(namespace).List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	items := make([]bo.KubeEvent, 0, len(list.Items))
	for index := range list.Items {
		event := &list.Items[index]
		if !matchesEventObject(event, kind, name, uid) {
			continue
		}
		items = append(items, toKubeEvent(event))
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].LastSeenAt != items[j].LastSeenAt {
			return items[i].LastSeenAt > items[j].LastSeenAt
		}
		if items[i].FirstSeenAt != items[j].FirstSeenAt {
			return items[i].FirstSeenAt > items[j].FirstSeenAt
		}
		return items[i].Reason < items[j].Reason
	})
	if limit <= 0 || limit > maxDetailEvents {
		limit = maxDetailEvents
	}
	if int32(len(items)) > limit {
		items = items[:limit]
	}
	return items, nil
}

// matchesEventObject 判断事件是否属于目标对象。UID 非空时只按 UID 匹配，
// 避免同名但已重建的对象拿到旧事件。
func matchesEventObject(event *corev1.Event, kind, name, uid string) bool {
	if uid != "" {
		return string(event.InvolvedObject.UID) == uid
	}
	if kind != "" && event.InvolvedObject.Kind != kind {
		return false
	}
	return name == "" || event.InvolvedObject.Name == name
}

// toKubeEvent 把事件转成业务对象。次数优先取 count，旧事件用 series.count。
func toKubeEvent(event *corev1.Event) bo.KubeEvent {
	count := event.Count
	if count == 0 && event.Series != nil {
		count = event.Series.Count
	}
	return bo.KubeEvent{
		Type:        event.Type,
		Reason:      event.Reason,
		Message:     event.Message,
		Count:       count,
		Source:      eventSource(event),
		FirstSeenAt: eventFirstSeen(event),
		LastSeenAt:  eventLastSeen(event),
	}
}

// eventSource 取事件的来源组件。新版事件用 reportingComponent，旧版用 source。
func eventSource(event *corev1.Event) string {
	parts := make([]string, 0, 2)
	for _, value := range []string{event.Source.Component, event.Source.Host} {
		if strings.TrimSpace(value) != "" {
			parts = append(parts, strings.TrimSpace(value))
		}
	}
	if len(parts) == 0 && strings.TrimSpace(event.ReportingController) != "" {
		parts = append(parts, strings.TrimSpace(event.ReportingController))
	}
	return strings.Join(parts, "/")
}

// eventFirstSeen 取事件首次发生时间，字段为空时回落到创建时间。
func eventFirstSeen(event *corev1.Event) int64 {
	if !event.FirstTimestamp.IsZero() {
		return unixTime(event.FirstTimestamp)
	}
	if !event.EventTime.IsZero() {
		return event.EventTime.Unix()
	}
	return unixTime(event.CreationTimestamp)
}

// eventLastSeen 取事件最近发生时间，字段为空时回落到创建时间。
func eventLastSeen(event *corev1.Event) int64 {
	if !event.LastTimestamp.IsZero() {
		return unixTime(event.LastTimestamp)
	}
	if !event.EventTime.IsZero() {
		return event.EventTime.Unix()
	}
	return unixTime(event.CreationTimestamp)
}
