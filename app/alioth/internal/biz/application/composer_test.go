package application

import (
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

func TestComposeGroupsResourcesByComponent(t *testing.T) {
	composer := NewComposer()
	components := composer.Compose([]bo.ApplicationResource{
		{UID: "r1", Item: "api", Component: "gateway"},
		{UID: "r2", Item: "db", Component: ""},
		{UID: "r3", Item: "worker", Component: "gateway"},
	})
	if len(components) != 2 {
		t.Fatalf("components = %d", len(components))
	}
	if components[0].Name != "gateway" || len(components[0].Resources) != 2 {
		t.Fatalf("first component = %+v", components[0])
	}
	if components[1].Name != defaultComponentName || len(components[1].Resources) != 1 {
		t.Fatalf("default component = %+v", components[1])
	}
	if components[0].Resources[1].Component != "gateway" {
		t.Fatalf("resource component not normalized: %+v", components[0].Resources[1])
	}
	// 没有资源的版本是 0 个服务，不再自动补 default。
	empty := composer.Compose(nil)
	if len(empty) != 0 {
		t.Fatalf("empty version components = %+v", empty)
	}
}

func TestComposerCreateSaveDeleteResource(t *testing.T) {
	composer := NewComposer()
	components := composer.Compose([]bo.ApplicationResource{
		{UID: "r1", Item: "api", Component: "gateway"},
	})
	components, err := composer.CreateResource(components, bo.ApplicationResource{UID: "r2", Item: "worker", Component: "worker"})
	if err != nil {
		t.Fatal(err)
	}
	if len(components) != 2 || len(components[1].Resources) != 1 {
		t.Fatalf("components = %+v", components)
	}
	if _, err := composer.CreateResource(components, bo.ApplicationResource{Item: "api", Component: "worker"}); !merr.IsConflict(err) {
		t.Fatalf("duplicate item error = %v", err)
	}
	// 迁移资源到新组件，原组件变空后删除。
	components, err = composer.SaveResource(components, "r2", bo.ApplicationResource{Item: "worker", Component: "queue"})
	if err != nil {
		t.Fatal(err)
	}
	if len(components) != 2 {
		t.Fatalf("components after save = %+v", components)
	}
	for _, component := range components {
		if component.Name == "worker" {
			t.Fatalf("empty component was not removed: %+v", components)
		}
	}
	components, err = composer.DeleteResource(components, "r1")
	if err != nil {
		t.Fatal(err)
	}
	if len(components) != 1 || components[0].Name != "queue" {
		t.Fatalf("components after delete = %+v", components)
	}
	// 删除最后一个资源后组件也删除，版本允许 0 个服务。
	components, err = composer.DeleteResource(components, "r2")
	if err != nil {
		t.Fatal(err)
	}
	if len(components) != 0 {
		t.Fatalf("last component must be removed: %+v", components)
	}
	if _, err := composer.DeleteResource(components, "missing"); !merr.IsNotFound(err) {
		t.Fatalf("missing resource error = %v", err)
	}
}

func TestComposerSaveAndDeleteComponent(t *testing.T) {
	composer := NewComposer()
	components := composer.Compose([]bo.ApplicationResource{
		{UID: "r1", Item: "api", Component: "gateway"},
	})
	components[0].UID = "c1"

	// 新建服务排在末尾，可以暂时没有资源。
	components, err := composer.SaveComponent(components, "", "worker", "worker 服务")
	if err != nil {
		t.Fatal(err)
	}
	if len(components) != 2 || components[1].Name != "worker" || components[1].SortOrder != 1 {
		t.Fatalf("created component = %+v", components)
	}
	components[1].UID = "c2"
	if components[1].Description != "worker 服务" || len(components[1].Resources) != 0 {
		t.Fatalf("created component detail = %+v", components[1])
	}
	if _, err := composer.SaveComponent(components, "", "worker", ""); !merr.IsConflict(err) {
		t.Fatalf("duplicate component name error = %v", err)
	}
	if _, err := composer.SaveComponent(components, "", "Worker_1", ""); !merr.IsInvalid(err) {
		t.Fatalf("invalid component name error = %v", err)
	}

	// 改名时把服务下所有资源的 component 一起改成新名。
	components, err = composer.SaveComponent(components, "c1", "edge", "入口")
	if err != nil {
		t.Fatal(err)
	}
	if components[0].Name != "edge" || components[0].Description != "入口" {
		t.Fatalf("renamed component = %+v", components[0])
	}
	if components[0].Resources[0].Component != "edge" {
		t.Fatalf("resource component not renamed: %+v", components[0].Resources[0])
	}
	// 资源数量不变，排序按位置重排。
	if components[0].ResourceCount != 1 || components[1].SortOrder != 1 {
		t.Fatalf("components after rename = %+v", components)
	}
	if _, err := composer.SaveComponent(components, "c1", "worker", ""); !merr.IsConflict(err) {
		t.Fatalf("rename to existing name error = %v", err)
	}

	// 删除服务连同其资源；删到 0 个服务也合法。
	components, err = composer.DeleteComponent(components, "c1")
	if err != nil {
		t.Fatal(err)
	}
	if len(components) != 1 || components[0].Name != "worker" {
		t.Fatalf("components after delete = %+v", components)
	}
	components, err = composer.DeleteComponent(components, "c2")
	if err != nil {
		t.Fatal(err)
	}
	if len(components) != 0 {
		t.Fatalf("components after delete last = %+v", components)
	}
	if _, err := composer.DeleteComponent(components, "missing"); !merr.IsNotFound(err) {
		t.Fatalf("missing component error = %v", err)
	}
}

func TestComposerValidateAndAllowedActions(t *testing.T) {
	composer := NewComposer()
	if err := composer.ValidateForPublish(nil); err == nil {
		t.Fatal("empty composition should not publish")
	}
	if err := composer.ValidateForPublish(composer.Compose(nil)); err == nil {
		t.Fatal("version without components should not publish")
	}
	emptyComponent := []bo.ApplicationVersionComponent{{Name: "worker"}}
	if err := composer.ValidateForPublish(emptyComponent); err == nil {
		t.Fatal("component without resources should not publish")
	}
	valid := composer.Compose([]bo.ApplicationResource{{Item: "api"}})
	if err := composer.ValidateForPublish(valid); err != nil {
		t.Fatal(err)
	}
	cases := map[string][]string{
		"DRAFT":      {"PUBLISH", "COPY", "DELETE"},
		"PUBLISHED":  {"DEPRECATE", "COPY", "RELEASE"},
		"DEPRECATED": {"COPY", "DELETE"},
	}
	for status, want := range cases {
		got := composer.AllowedActions(status)
		if len(got) != len(want) {
			t.Fatalf("%s actions = %v", status, got)
		}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("%s actions = %v", status, got)
			}
		}
	}
}
