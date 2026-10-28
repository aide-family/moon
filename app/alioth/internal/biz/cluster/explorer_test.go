package cluster

import (
	"fmt"
	"testing"

	"github.com/aide-family/moon/pkg/merr"
)

func TestSlicePage(t *testing.T) {
	items := make([]string, 0, 5)
	for index := 1; index <= 5; index++ {
		items = append(items, fmt.Sprintf("item-%d", index))
	}
	cases := []struct {
		name      string
		page      int32
		pageSize  int32
		wantLen   int
		wantFirst string
		wantPage  int
		wantSize  int
	}{
		{name: "first page", page: 1, pageSize: 2, wantLen: 2, wantFirst: "item-1", wantPage: 1, wantSize: 2},
		{name: "middle page", page: 2, pageSize: 2, wantLen: 2, wantFirst: "item-3", wantPage: 2, wantSize: 2},
		{name: "last partial page", page: 3, pageSize: 2, wantLen: 1, wantFirst: "item-5", wantPage: 3, wantSize: 2},
		{name: "out of range", page: 6, pageSize: 2, wantLen: 0, wantPage: 6, wantSize: 2},
		{name: "defaults", page: 0, pageSize: 0, wantLen: 5, wantFirst: "item-1", wantPage: 1, wantSize: 20},
		{name: "page size clamped", page: 1, pageSize: 1000, wantLen: 5, wantFirst: "item-1", wantPage: 1, wantSize: 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			paged, total, page, pageSize := slicePage(items, tc.page, tc.pageSize)
			if total != 5 {
				t.Fatalf("total = %d, want 5", total)
			}
			if len(paged) != tc.wantLen {
				t.Fatalf("len = %d, want %d", len(paged), tc.wantLen)
			}
			if page != tc.wantPage || pageSize != tc.wantSize {
				t.Fatalf("page=%d pageSize=%d, want %d/%d", page, pageSize, tc.wantPage, tc.wantSize)
			}
			if tc.wantFirst != "" && paged[0] != tc.wantFirst {
				t.Fatalf("first = %q, want %q", paged[0], tc.wantFirst)
			}
		})
	}
}

func TestSlicePageEmpty(t *testing.T) {
	paged, total, page, pageSize := slicePage([]string{}, 1, 20)
	if len(paged) != 0 || total != 0 || page != 1 || pageSize != 20 {
		t.Fatalf("empty page = %v total=%d page=%d size=%d", paged, total, page, pageSize)
	}
}

func TestNormalizeYAMLResource(t *testing.T) {
	cases := []struct {
		name          string
		kind          string
		namespace     string
		resourceName  string
		wantKind      string
		wantNamespace string
		wantName      string
		wantInvalid   bool
	}{
		{name: "namespaced resource", kind: " Pod ", namespace: " default ", resourceName: " web-1 ", wantKind: "Pod", wantNamespace: "default", wantName: "web-1"},
		{name: "node ignores namespace", kind: "Node", namespace: "ignored", resourceName: "node-a", wantKind: "Node", wantName: "node-a"},
		{name: "pv ignores namespace", kind: "PersistentVolume", namespace: "ignored", resourceName: "pv-a", wantKind: "PersistentVolume", wantName: "pv-a"},
		{name: "unknown kind", kind: "Job", namespace: "default", resourceName: "backup", wantInvalid: true},
		{name: "empty kind", kind: "  ", namespace: "default", resourceName: "web", wantInvalid: true},
		{name: "empty name", kind: "Pod", namespace: "default", resourceName: "  ", wantInvalid: true},
		{name: "missing namespace", kind: "Pod", namespace: " ", resourceName: "web", wantInvalid: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kind, namespace, resourceName, err := normalizeYAMLResource(tc.kind, tc.namespace, tc.resourceName)
			if tc.wantInvalid {
				if !merr.IsInvalid(err) {
					t.Fatalf("err = %v, want invalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeYAMLResource: %v", err)
			}
			if kind != tc.wantKind || namespace != tc.wantNamespace || resourceName != tc.wantName {
				t.Fatalf("got kind=%q namespace=%q name=%q, want %q/%q/%q", kind, namespace, resourceName, tc.wantKind, tc.wantNamespace, tc.wantName)
			}
		})
	}
}

func TestNormalizeNamespacedTarget(t *testing.T) {
	namespace, name, err := normalizeNamespacedTarget(" default ", " web ")
	if err != nil || namespace != "default" || name != "web" {
		t.Fatalf("got namespace=%q name=%q err=%v", namespace, name, err)
	}
	cases := []struct {
		name      string
		namespace string
		target    string
	}{
		{name: "missing namespace", namespace: "  ", target: "web"},
		{name: "missing name", namespace: "default", target: "  "},
		{name: "both missing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := normalizeNamespacedTarget(tc.namespace, tc.target); !merr.IsInvalid(err) {
				t.Fatalf("err = %v, want invalid", err)
			}
		})
	}
}
