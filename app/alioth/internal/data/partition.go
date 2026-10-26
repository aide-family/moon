package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	entschema "entgo.io/ent/dialect/sql/schema"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/permission"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

var partitionedTables = []string{"alert_events", "message_logs"}

// ApplySchema 创建表，并把空的告警事件和消息日志转成按 created_at 的周分区表。
// 已经分区时跳过这些表，避免生成器把主键改回单列 id。
func ApplySchema(ctx context.Context, client *ent.Client, db *sql.DB) error {
	if err := renameAlertDictionaries(ctx, db); err != nil {
		return err
	}
	if err := prepareAlertRuleDriverOnly(ctx, db); err != nil {
		return err
	}
	if err := prepareAlertEvalBatches(ctx, db); err != nil {
		return err
	}
	if err := prepareAlertEventIntervene(ctx, db); err != nil {
		return err
	}
	if err := prepareAlertEventSuppression(ctx, db); err != nil {
		return err
	}
	if err := prepareNotificationTaskChannels(ctx, db); err != nil {
		return err
	}
	if err := dropRuleNotificationGroups(ctx, db); err != nil {
		return err
	}
	if err := dropLegacyEdgeAudits(ctx, db); err != nil {
		return err
	}
	disabledSources, err := prepareDatasourceConnections(ctx, db)
	if err != nil {
		return err
	}
	if err := prepareNotificationColumns(ctx, db); err != nil {
		return err
	}
	opts := []entschema.MigrateOption{}
	for _, name := range partitionedTables {
		kind, err := relationKind(ctx, db, name)
		if err != nil {
			return err
		}
		if kind == "p" {
			opts = append(opts, entschema.WithHooks(skipTable(name)))
		}
	}
	if err := client.Schema.Create(ctx, opts...); err != nil {
		return err
	}
	if err := dropStalePermissions(ctx, client); err != nil {
		return err
	}
	if err := backfillNotificationGroupTargets(ctx, client); err != nil {
		return databaseErr(slog.Default(), err)
	}
	if err := ensureNotificationAuditFields(ctx, db); err != nil {
		return err
	}
	if err := recoverDisabledDatasourceFiring(ctx, client, disabledSources); err != nil {
		return databaseErr(slog.Default(), err)
	}
	for _, name := range partitionedTables {
		if err := partitionByCreatedAt(ctx, db, name); err != nil {
			return err
		}
		if err := ensureWeekPartitions(ctx, db, name, time.Now().UTC()); err != nil {
			return err
		}
		if err := ensurePartitionIndexes(ctx, db, name); err != nil {
			return err
		}
	}
	return nil
}

// prepareNotificationColumns 先给已经存在的表补上新增列。
// ent 在已有数据上直接 ADD COLUMN NOT NULL 不会自动回填默认值，会报 23502；
// 这里用 IF NOT EXISTS + DEFAULT 预建列，新库由 ent 自己建表时跳过。
func prepareNotificationColumns(ctx context.Context, db *sql.DB) error {
	columns := map[string][]string{
		"notification_groups": {
			`ALTER TABLE notification_groups ADD COLUMN IF NOT EXISTS channel_uids jsonb NOT NULL DEFAULT '[]'::jsonb`,
			`ALTER TABLE notification_groups ADD COLUMN IF NOT EXISTS template_uids jsonb NOT NULL DEFAULT '[]'::jsonb`,
			`ALTER TABLE notification_groups ADD COLUMN IF NOT EXISTS email_receivers character varying NOT NULL DEFAULT ''`,
			`ALTER TABLE notification_groups ADD COLUMN IF NOT EXISTS sms_receivers character varying NOT NULL DEFAULT ''`,
			`ALTER TABLE notification_groups ADD COLUMN IF NOT EXISTS email_member_uids jsonb NOT NULL DEFAULT '[]'::jsonb`,
			`ALTER TABLE notification_groups ADD COLUMN IF NOT EXISTS sms_member_uids jsonb NOT NULL DEFAULT '[]'::jsonb`,
		},
		"notification_subscriptions": {
			`ALTER TABLE notification_subscriptions ADD COLUMN IF NOT EXISTS datasource_level_uids jsonb NOT NULL DEFAULT '[]'::jsonb`,
			`ALTER TABLE notification_subscriptions ADD COLUMN IF NOT EXISTS repeat_interval_seconds integer NOT NULL DEFAULT 0`,
		},
		"notification_on_calls": {
			`ALTER TABLE notification_on_calls ADD COLUMN IF NOT EXISTS datasource_level_uids jsonb NOT NULL DEFAULT '[]'::jsonb`,
			`ALTER TABLE notification_on_calls ADD COLUMN IF NOT EXISTS repeat_interval_seconds integer NOT NULL DEFAULT 0`,
		},
		"notification_tasks": {
			`ALTER TABLE notification_tasks ADD COLUMN IF NOT EXISTS fingerprint character varying NOT NULL DEFAULT ''`,
		},
		"notification_templates": {
			`ALTER TABLE notification_templates ADD COLUMN IF NOT EXISTS app character varying NOT NULL DEFAULT 'OTHER'`,
		},
		"notification_presets": {
			`ALTER TABLE notification_presets ADD COLUMN IF NOT EXISTS app character varying NOT NULL DEFAULT 'OTHER'`,
		},
		"users": {
			`ALTER TABLE users ADD COLUMN IF NOT EXISTS phone character varying NOT NULL DEFAULT ''`,
		},
		"email_login_codes": {
			`ALTER TABLE email_login_codes ADD COLUMN IF NOT EXISTS purpose character varying NOT NULL DEFAULT ''`,
			`ALTER TABLE email_login_codes ADD COLUMN IF NOT EXISTS attempts bigint NOT NULL DEFAULT 0`,
			`ALTER TABLE email_login_codes ADD COLUMN IF NOT EXISTS bind_token_hash character varying NOT NULL DEFAULT ''`,
		},
	}
	for table, statements := range columns {
		var exists bool
		if err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, table).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			continue
		}
		for _, statement := range statements {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				return err
			}
		}
	}
	return nil
}

// backfillNotificationGroupTargets 把旧版单渠道通知组展开成新的
// channel_uids/template_uids 数组，升级后旧通知组仍可编辑和投递。
func backfillNotificationGroupTargets(ctx context.Context, client *ent.Client) error {
	groups, err := client.NotificationGroup.Query().All(ctx)
	if err != nil {
		return err
	}
	for _, group := range groups {
		if len(group.ChannelUids) > 0 || strings.TrimSpace(group.ChannelUID) == "" {
			continue
		}
		channelUIDs := []string{group.ChannelUID}
		templateUIDs := []string{group.TemplateUID}
		for _, webhookUID := range group.WebhookChannelUids {
			channelUIDs = append(channelUIDs, webhookUID)
			templateUIDs = append(templateUIDs, group.WebhookTemplateUID)
		}
		if _, err := group.Update().
			SetChannelUids(channelUIDs).
			SetTemplateUids(templateUIDs).
			SetEmailReceivers(group.Receivers).
			SetEmailMemberUids(group.MemberUids).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

// partitionIndexes 是分区父表必须保留的查询索引。
// CREATE TABLE ... (LIKE ...) 不会复制索引，转换后必须显式重建；这里用 IF NOT EXISTS 保证重复迁移安全。
var partitionIndexes = map[string][]string{
	"alert_events": {
		`CREATE INDEX IF NOT EXISTS alert_events_ns_status_created ON alert_events (namespace_uid, status, created_at)`,
		`CREATE INDEX IF NOT EXISTS alert_events_rule_uid ON alert_events (rule_uid)`,
		`CREATE INDEX IF NOT EXISTS alert_events_fingerprint ON alert_events (fingerprint)`,
	},
	"message_logs": {
		`CREATE INDEX IF NOT EXISTS message_logs_ns_created ON message_logs (namespace_uid, created_at)`,
		`CREATE INDEX IF NOT EXISTS message_logs_task_uid ON message_logs (task_uid)`,
	},
}

// ensureNotificationAuditFields 为已有表补充审计字段。message_logs 已分区时被 Ent 自动迁移跳过，
// 因此这里显式执行幂等 DDL；空字符串默认值保证旧数据和滚动升级期间的旧写入兼容。
func ensureNotificationAuditFields(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`ALTER TABLE notification_tasks ADD COLUMN IF NOT EXISTS sender VARCHAR NOT NULL DEFAULT ''`,
		`ALTER TABLE notification_tasks ADD COLUMN IF NOT EXISTS content VARCHAR NOT NULL DEFAULT ''`,
		`ALTER TABLE message_logs ADD COLUMN IF NOT EXISTS sender VARCHAR NOT NULL DEFAULT ''`,
		`ALTER TABLE message_logs ADD COLUMN IF NOT EXISTS subject VARCHAR NOT NULL DEFAULT ''`,
		`ALTER TABLE message_logs ADD COLUMN IF NOT EXISTS content VARCHAR NOT NULL DEFAULT ''`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			slog.Default().Error(fmt.Sprintf("notification audit schema: %v", err))
			return merr.Server("request failed")
		}
	}
	return nil
}

func ensurePartitionIndexes(ctx context.Context, db *sql.DB, name string) error {
	kind, err := relationKind(ctx, db, name)
	if err != nil {
		return err
	}
	if kind != "p" {
		return nil
	}
	for _, statement := range partitionIndexes[name] {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			slog.Default().Error(fmt.Sprintf("index %s: %v", name, err))
			return merr.Server("request failed")
		}
	}
	return nil
}

// renameAlertDictionaries 把旧的 alert_levels 表和索引改成告警字典的名字。
// ent 的 Schema.Create 不会改已有表名，不改的话会另建一张空表。
func renameAlertDictionaries(ctx context.Context, db *sql.DB) error {
	kind, err := relationKind(ctx, db, "alert_levels")
	if err != nil || kind != "r" {
		return err
	}
	statements := []string{
		`ALTER TABLE alert_levels RENAME TO alert_dictionaries`,
		`ALTER INDEX IF EXISTS alertlevel_uid RENAME TO alertdictionary_uid`,
		`ALTER INDEX IF EXISTS alertlevel_namespace_uid_name RENAME TO alertdictionary_namespace_uid_name`,
		`ALTER INDEX IF EXISTS alertlevel_namespace_uid RENAME TO alertdictionary_namespace_uid`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			slog.Default().Error(fmt.Sprintf("schema: %v", err))
			return merr.Server("request failed")
		}
	}
	return nil
}

// prepareAlertRuleDriverOnly 删除规则表的 type 列。规则只保存驱动，类型由驱动推导；
// 旧部署升级后列仍在，会让「规则不再有类型」的模型与库结构不一致，这里幂等删除。
func prepareAlertRuleDriverOnly(ctx context.Context, db *sql.DB) error {
	kind, err := relationKind(ctx, db, "alert_rules")
	if err != nil || kind != "r" {
		return err
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE alert_rules DROP COLUMN IF EXISTS type`); err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return merr.Server("request failed")
	}
	return nil
}

// prepareAlertEvalBatches 给旧表补上 datasource_uid。
// 这列是后加的非空字段，表里已有行时直接 ALTER 会因空值失败。
// 旧的 (rule_uid, eval_slot) 唯一索引也要去掉，改由 (rule_uid, datasource_uid, eval_slot) 约束。
func prepareAlertEvalBatches(ctx context.Context, db *sql.DB) error {
	kind, err := relationKind(ctx, db, "alert_eval_batches")
	if err != nil || kind != "r" {
		return err
	}
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'alert_eval_batches' AND column_name = 'datasource_uid'
	)`).Scan(&exists); err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return merr.Server("request failed")
	}
	if !exists {
		if _, err := db.ExecContext(ctx, `ALTER TABLE alert_eval_batches ADD COLUMN datasource_uid character varying NOT NULL DEFAULT ''`); err != nil {
			slog.Default().Error(fmt.Sprintf("schema: %v", err))
			return merr.Server("request failed")
		}
		if _, err := db.ExecContext(ctx, `ALTER TABLE alert_eval_batches ALTER COLUMN datasource_uid DROP DEFAULT`); err != nil {
			slog.Default().Error(fmt.Sprintf("schema: %v", err))
			return merr.Server("request failed")
		}
	}
	if _, err := db.ExecContext(ctx, `DROP INDEX IF EXISTS alertevalbatch_rule_uid_eval_slot`); err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return merr.Server("request failed")
	}
	return nil
}

// prepareAlertEventIntervene 给已分区的告警事件补上介入时间。
// 分区表会被迁移跳过，父表加列会落到各个分区。
func prepareAlertEventIntervene(ctx context.Context, db *sql.DB) error {
	kind, err := relationKind(ctx, db, "alert_events")
	if err != nil || kind == "" {
		return err
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE alert_events ADD COLUMN IF NOT EXISTS intervened_at timestamptz NULL`); err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return merr.Server("request failed")
	}
	return nil
}

// prepareAlertEventSuppression 给已分区的告警事件补上抑制到期时间。
// 分区表会被 Ent 自动迁移跳过，显式 DDL 会作用到父表和所有分区。
func prepareAlertEventSuppression(ctx context.Context, db *sql.DB) error {
	kind, err := relationKind(ctx, db, "alert_events")
	if err != nil || kind == "" {
		return err
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE alert_events ADD COLUMN IF NOT EXISTS suppressed_until timestamptz NULL`); err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return merr.Server("request failed")
	}
	return nil
}

func skipTable(name string) entschema.Hook {
	return func(next entschema.Creator) entschema.Creator {
		return entschema.CreateFunc(func(ctx context.Context, tables ...*entschema.Table) error {
			kept := make([]*entschema.Table, 0, len(tables))
			for _, table := range tables {
				if table.Name == name {
					continue
				}
				kept = append(kept, table)
			}
			return next.Create(ctx, kept...)
		})
	}
}

func partitionByCreatedAt(ctx context.Context, db *sql.DB, name string) error {
	kind, err := relationKind(ctx, db, name)
	if err != nil {
		return err
	}
	if kind == "p" || kind == "" {
		return nil
	}
	if kind != "r" {
		return merr.Server(name + " cannot be partitioned")
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM `+name).Scan(&count); err != nil {
		slog.Default().Error(fmt.Sprintf("count %s: %v", name, err))
		return merr.Server("request failed")
	}
	if count > 0 {
		return merr.Server(name + " is not empty")
	}
	heap := name + "_heap"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return merr.Server("request failed")
	}
	defer func() { _ = tx.Rollback() }()
	// 旧表改名后，约束、索引和 identity 序列的名字不会跟着变。
	// 新分区表要复用 table_pkey / table_id_seq 这些默认名，所以先把旧名字释放掉。
	statements := []string{
		fmt.Sprintf(`ALTER TABLE %s RENAME TO %s`, name, heap),
		fmt.Sprintf(`ALTER SEQUENCE IF EXISTS %s_id_seq RENAME TO %s_id_seq`, name, heap),
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			slog.Default().Error(fmt.Sprintf("partition %s: %v", name, err))
			return merr.Server("request failed")
		}
	}
	if err := dropPrimaryKey(ctx, tx, heap); err != nil {
		return err
	}
	if err := dropIndexes(ctx, tx, heap); err != nil {
		return err
	}
	create := []string{
		fmt.Sprintf(`CREATE TABLE %s (LIKE %s INCLUDING DEFAULTS INCLUDING IDENTITY) PARTITION BY RANGE (created_at)`, name, heap),
		fmt.Sprintf(`ALTER TABLE %s ADD PRIMARY KEY (created_at, id)`, name),
		fmt.Sprintf(`CREATE UNIQUE INDEX %s_uid_created_at ON %s (uid, created_at)`, name, name),
	}
	for _, statement := range create {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			slog.Default().Error(fmt.Sprintf("partition %s: %v", name, err))
			return merr.Server("request failed")
		}
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`DROP TABLE %s`, heap)); err != nil {
		slog.Default().Error(fmt.Sprintf("drop %s heap: %v", name, err))
		return merr.Server("request failed")
	}
	return tx.Commit()
}

// dropPrimaryKey 删除临时表的主键约束，释放默认的 <table>_pkey 索引名。
func dropPrimaryKey(ctx context.Context, tx *sql.Tx, table string) error {
	var name string
	err := tx.QueryRowContext(ctx, `SELECT conname FROM pg_constraint WHERE conrelid = $1::regclass AND contype = 'p'`, table).Scan(&name)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		slog.Default().Error(fmt.Sprintf("primary key %s: %v", table, err))
		return merr.Server("request failed")
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s DROP CONSTRAINT %q`, table, name)); err != nil {
		slog.Default().Error(fmt.Sprintf("drop primary key %s: %v", table, err))
		return merr.Server("request failed")
	}
	return nil
}

// dropIndexes 删除临时表上的普通索引。主键约束已经先删掉，剩余索引随 DROP TABLE 丢弃，这里只是释放索引名。
func dropIndexes(ctx context.Context, tx *sql.Tx, table string) error {
	rows, err := tx.QueryContext(ctx, `SELECT indexname FROM pg_indexes WHERE schemaname = 'public' AND tablename = $1`, table)
	if err != nil {
		slog.Default().Error(fmt.Sprintf("indexes %s: %v", table, err))
		return merr.Server("request failed")
	}
	var names []string
	for rows.Next() {
		var index string
		if err := rows.Scan(&index); err != nil {
			_ = rows.Close()
			return merr.Server("request failed")
		}
		names = append(names, index)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return merr.Server("request failed")
	}
	_ = rows.Close()
	for _, index := range names {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`DROP INDEX IF EXISTS %q`, index)); err != nil {
			slog.Default().Error(fmt.Sprintf("drop index %s: %v", index, err))
			return merr.Server("request failed")
		}
	}
	return nil
}

// partitionLeadWeeks 是维护任务预创建的未来周数。
const partitionLeadWeeks = 4

func ensureWeekPartitions(ctx context.Context, db *sql.DB, name string, now time.Time) error {
	return ensureUpcomingPartitions(ctx, db, name, now, 2)
}

func ensureUpcomingPartitions(ctx context.Context, db *sql.DB, name string, now time.Time, weeks int) error {
	kind, err := relationKind(ctx, db, name)
	if err != nil {
		return err
	}
	if kind != "p" {
		return nil
	}
	if weeks < 1 {
		weeks = 1
	}
	start, _ := weekStart(now)
	for i := 0; i < weeks; i++ {
		at := start.AddDate(0, 0, 7*i)
		from, until := weekStart(at)
		part := fmt.Sprintf("%s_y%04dw%02d", name, from.Year(), isoWeek(from))
		statement := fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')`,
			part, name, from.Format("2006-01-02"), until.Format("2006-01-02"),
		)
		if _, err := db.ExecContext(ctx, statement); err != nil {
			slog.Default().Error(fmt.Sprintf("ensure %s week: %v", name, err))
			return merr.Server("request failed")
		}
	}
	return nil
}

func ensureAlertEventWeeks(ctx context.Context, db *sql.DB, now time.Time) error {
	return ensureWeekPartitions(ctx, db, "alert_events", now)
}

func ensureMessageLogWeeks(ctx context.Context, db *sql.DB, now time.Time) error {
	return ensureWeekPartitions(ctx, db, "message_logs", now)
}

func relationKind(ctx context.Context, db *sql.DB, name string) (string, error) {
	var kind sql.NullString
	err := db.QueryRowContext(ctx, `SELECT c.relkind::text FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'public' AND c.relname = $1`, name).Scan(&kind)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return "", merr.Server("request failed")
	}
	return kind.String, nil
}

func weekStart(at time.Time) (time.Time, time.Time) {
	at = at.UTC()
	weekday := int(at.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(weekday - 1))
	return start, start.AddDate(0, 0, 7)
}

func isoWeek(start time.Time) int {
	_, week := start.ISOWeek()
	return week
}

// prepareNotificationTaskChannels 把任务去重索引从 (event, phase, group) 改成
// (event, phase, group, channel)。旧索引不删会挡住一个通知组投递多个 Webhook。
func prepareNotificationTaskChannels(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `DROP INDEX IF EXISTS notificationtask_event_uid_phase_group_uid`); err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return merr.Server("request failed")
	}
	return nil
}

// dropRuleNotificationGroups 移除规则直接绑定通知组的旧表。
// 事件路由改由 NotificationSubscription 负责，旧绑定不再参与任何查询。
func dropRuleNotificationGroups(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS alert_rule_notification_groups`); err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return merr.Server("request failed")
	}
	return nil
}

// dropLegacyEdgeAudits 删除已移除能力的审计表。迁移可重复执行。
func dropLegacyEdgeAudits(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS sftp_audits, k8s_proxy_audits`); err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return merr.Server("request failed")
	}
	return nil
}

// dropStalePermissions 删除代码目录里已经不存在的权限定义及其角色授权。
// 租户权限只由 perm.Catalog 定义，库里多出来的行都是旧版本遗留。
func dropStalePermissions(ctx context.Context, client *ent.Client) error {
	valid := make(map[string]struct{}, len(perm.Catalog()))
	for _, item := range perm.Catalog() {
		valid[item.Name] = struct{}{}
	}
	rows, err := client.Permission.Query().All(ctx)
	if err != nil {
		return databaseErr(slog.Default(), err)
	}
	stale := make([]string, 0, len(rows))
	for _, row := range rows {
		if _, ok := valid[row.Name]; !ok {
			stale = append(stale, row.UID)
		}
	}
	if len(stale) == 0 {
		return nil
	}
	if _, err := client.RolePermission.Delete().
		Where(rolepermission.PermissionUIDIn(stale...)).
		Exec(ctx); err != nil {
		return databaseErr(slog.Default(), err)
	}
	if _, err := client.Permission.Delete().
		Where(permission.UIDIn(stale...)).
		Exec(ctx); err != nil {
		return databaseErr(slog.Default(), err)
	}
	return nil
}
