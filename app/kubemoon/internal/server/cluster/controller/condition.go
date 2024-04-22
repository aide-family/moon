package controller

import (
	"github.com/aide-family/moon/api/cluster/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCondition(condType string, status metav1.ConditionStatus, reason, message string) *metav1.Condition {
	return &metav1.Condition{
		Type:               condType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

func GetCondition(status v1beta1.ClusterStatus, condType string) *metav1.Condition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

func GetOrNewCondition(status v1beta1.ClusterStatus, condType string) *metav1.Condition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return NewCondition(condType, metav1.ConditionUnknown, "", "")
}

func SetCondition(status *v1beta1.ClusterStatus, condition metav1.Condition) {
	for i, c := range status.Conditions {
		if c.Type == condition.Type {
			if c.Status != condition.Status {
				condition.LastTransitionTime = metav1.Now()
			}
			status.Conditions[i] = condition
			return
		}
	}
	status.Conditions = append(status.Conditions, condition)
}

func ReplaceCondition(status *v1beta1.ClusterStatus, condition metav1.Condition) {
	for i, c := range status.Conditions {
		if c.Type == condition.Type {
			status.Conditions[i] = condition
			return
		}
	}
	status.Conditions = append(status.Conditions, condition)
}

func RemoveCondition(status *v1beta1.ClusterStatus, condType string) {
	var newConditions []metav1.Condition
	for _, c := range status.Conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	status.Conditions = newConditions
}
