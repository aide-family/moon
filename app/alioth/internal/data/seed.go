package data

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationpreset"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/permission"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/role"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
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
	if err := syncSystemRoles(ctx, client, helper); err != nil {
		return err
	}
	return seedNotificationPresets(ctx, client, helper)
}

// seedNotificationPresets 写入系统预设模板。已存在同名预设时跳过，不覆盖管理员或历史数据。
func seedNotificationPresets(ctx context.Context, client *ent.Client, helper *slog.Logger) error {
	presets := []struct {
		name        string
		kind        string
		subject     string
		body        string
		description string
		sort        int
	}{
		{
			name: "邮件告警（默认）", kind: "EMAIL",
			subject:     "[{{.phase}}] {{.summary}}",
			body:        "告警标题：{{.summary}}\n告警描述：{{.description}}\n事件标识：{{.eventUid}}\n标签：{{.labels}}",
			description: "适用于 SMTP 渠道的纯文本邮件", sort: 10,
		},
		{
			name: "钉钉告警（文本）", kind: "WEBHOOK",
			subject:     "{{.summary}}",
			body:        "【{{.phase}}】{{.summary}}\n{{.description}}\n标签：{{.labels}}",
			description: "钉钉自定义机器人，渠道配置签名后自动加签", sort: 20,
		},
		{
			name: "飞书告警（文本）", kind: "WEBHOOK",
			subject:     "{{.summary}}",
			body:        "【{{.phase}}】{{.summary}}\n{{.description}}\n标签：{{.labels}}",
			description: "飞书自定义机器人，渠道配置签名后自动加签", sort: 30,
		},
		{
			name: "企业微信告警（文本）", kind: "WEBHOOK",
			subject:     "{{.summary}}",
			body:        "【{{.phase}}】{{.summary}}\n{{.description}}\n标签：{{.labels}}",
			description: "企业微信群机器人", sort: 40,
		},
		{
			name: "Webhook 通用模板", kind: "WEBHOOK",
			subject:     "{{.summary}}",
			body:        "{{.phase}} {{.summary}}\n{{.description}}\n事件：{{.eventUid}}",
			description: "OTHER 类型 Webhook，服务端会以 JSON 发送并附带签名", sort: 50,
		},
	}
	for _, item := range presets {
		exists, err := client.NotificationPreset.Query().Where(notificationpreset.NameEQ(item.name)).Exist(ctx)
		if err != nil {
			return databaseErr(helper, err)
		}
		if exists {
			continue
		}
		if _, err := client.NotificationPreset.Create().
			SetUID(id.New()).
			SetName(item.name).
			SetType(notificationpreset.Type(item.kind)).
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

func syncSystemRoles(ctx context.Context, client *ent.Client, helper *slog.Logger) error {
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
