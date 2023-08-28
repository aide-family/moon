# ************************************************************
# Sequel Ace SQL dump
# 版本号： 20016
#
# https://sequel-ace.com/
# https://github.com/Sequel-Ace/Sequel-Ace
#
# 主机: localhost (MySQL 8.0.33)
# 数据库: prom
# 生成时间: 2023-08-28 03:45:58 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE='NO_AUTO_VALUE_ON_ZERO', SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# 转储表 prom_alarm_histories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_histories`;

CREATE TABLE `prom_alarm_histories` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `node` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'node名称',
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '告警消息状态, 报警和恢复',
  `info` json NOT NULL COMMENT '原始告警消息',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `start_at` bigint unsigned NOT NULL COMMENT '报警开始时间',
  `end_at` bigint unsigned NOT NULL COMMENT '报警恢复时间',
  `duration` bigint unsigned NOT NULL DEFAULT '0' COMMENT '持续时间时间戳, 没有恢复, 时间戳是0',
  `strategy_id` int unsigned NOT NULL DEFAULT '0' COMMENT '规则ID, 用于查询时候',
  `level_id` int unsigned NOT NULL DEFAULT '0' COMMENT '报警等级ID',
  `md5` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'md5',
  PRIMARY KEY (`id`),
  KEY `idx__strategy_id` (`strategy_id`),
  KEY `idx__level_id` (`level_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;



# 转储表 prom_alarm_page_histories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_page_histories`;

CREATE TABLE `prom_alarm_page_histories` (
  `alarm_page_id` int unsigned NOT NULL COMMENT '报警页面ID',
  `history_id` int unsigned NOT NULL COMMENT '历史ID',
  UNIQUE KEY `idx__page_id__history_id` (`alarm_page_id`,`history_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;



# 转储表 prom_alarm_pages
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_pages`;

CREATE TABLE `prom_alarm_pages` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '报警页面名称',
  `remark` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '描述信息',
  `icon` varchar(1024) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '图表',
  `color` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'tab颜色',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态,1启用;2禁用',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `prom_alarm_pages` WRITE;
/*!40000 ALTER TABLE `prom_alarm_pages` DISABLE KEYS */;

INSERT INTO `prom_alarm_pages` (`id`, `name`, `remark`, `icon`, `color`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES
	(1,'system','基础设施','','red',1,'2023-08-10 06:44:59','2023-08-10 06:45:14',NULL),
	(2,'network','网络','','blue',1,'2023-08-16 14:13:33','2023-08-16 14:13:33',NULL),
	(3,'logic','业务告警','','green',1,'2023-08-16 14:18:01','2023-08-16 14:19:21',NULL);

/*!40000 ALTER TABLE `prom_alarm_pages` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_dict
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_dict`;

CREATE TABLE `prom_dict` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '字典名称',
  `category` tinyint NOT NULL DEFAULT '0' COMMENT '字典类型',
  `color` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '#165DFF' COMMENT '字典tag颜色',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `remark` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '字典备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__name__category` (`name`,`category`),
  KEY `idx__category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `prom_dict` WRITE;
/*!40000 ALTER TABLE `prom_dict` DISABLE KEYS */;

INSERT INTO `prom_dict` (`id`, `name`, `category`, `color`, `status`, `remark`, `created_at`, `updated_at`, `deleted_at`)
VALUES
	(1,'基础设施',1,'#165DFF',1,'system core base metrics','2022-01-01 01:01:01','2023-08-10 05:45:21',NULL),
	(2,'网络',1,'rgb(var(--red-5))',1,'network','2023-08-10 05:44:40','2023-08-12 13:38:23',NULL),
	(3,'高优先级',4,'rgb(var(--red-7))',1,'high','2023-08-10 06:42:18','2023-08-20 12:43:08',NULL),
	(4,'中优先级',4,'rgb(var(--red-4))',1,'mid','2023-08-10 06:42:51','2023-08-20 12:43:21',NULL),
	(5,'core',3,'#165DFF',1,'core','2023-08-10 06:55:13','2023-08-10 06:55:24',NULL),
	(6,'网络类型',3,'#165DFF',1,'network','2023-08-10 06:55:13','2023-08-10 06:55:24',NULL),
	(7,'中低优先级',4,'rgb(var(--orange-6))',1,'mid','2023-08-10 06:42:51','2023-08-20 12:43:21',NULL),
	(8,'低优先级',4,'rgb(var(--arcoblue-4))',1,'mid','2023-08-10 06:42:51','2023-08-20 12:43:21',NULL);

/*!40000 ALTER TABLE `prom_dict` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_group_categories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_group_categories`;

CREATE TABLE `prom_group_categories` (
  `dict_id` int unsigned NOT NULL,
  `prom_group_id` int unsigned NOT NULL,
  UNIQUE KEY `idx__dict_id__group_id` (`dict_id`,`prom_group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;



# 转储表 prom_groups
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_groups`;

CREATE TABLE `prom_groups` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '规则组名称',
  `strategy_count` bigint NOT NULL DEFAULT '0' COMMENT '规则数量',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态1:启用;2禁用',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '描述信息',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `prom_groups` WRITE;
/*!40000 ALTER TABLE `prom_groups` DISABLE KEYS */;

INSERT INTO `prom_groups` (`id`, `name`, `strategy_count`, `status`, `remark`, `created_at`, `updated_at`, `deleted_at`)
VALUES
	(1,'prom_system',6,1,'prom系统告警规则组','2023-08-10 13:33:01','2023-08-27 17:10:04',NULL),
	(2,'system_alert',-1,2,'','2023-08-10 05:48:22','2023-08-27 09:05:53','2023-08-27 17:05:53'),
	(3,'system_alert_3',1,2,'','2023-08-10 06:24:50','2023-08-27 09:06:04','2023-08-27 17:06:04'),
	(4,'systemalertxxx',4,1,'','2023-08-10 06:24:50','2023-08-17 21:39:30',NULL),
	(5,'system_5',5,1,'','2023-08-10 06:24:50','2023-08-17 13:43:44',NULL);

/*!40000 ALTER TABLE `prom_groups` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_strategies
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategies`;

CREATE TABLE `prom_strategies` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `group_id` int unsigned NOT NULL DEFAULT '0' COMMENT '所属规则组ID',
  `alert` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '规则名称',
  `expr` text COLLATE utf8mb4_general_ci NOT NULL COMMENT 'prom ql',
  `for` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '10s' COMMENT '持续时间',
  `labels` json NOT NULL COMMENT '标签',
  `annotations` json NOT NULL COMMENT '告警文案',
  `alert_level_id` int NOT NULL DEFAULT '0' COMMENT '告警等级dict ID ',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态: 1启用;2禁用',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx__alart_level_id` (`alert_level_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `prom_strategies` WRITE;
/*!40000 ALTER TABLE `prom_strategies` DISABLE KEYS */;

INSERT INTO `prom_strategies` (`id`, `group_id`, `alert`, `expr`, `for`, `labels`, `annotations`, `alert_level_id`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES
	(1,3,'core_alert-1','up == 1','60s','{\"job\": \"{{$job}}\", \"endpoint\": \"{{$endpoint}}\", \"instance\": \"{{$instance}}\"}','{\"title\": \"server up {{$Value}}\", \"remark\": \"server not up, instance is {{$instance}}\", \"describe\": \"server not up, instance is {{$instance}}\"}',3,1,'2023-08-10 06:45:32','2023-08-27 08:27:20','2023-08-27 16:27:20'),
	(2,3,'core_alert','(sum by (instance) (apache_workers{state=\"busy\"}) / sum by (instance) (apache_scoreboard) ) * 100 < 80','1h','{\"a\": \"a\"}','{\"title\": \"core {{$Value}}\"}',4,1,'2023-08-10 06:56:59','2023-08-27 08:26:22','2023-08-27 16:26:22'),
	(3,2,'core_alert','up == 1','10m','{\"a\": \"a\"}','{\"title\": \"core {{$Value}}\"}',7,2,'2023-08-10 07:21:11','2023-08-27 08:26:26','2023-08-27 16:26:26'),
	(4,2,'core_alert-1xxx','up == 1','1m','{\"a\": \"a\"}','{\"title\": \"core {{$Value}}\"}',8,1,'2023-08-10 07:23:40','2023-08-27 08:26:30','2023-08-27 16:26:30'),
	(5,2,'core_alert','up == 1','1s','{\"a\": \"a\"}','{\"title\": \"core {{$Value}}\"}',3,2,'2023-08-10 07:58:46','2023-08-27 08:26:17','2023-08-27 16:26:18'),
	(6,1,'test_alert_1','go_memstats_sys_bytes{instance=~\"prom:9090\"}','60m','{\"severity\": \"critical\"}','{\"title\": \"{{ $labels.instance }}go_memstats_sys_bytes is {{ $value }}\", \"remark\": \"{{ $labels.instance }}go_memstats_sys_bytes is {{ $value }}, job is {{ $labels.job }}\"}',8,1,'2023-08-25 14:48:10','2023-08-27 08:27:12','2023-08-27 16:27:13'),
	(7,1,'test','process_cpu_seconds_total == 1','60m','{}','{\"title\": \"{{ $label.instance }} cpu\", \"remark\": \"{{ $lable.instance }} cpu value {{ $value }}\"}',4,1,'2023-08-27 05:09:35','2023-08-27 08:27:16','2023-08-27 16:27:16'),
	(8,1,'PrometheusJobMissing','absent(up{job=\"prometheus\"})','10s','{\"severity\": \"warning\"}','{\"title\": \"Prometheus job missing (instance {{ $labels.instance }})\", \"remark\": \"A Prometheus job has disappeared\\\\n  VALUE = {{ $value }}\\\\n  LABELS = {{ $labels }}\"}',4,1,'2023-08-27 08:21:22','2023-08-27 08:21:22',NULL),
	(9,1,'PrometheusTargetMissing','up == 0','10s','{\"severity\": \"critical\"}','{\"title\": \"Prometheus target missing (instance {{ $labels.instance }})\", \"remark\": \"A Prometheus target has disappeared. An exporter might be crashed.\\\\n  VALUE = {{ $value }}\\\\n  LABELS = {{ $labels }}\"}',3,1,'2023-08-27 08:29:52','2023-08-27 08:29:52',NULL),
	(10,1,'PrometheusAllTargetsMissing','sum by (job) (up) == 0','10s','{\"severity\": \"critical\"}','{\"title\": \"Prometheus all targets missing (instance {{ $labels.instance }})\", \"remark\": \"A Prometheus job does not have living target anymore.\\\\n  VALUE = {{ $value }}\\\\n  LABELS = {{ $labels }}\"}',3,1,'2023-08-27 08:32:04','2023-08-27 08:32:04',NULL),
	(11,1,'PrometheusConfigurationReloadFailure','prometheus_config_last_reload_successful','10s','{\"severity\": \"warning\"}','{\"title\": \"Prometheus configuration reload failure (instance {{ $labels.instance }})\", \"remark\": \"Prometheus configuration reload error\\\\n  VALUE = {{ $value }}\\\\n  LABELS = {{ $labels }}\"}',7,1,'2023-08-27 08:40:28','2023-08-27 19:54:36',NULL),
	(12,1,'PrometheusTooManyRestarts','changes(process_start_time_seconds{job=~\"prometheus|pushgateway|alertmanager\"}[15m]) > 2','60s','{\"severity\": \"warning\"}','{\"title\": \"Prometheus too many restarts (instance {{ $labels.instance }})\", \"remark\": \"Prometheus has restarted more than twice in the last 15 minutes. It might be crashlooping.\\\\n  VALUE = {{ $value }}\\\\n  LABELS = {{ $labels }}\"}',4,1,'2023-08-27 08:54:11','2023-08-27 08:54:11',NULL),
	(13,1,'PrometheusNotConnectedToAlertmanager','prometheus_notifications_alertmanagers_discovered ','60s','{}','{\"title\": \"Prometheus not connected to alertmanager (instance {{ $labels.instance }})\", \"remark\": \"Prometheus cannot connect the alertmanager\\\\n  VALUE = {{ $value }}\\\\n  LABELS = {{ $labels }}\"}',8,1,'2023-08-27 09:05:28','2023-08-27 17:30:16',NULL),
	(14,1,'xxx','xx','60m','{}','{\"title\": \"xx\", \"remark\": \"xx\"}',3,1,'2023-08-27 09:09:51','2023-08-27 09:10:04','2023-08-27 17:10:04');

/*!40000 ALTER TABLE `prom_strategies` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_strategy_alarm_pages
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_alarm_pages`;

CREATE TABLE `prom_strategy_alarm_pages` (
  `alarm_page_id` int unsigned NOT NULL,
  `prom_strategy_id` int unsigned NOT NULL,
  UNIQUE KEY `idx__prom_strategy_id__dict_id` (`alarm_page_id`,`prom_strategy_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `prom_strategy_alarm_pages` WRITE;
/*!40000 ALTER TABLE `prom_strategy_alarm_pages` DISABLE KEYS */;

INSERT INTO `prom_strategy_alarm_pages` (`alarm_page_id`, `prom_strategy_id`)
VALUES
	(1,1),
	(1,2),
	(1,3),
	(1,4),
	(1,5),
	(1,6),
	(1,7),
	(1,8),
	(1,9),
	(1,10),
	(1,11),
	(1,12),
	(1,13),
	(1,14),
	(2,2),
	(2,3),
	(3,2),
	(3,3);

/*!40000 ALTER TABLE `prom_strategy_alarm_pages` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_strategy_categories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_categories`;

CREATE TABLE `prom_strategy_categories` (
  `dict_id` int unsigned NOT NULL,
  `prom_strategy_id` int unsigned NOT NULL,
  UNIQUE KEY `idx__prom_strategy_id__dict_id` (`dict_id`,`prom_strategy_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `prom_strategy_categories` WRITE;
/*!40000 ALTER TABLE `prom_strategy_categories` DISABLE KEYS */;

INSERT INTO `prom_strategy_categories` (`dict_id`, `prom_strategy_id`)
VALUES
	(5,2),
	(5,3),
	(5,4),
	(5,5),
	(5,6),
	(5,7),
	(5,13),
	(6,1),
	(6,7);

/*!40000 ALTER TABLE `prom_strategy_categories` ENABLE KEYS */;
UNLOCK TABLES;



/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
