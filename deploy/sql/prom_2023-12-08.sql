# ************************************************************
# Sequel Ace SQL dump
# 版本号： 20016
#
# https://sequel-ace.com/
# https://github.com/Sequel-Ace/Sequel-Ace
#
# 主机: 124.223.104.203 (MySQL 8.0.34)
# 数据库: prom
# 生成时间: 2023-12-08 03:11:10 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE='NO_AUTO_VALUE_ON_ZERO', SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# 转储表 casbin_rule
# ------------------------------------------------------------

DROP TABLE IF EXISTS `casbin_rule`;

CREATE TABLE `casbin_rule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ptype` varchar(100) DEFAULT NULL,
  `v0` varchar(100) DEFAULT NULL,
  `v1` varchar(100) DEFAULT NULL,
  `v2` varchar(100) DEFAULT NULL,
  `v3` varchar(100) DEFAULT NULL,
  `v4` varchar(100) DEFAULT NULL,
  `v5` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_casbin_rule` (`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `casbin_rule` WRITE;
/*!40000 ALTER TABLE `casbin_rule` DISABLE KEYS */;

INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`)
VALUES
	(1,'p','1','*','*','','',''),
	(82,'p','2','/api/v1/dict/batch/delete','POST','','',''),
	(79,'p','2','/api/v1/dict/create','POST','','',''),
	(83,'p','2','/api/v1/dict/delete','POST','','',''),
	(84,'p','2','/api/v1/dict/get','POST','','',''),
	(80,'p','2','/api/v1/dict/list','POST','','',''),
	(85,'p','2','/api/v1/dict/select','POST','','',''),
	(81,'p','2','/api/v1/dict/status/update/batch','POST','','',''),
	(86,'p','2','/api/v1/dict/update','POST','','',''),
	(73,'p','2','/api/v1/user/create','POST','','',''),
	(74,'p','2','/api/v1/user/delete','POST','','',''),
	(75,'p','2','/api/v1/user/list','POST','','',''),
	(76,'p','2','/api/v1/user/roles/relate','POST','','',''),
	(77,'p','2','/api/v1/user/select','POST','','',''),
	(78,'p','2','/api/v1/user/status/edit','POST','','',''),
	(38,'p','3','/api/v1/system/api/get','POST','','',''),
	(36,'p','3','/api/v1/user/delete','POST','','',''),
	(37,'p','3','/api/v1/user/list','POST','','','');

/*!40000 ALTER TABLE `casbin_rule` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_alarm_been_notify_chat_groups
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_been_notify_chat_groups`;

CREATE TABLE `prom_alarm_been_notify_chat_groups` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `realtime_alarm_id` int unsigned NOT NULL COMMENT '告警ID',
  `chat_group_id` int unsigned NOT NULL COMMENT '通知组ID',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `msg` text NOT NULL COMMENT '通知的消息',
  `prom_alarm_notify_id` int unsigned NOT NULL COMMENT '通知ID',
  PRIMARY KEY (`id`),
  KEY `idx__realtime_alarm_id` (`realtime_alarm_id`),
  KEY `idx__chat_group_id` (`chat_group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_alarm_been_notify_members
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_been_notify_members`;

CREATE TABLE `prom_alarm_been_notify_members` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `realtime_alarm_id` int unsigned NOT NULL COMMENT '告警ID',
  `notify_types` json NOT NULL COMMENT '通知方式',
  `member_id` int unsigned NOT NULL COMMENT '通知人员ID',
  `msg` text NOT NULL COMMENT '通知的消息',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `prom_alarm_notify_id` int unsigned NOT NULL COMMENT '通知ID',
  PRIMARY KEY (`id`),
  KEY `idx__realtime_alarm_id` (`realtime_alarm_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_alarm_chat_groups
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_chat_groups`;

CREATE TABLE `prom_alarm_chat_groups` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `remark` varchar(255) NOT NULL COMMENT '备注',
  `name` varchar(64) NOT NULL COMMENT '名称',
  `hook` varchar(255) NOT NULL COMMENT '钩子地址',
  `notify_app` tinyint NOT NULL DEFAULT '1' COMMENT '通知方式',
  `hook_name` varchar(64) NOT NULL COMMENT '钩子名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_alarm_histories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_histories`;

CREATE TABLE `prom_alarm_histories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `instance` varchar(64) NOT NULL COMMENT 'instance名称',
  `status` varchar(16) NOT NULL COMMENT '告警消息状态, 报警和恢复',
  `info` json NOT NULL COMMENT '原始告警消息',
  `start_at` bigint NOT NULL COMMENT '报警开始时间',
  `end_at` bigint NOT NULL COMMENT '报警恢复时间',
  `duration` bigint NOT NULL COMMENT '持续时间时间戳, 没有恢复, 时间戳是0',
  `strategy_id` int unsigned NOT NULL COMMENT '规则ID, 用于查询时候',
  `level_id` int unsigned NOT NULL COMMENT '报警等级ID',
  `md5` char(32) NOT NULL COMMENT 'md5',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__md5` (`md5`),
  KEY `idx__strategy_id` (`strategy_id`),
  KEY `idx__level_id` (`level_id`),
  KEY `idx__instance` (`instance`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_alarm_histories` WRITE;
/*!40000 ALTER TABLE `prom_alarm_histories` DISABLE KEYS */;

INSERT INTO `prom_alarm_histories` (`id`, `created_at`, `updated_at`, `deleted_at`, `instance`, `status`, `info`, `start_at`, `end_at`, `duration`, `strategy_id`, `level_id`, `md5`)
VALUES
	(1,'2023-12-05 09:00:24','2023-12-05 09:09:40',0,'prom:9090','1','{\"endsAt\": 1700195031, \"labels\": {\"job\": \"prometheus_server\", \"instance\": \"prom:9090\", \"severity\": \"warning\", \"alertname\": \"up\"}, \"status\": \"firing\", \"startsAt\": 1697516631, \"annotations\": {\"title\": \"Instance prom:9090 down\", \"description\": \"prom:9090 of job prometheus_server has been down for more than 1 minute.\"}, \"fingerprint\": \"a2f1d053912a1d40\", \"generatorURL\": \"http://8ce022f7e801:9090/graph?g0.expr=up+%3D%3D+1&g0.tab=1\"}',1697516631,1700195031,2678400,0,0,'a2f1d053912a1d40'),
	(2,'2023-12-05 09:00:24','2023-12-05 09:14:26',0,'pushgateway:9091','1','{\"endsAt\": -62135596800, \"labels\": {\"job\": \"pushgateway\", \"instance\": \"pushgateway:9091\", \"severity\": \"warning\", \"alertname\": \"up\"}, \"status\": \"firing\", \"startsAt\": 1697516631, \"annotations\": {\"title\": \"Instance pushgateway:9091 down\", \"description\": \"pushgateway:9091 of job pushgateway has been down for more than 1 minute.\"}, \"fingerprint\": \"dfd0a50a1843cd59\", \"generatorURL\": \"http://8ce022f7e801:9090/graph?g0.expr=up+%3D%3D+1&g0.tab=1\"}',1697516631,-62135596800,-9223372036,0,0,'dfd0a50a1843cd59');

/*!40000 ALTER TABLE `prom_alarm_histories` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_alarm_intervenes
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_intervenes`;

CREATE TABLE `prom_alarm_intervenes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `realtime_alarm_id` int unsigned NOT NULL COMMENT '告警ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `intervened_at` bigint NOT NULL COMMENT '干预时间',
  `remark` varchar(255) NOT NULL COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx__user_id` (`user_id`),
  KEY `idx__realtime_alarm_id` (`realtime_alarm_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_alarm_notifies
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_notifies`;

CREATE TABLE `prom_alarm_notifies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL COMMENT '通知名称',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `remark` varchar(255) NOT NULL COMMENT '备注',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_alarm_notify_members
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_notify_members`;

CREATE TABLE `prom_alarm_notify_members` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `prom_alarm_notify_id` int unsigned NOT NULL COMMENT '通知ID',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `notify_types` json NOT NULL COMMENT '通知方式',
  `member_id` int unsigned NOT NULL COMMENT '成员ID',
  PRIMARY KEY (`id`),
  KEY `idx__prom_alarm_notify_id` (`prom_alarm_notify_id`),
  KEY `idx__member_id` (`member_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_alarm_page_histories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_page_histories`;

CREATE TABLE `prom_alarm_page_histories` (
  `prom_alarm_page_id` bigint unsigned NOT NULL,
  `prom_alarm_history_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`prom_alarm_page_id`,`prom_alarm_history_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_alarm_page_realtime_alarms
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_page_realtime_alarms`;

CREATE TABLE `prom_alarm_page_realtime_alarms` (
  `prom_alarm_page_id` bigint unsigned NOT NULL,
  `prom_alarm_realtime_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`prom_alarm_page_id`,`prom_alarm_realtime_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_alarm_pages
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_pages`;

CREATE TABLE `prom_alarm_pages` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL COMMENT '报警页面名称',
  `remark` varchar(255) NOT NULL COMMENT '描述信息',
  `icon` varchar(1024) NOT NULL COMMENT '图表',
  `color` varchar(64) NOT NULL COMMENT 'tab颜色',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态,1启用',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_alarm_realtime
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_realtime`;

CREATE TABLE `prom_alarm_realtime` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `strategy_id` int unsigned NOT NULL COMMENT '策略ID',
  `instance` varchar(64) NOT NULL COMMENT 'instance名称',
  `note` varchar(255) NOT NULL COMMENT '告警内容',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '告警状态: 1告警',
  `event_at` bigint NOT NULL COMMENT '告警时间',
  `notified_at` bigint NOT NULL DEFAULT '0' COMMENT '通知时间',
  `history_id` int unsigned NOT NULL COMMENT '历史记录ID',
  `level_id` int unsigned NOT NULL COMMENT '告警等级ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__history_id` (`history_id`),
  KEY `idx__strategy_id` (`strategy_id`),
  KEY `idx__instance` (`instance`),
  KEY `idx__level_id` (`level_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_alarm_realtime` WRITE;
/*!40000 ALTER TABLE `prom_alarm_realtime` DISABLE KEYS */;

INSERT INTO `prom_alarm_realtime` (`id`, `created_at`, `updated_at`, `deleted_at`, `strategy_id`, `instance`, `note`, `status`, `event_at`, `notified_at`, `history_id`, `level_id`)
VALUES
	(1,'2023-12-05 09:00:25','2023-12-05 09:09:41',0,0,'prom:9090','prom:9090 of job prometheus_server has been down for more than 1 minute.',1,1697516631,0,1,0),
	(2,'2023-12-05 09:00:25','2023-12-05 09:14:26',0,0,'pushgateway:9091','pushgateway:9091 of job pushgateway has been down for more than 1 minute.',1,1697516631,0,2,0);

/*!40000 ALTER TABLE `prom_alarm_realtime` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_alarm_suppress
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_suppress`;

CREATE TABLE `prom_alarm_suppress` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `realtime_alarm_id` int unsigned NOT NULL COMMENT '告警ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `suppressed_at` bigint NOT NULL COMMENT '抑制时间',
  `remark` varchar(255) NOT NULL COMMENT '备注',
  `duration` bigint NOT NULL COMMENT '抑制时长',
  PRIMARY KEY (`id`),
  KEY `idx__realtime_alarm_id` (`realtime_alarm_id`),
  KEY `idx__user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_dict
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_dict`;

CREATE TABLE `prom_dict` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL COMMENT '字典名称',
  `category` tinyint NOT NULL COMMENT '字典类型',
  `color` varchar(32) NOT NULL DEFAULT '#165DFF' COMMENT '字典tag颜色',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `remark` varchar(255) NOT NULL COMMENT '字典备注',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__name__category` (`name`,`category`),
  KEY `idx__category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_dict` WRITE;
/*!40000 ALTER TABLE `prom_dict` DISABLE KEYS */;

INSERT INTO `prom_dict` (`id`, `created_at`, `updated_at`, `deleted_at`, `name`, `category`, `color`, `status`, `remark`)
VALUES
	(1,'2023-12-04 12:52:30','2023-12-06 22:32:28',0,'xxx1',3,'#c00',1,'xxxxxx');

/*!40000 ALTER TABLE `prom_dict` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_group_categories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_group_categories`;

CREATE TABLE `prom_group_categories` (
  `prom_strategy_group_id` bigint unsigned NOT NULL,
  `prom_dict_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`prom_strategy_group_id`,`prom_dict_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_notify_chat_groups
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_notify_chat_groups`;

CREATE TABLE `prom_notify_chat_groups` (
  `prom_alarm_notify_id` bigint unsigned NOT NULL,
  `prom_alarm_chat_group_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`prom_alarm_notify_id`,`prom_alarm_chat_group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_strategies
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategies`;

CREATE TABLE `prom_strategies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `group_id` int unsigned NOT NULL COMMENT '所属规则组ID',
  `alert` varchar(64) NOT NULL COMMENT '规则名称',
  `expr` text NOT NULL COMMENT 'prom ql',
  `for` varchar(64) NOT NULL DEFAULT '10s' COMMENT '持续时间',
  `labels` json NOT NULL COMMENT '标签',
  `annotations` json NOT NULL COMMENT '告警文案',
  `alert_level_id` int unsigned NOT NULL COMMENT '告警等级dict ID',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态: 1启用',
  `remark` varchar(255) NOT NULL COMMENT '描述信息',
  `max_suppress` bigint NOT NULL DEFAULT '0' COMMENT '最大抑制时长(s)',
  `send_recover` tinyint NOT NULL DEFAULT '0' COMMENT '是否发送告警恢复通知',
  `send_interval` bigint NOT NULL DEFAULT '0' COMMENT '发送告警时间间隔(s), 默认为for的10倍时间分钟, 用于长时间未消警情况',
  PRIMARY KEY (`id`),
  KEY `idx__alert_level_id` (`alert_level_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_strategy_alarm_pages
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_alarm_pages`;

CREATE TABLE `prom_strategy_alarm_pages` (
  `prom_strategy_id` bigint unsigned NOT NULL,
  `alarm_page_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`prom_strategy_id`,`alarm_page_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_strategy_categories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_categories`;

CREATE TABLE `prom_strategy_categories` (
  `prom_strategy_id` bigint unsigned NOT NULL,
  `dict_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`prom_strategy_id`,`dict_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_strategy_groups
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_groups`;

CREATE TABLE `prom_strategy_groups` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL COMMENT '规则组名称',
  `strategy_count` bigint NOT NULL COMMENT '规则数量',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态1:启用',
  `remark` varchar(255) NOT NULL COMMENT '描述信息',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_strategy_notifies
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_notifies`;

CREATE TABLE `prom_strategy_notifies` (
  `prom_strategy_id` bigint unsigned NOT NULL,
  `prom_alarm_notify_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`prom_strategy_id`,`prom_alarm_notify_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_strategy_notify_upgrades
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_notify_upgrades`;

CREATE TABLE `prom_strategy_notify_upgrades` (
  `prom_strategy_id` bigint unsigned NOT NULL,
  `prom_alarm_notify_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`prom_strategy_id`,`prom_alarm_notify_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 sys_apis
# ------------------------------------------------------------

DROP TABLE IF EXISTS `sys_apis`;

CREATE TABLE `sys_apis` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `name` varchar(64) NOT NULL COMMENT 'api名称',
  `path` varchar(255) NOT NULL COMMENT 'api路径',
  `method` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'POST' COMMENT '请求方法',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '这个API没有说明, 赶紧补充吧' COMMENT '备注',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `module` int NOT NULL DEFAULT '0' COMMENT '模块',
  `domain` int NOT NULL DEFAULT '0' COMMENT '领域',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__name` (`name`),
  UNIQUE KEY `idx__path` (`path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `sys_apis` WRITE;
/*!40000 ALTER TABLE `sys_apis` DISABLE KEYS */;

INSERT INTO `sys_apis` (`id`, `created_at`, `updated_at`, `name`, `path`, `method`, `status`, `remark`, `deleted_at`, `module`, `domain`)
VALUES
	(1,'2023-12-02 05:50:47','2023-12-03 13:24:41','创建用户','/api/v1/user/create','POST',1,'用于管理员创建新用户所用',0,0,0),
	(2,'2023-12-02 12:30:00','2023-12-02 12:30:00','删除用户','/api/v1/user/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(3,'2023-12-02 12:30:00','2023-12-02 12:30:00','获取用户列表','/api/v1/user/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(4,'2023-12-02 12:30:00','2023-12-02 12:30:00','用户关联角色','/api/v1/user/roles/relate','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(5,'2023-12-02 12:30:00','2023-12-02 12:30:00','获取用户下拉列表','/api/v1/user/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(6,'2023-12-02 12:30:00','2023-12-02 12:30:00','批量修改用户状态','/api/v1/user/status/edit','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(7,'2023-12-02 12:30:00','2023-12-02 12:30:00','更新用户基础信息','/api/v1/user/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(8,'2023-12-02 12:30:00','2023-12-02 12:30:00','修改密码','/api/v1/user/password/edit','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(9,'2023-12-02 12:30:00','2023-12-02 12:30:00','获取用户详情','/api/v1/user/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(10,'2023-12-02 12:32:49','2023-12-02 12:32:49','创建角色','/api/v1/role/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(11,'2023-12-02 12:32:49','2023-12-02 12:32:49','删除角色','/api/v1/role/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(12,'2023-12-02 12:32:49','2023-12-02 12:32:49','获取角色详情','/api/v1/role/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(13,'2023-12-02 12:32:49','2023-12-02 12:32:49','获取角色列表','/api/v1/role/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(14,'2023-12-02 12:32:49','2023-12-02 12:32:49','角色关联API','/api/v1/role/relate/api','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(15,'2023-12-02 12:32:49','2023-12-02 12:32:49','获取角色下拉列表','/api/v1/role/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(16,'2023-12-02 12:32:49','2023-12-02 12:32:49','更新角色','/api/v1/role/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(17,'2023-12-02 12:35:40','2023-12-02 12:35:40','创建API','/api/v1/system/api/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(18,'2023-12-02 12:35:40','2023-12-02 12:35:40','删除API','/api/v1/system/api/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(19,'2023-12-02 12:35:40','2023-12-02 12:35:40','获取API详情','/api/v1/system/api/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(20,'2023-12-02 12:35:40','2023-12-02 12:35:40','获取API列表','/api/v1/system/api/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(21,'2023-12-02 12:35:40','2023-12-02 13:15:43','获取API下拉列表','/api/v1/system/api/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(22,'2023-12-02 12:35:40','2023-12-02 12:35:40','更新API数据','/api/v1/system/api/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(23,'2023-12-03 05:51:05','2023-12-03 05:51:05','测试','/test','POST',1,'测试详情',0,0,0),
	(24,'2023-12-03 08:44:47','2023-12-03 08:44:47','创建字典','/api/v1/dict/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(25,'2023-12-03 08:45:11','2023-12-03 08:45:11','字典列表','/api/v1/dict/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(26,'2023-12-04 13:13:55','2023-12-04 22:11:33','更新字典状态','/api/v1/dict/status/update/batch','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(27,'2023-12-04 14:08:31','2023-12-04 22:11:16','删除字典列表','/api/v1/dict/batch/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(29,'2023-12-04 14:13:41','2023-12-04 14:13:41','删除字典','/api/v1/dict/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(30,'2023-12-04 14:14:09','2023-12-04 14:14:09','获取字典详情','/api/v1/dict/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(31,'2023-12-04 14:15:01','2023-12-04 14:15:01','获取字典下拉列表','/api/v1/dict/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(32,'2023-12-04 14:15:35','2023-12-04 14:15:35','更新字典','/api/v1/dict/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(33,'2023-12-05 09:16:39','2023-12-05 09:16:39','实时告警列表','/api/v1/alarm/realtime/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0);

/*!40000 ALTER TABLE `sys_apis` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 sys_role_apis
# ------------------------------------------------------------

DROP TABLE IF EXISTS `sys_role_apis`;

CREATE TABLE `sys_role_apis` (
  `sys_role_id` bigint unsigned NOT NULL,
  `sys_api_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`sys_role_id`,`sys_api_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `sys_role_apis` WRITE;
/*!40000 ALTER TABLE `sys_role_apis` DISABLE KEYS */;

INSERT INTO `sys_role_apis` (`sys_role_id`, `sys_api_id`)
VALUES
	(2,1),
	(2,2),
	(2,3),
	(2,4),
	(2,5),
	(2,6),
	(2,24),
	(2,25),
	(2,26),
	(2,27),
	(2,29),
	(2,30),
	(2,31),
	(2,32),
	(3,2),
	(3,3),
	(3,19);

/*!40000 ALTER TABLE `sys_role_apis` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 sys_roles
# ------------------------------------------------------------

DROP TABLE IF EXISTS `sys_roles`;

CREATE TABLE `sys_roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `remark` varchar(255) NOT NULL COMMENT '备注',
  `name` varchar(64) NOT NULL COMMENT '角色名称',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `sys_roles` WRITE;
/*!40000 ALTER TABLE `sys_roles` DISABLE KEYS */;

INSERT INTO `sys_roles` (`id`, `created_at`, `updated_at`, `remark`, `name`, `status`, `deleted_at`)
VALUES
	(1,'2023-12-02 05:52:02','2023-12-03 11:59:59','超级管理员角色','超级管理员',1,0),
	(2,'2023-12-03 03:47:28','2023-12-04 22:48:24','普通管理员角色','普通管理员',1,0),
	(3,'2023-12-03 07:47:23','2023-12-03 16:28:46','普通人物角色','普通人物',1,0);

/*!40000 ALTER TABLE `sys_roles` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 sys_user_roles
# ------------------------------------------------------------

DROP TABLE IF EXISTS `sys_user_roles`;

CREATE TABLE `sys_user_roles` (
  `sys_role_id` bigint unsigned NOT NULL,
  `sys_user_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`sys_role_id`,`sys_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `sys_user_roles` WRITE;
/*!40000 ALTER TABLE `sys_user_roles` DISABLE KEYS */;

INSERT INTO `sys_user_roles` (`sys_role_id`, `sys_user_id`)
VALUES
	(1,1),
	(1,2);

/*!40000 ALTER TABLE `sys_user_roles` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 sys_users
# ------------------------------------------------------------

DROP TABLE IF EXISTS `sys_users`;

CREATE TABLE `sys_users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `username` varchar(64) NOT NULL COMMENT '用户名',
  `nickname` varchar(64) NOT NULL COMMENT '昵称',
  `password` varchar(255) NOT NULL COMMENT '密码',
  `email` varchar(64) NOT NULL COMMENT '邮箱',
  `phone` varchar(64) NOT NULL COMMENT '手机号',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `remark` varchar(255) NOT NULL COMMENT '备注',
  `avatar` varchar(255) NOT NULL COMMENT '头像',
  `salt` varchar(16) NOT NULL COMMENT '盐',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `gender` tinyint NOT NULL DEFAULT '0' COMMENT '性别',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__username` (`username`),
  UNIQUE KEY `idx__email` (`email`),
  UNIQUE KEY `idx__phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `sys_users` WRITE;
/*!40000 ALTER TABLE `sys_users` DISABLE KEYS */;

INSERT INTO `sys_users` (`id`, `created_at`, `updated_at`, `username`, `nickname`, `password`, `email`, `phone`, `status`, `remark`, `avatar`, `salt`, `deleted_at`, `gender`)
VALUES
	(1,'1970-01-01 08:00:00','2023-12-06 00:17:17','biao.hu','梧桐','zUnQGjDhAd7ChhVLa9R+Dw==','aidedevops@163.com','18275077343',1,'','http://img.touxiangwu.com/uploads/allimg/2022010212/0k31jrfuezv.jpg','2441c436f3fa3613',0,1),
	(2,'1970-01-01 08:00:00','2023-12-06 21:16:48','lier','莉莉','xen5WnrlweLWODeArERA+g==','aidelier@163.com','15585315449',1,'','https://img1.baidu.com/it/u=1090762042,1225374156&fm=253&fmt=auto&app=138&f=JPEG?w=500&h=500','5942d93b2c07ff52',0,2),
	(3,'2023-12-02 06:00:20','2023-12-06 21:17:31','test','测试号','53ZnDtGGF4JyC+fCFJsKHA==','aidetest@163.com','15585310001',2,'','https://img1.baidu.com/it/u=1285375996,3783960243&fm=253&fmt=auto&app=138&f=JPEG?w=500&h=500','0addd2e0f4145e3b',0,2),
	(4,'2023-12-02 09:20:16','2023-12-06 21:28:48','sss','','C6xWQLqxBnZOv813W7RHuw==','123@163.com','13355777441',1,'','','8c12d009db56e0ea',0,1),
	(5,'2023-12-02 09:29:02','2023-12-02 12:34:32','sfasfas','','Bzlv+7/jQnmoNslxPHq2TQ==','12312@qq.com','18812312311',2,'','','96b530aedc68d976',1701520472,0);

/*!40000 ALTER TABLE `sys_users` ENABLE KEYS */;
UNLOCK TABLES;



/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
