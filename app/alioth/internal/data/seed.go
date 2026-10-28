package data

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
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
	if err := syncTenantRoles(ctx, client, helper); err != nil {
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

func syncTenantRoles(ctx context.Context, client *ent.Client, helper *slog.Logger) error {
	rows, err := client.Permission.Query().All(ctx)
	if err != nil {
		return databaseErr(helper, err)
	}
	byName := make(map[string]string, len(rows))
	for _, row := range rows {
		byName[row.Name] = row.UID
	}
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
		for _, code := range perm.RolePermissions(item.Name) {
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
