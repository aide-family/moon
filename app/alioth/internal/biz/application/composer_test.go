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
	empty := composer.Compose(nil)
	if len(empty) != 1 || empty[0].Name != defaultComponentName {
		t.Fatalf("empty version component = %+v", empty)
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
	// 删除最后一个资源也必须保留组件。
	components, err = composer.DeleteResource(components, "r2")
	if err != nil {
		t.Fatal(err)
	}
	if len(components) != 1 || len(components[0].Resources) != 0 {
		t.Fatalf("last component must be kept: %+v", components)
	}
	if _, err := composer.DeleteResource(components, "missing"); !merr.IsNotFound(err) {
		t.Fatalf("missing resource error = %v", err)
	}
}

func TestComposerValidateAndAllowedActions(t *testing.T) {
	composer := NewComposer()
	if err := composer.ValidateForPublish(nil); err == nil {
		t.Fatal("empty composition should not publish")
	}
	if err := composer.ValidateForPublish(composer.Compose(nil)); err == nil {
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
