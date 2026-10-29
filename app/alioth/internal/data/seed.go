package data

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackageitem"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespacefeature"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationpreset"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/permission"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

func seed(ctx context.Context, client *ent.Client, helper *slog.Logger) error {
	for _, item := range perm.Catalog() {
		exists, err := client.Permission.Query().Where(permission.NameEQ(item.Name)).Exist(ctx)
		if err != nil {
			return databaseErr(helper, err)
		}
		if exists {
			continue
		}
		if _, err := client.Permission.Create().
			SetUID(id.New()).
			SetName(item.Name).
			SetResource(item.Resource).
			SetAction(item.Action).
			Save(ctx); err != nil {
			return databaseErr(helper, err)
		}
	}
	if err := ensureBasePackage(ctx, client, helper); err != nil {
		return err
	}
	if err := migrateLegacyNamespaceFeatures(ctx, client, helper); err != nil {
		return err
	}
	if err := syncTenantRoles(ctx, client, helper); err != nil {
		return err
	}
	if err := migrateCustomRoleK8sPermissions(ctx, client, helper); err != nil {
		return err
	}
	if err := ensureSuperAdmin(ctx, client, helper); err != nil {
		return err
	}
	return seedNotificationPresets(ctx, client, helper)
}

// seedNotificationPresets 写入系统预设模板。已存在同名预设时跳过，不覆盖管理员或历史数据。
func seedNotificationPresets(ctx context.Context, client *ent.Client, helper *slog.Logger) error {
	presets := []struct {
		name        string
		kind        string
		app         string
		subject     string
		body        string
		description string
		sort        int
	}{
		{
			name: "邮件告警（默认）", kind: "EMAIL", app: "OTHER",
			subject:     "[{{.phase}}] {{.summary}}",
			body:        "告警标题：{{.summary}}\n告警描述：{{.description}}\n事件标识：{{.eventUid}}\n标签：{{.labels}}",
			description: "适用于 SMTP 渠道的纯文本邮件", sort: 10,
		},
		{
			name: "钉钉告警（文本）", kind: "WEBHOOK", app: "DINGTALK",
			subject:     "{{.summary}}",
			body:        "【{{.phase}}】{{.summary}}\n{{.description}}\n标签：{{.labels}}",
			description: "钉钉自定义机器人，渠道配置签名后自动加签", sort: 20,
		},
		{
			name: "飞书告警（文本）", kind: "WEBHOOK", app: "FEISHU",
			subject:     "{{.summary}}",
			body:        "【{{.phase}}】{{.summary}}\n{{.description}}\n标签：{{.labels}}",
			description: "飞书自定义机器人，渠道配置签名后自动加签", sort: 30,
		},
		{
			name: "企业微信告警（文本）", kind: "WEBHOOK", app: "WECHAT",
			subject:     "{{.summary}}",
			body:        "【{{.phase}}】{{.summary}}\n{{.description}}\n标签：{{.labels}}",
			description: "企业微信群机器人", sort: 40,
		},
		{
			name: "Webhook 通用模板", kind: "WEBHOOK", app: "OTHER",
			subject:     "{{.summary}}",
			body:        "{{.phase}} {{.summary}}\n{{.description}}\n事件：{{.eventUid}}",
			description: "OTHER 类型 Webhook，服务端会以 JSON 发送并附带签名", sort: 50,
		},
		{
			name: "短信告警（默认）", kind: "SMS_ALICLOUD", app: "OTHER",
			subject:     "",
			body:        "【{{.phase}}】{{.summary}}，值 {{.value}}",
			description: "短信渠道使用的基础模板。当前版本只保存配置，发送实现尚未接入", sort: 60,
		},
	}
	for _, item := range presets {
		exists, err := client.NotificationPreset.Query().Where(notificationpreset.NameEQ(item.name)).Exist(ctx)
		if err != nil {
			return databaseErr(helper, err)
		}
		if exists {
			// 预设的 app 是后加字段：老数据按名称补齐，避免企业微信预设仍然只能被通用模板匹配。
			if _, err := client.NotificationPreset.Update().
				Where(notificationpreset.NameEQ(item.name)).
				SetApp(notificationpreset.App(item.app)).
				Save(ctx); err != nil {
				return databaseErr(helper, err)
			}
			continue
		}
		if _, err := client.NotificationPreset.Create().
			SetUID(id.New()).
			SetName(item.name).
			SetType(notificationpreset.Type(item.kind)).
			SetApp(notificationpreset.App(item.app)).
			SetSubject(item.subject).
			SetBody(item.body).
			SetDescription(item.description).
			SetSort(item.sort).
			SetStatus(notificationpreset.StatusACTIVE).
			Save(ctx); err != nil {
			return databaseErr(helper, fmt.Errorf("seed preset %s: %w", item.name, err))
		}
	}
	return nil
}

const (
	basePackageName   = "基础套餐"
	basePackageRemark = "新租户默认功能：告警、告警规则、探测、数据源、通知、成员与角色。"
)

// ensureBasePackage 保证基础套餐存在。已经存在时不做任何修改，管理员改过的
// 名称、备注和功能都保留。多副本同时启动时由 is_base 的部分唯一索引兜底。
func ensureBasePackage(ctx context.Context, client *ent.Client, helper *slog.Logger) error {
	exists, err := client.FeaturePackage.Query().
		Where(featurepackage.IsBase(true), featurepackage.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return databaseErr(helper, err)
	}
	if exists {
		return nil
	}
	tx, err := client.Tx(ctx)
	if err != nil {
		return databaseErr(helper, err)
	}
	defer func() { _ = tx.Rollback() }()
	created, err := tx.FeaturePackage.Create().
		SetUID(id.New()).
		SetName(basePackageName).
		SetRemark(basePackageRemark).
		SetIsBase(true).
		SetStatus(featurepackage.StatusACTIVE).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		// 另一个副本已经写入基础套餐：按已存在处理，不覆盖它的配置。
		exists, existsErr := client.FeaturePackage.Query().
			Where(featurepackage.IsBase(true), featurepackage.DeletedAtIsNil()).
			Exist(ctx)
		if existsErr == nil && exists {
			return nil
		}
		return databaseErr(helper, err)
	}
	for _, code := range perm.BaseFeatures() {
		if _, err := tx.FeaturePackageItem.Create().
			SetUID(id.New()).
			SetPackageUID(created.UID).
			SetFeature(code).
			Save(ctx); err != nil {
			return databaseErr(helper, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return databaseErr(helper, err)
	}
	return nil
}

// migrateLegacyNamespaceFeatures 把升级前创建的租户迁到基础套餐，并加开
// Kubernetes、机器与 SSH，保持升级前的能力不变。package_uid 为空只可能是存量租户：
// 迁移后不会再次命中，可以重复执行。
func migrateLegacyNamespaceFeatures(ctx context.Context, client *ent.Client, helper *slog.Logger) error {
	base, err := client.FeaturePackage.Query().
		Where(featurepackage.IsBase(true), featurepackage.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return databaseErr(helper, err)
	}
	rows, err := client.Namespace.Query().
		Where(namespace.PackageUIDEQ(""), namespace.DeletedAtIsNil()).
		All(ctx)
	if err != nil {
		return databaseErr(helper, err)
	}
	for _, row := range rows {
		if err := migrateLegacyNamespace(ctx, client, base.UID, row.UID); err != nil {
			return databaseErr(helper, err)
		}
	}
	return nil
}

// migrateLegacyNamespace 在同一个事务里换套餐并写加开差异。差异先查后写，
// 允许上次迁移只完成一半后重试。
func migrateLegacyNamespace(ctx context.Context, client *ent.Client, baseUID, namespaceUID string) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	// 先写差异再换套餐：中途崩溃后重试时 package_uid 仍为空，会重新走这段，
	// 已经存在的差异由存在性检查跳过，不会漏掉加开的功能。
	for _, code := range []string{perm.FeatureCluster, perm.FeatureAgent, perm.FeatureSSH} {
		exists, err := tx.NamespaceFeature.Query().
			Where(namespacefeature.NamespaceUIDEQ(namespaceUID), namespacefeature.FeatureEQ(code)).
			Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err := tx.NamespaceFeature.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetFeature(code).
			SetEnabled(true).
			Save(ctx); err != nil {
			return err
		}
	}
	n, err := tx.Namespace.Update().
		Where(namespace.UIDEQ(namespaceUID), namespace.PackageUIDEQ(""), namespace.DeletedAtIsNil()).
		SetPackageUID(baseUID).
		Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		// 另一个迁移者已经换过套餐；它同样先写了差异，这里直接回滚本次写入。
		_ = tx.Rollback()
		return nil
	}
	return tx.Commit()
}

// tenantEffectiveFeatures 计算租户当前的有效功能：套餐功能叠加租户差异。
// 租户被删除或没有套餐时按空功能处理，预置角色只保留 namespace:*。
func tenantEffectiveFeatures(ctx context.Context, client *ent.Client, helper *slog.Logger, namespaceUID string) ([]string, error) {
	row, err := client.Namespace.Query().Where(namespace.UIDEQ(namespaceUID), namespace.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, databaseErr(helper, err)
	}
	return effectiveFeaturesForNamespace(ctx, client, helper, namespaceUID, row.PackageUID)
}

func effectiveFeaturesForNamespace(ctx context.Context, client *ent.Client, helper *slog.Logger, namespaceUID, packageUID string) ([]string, error) {
	packageFeatures := make([]string, 0, len(perm.FeatureCodes()))
	if packageUID != "" {
		rows, err := client.FeaturePackageItem.Query().Where(featurepackageitem.PackageUIDEQ(packageUID)).All(ctx)
		if err != nil {
			return nil, databaseErr(helper, err)
		}
		for _, row := range rows {
			packageFeatures = append(packageFeatures, row.Feature)
		}
	}
	rows, err := client.NamespaceFeature.Query().Where(namespacefeature.NamespaceUIDEQ(namespaceUID)).All(ctx)
	if err != nil {
		return nil, databaseErr(helper, err)
	}
	overrides := make(map[string]bool, len(rows))
	for _, row := range rows {
		overrides[row.Feature] = row.Enabled
	}
	return perm.EffectiveFeatures(packageFeatures, overrides), nil
}

func syncTenantRoles(ctx context.Context, client *ent.Client, helper *slog.Logger) error {
	rows, err := client.Permission.Query().All(ctx)
	if err != nil {
		return databaseErr(helper, err)
	}
	byName := make(map[string]string, len(rows))
	for _, row := range rows {
		byName[row.Name] = row.UID
	}
	featuresByNamespace := make(map[string][]string)
	roles, err := client.Role.Query().Where(role.IsSystem(true), role.DeletedAtIsNil()).All(ctx)
	if err != nil {
		return databaseErr(helper, err)
	}
	for _, item := range roles {
		// 系统角色的预设备注只在空备注时补齐，管理员改过的值不覆盖。
		if item.Remark == "" {
			if remark := perm.DefaultRoleRemark(item.Name); remark != "" {
				if _, err := client.Role.Update().
					Where(role.UIDEQ(item.UID), role.RemarkEQ("")).
					SetRemark(remark).
					Save(ctx); err != nil {
					return databaseErr(helper, err)
				}
			}
		}
		links, err := client.RolePermission.Query().Where(rolepermission.RoleUIDEQ(item.UID)).All(ctx)
		if err != nil {
			return databaseErr(helper, err)
		}
		have := make(map[string]struct{}, len(links))
		for _, link := range links {
			have[link.PermissionUID] = struct{}{}
		}
		features, ok := featuresByNamespace[item.NamespaceUID]
		if !ok {
			features, err = tenantEffectiveFeatures(ctx, client, helper, item.NamespaceUID)
			if err != nil {
				return err
			}
			featuresByNamespace[item.NamespaceUID] = features
		}
		// 只补租户已开通功能内的权限；未开通功能即使预置角色目录里有也不写。
		for _, code := range perm.RolePermissionsForFeatures(item.Name, features) {
			permissionUID := byName[code]
			if permissionUID == "" {
				continue
			}
			if _, ok := have[permissionUID]; ok {
				continue
			}
			if _, err := client.RolePermission.Create().
				SetUID(id.New()).
				SetNamespaceUID(item.NamespaceUID).
				SetRoleUID(item.UID).
				SetPermissionUID(permissionUID).
				Save(ctx); err != nil {
				return databaseErr(helper, err)
			}
		}
	}
	return nil
}

// migrateCustomRoleK8sPermissions 把集群级权限展开成 Kubernetes 资源级权限。
// Kubernetes 控制台改为按资源授权后，旧的自定义角色只有 cluster:read/cluster:write：
// 有 cluster:read 的补全部 k8s 读权限和 Pod 日志，有 cluster:write 的补全部 k8s 写权限。
// 只处理当前没有任何 k8s:* 权限的角色，补齐后不会再命中，可以重复执行。
// 系统角色由 syncTenantRoles 按预置角色目录直接补齐，不走这里。
func migrateCustomRoleK8sPermissions(ctx context.Context, client *ent.Client, helper *slog.Logger) error {
	rows, err := client.Permission.Query().All(ctx)
	if err != nil {
		return databaseErr(helper, err)
	}
	byName := make(map[string]string, len(rows))
	for _, row := range rows {
		byName[row.Name] = row.UID
	}
	roles, err := client.Role.Query().Where(role.IsSystem(false), role.DeletedAtIsNil()).All(ctx)
	if err != nil {
		return databaseErr(helper, err)
	}
	if len(roles) == 0 {
		return nil
	}
	roleUIDs := make([]string, len(roles))
	for index, item := range roles {
		roleUIDs[index] = item.UID
	}
	links, err := client.RolePermission.Query().Where(rolepermission.RoleUIDIn(roleUIDs...)).All(ctx)
	if err != nil {
		return databaseErr(helper, err)
	}
	heldByRole := make(map[string]map[string]struct{}, len(links))
	for _, link := range links {
		names := heldByRole[link.RoleUID]
		if names == nil {
			names = make(map[string]struct{})
			heldByRole[link.RoleUID] = names
		}
		names[byName[link.PermissionUID]] = struct{}{}
	}
	for _, item := range roles {
		for _, code := range legacyRoleK8sPermissions(heldByRole[item.UID]) {
			permissionUID := byName[code]
			if permissionUID == "" {
				continue
			}
			if _, err := client.RolePermission.Create().
				SetUID(id.New()).
				SetNamespaceUID(item.NamespaceUID).
				SetRoleUID(item.UID).
				SetPermissionUID(permissionUID).
				Save(ctx); err != nil {
				return databaseErr(helper, err)
			}
		}
	}
	return nil
}

// legacyRoleK8sPermissions 计算旧的自定义角色需要补齐的 Kubernetes 权限。
// names 是角色当前持有的权限名：已经有任何 k8s:* 权限时返回 nil，补齐后不会重复命中；
// 否则有 cluster:read 补全部 k8s 读权限和 Pod 日志，有 cluster:write 补全部 k8s 写权限。
func legacyRoleK8sPermissions(names map[string]struct{}) []string {
	for name := range names {
		if strings.HasPrefix(name, "k8s:") {
			return nil
		}
	}
	codes := make([]string, 0, len(perm.K8sReadPermissions())+len(perm.K8sWritePermissions()))
	if _, ok := names[perm.ClusterRead]; ok {
		codes = append(codes, perm.K8sReadPermissions()...)
	}
	if _, ok := names[perm.ClusterWrite]; ok {
		codes = append(codes, perm.K8sWritePermissions()...)
	}
	return codes
}

// ensureSuperAdmin 为已有部署补上超级管理员：第一个注册的用户。
// 新库还没有用户时不写入；第一个用户由 CreateUser 在创建时确定。
func ensureSuperAdmin(ctx context.Context, client *ent.Client, helper *slog.Logger) error {
	exists, err := client.User.Query().Where(user.SystemRoleEQ(user.SystemRoleSUPER_ADMIN)).Exist(ctx)
	if err != nil {
		return databaseErr(helper, err)
	}
	if exists {
		return nil
	}
	row, err := client.User.Query().Order(ent.Asc(user.FieldCreatedAt), ent.Asc(user.FieldUID)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return databaseErr(helper, err)
	}
	if _, err := client.User.UpdateOne(row).SetSystemRole(user.SystemRoleSUPER_ADMIN).Save(ctx); err != nil {
		return databaseErr(helper, err)
	}
	return nil
}
