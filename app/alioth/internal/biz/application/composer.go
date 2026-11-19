package application

import (
	"regexp"
	"strings"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// Composer 维护版本内「服务（组件）+ 资源」的归一化视图：分组、排序、服务与资源
// 增删改。它不碰存储，状态机与权限留在 Catalog。
type Composer struct{}

func NewComposer() *Composer {
	return &Composer{}
}

// Compose 按资源自带的 Component 分组；空名归入 default。组件顺序按首次出现，
// 资源顺序沿用输入顺序。版本没有资源时返回空组件列表：服务由用户显式创建，
// 不再自动补 default。
func (composer *Composer) Compose(resources []bo.ApplicationResource) []bo.ApplicationVersionComponent {
	components := make([]bo.ApplicationVersionComponent, 0)
	index := map[string]int{}
	for _, resource := range resources {
		name := normalizeComponentName(resource.Component)
		position, ok := index[name]
		if !ok {
			position = len(components)
			index[name] = position
			components = append(components, bo.ApplicationVersionComponent{
				Name: name, DisplayName: name, SortOrder: int32(position),
				Resources: make([]bo.ApplicationResource, 0),
			})
		}
		resource.Component = name
		components[position].Resources = append(components[position].Resources, resource)
	}
	return normalizeComponentOrders(components)
}

// Flatten 按组件顺序展开资源。
func (composer *Composer) Flatten(components []bo.ApplicationVersionComponent) []bo.ApplicationResource {
	resources := make([]bo.ApplicationResource, 0)
	for _, component := range components {
		resources = append(resources, component.Resources...)
	}
	return resources
}

// FindResource 按 uid 查找版本内的资源。
func (composer *Composer) FindResource(components []bo.ApplicationVersionComponent, resourceUID string) (bo.ApplicationResource, error) {
	for _, component := range components {
		for _, resource := range component.Resources {
			if resource.UID == resourceUID {
				return resource, nil
			}
		}
	}
	return bo.ApplicationResource{}, merr.NotFound("version resource not found")
}

// ValidateForPublish 要求至少一个组件且每个组件至少一个资源。版本内容在草稿
// 阶段允许 0 个服务、0 个资源，只有上架时才校验。
func (composer *Composer) ValidateForPublish(components []bo.ApplicationVersionComponent) error {
	if len(components) == 0 {
		return merr.Invalid("version must have at least one component")
	}
	for _, component := range components {
		if len(component.Resources) == 0 {
			return merr.Invalid("component " + component.Name + " must have at least one resource")
		}
	}
	return nil
}

// CreateResource 把新资源放进目标组件；组件不存在时按名称新建。
func (composer *Composer) CreateResource(components []bo.ApplicationVersionComponent, resource bo.ApplicationResource) ([]bo.ApplicationVersionComponent, error) {
	item := strings.TrimSpace(resource.Item)
	for _, component := range components {
		for _, existing := range component.Resources {
			if existing.Item == item {
				return nil, merr.Conflict("resource item already exists in this version")
			}
		}
	}
	return composer.place(components, resource), nil
}

// SaveResource 用新内容替换指定 uid 的资源，允许迁移到另一个组件。
func (composer *Composer) SaveResource(components []bo.ApplicationVersionComponent, resourceUID string, resource bo.ApplicationResource) ([]bo.ApplicationVersionComponent, error) {
	found := false
	for position := range components {
		for resourcePosition := range components[position].Resources {
			existing := &components[position].Resources[resourcePosition]
			if existing.UID != resourceUID {
				continue
			}
			found = true
			resource.UID = existing.UID
			resource.Item = existing.Item
			components[position].Resources = append(components[position].Resources[:resourcePosition], components[position].Resources[resourcePosition+1:]...)
			break
		}
		if found {
			break
		}
	}
	if !found {
		return nil, merr.NotFound("version resource not found")
	}
	components = dropEmptyComponents(components)
	return composer.place(components, resource), nil
}

// DeleteResource 删除资源；组件变空时删除，版本可以因此变成 0 个服务。
func (composer *Composer) DeleteResource(components []bo.ApplicationVersionComponent, resourceUID string) ([]bo.ApplicationVersionComponent, error) {
	found := false
	for position := range components {
		for resourcePosition := range components[position].Resources {
			if components[position].Resources[resourcePosition].UID != resourceUID {
				continue
			}
			found = true
			components[position].Resources = append(
				components[position].Resources[:resourcePosition],
				components[position].Resources[resourcePosition+1:]...,
			)
			break
		}
		if found {
			break
		}
	}
	if !found {
		return nil, merr.NotFound("version resource not found")
	}
	return normalizeComponentOrders(dropEmptyComponents(components)), nil
}

// SaveComponent 新增或更新一个服务（组件）。componentUID 为空时按末尾顺序新建；
// 非空时改名字与备注。改名会同步该服务下所有资源的 Component，避免资源仍指向
// 旧名。名称必须是小写 DNS-1123 label，且在同一版本内唯一。
func (composer *Composer) SaveComponent(components []bo.ApplicationVersionComponent, componentUID, name, description string) ([]bo.ApplicationVersionComponent, error) {
	name = strings.TrimSpace(name)
	if err := validateComponentName(name); err != nil {
		return nil, err
	}
	description = strings.TrimSpace(description)
	if len([]rune(description)) > maxComponentDesc {
		return nil, merr.Invalid("component description must not exceed 2048 characters")
	}
	position := -1
	for index := range components {
		sameComponent := componentUID != "" && components[index].UID == componentUID
		if components[index].Name == name && !sameComponent {
			return nil, merr.Conflict("component name already exists in this version")
		}
		if sameComponent {
			position = index
		}
	}
	if componentUID == "" {
		components = append(components, bo.ApplicationVersionComponent{
			Name: name, DisplayName: name, Description: description,
			Resources: make([]bo.ApplicationResource, 0),
		})
		return normalizeComponentOrders(components), nil
	}
	if position < 0 {
		return nil, merr.NotFound("version component not found")
	}
	components[position].Name = name
	components[position].DisplayName = name
	components[position].Description = description
	for resourcePosition := range components[position].Resources {
		components[position].Resources[resourcePosition].Component = name
	}
	return normalizeComponentOrders(components), nil
}

// FindComponent 按 uid 查找版本内的服务。
func (composer *Composer) FindComponent(components []bo.ApplicationVersionComponent, componentUID string) (bo.ApplicationVersionComponent, error) {
	for _, component := range components {
		if componentUID != "" && component.UID == componentUID {
			return component, nil
		}
	}
	return bo.ApplicationVersionComponent{}, merr.NotFound("version component not found")
}

// DeleteComponent 删除一个服务及其下的全部资源；版本可以因此变成 0 个服务。
func (composer *Composer) DeleteComponent(components []bo.ApplicationVersionComponent, componentUID string) ([]bo.ApplicationVersionComponent, error) {
	kept := make([]bo.ApplicationVersionComponent, 0, len(components))
	found := false
	for _, component := range components {
		if componentUID != "" && component.UID == componentUID {
			found = true
			continue
		}
		kept = append(kept, component)
	}
	if !found {
		return nil, merr.NotFound("version component not found")
	}
	return normalizeComponentOrders(kept), nil
}

// AllowedActions 按状态给出可执行动作；RELEASE 只表示版本可发布部署，具体集群
// 参数由前端按部署组选择。
func (composer *Composer) AllowedActions(status string) []string {
	actions := make([]string, 0, 5)
	switch status {
	case "DRAFT":
		actions = append(actions, "PUBLISH", "COPY", "DELETE")
	case "PUBLISHED":
		actions = append(actions, "DEPRECATE", "COPY", "RELEASE")
	case "DEPRECATED":
		actions = append(actions, "COPY", "DELETE")
	}
	return actions
}

// place 把资源放到目标组件；组件不存在时按名称新建。
func (composer *Composer) place(components []bo.ApplicationVersionComponent, resource bo.ApplicationResource) []bo.ApplicationVersionComponent {
	name := normalizeComponentName(resource.Component)
	resource.Component = name
	for position := range components {
		if components[position].Name != name {
			continue
		}
		components[position].Resources = append(components[position].Resources, resource)
		return normalizeComponentOrders(components)
	}
	components = append(components, bo.ApplicationVersionComponent{
		Name: name, DisplayName: name, SortOrder: int32(len(components)),
		Resources: []bo.ApplicationResource{resource},
	})
	return normalizeComponentOrders(components)
}

// dropEmptyComponents 删除没有资源的组件。版本允许 0 个服务，不再补齐 default。
func dropEmptyComponents(components []bo.ApplicationVersionComponent) []bo.ApplicationVersionComponent {
	kept := make([]bo.ApplicationVersionComponent, 0, len(components))
	for _, component := range components {
		if len(component.Resources) == 0 {
			continue
		}
		kept = append(kept, component)
	}
	return kept
}

func normalizeComponentOrders(components []bo.ApplicationVersionComponent) []bo.ApplicationVersionComponent {
	for position := range components {
		components[position].SortOrder = int32(position)
		components[position].ResourceCount = int32(len(components[position].Resources))
		if components[position].DisplayName == "" {
			components[position].DisplayName = components[position].Name
		}
	}
	return components
}

func normalizeComponentName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return defaultComponentName
	}
	return name
}

// componentNamePattern 是小写 DNS-1123 label 的形状，服务名用它做资源分组名。
var componentNamePattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// validateComponentName 要求服务名是小写 DNS-1123 label，长度不超过 63。
func validateComponentName(name string) error {
	if name == "" || len(name) > maxComponentRunes {
		return merr.Invalid("component name must be a lowercase DNS-1123 label and not exceed 63 characters")
	}
	if !componentNamePattern.MatchString(name) {
		return merr.Invalid("component name must be a lowercase DNS-1123 label")
	}
	return nil
}

// defaultComponentName 是没有 component 输入时的兜底服务名。
const defaultComponentName = "default"
