# ************************************************************
# Sequel Ace SQL dump
# 版本号： 20016
#
# https://sequel-ace.com/
# https://github.com/Sequel-Ace/Sequel-Ace
#
# 主机: 124.223.104.203 (MySQL 8.0.27)
# 数据库: prometheus-manager
# 生成时间: 2024-01-31 14:02:26 +0000
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
) ENGINE=InnoDB AUTO_INCREMENT=169 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

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
	(114,'p','3','/api/v1/alarm_page/alarm/count','POST','','',''),
	(115,'p','3','/api/v1/alarm_page/create','POST','','',''),
	(116,'p','3','/api/v1/alarm_page/get','POST','','',''),
	(117,'p','3','/api/v1/alarm_page/list','POST','','',''),
	(118,'p','3','/api/v1/alarm_page/select','POST','','',''),
	(119,'p','3','/api/v1/alarm_page/update','POST','','',''),
	(161,'p','3','/api/v1/alarm/history/get','POST','','',''),
	(160,'p','3','/api/v1/alarm/history/list','POST','','',''),
	(144,'p','3','/api/v1/alarm/realtime/detail','POST','','',''),
	(145,'p','3','/api/v1/alarm/realtime/intervene','POST','','',''),
	(159,'p','3','/api/v1/alarm/realtime/list','POST','','',''),
	(146,'p','3','/api/v1/alarm/realtime/suppress','POST','','',''),
	(147,'p','3','/api/v1/alarm/realtime/upgrade','POST','','',''),
	(120,'p','3','/api/v1/chat/group/create','POST','','',''),
	(121,'p','3','/api/v1/chat/group/get','POST','','',''),
	(122,'p','3','/api/v1/chat/group/list','POST','','',''),
	(123,'p','3','/api/v1/chat/group/select','POST','','',''),
	(124,'p','3','/api/v1/chat/group/update','POST','','',''),
	(155,'p','3','/api/v1/dict/create','POST','','',''),
	(157,'p','3','/api/v1/dict/get','POST','','',''),
	(156,'p','3','/api/v1/dict/list','POST','','',''),
	(158,'p','3','/api/v1/dict/select','POST','','',''),
	(125,'p','3','/api/v1/endpoint/append','POST','','',''),
	(126,'p','3','/api/v1/endpoint/detail','POST','','',''),
	(127,'p','3','/api/v1/endpoint/edit','POST','','',''),
	(128,'p','3','/api/v1/endpoint/list','POST','','',''),
	(129,'p','3','/api/v1/endpoint/select','POST','','',''),
	(139,'p','3','/api/v1/prom/notify/create','POST','','',''),
	(140,'p','3','/api/v1/prom/notify/get','POST','','',''),
	(141,'p','3','/api/v1/prom/notify/list','POST','','',''),
	(143,'p','3','/api/v1/prom/notify/select','POST','','',''),
	(142,'p','3','/api/v1/prom/notify/update','POST','','',''),
	(149,'p','3','/api/v1/role/get','POST','','',''),
	(150,'p','3','/api/v1/role/list','POST','','',''),
	(151,'p','3','/api/v1/role/select','POST','','',''),
	(164,'p','3','/api/v1/strategy/create','POST','','',''),
	(163,'p','3','/api/v1/strategy/detail','POST','','',''),
	(130,'p','3','/api/v1/strategy/group/all/list','POST','','',''),
	(131,'p','3','/api/v1/strategy/group/create','POST','','',''),
	(132,'p','3','/api/v1/strategy/group/export','POST','','',''),
	(133,'p','3','/api/v1/strategy/group/get','POST','','',''),
	(134,'p','3','/api/v1/strategy/group/import','POST','','',''),
	(135,'p','3','/api/v1/strategy/group/list','POST','','',''),
	(136,'p','3','/api/v1/strategy/group/select','POST','','',''),
	(137,'p','3','/api/v1/strategy/group/status/batch/update','POST','','',''),
	(138,'p','3','/api/v1/strategy/group/update','POST','','',''),
	(162,'p','3','/api/v1/strategy/list','POST','','',''),
	(168,'p','3','/api/v1/strategy/notify/object/bind','POST','','',''),
	(167,'p','3','/api/v1/strategy/select','POST','','',''),
	(166,'p','3','/api/v1/strategy/status/batch/update','POST','','',''),
	(165,'p','3','/api/v1/strategy/update','POST','','',''),
	(152,'p','3','/api/v1/system/api/get','POST','','',''),
	(153,'p','3','/api/v1/system/api/list','POST','','',''),
	(154,'p','3','/api/v1/system/api/select','POST','','',''),
	(148,'p','3','/api/v1/user/password/edit','POST','','','');

/*!40000 ALTER TABLE `casbin_rule` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 endpoints
# ------------------------------------------------------------

DROP TABLE IF EXISTS `endpoints`;

CREATE TABLE `endpoints` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL COMMENT '名称',
  `endpoint` varchar(255) NOT NULL COMMENT '地址',
  `remark` varchar(255) NOT NULL DEFAULT '这个Endpoint没有说明, 赶紧补充吧' COMMENT '备注',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态: 1启用',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__endpoint` (`endpoint`),
  UNIQUE KEY `idx__name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `endpoints` WRITE;
/*!40000 ALTER TABLE `endpoints` DISABLE KEYS */;

INSERT INTO `endpoints` (`id`, `created_at`, `updated_at`, `deleted_at`, `name`, `endpoint`, `remark`, `status`)
VALUES
	(1,'2023-12-27 03:00:36','2024-01-30 09:29:37',0,'Prometheus','http://124.223.10.2:9090','promehtue 开发环境',2),
	(3,'2023-12-27 03:00:36','2024-01-08 01:18:06',0,'PrometheusServer','https://prom-server.aide-cloud.cn','promehtue 开发环境',1),
	(4,'2023-12-27 03:00:36','2024-01-27 13:09:18',0,'localhost','http://localhost:9090','promehtue 开发环境',2);

/*!40000 ALTER TABLE `endpoints` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 external_customer_hooks
# ------------------------------------------------------------

DROP TABLE IF EXISTS `external_customer_hooks`;

CREATE TABLE `external_customer_hooks` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `hook` varchar(255) NOT NULL COMMENT '钩子地址',
  `hook_name` varchar(64) NOT NULL COMMENT '钩子名称',
  `notify_app` tinyint NOT NULL DEFAULT '1' COMMENT '通知方式',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `remark` varchar(255) NOT NULL COMMENT '备注',
  `customer_id` int unsigned NOT NULL COMMENT '外部客户ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 external_customers
# ------------------------------------------------------------

DROP TABLE IF EXISTS `external_customers`;

CREATE TABLE `external_customers` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL COMMENT '外部客户名称',
  `address` varchar(255) NOT NULL COMMENT '外部客户地址',
  `contact` varchar(64) NOT NULL COMMENT '外部客户联系人',
  `phone` varchar(64) NOT NULL COMMENT '外部客户联系电话',
  `email` varchar(64) NOT NULL COMMENT '外部客户联系邮箱',
  `remark` varchar(255) NOT NULL COMMENT '外部客户备注',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '外部客户状态',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



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
  `template` text NOT NULL COMMENT '消息模板',
  `title` varchar(64) NOT NULL DEFAULT '' COMMENT '消息标题',
  `secret` varchar(128) NOT NULL DEFAULT '' COMMENT '通信密钥',
  `create_by` int unsigned NOT NULL DEFAULT '0' COMMENT '创建人ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_alarm_chat_groups` WRITE;
/*!40000 ALTER TABLE `prom_alarm_chat_groups` DISABLE KEYS */;

INSERT INTO `prom_alarm_chat_groups` (`id`, `created_at`, `updated_at`, `deleted_at`, `status`, `remark`, `name`, `hook`, `notify_app`, `hook_name`, `template`, `title`, `secret`, `create_by`)
VALUES
	(1,'2024-01-24 10:16:54','2024-01-28 14:13:04',0,1,'企微告警hook','企业微信机器人','https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=90a2ff5f-c39a-4a14-b5c1-3d91084954a3',2,'测试告警','## prometheus监控告警【{{ $status }}】\n\n* 告警时间: {{ $startsAt }}\n* 恢复时间: {{ $endsAt }}\n* 告警标题: {{ $annotations.title }}\n* 告警内容: {{ $annotations.description }}\n* 唯一指纹: {{ $fingerprint }}\n* 告警标识\n    > 规则名称: {{ $labels.alertname }}\n    > 机器名称: {{ $labels.endpoint }}\n    > 实例名称: {{ $labels.instance }}','机器名称: {{ $labels.endpoint }}','',1),
	(2,'2024-01-25 03:55:21','2024-01-28 14:13:41',0,1,'飞书告警hook','飞书机器人','https://open.feishu.cn/open-apis/bot/v2/hook/32b3d11e-3c20-45a9-a6b4-03e2bba19b39',3,'监控告警','{\n  \"msg_type\": \"interactive\",\n  \"card\": {\n    \"elements\": [\n      {\n        \"tag\": \"div\",\n        \"text\": {\n          \"content\": \"* 告警时间: {{ $startsAt }}\\n* 恢复时间: {{ $endsAt }}\\n* 告警标题: {{ $annotations.title }}\\n* 告警内容: {{ $annotations.description }}\\n* 唯一指纹: {{ $fingerprint }}\\n* 告警标识\\n    > 规则名称: {{ $labels.alertname }}\\n    > 机器名称: {{ $labels.endpoint }}\\n    > 实例名称: {{ $labels.instance }}\",\n          \"tag\": \"lark_md\"\n        }\n      }\n    ],\n    \"header\": {\n      \"title\": {\n        \"content\": \"prometheus监控告警【{{ $status }}】\",\n        \"tag\": \"plain_text\"\n      }\n    }\n  }\n}','机器名称: {{ $labels.endpoint }}','NYVCV0mYz6nvmiDQlr9LRd',1),
	(3,'2024-01-31 12:39:08','2024-01-31 12:39:08',0,1,'','企业微信告警测试','https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=c40f025a-0b7f-4929-ad9d-6a0e84cb51f8',2,'告警机器人1','## prometheus监控告警【{{ $status }}】\n\n* 告警时间: {{ $startsAt }}\n* 恢复时间: {{ $endsAt }}\n* 告警标题: {{ $annotations.title }}\n* 告警内容: {{ $annotations.description }}\n* 唯一指纹: {{ $fingerprint }}\n* 告警标识\n    > 规则名称: {{ $labels.alertname }}\n    > 机器名称: {{ $labels.endpoint }}\n    > 实例名称: {{ $labels.instance }}','机器名称: {{ $labels.endpoint }}','',1);

/*!40000 ALTER TABLE `prom_alarm_chat_groups` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_alarm_external_notify_obj_external_customer_hooks
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_external_notify_obj_external_customer_hooks`;

CREATE TABLE `prom_alarm_external_notify_obj_external_customer_hooks` (
  `external_notify_obj_id` int unsigned NOT NULL,
  `external_customer_hook_id` int unsigned NOT NULL,
  PRIMARY KEY (`external_notify_obj_id`,`external_customer_hook_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_alarm_external_notify_obj_external_customers
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_external_notify_obj_external_customers`;

CREATE TABLE `prom_alarm_external_notify_obj_external_customers` (
  `external_notify_obj_id` int unsigned NOT NULL,
  `external_customer_id` int unsigned NOT NULL,
  PRIMARY KEY (`external_notify_obj_id`,`external_customer_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# 转储表 prom_alarm_external_notify_objs
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_external_notify_objs`;

CREATE TABLE `prom_alarm_external_notify_objs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL COMMENT '外部通知对象名称',
  `remark` varchar(255) NOT NULL COMMENT '外部通知对象说明',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
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
) ENGINE=InnoDB AUTO_INCREMENT=130205 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_alarm_histories` WRITE;
/*!40000 ALTER TABLE `prom_alarm_histories` DISABLE KEYS */;

INSERT INTO `prom_alarm_histories` (`id`, `created_at`, `updated_at`, `deleted_at`, `instance`, `status`, `info`, `start_at`, `end_at`, `duration`, `strategy_id`, `level_id`, `md5`)
VALUES
	(1,'2024-01-24 06:44:45','2024-01-24 06:47:45',0,'pushgateway:9091','1','{\"endsAt\": \"\", \"labels\": {\"job\": \"pushgateway\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"pushgateway:9091\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-24T14:44:35+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 process_cpu_seconds_total\", \"description\": \"pushgateway:9091 process_cpu_seconds_total value is 1197.28\"}, \"fingerprint\": \"484d593da0623094ef2b917437581cad\", \"generatorURL\": \"\"}',1706078675,0,0,3,3,'484d593da0623094ef2b917437581cad'),
	(2,'2024-01-24 06:44:45','2024-01-24 06:47:25',0,'pushgateway:9091','1','{\"endsAt\": \"\", \"labels\": {\"job\": \"pushgateway\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"pushgateway:9091\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-24T14:44:35+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 process_cpu_seconds_total\", \"description\": \"pushgateway:9091 process_cpu_seconds_total value is 215.91\"}, \"fingerprint\": \"a2bdc0f9e2dcc3503cceab0fe5a6f603\", \"generatorURL\": \"\"}',1706078675,0,0,3,3,'a2bdc0f9e2dcc3503cceab0fe5a6f603'),
	(3,'2024-01-24 06:44:45','2024-01-24 06:47:35',0,'prom:9090','1','{\"endsAt\": \"\", \"labels\": {\"job\": \"prometheus_server\", \"sverity\": \"4\", \"__name__\": \"prometheus_notifications_dropped_total\", \"instance\": \"prom:9090\", \"alertname\": \"prometheus_notifications_dropped_total\", \"__alert_id__\": \"4\", \"__group_id__\": \"1\", \"__level_id__\": \"4\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-24T14:44:35+08:00\", \"annotations\": {\"title\": \"prom:9090 prometheus_notifications_dropped_total\", \"description\": \"prom:9090 prometheus_notifications_dropped_total values is 38300\"}, \"fingerprint\": \"26a205bac05d0112f971ac278e3500ce\", \"generatorURL\": \"\"}',1706078675,0,0,4,4,'26a205bac05d0112f971ac278e3500ce'),
	(4,'2024-01-24 06:44:45','2024-01-24 06:47:45',0,'pushgateway:9091','1','{\"endsAt\": \"\", \"labels\": {\"job\": \"pushgateway\", \"sverity\": \"2\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"pushgateway:9091\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"2\", \"__group_id__\": \"2\", \"__level_id__\": \"2\", \"__group_name__\": \"alarm_2\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-24T14:44:35+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 process_cpu_seconds_total > 90\", \"description\": \"pushgateway:9091 process_cpu_seconds_total value is 1197.28\"}, \"fingerprint\": \"99f2547fda8a0416b7cd8ee7d0fe7fb1\", \"generatorURL\": \"\"}',1706078675,0,0,2,2,'99f2547fda8a0416b7cd8ee7d0fe7fb1'),
	(5,'2024-01-24 06:44:45','2024-01-24 06:47:25',0,'pushgateway:9091','1','{\"endsAt\": \"\", \"labels\": {\"job\": \"pushgateway\", \"sverity\": \"2\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"pushgateway:9091\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"2\", \"__group_id__\": \"2\", \"__level_id__\": \"2\", \"__group_name__\": \"alarm_2\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-24T14:44:35+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 process_cpu_seconds_total > 90\", \"description\": \"pushgateway:9091 process_cpu_seconds_total value is 215.91\"}, \"fingerprint\": \"b012f8497dd7856fd1d2ae29229a112a\", \"generatorURL\": \"\"}',1706078675,0,0,2,2,'b012f8497dd7856fd1d2ae29229a112a'),
	(6,'2024-01-24 06:44:45','2024-01-24 06:44:45',0,'pushgateway:9091','1','{\"endsAt\": \"\", \"labels\": {\"job\": \"pushgateway\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"instance\": \"pushgateway:9091\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-24T14:44:35+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 go_memstats_sys_bytes > 0\", \"description\": \"pushgateway:9091 go_memstats_sys_bytes value is 118677784\"}, \"fingerprint\": \"22f34c262d804041555e652f00ab7d27\", \"generatorURL\": \"\"}',1706078675,0,0,1,1,'22f34c262d804041555e652f00ab7d27'),
	(7,'2024-01-24 06:44:45','2024-01-24 06:44:45',0,'pushgateway:9091','1','{\"endsAt\": \"\", \"labels\": {\"job\": \"pushgateway\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"instance\": \"pushgateway:9091\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-24T14:44:35+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 go_memstats_sys_bytes > 0\", \"description\": \"pushgateway:9091 go_memstats_sys_bytes value is 74793992\"}, \"fingerprint\": \"0a70eaeebf8aa65ea231403fc9327ab1\", \"generatorURL\": \"\"}',1706078675,0,0,1,1,'0a70eaeebf8aa65ea231403fc9327ab1'),
	(8,'2024-01-24 06:44:45','2024-01-24 06:47:45',0,'pushgateway:9091','1','{\"endsAt\": \"\", \"labels\": {\"job\": \"pushgateway\", \"sverity\": \"1\", \"__name__\": \"go_memstats_stack_sys_bytes\", \"instance\": \"pushgateway:9091\", \"alertname\": \"go_memstats_stack_sys_bytes\", \"__alert_id__\": \"9\", \"__group_id__\": \"2\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_2\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-24T14:44:35+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 go_memstats_stack_sys_bytes > 0\", \"description\": \"pushgateway:9091 go_memstats_stack_sys_bytes value is 1081344\"}, \"fingerprint\": \"ce9d86313ca6bbea394fd07c298a1fb5\", \"generatorURL\": \"\"}',1706078675,0,0,9,1,'ce9d86313ca6bbea394fd07c298a1fb5'),
	(9,'2024-01-24 06:44:45','2024-01-24 06:47:05',0,'pushgateway:9091','1','{\"endsAt\": \"\", \"labels\": {\"job\": \"pushgateway\", \"sverity\": \"1\", \"__name__\": \"go_memstats_stack_sys_bytes\", \"instance\": \"pushgateway:9091\", \"alertname\": \"go_memstats_stack_sys_bytes\", \"__alert_id__\": \"9\", \"__group_id__\": \"2\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_2\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-24T14:44:35+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 go_memstats_stack_sys_bytes > 0\", \"description\": \"pushgateway:9091 go_memstats_stack_sys_bytes value is 557056\"}, \"fingerprint\": \"2943420f6a6b33b773169296bb14f82b\", \"generatorURL\": \"\"}',1706078675,0,0,9,1,'2943420f6a6b33b773169296bb14f82b'),
	(10,'2024-01-24 06:44:45','2024-01-24 06:47:45',0,'pushgateway:9091','1','{\"endsAt\": \"\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-24T14:44:35+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.86666666666679\"}, \"fingerprint\": \"48771e1223824c5fbb48d45c980b49da\", \"generatorURL\": \"\"}',1706078675,0,0,8,6,'48771e1223824c5fbb48d45c980b49da'),
	(11,'2024-01-24 06:44:45','2024-01-24 06:47:25',0,'pushgateway:9091','1','{\"endsAt\": \"\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-24T14:44:35+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.98000000000002\"}, \"fingerprint\": \"de6ee4754ba03237911ed704dde878f6\", \"generatorURL\": \"\"}',1706078675,0,0,8,6,'de6ee4754ba03237911ed704dde878f6'),
	(221,'2024-01-24 06:50:42','2024-01-25 12:29:31',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-25T20:29:31+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:50:32+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 process_cpu_seconds_total\", \"description\": \"pushgateway:9091 process_cpu_seconds_total value is 199.86\"}, \"fingerprint\": \"94e428a96f4e91982ec7069391768124\", \"generatorURL\": \"\"}',1706079032,1706185771,106739,3,3,'94e428a96f4e91982ec7069391768124'),
	(222,'2024-01-24 06:50:42','2024-01-25 12:29:31',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-25T20:29:31+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:50:32+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 process_cpu_seconds_total\", \"description\": \"pushgateway:9091 process_cpu_seconds_total value is 16.33\"}, \"fingerprint\": \"d54bac5492b410f074d4874270d5fe0b\", \"generatorURL\": \"\"}',1706079032,1706185771,106739,3,3,'d54bac5492b410f074d4874270d5fe0b'),
	(223,'2024-01-24 06:50:42','2024-01-24 07:51:10',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-24T15:51:09+08:00\", \"labels\": {\"job\": \"pushgateway\", \"sverity\": \"2\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"pushgateway:9091\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"2\", \"__group_id__\": \"2\", \"__level_id__\": \"2\", \"__group_name__\": \"alarm_2\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:50:32+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 process_cpu_seconds_total > 90\", \"description\": \"pushgateway:9091 process_cpu_seconds_total value is 1203.65\"}, \"fingerprint\": \"8f952bbb6c9b2941186136fa63505b26\", \"generatorURL\": \"\"}',1706079032,1706082669,3637,2,2,'8f952bbb6c9b2941186136fa63505b26'),
	(224,'2024-01-24 06:50:42','2024-01-24 07:51:10',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-24T15:51:09+08:00\", \"labels\": {\"job\": \"pushgateway\", \"sverity\": \"2\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"pushgateway:9091\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"2\", \"__group_id__\": \"2\", \"__level_id__\": \"2\", \"__group_name__\": \"alarm_2\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:50:32+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 process_cpu_seconds_total > 90\", \"description\": \"pushgateway:9091 process_cpu_seconds_total value is 216.48\"}, \"fingerprint\": \"01089a490914099397f3e5d99030979c\", \"generatorURL\": \"\"}',1706079032,1706082669,3637,2,2,'01089a490914099397f3e5d99030979c'),
	(225,'2024-01-24 06:50:42','2024-01-24 07:51:10',0,'prom:9090','2','{\"endsAt\": \"2024-01-24T15:51:09+08:00\", \"labels\": {\"job\": \"prometheus_server\", \"sverity\": \"4\", \"__name__\": \"prometheus_notifications_dropped_total\", \"instance\": \"prom:9090\", \"alertname\": \"prometheus_notifications_dropped_total\", \"__alert_id__\": \"4\", \"__group_id__\": \"1\", \"__level_id__\": \"4\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:50:32+08:00\", \"annotations\": {\"title\": \"prom:9090 prometheus_notifications_dropped_total\", \"description\": \"prom:9090 prometheus_notifications_dropped_total values is 38400\"}, \"fingerprint\": \"8549807c3d10a2791fa47a2b088c6d75\", \"generatorURL\": \"\"}',1706079032,1706082669,3637,4,4,'8549807c3d10a2791fa47a2b088c6d75'),
	(226,'2024-01-24 06:50:42','2024-01-25 12:29:31',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-25T20:29:31+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:50:32+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 go_memstats_sys_bytes > 0\", \"description\": \"pushgateway:9091 go_memstats_sys_bytes value is 130736408\"}, \"fingerprint\": \"a94797f5ae5bb656eebf65ce330eee65\", \"generatorURL\": \"\"}',1706079032,1706185771,106739,1,1,'a94797f5ae5bb656eebf65ce330eee65'),
	(227,'2024-01-24 06:50:42','2024-01-25 12:29:31',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-25T20:29:31+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:50:32+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 go_memstats_sys_bytes > 0\", \"description\": \"pushgateway:9091 go_memstats_sys_bytes value is 74793992\"}, \"fingerprint\": \"2b72098fc836f5923ff7dae2e4fd943a\", \"generatorURL\": \"\"}',1706079032,1706185771,106739,1,1,'2b72098fc836f5923ff7dae2e4fd943a'),
	(228,'2024-01-24 06:50:43','2024-01-24 06:52:32',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-24T14:52:32+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:50:32+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.84444444444408\"}, \"fingerprint\": \"40e06a2987c60cd999dd4457cdaae729\", \"generatorURL\": \"\"}',1706079032,1706079152,120,8,6,'40e06a2987c60cd999dd4457cdaae729'),
	(229,'2024-01-24 06:50:43','2024-01-24 06:52:32',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-24T14:52:32+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:50:32+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.98000000000002\"}, \"fingerprint\": \"91c02ab074dd00d3aceb9120943840d7\", \"generatorURL\": \"\"}',1706079032,1706079152,120,8,6,'91c02ab074dd00d3aceb9120943840d7'),
	(230,'2024-01-24 06:50:43','2024-01-25 12:29:31',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-25T20:29:31+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"1\", \"__name__\": \"go_memstats_stack_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"go_memstats_stack_sys_bytes\", \"__alert_id__\": \"9\", \"__group_id__\": \"2\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_2\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:50:32+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 go_memstats_stack_sys_bytes > 0\", \"description\": \"pushgateway:9091 go_memstats_stack_sys_bytes value is 1474560\"}, \"fingerprint\": \"f5e9b2b64dd9ae6e0065adc49fbda507\", \"generatorURL\": \"\"}',1706079032,1706185771,106739,9,1,'f5e9b2b64dd9ae6e0065adc49fbda507'),
	(231,'2024-01-24 06:50:43','2024-01-25 12:29:31',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-25T20:29:31+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"1\", \"__name__\": \"go_memstats_stack_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"go_memstats_stack_sys_bytes\", \"__alert_id__\": \"9\", \"__group_id__\": \"2\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_2\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:50:32+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 go_memstats_stack_sys_bytes > 0\", \"description\": \"pushgateway:9091 go_memstats_stack_sys_bytes value is 524288\"}, \"fingerprint\": \"a673c3260c3d78c7dc9a5cb1ea49a70c\", \"generatorURL\": \"\"}',1706079032,1706185771,106739,9,1,'a673c3260c3d78c7dc9a5cb1ea49a70c'),
	(427,'2024-01-24 06:53:52','2024-01-24 07:27:00',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-24T15:26:59+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:53:42+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.79999999999968\"}, \"fingerprint\": \"566298b92840febbc5857c7fdc0169a0\", \"generatorURL\": \"\"}',1706079222,1706081219,1997,8,6,'566298b92840febbc5857c7fdc0169a0'),
	(428,'2024-01-24 06:53:52','2024-01-24 07:27:00',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-24T15:26:59+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T14:53:42+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是100\"}, \"fingerprint\": \"3f1916f851ecb158d2ddbede79a73458\", \"generatorURL\": \"\"}',1706079222,1706081219,1997,8,6,'3f1916f851ecb158d2ddbede79a73458'),
	(2407,'2024-01-24 07:27:50','2024-01-24 07:51:00',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-24T15:50:59+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T15:27:39+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.79999999999968\"}, \"fingerprint\": \"1f3dd5d34c05e1560f13fb2099c4dfc5\", \"generatorURL\": \"\"}',1706081259,1706082659,1400,8,6,'1f3dd5d34c05e1560f13fb2099c4dfc5'),
	(2408,'2024-01-24 07:27:50','2024-01-24 07:51:00',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-24T15:50:59+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T15:27:39+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.98000000000002\"}, \"fingerprint\": \"0549b522743a29e26129461873ec8ac4\", \"generatorURL\": \"\"}',1706081259,1706082659,1400,8,6,'0549b522743a29e26129461873ec8ac4'),
	(3898,'2024-01-24 07:51:30','2024-01-25 12:29:41',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-25T20:29:40+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T15:51:19+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.84445481412345\"}, \"fingerprint\": \"8a8276f0c78c77d692ddd74fd316b38e\", \"generatorURL\": \"\"}',1706082679,1706185780,103101,8,6,'8a8276f0c78c77d692ddd74fd316b38e'),
	(3899,'2024-01-24 07:51:30','2024-01-25 12:29:41',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-25T20:29:40+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T15:51:19+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.98\"}, \"fingerprint\": \"64b106f45d44a1e6a8e50a1af126f871\", \"generatorURL\": \"\"}',1706082679,1706185780,103101,8,6,'64b106f45d44a1e6a8e50a1af126f871'),
	(3938,'2024-01-24 07:52:20','2024-01-25 12:29:32',0,'prom:9090','2','{\"endsAt\": \"2024-01-25T20:29:31+08:00\", \"labels\": {\"job\": \"prometheus_server\", \"sverity\": \"4\", \"__name__\": \"prometheus_notifications_dropped_total\", \"instance\": \"prom:9090\", \"alertname\": \"prometheus_notifications_dropped_total\", \"__alert_id__\": \"4\", \"__group_id__\": \"1\", \"__level_id__\": \"4\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-24T15:52:09+08:00\", \"annotations\": {\"title\": \"prom:9090 prometheus_notifications_dropped_total\", \"description\": \"prom:9090 prometheus_notifications_dropped_total values is 4122\"}, \"fingerprint\": \"deeb257f8466319ac8e8f2ef66601b24\", \"generatorURL\": \"\"}',1706082729,1706185771,103042,4,4,'deeb257f8466319ac8e8f2ef66601b24'),
	(46609,'2024-01-24 21:02:30','2024-01-25 12:29:31',0,'prom:9090','2','{\"endsAt\": \"2024-01-25T20:29:31+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"prometheus_server\", \"sverity\": \"2\", \"__name__\": \"process_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"prom:9090\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"2\", \"__group_id__\": \"2\", \"__level_id__\": \"2\", \"__group_name__\": \"alarm_2\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-25T05:02:19+08:00\", \"annotations\": {\"title\": \"prom:9090 process_cpu_seconds_total > 90\", \"description\": \"prom:9090 process_cpu_seconds_total value is 199.86\"}, \"fingerprint\": \"c092fd88a63ab44c8e1e709770f76a83\", \"generatorURL\": \"\"}',1706130139,1706185771,55632,2,2,'c092fd88a63ab44c8e1e709770f76a83'),
	(102031,'2024-01-25 12:36:21','2024-01-25 12:45:30',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-25T20:45:29+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-25T20:36:10+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 go_memstats_sys_bytes > 0\", \"description\": \"pushgateway:9091 go_memstats_sys_bytes value is 20268040\"}, \"fingerprint\": \"f6e5f142f082eb9e718236a5a2481102\", \"generatorURL\": \"\"}',1706186170,1706186729,559,1,1,'f6e5f142f082eb9e718236a5a2481102'),
	(102032,'2024-01-25 12:36:21','2024-01-25 12:45:30',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-25T20:45:29+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-25T20:36:10+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 go_memstats_sys_bytes > 0\", \"description\": \"pushgateway:9091 go_memstats_sys_bytes value is 130736408\"}, \"fingerprint\": \"dd8e7a03f583947e4d1e042a3ffdf5f1\", \"generatorURL\": \"\"}',1706186170,1706186729,559,1,1,'dd8e7a03f583947e4d1e042a3ffdf5f1'),
	(102033,'2024-01-25 12:36:21','2024-01-25 12:45:30',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-25T20:45:29+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-25T20:36:10+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 go_memstats_sys_bytes > 0\", \"description\": \"pushgateway:9091 go_memstats_sys_bytes value is 74793992\"}, \"fingerprint\": \"b243d2e4092bae522fdb94beaafc1abc\", \"generatorURL\": \"\"}',1706186170,1706186729,559,1,1,'b243d2e4092bae522fdb94beaafc1abc'),
	(102196,'2024-01-25 12:47:50','2024-01-27 12:35:59',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-27T20:35:58+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-25T20:47:39+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20268040\"}, \"fingerprint\": \"6f005b148ef8158dcd7bc8e881126770\", \"generatorURL\": \"\"}',1706186859,1706358958,172099,1,1,'6f005b148ef8158dcd7bc8e881126770'),
	(119392,'2024-01-28 06:14:22','2024-01-30 08:05:21',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-30T16:05:21+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-28T14:14:12+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"23268fc7a4aa33d7c129966e320c1b22\", \"generatorURL\": \"\"}',1706422452,1706601921,179469,1,1,'23268fc7a4aa33d7c129966e320c1b22'),
	(128783,'2024-01-29 08:19:53','2024-01-31 09:58:18',0,'node_exporter:9100','2','{\"endsAt\": \"\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-29T16:19:42+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"42589f3dd70ee2e35936572400a36c14\", \"generatorURL\": \"\"}',1706516382,0,0,1,1,'42589f3dd70ee2e35936572400a36c14'),
	(128919,'2024-01-29 08:42:50','2024-01-31 09:58:15',0,'node_exporter:9100','2','{\"endsAt\": \"\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-29T16:42:39+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"9c57b10dbcf6c90a852c61004ca715b8\", \"generatorURL\": \"\"}',1706517759,0,0,1,1,'9c57b10dbcf6c90a852c61004ca715b8'),
	(128976,'2024-01-29 08:52:41','2024-01-31 09:58:12',0,'node_exporter:9100','2','{\"endsAt\": \"\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-29T16:52:30+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"8a76afd69533c8de573514e31d2528e6\", \"generatorURL\": \"\"}',1706518350,0,0,1,1,'8a76afd69533c8de573514e31d2528e6'),
	(129023,'2024-01-29 09:00:57','2024-01-29 09:09:00',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-29T17:08:59+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-29T17:00:46+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"bc54f96ae5cb81e0998249d3a331ff2e\", \"generatorURL\": \"\"}',1706518846,1706519339,493,1,1,'bc54f96ae5cb81e0998249d3a331ff2e'),
	(129072,'2024-01-30 08:05:51','2024-01-30 08:06:11',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-30T16:06:11+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-30T16:05:41+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"65f20db607f2a474368320825e73b155\", \"generatorURL\": \"\"}',1706601941,1706601971,30,1,1,'65f20db607f2a474368320825e73b155'),
	(129075,'2024-01-30 08:53:01','2024-01-30 08:53:41',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-30T16:53:41+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-30T16:52:51+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"9daa7c89cd6432e98a69a3a3be80b5f5\", \"generatorURL\": \"\"}',1706604771,1706604821,50,1,1,'9daa7c89cd6432e98a69a3a3be80b5f5'),
	(129080,'2024-01-30 14:29:51','2024-01-30 14:30:11',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-30T22:30:11+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-30T22:29:41+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"7b4e86c724b1377c5abdeb83a329be9f\", \"generatorURL\": \"\"}',1706624981,1706625011,30,1,1,'7b4e86c724b1377c5abdeb83a329be9f'),
	(129083,'2024-01-30 14:30:51','2024-01-30 14:32:11',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-30T22:32:11+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-30T22:30:41+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"1d63242ff913b95d7a89ed48fb0a81e8\", \"generatorURL\": \"\"}',1706625041,1706625131,90,1,1,'1d63242ff913b95d7a89ed48fb0a81e8'),
	(129092,'2024-01-31 01:47:35','2024-01-31 01:57:05',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-31T09:57:05+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T09:44:35+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"4c4dd59128c55281167742b7f69ebbca\", \"generatorURL\": \"\"}',1706665475,1706666225,750,1,1,'4c4dd59128c55281167742b7f69ebbca'),
	(129150,'2024-01-31 10:01:55','2024-01-31 10:28:05',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-31T18:28:05+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T17:58:55+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"2eb82b659b15fee0d11b2bb9f6f81bee\", \"generatorURL\": \"\"}',1706695135,1706696885,1750,1,1,'2eb82b659b15fee0d11b2bb9f6f81bee'),
	(129308,'2024-01-31 12:22:45','2024-01-31 12:25:15',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-31T20:25:15+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:22:35+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.86666666666679\"}, \"fingerprint\": \"36c6ac4d34a332157ee6f434ac6e024f\", \"generatorURL\": \"\"}',1706703755,1706703915,160,8,6,'36c6ac4d34a332157ee6f434ac6e024f'),
	(129309,'2024-01-31 12:22:45','2024-01-31 12:25:15',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-31T20:25:15+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:22:35+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.86666666666653\"}, \"fingerprint\": \"b0cabc9b6e9dcd4ef3c9128b98d5e5c3\", \"generatorURL\": \"\"}',1706703755,1706703915,160,8,6,'b0cabc9b6e9dcd4ef3c9128b98d5e5c3'),
	(129310,'2024-01-31 12:22:45','2024-01-31 12:25:15',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-31T20:25:15+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:22:35+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是100\"}, \"fingerprint\": \"f7a219163cceb07532fc0c05bf36dc4d\", \"generatorURL\": \"\"}',1706703755,1706703915,160,8,6,'f7a219163cceb07532fc0c05bf36dc4d'),
	(129311,'2024-01-31 12:22:45','2024-01-31 12:25:25',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-31T20:25:25+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:22:35+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 process_cpu_seconds_total\", \"description\": \"pushgateway:9091 process_cpu_seconds_total value is 654.3\"}, \"fingerprint\": \"04a1c12f7331b9fda909260aea30fd82\", \"generatorURL\": \"\"}',1706703755,1706703925,170,3,3,'04a1c12f7331b9fda909260aea30fd82'),
	(129312,'2024-01-31 12:22:45','2024-01-31 12:25:25',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-31T20:25:25+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:22:35+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 process_cpu_seconds_total\", \"description\": \"pushgateway:9091 process_cpu_seconds_total value is 757.96\"}, \"fingerprint\": \"46aa20667d6392f2ca81c3d56cecbe72\", \"generatorURL\": \"\"}',1706703755,1706703925,170,3,3,'46aa20667d6392f2ca81c3d56cecbe72'),
	(129313,'2024-01-31 12:22:45','2024-01-31 12:25:25',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-31T20:25:25+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"pushgateway\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"pushgateway:9091\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:22:35+08:00\", \"annotations\": {\"title\": \"pushgateway:9091 process_cpu_seconds_total\", \"description\": \"pushgateway:9091 process_cpu_seconds_total value is 100.39\"}, \"fingerprint\": \"fbc05e7225d383ba8f9c40919f900e8d\", \"generatorURL\": \"\"}',1706703755,1706703925,170,3,3,'fbc05e7225d383ba8f9c40919f900e8d'),
	(129419,'2024-01-31 12:25:55','2024-01-31 12:29:16',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-31T20:29:15+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:22:35+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"1e55655d439410f40c09b5632fdfa527\", \"generatorURL\": \"\"}',1706703755,1706704155,400,1,1,'1e55655d439410f40c09b5632fdfa527'),
	(129429,'2024-01-31 12:27:35','2024-01-31 12:29:15',0,'prom:9090','2','{\"endsAt\": \"2024-01-31T20:29:15+08:00\", \"labels\": {\"job\": \"prometheus_server\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"prom:9090\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:27:25+08:00\", \"annotations\": {\"title\": \"prom:9090 process_cpu_seconds_total\", \"description\": \"prom:9090 process_cpu_seconds_total value is 758.25\"}, \"fingerprint\": \"956c35690aa65652493d0f8f6dda9ee0\", \"generatorURL\": \"\"}',1706704045,1706704155,110,3,3,'956c35690aa65652493d0f8f6dda9ee0'),
	(129451,'2024-01-31 12:31:35','2024-01-31 12:37:45',0,'prom:9090','2','{\"endsAt\": \"2024-01-31T20:37:45+08:00\", \"labels\": {\"job\": \"prometheus_server\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"prom:9090\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:31:25+08:00\", \"annotations\": {\"title\": \"prom:9090 process_cpu_seconds_total\", \"description\": \"prom:9090 process_cpu_seconds_total value is 758.9\"}, \"fingerprint\": \"1243798afd8f2047dc49ffb56a4db409\", \"generatorURL\": \"\"}',1706704285,1706704665,380,3,3,'1243798afd8f2047dc49ffb56a4db409'),
	(129474,'2024-01-31 12:35:25','2024-01-31 12:37:46',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-31T20:37:45+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:32:25+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"249696d28281936d4b864f93b5209493\", \"generatorURL\": \"\"}',1706704345,1706704665,320,1,1,'249696d28281936d4b864f93b5209493'),
	(129504,'2024-01-31 12:52:18','2024-01-31 12:53:18',0,'prom:9090','2','{\"endsAt\": \"2024-01-31T20:53:18+08:00\", \"labels\": {\"job\": \"prometheus_server\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"prom:9090\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:52:08+08:00\", \"annotations\": {\"title\": \"prom:9090 process_cpu_seconds_total\", \"description\": \"prom:9090 process_cpu_seconds_total value is 759.82\"}, \"fingerprint\": \"3a82d44f118794270b346e021a6b99d8\", \"generatorURL\": \"\"}',1706705528,1706705598,70,3,3,'3a82d44f118794270b346e021a6b99d8'),
	(129511,'2024-01-31 12:53:38','2024-01-31 12:57:08',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-31T20:57:08+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:53:28+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.888888888889\"}, \"fingerprint\": \"32a0139b8bacad63d96382b9ce1c687a\", \"generatorURL\": \"\"}',1706705608,1706705828,220,8,6,'32a0139b8bacad63d96382b9ce1c687a'),
	(129512,'2024-01-31 12:53:38','2024-01-31 12:57:08',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-31T20:57:08+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:53:28+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.84444444444459\"}, \"fingerprint\": \"6ff3af08d2a839afb42ef427461eaa0e\", \"generatorURL\": \"\"}',1706705608,1706705828,220,8,6,'6ff3af08d2a839afb42ef427461eaa0e'),
	(129513,'2024-01-31 12:53:38','2024-01-31 12:57:08',0,'pushgateway:9091','2','{\"endsAt\": \"2024-01-31T20:57:08+08:00\", \"labels\": {\"sverity\": \"1\", \"instance\": \"pushgateway:9091\", \"alertname\": \"cpu_90\", \"__alert_id__\": \"8\", \"__group_id__\": \"3\", \"__level_id__\": \"6\", \"__group_name__\": \"system_alarm\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:53:28+08:00\", \"annotations\": {\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是99.98000000000002\"}, \"fingerprint\": \"dd3a9284a6258c30a636fec7cf012fbe\", \"generatorURL\": \"\"}',1706705608,1706705828,220,8,6,'dd3a9284a6258c30a636fec7cf012fbe'),
	(129517,'2024-01-31 12:53:48','2024-01-31 12:57:10',0,'prom:9090','2','{\"endsAt\": \"2024-01-31T20:57:08+08:00\", \"labels\": {\"job\": \"prometheus_server\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"prom:9090\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:53:38+08:00\", \"annotations\": {\"title\": \"prom:9090 process_cpu_seconds_total\", \"description\": \"prom:9090 process_cpu_seconds_total value is 760.15\"}, \"fingerprint\": \"8183c5d170ea14327314c8f258498800\", \"generatorURL\": \"\"}',1706705618,1706705828,210,3,3,'8183c5d170ea14327314c8f258498800'),
	(129582,'2024-01-31 12:56:38','2024-01-31 12:57:10',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-31T20:57:08+08:00\", \"labels\": {\"env\": \"prod\", \"job\": \"node_exporter_service\", \"sverity\": \"1\", \"__name__\": \"go_memstats_sys_bytes\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"go_memstats_sys_bytes\", \"__alert_id__\": \"1\", \"__group_id__\": \"1\", \"__level_id__\": \"1\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T20:53:38+08:00\", \"annotations\": {\"title\": \"node_exporter:9100 go_memstats_sys_bytes > 0\", \"description\": \"node_exporter:9100 go_memstats_sys_bytes value is 20530184\"}, \"fingerprint\": \"c90f60342424b41c4bb7f5e6c4e3b920\", \"generatorURL\": \"\"}',1706705618,1706705828,210,1,1,'c90f60342424b41c4bb7f5e6c4e3b920'),
	(129608,'2024-01-31 13:12:38','2024-01-31 13:13:08',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-31T21:13:08+08:00\", \"labels\": {\"cpu\": \"1\", \"env\": \"prod\", \"job\": \"node_exporter_service\", \"xx1\": \"xxx\", \"mode\": \"idle\", \"sverity\": \"1\", \"__name__\": \"node_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"node_cpu_seconds_total\", \"__alert_id__\": \"10\", \"__group_id__\": \"1\", \"__level_id__\": \"5\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T21:12:28+08:00\", \"annotations\": {\"title\": \"node_cpu_seconds_total instance is node_exporter:9100\", \"description\": \"node_cpu_seconds_total value is 2111990.33\"}, \"fingerprint\": \"ef61f2e184f0ae393df0dc267e6752bf\", \"generatorURL\": \"\"}',1706706748,1706706788,40,10,5,'ef61f2e184f0ae393df0dc267e6752bf'),
	(129609,'2024-01-31 13:12:38','2024-01-31 13:13:08',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-31T21:13:08+08:00\", \"labels\": {\"cpu\": \"1\", \"env\": \"prod\", \"job\": \"node_exporter_service\", \"xx1\": \"xxx\", \"mode\": \"idle\", \"sverity\": \"1\", \"__name__\": \"node_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"node_cpu_seconds_total\", \"__alert_id__\": \"10\", \"__group_id__\": \"1\", \"__level_id__\": \"5\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T21:12:28+08:00\", \"annotations\": {\"title\": \"node_cpu_seconds_total instance is node_exporter:9100\", \"description\": \"node_cpu_seconds_total value is 2110655.48\"}, \"fingerprint\": \"5cc01564eb2c4c349de8a23024e15bee\", \"generatorURL\": \"\"}',1706706748,1706706788,40,10,5,'5cc01564eb2c4c349de8a23024e15bee'),
	(129618,'2024-01-31 13:14:18','2024-01-31 13:15:08',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-31T21:15:08+08:00\", \"labels\": {\"cpu\": \"1\", \"env\": \"prod\", \"job\": \"node_exporter_service\", \"xx1\": \"xxx\", \"mode\": \"idle\", \"sverity\": \"1\", \"__name__\": \"node_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"node_cpu_seconds_total\", \"__alert_id__\": \"10\", \"__group_id__\": \"1\", \"__level_id__\": \"5\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T21:14:08+08:00\", \"annotations\": {\"title\": \"node_cpu_seconds_total instance is node_exporter:9100\", \"description\": \"node_cpu_seconds_total value is 2112095.76\"}, \"fingerprint\": \"c3f1501c17c3373d5d0919bdeda5e7f4\", \"generatorURL\": \"\"}',1706706848,1706706908,60,10,5,'c3f1501c17c3373d5d0919bdeda5e7f4'),
	(129619,'2024-01-31 13:14:18','2024-01-31 13:15:08',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-31T21:15:08+08:00\", \"labels\": {\"cpu\": \"1\", \"env\": \"prod\", \"job\": \"node_exporter_service\", \"xx1\": \"xxx\", \"mode\": \"idle\", \"sverity\": \"1\", \"__name__\": \"node_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"node_cpu_seconds_total\", \"__alert_id__\": \"10\", \"__group_id__\": \"1\", \"__level_id__\": \"5\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T21:14:08+08:00\", \"annotations\": {\"title\": \"node_cpu_seconds_total instance is node_exporter:9100\", \"description\": \"node_cpu_seconds_total value is 2110760.89\"}, \"fingerprint\": \"9c705fc646039870338d2e74eb62ba03\", \"generatorURL\": \"\"}',1706706848,1706706908,60,10,5,'9c705fc646039870338d2e74eb62ba03'),
	(129632,'2024-01-31 13:15:58','2024-01-31 13:18:19',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-31T21:18:18+08:00\", \"labels\": {\"cpu\": \"1\", \"env\": \"prod\", \"job\": \"node_exporter_service\", \"xx1\": \"xxx\", \"mode\": \"idle\", \"sverity\": \"1\", \"__name__\": \"node_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"node_cpu_seconds_total\", \"__alert_id__\": \"10\", \"__group_id__\": \"1\", \"__level_id__\": \"5\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T21:15:48+08:00\", \"annotations\": {\"title\": \"node_cpu_seconds_total instance is node_exporter:9100\", \"description\": \"node_cpu_seconds_total value is 2112263.19\"}, \"fingerprint\": \"0fa8368c3cba0227a026476d4203c690\", \"generatorURL\": \"\"}',1706706948,1706707098,150,10,5,'0fa8368c3cba0227a026476d4203c690'),
	(129633,'2024-01-31 13:15:58','2024-01-31 13:18:19',0,'node_exporter:9100','2','{\"endsAt\": \"2024-01-31T21:18:18+08:00\", \"labels\": {\"cpu\": \"1\", \"env\": \"prod\", \"job\": \"node_exporter_service\", \"xx1\": \"xxx\", \"mode\": \"idle\", \"sverity\": \"1\", \"__name__\": \"node_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"node_cpu_seconds_total\", \"__alert_id__\": \"10\", \"__group_id__\": \"1\", \"__level_id__\": \"5\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T21:15:48+08:00\", \"annotations\": {\"title\": \"node_cpu_seconds_total instance is node_exporter:9100\", \"description\": \"node_cpu_seconds_total value is 2110927.97\"}, \"fingerprint\": \"ae3d32ca5bdf7b160c5e076b3c3e17c7\", \"generatorURL\": \"\"}',1706706948,1706707098,150,10,5,'ae3d32ca5bdf7b160c5e076b3c3e17c7'),
	(129642,'2024-01-31 13:16:48','2024-01-31 13:18:18',0,'prom:9090','2','{\"endsAt\": \"2024-01-31T21:18:18+08:00\", \"labels\": {\"job\": \"prometheus_server\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"prom:9090\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T21:16:38+08:00\", \"annotations\": {\"title\": \"prom:9090 process_cpu_seconds_total\", \"description\": \"prom:9090 process_cpu_seconds_total value is 761.67\"}, \"fingerprint\": \"5cd1af047eef95e116bf70b301d8096f\", \"generatorURL\": \"\"}',1706706998,1706707098,100,3,3,'5cd1af047eef95e116bf70b301d8096f'),
	(129674,'2024-01-31 13:19:18','2024-01-31 14:02:28',0,'node_exporter:9100','1','{\"endsAt\": \"\", \"labels\": {\"cpu\": \"1\", \"env\": \"prod\", \"job\": \"node_exporter_service\", \"xx1\": \"xxx\", \"mode\": \"idle\", \"sverity\": \"1\", \"__name__\": \"node_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"node_cpu_seconds_total\", \"__alert_id__\": \"10\", \"__group_id__\": \"1\", \"__level_id__\": \"5\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-31T21:19:08+08:00\", \"annotations\": {\"title\": \"node_cpu_seconds_total instance is node_exporter:9100\", \"description\": \"node_cpu_seconds_total value is 2114553.29\"}, \"fingerprint\": \"fed734ea22d74d2a6fd0a7672bfa6713\", \"generatorURL\": \"\"}',1706707148,0,0,10,5,'fed734ea22d74d2a6fd0a7672bfa6713'),
	(129675,'2024-01-31 13:19:18','2024-01-31 14:02:28',0,'node_exporter:9100','1','{\"endsAt\": \"\", \"labels\": {\"cpu\": \"1\", \"env\": \"prod\", \"job\": \"node_exporter_service\", \"xx1\": \"xxx\", \"mode\": \"idle\", \"sverity\": \"1\", \"__name__\": \"node_cpu_seconds_total\", \"endpoint\": \"hello-world-1-2-3\", \"instance\": \"node_exporter:9100\", \"alertname\": \"node_cpu_seconds_total\", \"__alert_id__\": \"10\", \"__group_id__\": \"1\", \"__level_id__\": \"5\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"firing\", \"startsAt\": \"2024-01-31T21:19:08+08:00\", \"annotations\": {\"title\": \"node_cpu_seconds_total instance is node_exporter:9100\", \"description\": \"node_cpu_seconds_total value is 2113216.76\"}, \"fingerprint\": \"e2903dc3868eb1191b7a910eca2e8b59\", \"generatorURL\": \"\"}',1706707148,0,0,10,5,'e2903dc3868eb1191b7a910eca2e8b59'),
	(129676,'2024-01-31 13:19:18','2024-01-31 13:20:58',0,'prom:9090','2','{\"endsAt\": \"2024-01-31T21:20:58+08:00\", \"labels\": {\"job\": \"prometheus_server\", \"sverity\": \"3\", \"__name__\": \"process_cpu_seconds_total\", \"instance\": \"prom:9090\", \"alertname\": \"process_cpu_seconds_total\", \"__alert_id__\": \"3\", \"__group_id__\": \"1\", \"__level_id__\": \"3\", \"__group_name__\": \"alarm_test_group\"}, \"status\": \"resolved\", \"startsAt\": \"2024-01-31T21:19:08+08:00\", \"annotations\": {\"title\": \"prom:9090 process_cpu_seconds_total\", \"description\": \"prom:9090 process_cpu_seconds_total value is 761.87\"}, \"fingerprint\": \"326f0ae42c8fa1201d8a3a71779efef9\", \"generatorURL\": \"\"}',1706707148,1706707258,110,3,3,'326f0ae42c8fa1201d8a3a71779efef9');

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
  `create_by` int unsigned NOT NULL DEFAULT '0' COMMENT '创建人ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_alarm_notifies` WRITE;
/*!40000 ALTER TABLE `prom_alarm_notifies` DISABLE KEYS */;

INSERT INTO `prom_alarm_notifies` (`id`, `created_at`, `updated_at`, `deleted_at`, `name`, `status`, `remark`, `create_by`)
VALUES
	(1,'2024-01-24 10:15:35','2024-01-31 20:51:33',0,'第一个报警组',1,'用于测试监控告警通知',1);

/*!40000 ALTER TABLE `prom_alarm_notifies` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_alarm_notify_external_notify_objs
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_notify_external_notify_objs`;

CREATE TABLE `prom_alarm_notify_external_notify_objs` (
  `prom_alarm_notify_id` int unsigned NOT NULL,
  `external_notify_obj_id` int unsigned NOT NULL,
  PRIMARY KEY (`prom_alarm_notify_id`,`external_notify_obj_id`)
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
) ENGINE=InnoDB AUTO_INCREMENT=41 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_alarm_pages` WRITE;
/*!40000 ALTER TABLE `prom_alarm_pages` DISABLE KEYS */;

INSERT INTO `prom_alarm_pages` (`id`, `created_at`, `updated_at`, `deleted_at`, `name`, `remark`, `icon`, `color`, `status`)
VALUES
	(1,'2023-12-29 15:32:58','2024-01-26 21:23:46',0,'实时告警','xsfasf','icon-shishijiankong1','#ef0a0a',1),
	(2,'2023-12-29 16:03:15','2023-12-30 01:11:43',0,'测试页面','','icon-ceshi','green',1),
	(3,'2023-12-29 16:07:50','2024-01-27 20:39:20',0,'值班监控','faswr21re1','icon-zhibanjiankong','#b814ec',1),
	(4,'2023-12-29 17:21:27','2024-01-26 17:34:30',0,'监控-1','没有说明信息','','#f1cd17',1),
	(5,'2023-12-29 17:21:27','2024-01-26 17:34:41',0,'监控-2','没有说明信息','','#e16161',1),
	(6,'2023-12-29 17:22:44','2024-01-26 17:34:51',0,'监控-3','没有说明信息','','#d215ea',1),
	(7,'2023-12-29 17:22:44','2024-01-26 17:35:01',0,'监控-4','没有说明信息','','#5317df',1),
	(8,'2023-12-29 17:22:44','2023-12-29 17:22:44',0,'监控-5','没有说明信息','','',1),
	(9,'2023-12-29 17:22:44','2023-12-29 17:22:44',0,'监控-6','没有说明信息','','',1),
	(10,'2023-12-29 17:22:44','2023-12-29 17:22:44',0,'监控-7','没有说明信息','','',1),
	(11,'2023-12-29 17:22:44','2023-12-29 17:22:44',0,'监控-8','没有说明信息','','',1),
	(12,'2023-12-29 17:22:44','2023-12-29 17:22:44',0,'监控-9','没有说明信息','','',1),
	(27,'2023-12-29 17:55:17','2023-12-29 17:55:17',0,'监控-10','没有说明信息','','',1);

/*!40000 ALTER TABLE `prom_alarm_pages` ENABLE KEYS */;
UNLOCK TABLES;


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
) ENGINE=InnoDB AUTO_INCREMENT=130205 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_alarm_realtime` WRITE;
/*!40000 ALTER TABLE `prom_alarm_realtime` DISABLE KEYS */;

INSERT INTO `prom_alarm_realtime` (`id`, `created_at`, `updated_at`, `deleted_at`, `strategy_id`, `instance`, `note`, `status`, `event_at`, `notified_at`, `history_id`, `level_id`)
VALUES
	(221,'2024-01-24 06:50:42','2024-01-25 12:29:31',0,3,'pushgateway:9091','title: pushgateway:9091 process_cpu_seconds_total;\ndescription: pushgateway:9091 process_cpu_seconds_total value is 199.86;\n',2,1706079032,0,221,3),
	(222,'2024-01-24 06:50:42','2024-01-25 12:29:31',0,3,'pushgateway:9091','title: pushgateway:9091 process_cpu_seconds_total;\ndescription: pushgateway:9091 process_cpu_seconds_total value is 16.33;\n',2,1706079032,0,222,3),
	(223,'2024-01-24 06:50:42','2024-01-24 07:51:10',0,2,'pushgateway:9091','title: pushgateway:9091 process_cpu_seconds_total > 90;\ndescription: pushgateway:9091 process_cpu_seconds_total value is 216.48;\n',2,1706079032,0,224,2),
	(224,'2024-01-24 06:50:42','2024-01-24 07:51:10',0,2,'pushgateway:9091','title: pushgateway:9091 process_cpu_seconds_total > 90;\ndescription: pushgateway:9091 process_cpu_seconds_total value is 1203.65;\n',2,1706079032,0,223,2),
	(225,'2024-01-24 06:50:42','2024-01-24 07:51:10',0,4,'prom:9090','title: prom:9090 prometheus_notifications_dropped_total;\ndescription: prom:9090 prometheus_notifications_dropped_total values is 38400;\n',2,1706079032,0,225,4),
	(226,'2024-01-24 06:50:43','2024-01-25 12:29:31',0,1,'pushgateway:9091','title: pushgateway:9091 go_memstats_sys_bytes > 0;\ndescription: pushgateway:9091 go_memstats_sys_bytes value is 74793992;\n',2,1706079032,0,227,1),
	(227,'2024-01-24 06:50:43','2024-01-25 12:29:31',0,1,'pushgateway:9091','title: pushgateway:9091 go_memstats_sys_bytes > 0;\ndescription: pushgateway:9091 go_memstats_sys_bytes value is 130736408;\n',2,1706079032,0,226,1),
	(228,'2024-01-24 06:50:43','2024-01-24 06:52:32',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是99.84444444444408;\n',2,1706079032,0,228,6),
	(229,'2024-01-24 06:50:43','2024-01-24 06:52:32',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是99.98000000000002;\n',2,1706079032,0,229,6),
	(230,'2024-01-24 06:50:43','2024-01-25 12:29:31',0,9,'pushgateway:9091','title: pushgateway:9091 go_memstats_stack_sys_bytes > 0;\ndescription: pushgateway:9091 go_memstats_stack_sys_bytes value is 524288;\n',2,1706079032,0,231,1),
	(231,'2024-01-24 06:50:43','2024-01-25 12:29:31',0,9,'pushgateway:9091','description: pushgateway:9091 go_memstats_stack_sys_bytes value is 1474560;\ntitle: pushgateway:9091 go_memstats_stack_sys_bytes > 0;\n',2,1706079032,0,230,1),
	(427,'2024-01-24 06:53:53','2024-01-24 07:27:00',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是100;\n',2,1706079222,0,428,6),
	(428,'2024-01-24 06:53:53','2024-01-24 07:27:00',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是99.79999999999968;\n',2,1706079222,0,427,6),
	(2407,'2024-01-24 07:27:50','2024-01-24 07:51:00',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是99.98000000000002;\n',2,1706081259,0,2408,6),
	(2408,'2024-01-24 07:27:50','2024-01-24 07:51:00',0,8,'pushgateway:9091','description: 持续1min的内存使用率大于90%, 当前值是99.79999999999968;\ntitle: 持续1min的内存使用率大于90%;\n',2,1706081259,0,2407,6),
	(3898,'2024-01-24 07:51:30','2024-01-25 12:29:41',0,8,'pushgateway:9091','description: 持续1min的内存使用率大于90%, 当前值是99.98;\ntitle: 持续1min的内存使用率大于90%;\n',2,1706082679,0,3899,6),
	(3899,'2024-01-24 07:51:30','2024-01-25 12:29:41',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是99.84445481412345;\n',2,1706082679,0,3898,6),
	(3938,'2024-01-24 07:52:20','2024-01-25 12:29:32',0,4,'prom:9090','title: prom:9090 prometheus_notifications_dropped_total;\ndescription: prom:9090 prometheus_notifications_dropped_total values is 4122;\n',2,1706082729,0,3938,4),
	(46609,'2024-01-24 21:02:30','2024-01-25 12:29:31',0,2,'prom:9090','title: prom:9090 process_cpu_seconds_total > 90;\ndescription: prom:9090 process_cpu_seconds_total value is 199.86;\n',2,1706130139,0,46609,2),
	(102031,'2024-01-25 12:36:21','2024-01-25 12:45:30',0,1,'pushgateway:9091','title: pushgateway:9091 go_memstats_sys_bytes > 0;\ndescription: pushgateway:9091 go_memstats_sys_bytes value is 74793992;\n',2,1706186170,0,102033,1),
	(102032,'2024-01-25 12:36:21','2024-01-25 12:45:30',0,1,'pushgateway:9091','title: pushgateway:9091 go_memstats_sys_bytes > 0;\ndescription: pushgateway:9091 go_memstats_sys_bytes value is 130736408;\n',2,1706186170,0,102032,1),
	(102033,'2024-01-25 12:36:21','2024-01-25 12:45:30',0,1,'pushgateway:9091','title: pushgateway:9091 go_memstats_sys_bytes > 0;\ndescription: pushgateway:9091 go_memstats_sys_bytes value is 20268040;\n',2,1706186170,0,102031,1),
	(102196,'2024-01-25 12:47:50','2024-01-27 12:35:59',0,1,'node_exporter:9100','title: node_exporter:9100 go_memstats_sys_bytes > 0;\ndescription: node_exporter:9100 go_memstats_sys_bytes value is 20268040;\n',2,1706186859,0,102196,1),
	(119392,'2024-01-28 06:14:22','2024-01-30 08:05:21',0,1,'node_exporter:9100','description: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\ntitle: node_exporter:9100 go_memstats_sys_bytes > 0;\n',2,1706422452,0,119392,1),
	(128783,'2024-01-29 08:19:53','2024-01-31 09:57:44',0,1,'node_exporter:9100','title: node_exporter:9100 go_memstats_sys_bytes > 0;\ndescription: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\n',2,1706516382,0,128783,1),
	(128919,'2024-01-29 08:42:50','2024-01-31 09:57:48',0,1,'node_exporter:9100','title: node_exporter:9100 go_memstats_sys_bytes > 0;\ndescription: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\n',2,1706517759,0,128919,1),
	(128976,'2024-01-29 08:52:41','2024-01-31 09:57:52',0,1,'node_exporter:9100','title: node_exporter:9100 go_memstats_sys_bytes > 0;\ndescription: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\n',2,1706518350,0,128976,1),
	(129023,'2024-01-29 09:00:57','2024-01-29 09:09:00',0,1,'node_exporter:9100','description: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\ntitle: node_exporter:9100 go_memstats_sys_bytes > 0;\n',2,1706518846,0,129023,1),
	(129072,'2024-01-30 08:05:51','2024-01-30 08:06:11',0,1,'node_exporter:9100','title: node_exporter:9100 go_memstats_sys_bytes > 0;\ndescription: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\n',2,1706601941,0,129072,1),
	(129075,'2024-01-30 08:53:01','2024-01-30 08:53:41',0,1,'node_exporter:9100','title: node_exporter:9100 go_memstats_sys_bytes > 0;\ndescription: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\n',2,1706604771,0,129075,1),
	(129080,'2024-01-30 14:29:51','2024-01-30 14:30:11',0,1,'node_exporter:9100','description: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\ntitle: node_exporter:9100 go_memstats_sys_bytes > 0;\n',2,1706624981,0,129080,1),
	(129083,'2024-01-30 14:30:51','2024-01-30 14:32:11',0,1,'node_exporter:9100','title: node_exporter:9100 go_memstats_sys_bytes > 0;\ndescription: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\n',2,1706625041,0,129083,1),
	(129092,'2024-01-31 01:47:35','2024-01-31 01:57:05',0,1,'node_exporter:9100','title: node_exporter:9100 go_memstats_sys_bytes > 0;\ndescription: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\n',2,1706665475,0,129092,1),
	(129150,'2024-01-31 10:01:55','2024-01-31 10:28:05',0,1,'node_exporter:9100','description: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\ntitle: node_exporter:9100 go_memstats_sys_bytes > 0;\n',2,1706695135,0,129150,1),
	(129308,'2024-01-31 12:22:45','2024-01-31 12:25:15',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是99.86666666666679;\n',2,1706703755,0,129308,6),
	(129309,'2024-01-31 12:22:45','2024-01-31 12:25:15',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是99.86666666666653;\n',2,1706703755,0,129309,6),
	(129310,'2024-01-31 12:22:45','2024-01-31 12:25:15',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是100;\n',2,1706703755,0,129310,6),
	(129311,'2024-01-31 12:22:45','2024-01-31 12:25:27',0,3,'pushgateway:9091','description: pushgateway:9091 process_cpu_seconds_total value is 654.3;\ntitle: pushgateway:9091 process_cpu_seconds_total;\n',2,1706703755,0,129311,3),
	(129312,'2024-01-31 12:22:45','2024-01-31 12:25:27',0,3,'pushgateway:9091','title: pushgateway:9091 process_cpu_seconds_total;\ndescription: pushgateway:9091 process_cpu_seconds_total value is 757.96;\n',2,1706703755,0,129312,3),
	(129313,'2024-01-31 12:22:45','2024-01-31 12:25:27',0,3,'pushgateway:9091','description: pushgateway:9091 process_cpu_seconds_total value is 100.39;\ntitle: pushgateway:9091 process_cpu_seconds_total;\n',2,1706703755,0,129313,3),
	(129419,'2024-01-31 12:25:55','2024-01-31 12:29:16',0,1,'node_exporter:9100','title: node_exporter:9100 go_memstats_sys_bytes > 0;\ndescription: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\n',2,1706703755,0,129419,1),
	(129429,'2024-01-31 12:27:35','2024-01-31 12:29:15',0,3,'prom:9090','title: prom:9090 process_cpu_seconds_total;\ndescription: prom:9090 process_cpu_seconds_total value is 758.25;\n',2,1706704045,0,129429,3),
	(129451,'2024-01-31 12:31:35','2024-01-31 12:37:45',0,3,'prom:9090','title: prom:9090 process_cpu_seconds_total;\ndescription: prom:9090 process_cpu_seconds_total value is 758.9;\n',2,1706704285,0,129451,3),
	(129474,'2024-01-31 12:35:25','2024-01-31 12:37:46',0,1,'node_exporter:9100','title: node_exporter:9100 go_memstats_sys_bytes > 0;\ndescription: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\n',2,1706704345,0,129474,1),
	(129504,'2024-01-31 12:52:18','2024-01-31 12:53:18',0,3,'prom:9090','title: prom:9090 process_cpu_seconds_total;\ndescription: prom:9090 process_cpu_seconds_total value is 759.82;\n',2,1706705528,0,129504,3),
	(129511,'2024-01-31 12:53:38','2024-01-31 12:57:08',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是99.888888888889;\n',2,1706705608,0,129511,6),
	(129512,'2024-01-31 12:53:38','2024-01-31 12:57:08',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是99.84444444444459;\n',2,1706705608,0,129512,6),
	(129513,'2024-01-31 12:53:38','2024-01-31 12:57:08',0,8,'pushgateway:9091','title: 持续1min的内存使用率大于90%;\ndescription: 持续1min的内存使用率大于90%, 当前值是99.98000000000002;\n',2,1706705608,0,129513,6),
	(129517,'2024-01-31 12:53:48','2024-01-31 12:57:10',0,3,'prom:9090','title: prom:9090 process_cpu_seconds_total;\ndescription: prom:9090 process_cpu_seconds_total value is 760.15;\n',2,1706705618,0,129517,3),
	(129582,'2024-01-31 12:56:38','2024-01-31 12:57:10',0,1,'node_exporter:9100','title: node_exporter:9100 go_memstats_sys_bytes > 0;\ndescription: node_exporter:9100 go_memstats_sys_bytes value is 20530184;\n',2,1706705618,0,129582,1),
	(129608,'2024-01-31 13:12:38','2024-01-31 13:13:08',0,10,'node_exporter:9100','title: node_cpu_seconds_total instance is node_exporter:9100;\ndescription: node_cpu_seconds_total value is 2110655.48;\n',2,1706706748,0,129609,5),
	(129609,'2024-01-31 13:12:38','2024-01-31 13:13:08',0,10,'node_exporter:9100','title: node_cpu_seconds_total instance is node_exporter:9100;\ndescription: node_cpu_seconds_total value is 2111990.33;\n',2,1706706748,0,129608,5),
	(129618,'2024-01-31 13:14:18','2024-01-31 13:15:08',0,10,'node_exporter:9100','title: node_cpu_seconds_total instance is node_exporter:9100;\ndescription: node_cpu_seconds_total value is 2110760.89;\n',2,1706706848,0,129619,5),
	(129619,'2024-01-31 13:14:18','2024-01-31 13:15:08',0,10,'node_exporter:9100','title: node_cpu_seconds_total instance is node_exporter:9100;\ndescription: node_cpu_seconds_total value is 2112095.76;\n',2,1706706848,0,129618,5),
	(129632,'2024-01-31 13:15:58','2024-01-31 13:18:19',0,10,'node_exporter:9100','title: node_cpu_seconds_total instance is node_exporter:9100;\ndescription: node_cpu_seconds_total value is 2112263.19;\n',2,1706706948,0,129632,5),
	(129633,'2024-01-31 13:15:58','2024-01-31 13:18:19',0,10,'node_exporter:9100','title: node_cpu_seconds_total instance is node_exporter:9100;\ndescription: node_cpu_seconds_total value is 2110927.97;\n',2,1706706948,0,129633,5),
	(129642,'2024-01-31 13:16:48','2024-01-31 13:18:18',0,3,'prom:9090','title: prom:9090 process_cpu_seconds_total;\ndescription: prom:9090 process_cpu_seconds_total value is 761.67;\n',2,1706706998,0,129642,3),
	(129674,'2024-01-31 13:19:18','2024-01-31 14:02:28',0,10,'node_exporter:9100','title: node_cpu_seconds_total instance is node_exporter:9100;\ndescription: node_cpu_seconds_total value is 2113216.76;\n',1,1706707148,0,129675,5),
	(129675,'2024-01-31 13:19:18','2024-01-31 14:02:28',0,10,'node_exporter:9100','description: node_cpu_seconds_total value is 2114553.29;\ntitle: node_cpu_seconds_total instance is node_exporter:9100;\n',1,1706707148,0,129674,5),
	(129676,'2024-01-31 13:19:18','2024-01-31 13:20:58',0,3,'prom:9090','title: prom:9090 process_cpu_seconds_total;\ndescription: prom:9090 process_cpu_seconds_total value is 761.87;\n',2,1706707148,0,129676,3);

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
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_dict` WRITE;
/*!40000 ALTER TABLE `prom_dict` DISABLE KEYS */;

INSERT INTO `prom_dict` (`id`, `created_at`, `updated_at`, `deleted_at`, `name`, `category`, `color`, `status`, `remark`)
VALUES
	(1,'2023-12-29 14:24:59','2024-01-23 21:26:43',0,'一等优先级',5,'#cf1322',1,'最高等级-1'),
	(2,'2023-12-29 14:34:12','2024-01-23 21:26:43',0,'二等优先级',5,'#d4380d',1,''),
	(3,'2023-12-29 14:34:51','2024-01-23 21:26:43',0,'三等优先级',5,'#d46b08',1,''),
	(4,'2023-12-29 14:35:32','2024-01-23 21:26:43',0,'四等优先级',5,'#d48806',1,''),
	(5,'2023-12-29 14:36:26','2024-01-23 21:26:43',0,'五等优先级',5,'#d4b106',1,''),
	(6,'2023-12-29 14:36:43','2024-01-23 21:26:43',0,'六等优先级',5,'#7cb305',1,''),
	(7,'2023-12-29 14:38:09','2024-01-23 21:27:45',0,'系统策略',3,'#165DFF',1,''),
	(8,'2023-12-30 12:28:43','2023-12-30 12:28:43',0,'系统策略组',3,'#165DFF',1,''),
	(9,'2024-01-08 14:12:32','2024-01-23 21:27:37',0,'业务策略',3,'#ff4d4f',1,''),
	(10,'2024-01-08 14:12:56','2024-01-23 21:27:30',0,'客户策略',3,'#bae637',1,''),
	(11,'2024-01-27 11:07:51','2024-01-27 11:07:51',0,'11',1,'#165DFF',1,'');

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

LOCK TABLES `prom_notify_chat_groups` WRITE;
/*!40000 ALTER TABLE `prom_notify_chat_groups` DISABLE KEYS */;

INSERT INTO `prom_notify_chat_groups` (`prom_alarm_notify_id`, `prom_alarm_chat_group_id`)
VALUES
	(1,1),
	(1,2),
	(1,3);

/*!40000 ALTER TABLE `prom_notify_chat_groups` ENABLE KEYS */;
UNLOCK TABLES;


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
  `max_suppress` varchar(255) NOT NULL DEFAULT '1m' COMMENT '最大抑制时长(s)',
  `send_recover` tinyint NOT NULL DEFAULT '0' COMMENT '是否发送告警恢复通知',
  `send_interval` varchar(255) NOT NULL DEFAULT '1m' COMMENT '发送告警时间间隔(s), 默认为for的10倍时间分钟, 用于长时间未消警情况',
  `endpoint_id` int unsigned NOT NULL DEFAULT '0' COMMENT '数据源ID',
  `create_by` int unsigned NOT NULL DEFAULT '0' COMMENT '创建人ID',
  PRIMARY KEY (`id`),
  KEY `idx__alert_level_id` (`alert_level_id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_strategies` WRITE;
/*!40000 ALTER TABLE `prom_strategies` DISABLE KEYS */;

INSERT INTO `prom_strategies` (`id`, `created_at`, `updated_at`, `deleted_at`, `group_id`, `alert`, `expr`, `for`, `labels`, `annotations`, `alert_level_id`, `status`, `remark`, `max_suppress`, `send_recover`, `send_interval`, `endpoint_id`, `create_by`)
VALUES
	(1,'2023-12-26 13:53:59','2024-01-31 20:57:07',0,1,'go_memstats_sys_bytes','go_memstats_sys_bytes{endpoint=\"hello-world-1-2-3\"} > 0','3m','{\"sverity\": \"1\"}','{\"title\": \"{{ $labels.instance }} go_memstats_sys_bytes > 0\", \"description\": \"{{ $labels.instance }} go_memstats_sys_bytes value is {{ $value }}\"}',1,2,'','1m',2,'33m',3,0),
	(2,'2023-12-26 15:08:49','2024-01-25 20:29:27',0,2,'process_cpu_seconds_total','process_cpu_seconds_total{} > 90','10s','{\"sverity\": \"2\"}','{\"title\": \"{{ $labels.instance }} process_cpu_seconds_total > 90\", \"description\": \"{{ $labels.instance }} process_cpu_seconds_total value is {{ $value }}\"}',2,2,'','1m',2,'2m',3,0),
	(3,'2023-12-27 03:09:30','2024-01-31 21:20:53',0,1,'process_cpu_seconds_total','process_cpu_seconds_total > 800','10s','{\"sverity\": \"3\"}','{\"title\": \"{{ $labels.instance }} process_cpu_seconds_total\", \"description\": \"{{ $labels.instance }} process_cpu_seconds_total value is {{ $value }}\"}',3,1,'string','',2,'',3,0),
	(4,'2023-12-27 03:44:19','2024-01-25 20:29:24',0,1,'prometheus_notifications_dropped_total','prometheus_notifications_dropped_total > 0','10s','{\"sverity\": \"4\"}','{\"title\": \"{{ $labels.instance }} prometheus_notifications_dropped_total\", \"description\": \"{{ $labels.instance }} prometheus_notifications_dropped_total values is {{ $value }}\"}',4,2,'prometheus_notifications_dropped_total指标监控规则','2s',2,'12d',3,0),
	(7,'2024-01-08 03:26:21','2024-01-31 20:31:09',0,2,'MysqlHighThreadsRunning','avg by (instance) (rate(mysql_global_status_threads_running[1m])) / avg by (instance) (mysql_global_variables_max_connections) * 100 > 60','10s','{\"sverity\": \"5\"}','{\"title\": \"MySQL high threads running (instance {{ $labels.instance }})\", \"description\": \"More than 60% of MySQL connections are in running state on {{ $labels.instance }}\\\\n  VALUE = {{ $value }}\\\\n  LABELS = {{ $labels }}\"}',5,2,'More than 60% of MySQL connections are in running state on {{ $labels.instance }}','21h',2,'222m',3,0),
	(8,'2024-01-11 05:13:39','2024-01-31 20:57:04',0,3,'cpu_90','(1 - avg(rate(process_cpu_seconds_total{}[1m])) by (instance))*100 > 90','10s','{\"sverity\": \"1\"}','{\"title\": \"持续1min的内存使用率大于90%\", \"description\": \"持续1min的内存使用率大于90%, 当前值是{{ $value }}\"}',6,2,'xxxx','1m',2,'1m',3,0),
	(9,'2024-01-23 12:37:30','2024-01-25 20:29:30',0,2,'go_memstats_stack_sys_bytes','go_memstats_stack_sys_bytes > 0','10s','{\"sverity\": \"1\"}','{\"title\": \"{{ $labels.instance }} go_memstats_stack_sys_bytes > 0\", \"description\": \"{{ $labels.instance }} go_memstats_stack_sys_bytes value is {{ $value }}\"}',1,2,'测试新增规则告警','1m',2,'1h',3,0),
	(10,'2024-01-31 13:12:24','2024-01-31 21:15:43',0,1,'node_cpu_seconds_total','node_cpu_seconds_total > 2109096','10s','{\"xx1\": \"xxx\", \"sverity\": \"1\"}','{\"title\": \"node_cpu_seconds_total instance is {{ $labels.instance }}\", \"description\": \"node_cpu_seconds_total value is {{ $value }}\"}',5,1,'','10h',2,'30m',3,1);

/*!40000 ALTER TABLE `prom_strategies` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_strategy_alarm_pages
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_alarm_pages`;

CREATE TABLE `prom_strategy_alarm_pages` (
  `prom_strategy_id` bigint unsigned NOT NULL,
  `alarm_page_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`prom_strategy_id`,`alarm_page_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_strategy_alarm_pages` WRITE;
/*!40000 ALTER TABLE `prom_strategy_alarm_pages` DISABLE KEYS */;

INSERT INTO `prom_strategy_alarm_pages` (`prom_strategy_id`, `alarm_page_id`)
VALUES
	(1,1),
	(1,2),
	(1,3),
	(2,1),
	(3,4),
	(4,1),
	(4,5),
	(7,3),
	(7,4),
	(8,3),
	(9,2),
	(10,3),
	(10,4),
	(10,5);

/*!40000 ALTER TABLE `prom_strategy_alarm_pages` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_strategy_categories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_categories`;

CREATE TABLE `prom_strategy_categories` (
  `prom_strategy_id` bigint unsigned NOT NULL,
  `dict_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`prom_strategy_id`,`dict_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_strategy_categories` WRITE;
/*!40000 ALTER TABLE `prom_strategy_categories` DISABLE KEYS */;

INSERT INTO `prom_strategy_categories` (`prom_strategy_id`, `dict_id`)
VALUES
	(1,10),
	(2,10),
	(3,10),
	(4,7),
	(7,10),
	(8,7),
	(9,9),
	(10,7);

/*!40000 ALTER TABLE `prom_strategy_categories` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_strategy_groups
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_groups`;

CREATE TABLE `prom_strategy_groups` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` bigint NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL COMMENT '规则组名称',
  `strategy_count` int NOT NULL COMMENT '规则数量',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态1:启用',
  `remark` varchar(255) NOT NULL COMMENT '描述信息',
  `enable_strategy_count` int NOT NULL DEFAULT '0' COMMENT '启用策略数量',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx__name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_strategy_groups` WRITE;
/*!40000 ALTER TABLE `prom_strategy_groups` DISABLE KEYS */;

INSERT INTO `prom_strategy_groups` (`id`, `created_at`, `updated_at`, `deleted_at`, `name`, `strategy_count`, `status`, `remark`, `enable_strategy_count`)
VALUES
	(1,'2023-12-27 03:03:14','2024-01-31 21:19:07',0,'alarm_test_group',4,1,'告警测试用',2),
	(2,'2024-01-08 14:55:57','2024-01-31 20:31:09',0,'alarm_2',3,1,'',0),
	(3,'2024-01-11 05:15:16','2024-01-31 20:57:04',0,'system_alarm',1,1,'',0);

/*!40000 ALTER TABLE `prom_strategy_groups` ENABLE KEYS */;
UNLOCK TABLES;


# 转储表 prom_strategy_notifies
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_notifies`;

CREATE TABLE `prom_strategy_notifies` (
  `prom_strategy_id` bigint unsigned NOT NULL,
  `prom_alarm_notify_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`prom_strategy_id`,`prom_alarm_notify_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `prom_strategy_notifies` WRITE;
/*!40000 ALTER TABLE `prom_strategy_notifies` DISABLE KEYS */;

INSERT INTO `prom_strategy_notifies` (`prom_strategy_id`, `prom_alarm_notify_id`)
VALUES
	(1,1),
	(3,1),
	(7,1),
	(8,1),
	(10,1);

/*!40000 ALTER TABLE `prom_strategy_notifies` ENABLE KEYS */;
UNLOCK TABLES;


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
) ENGINE=InnoDB AUTO_INCREMENT=87 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `sys_apis` WRITE;
/*!40000 ALTER TABLE `sys_apis` DISABLE KEYS */;

INSERT INTO `sys_apis` (`id`, `created_at`, `updated_at`, `name`, `path`, `method`, `status`, `remark`, `deleted_at`, `module`, `domain`)
VALUES
	(1,'2024-01-27 09:08:59','2024-01-27 09:08:59','统计各告警页面告警的数量','/api/v1/alarm_page/alarm/count','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(2,'2024-01-27 09:08:59','2024-01-27 09:08:59','批量删除告警页面','/api/v1/alarm_page/batch/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(3,'2024-01-27 09:08:59','2024-01-27 09:08:59','创建告警页面','/api/v1/alarm_page/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(4,'2024-01-27 09:08:59','2024-01-27 09:08:59','删除告警页面','/api/v1/alarm_page/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(5,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取告警页面','/api/v1/alarm_page/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(6,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取告警页面列表','/api/v1/alarm_page/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(7,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取告警页面下拉列表','/api/v1/alarm_page/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(8,'2024-01-27 09:08:59','2024-01-27 09:08:59','批量更新告警页面状态','/api/v1/alarm_page/status/batch/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(9,'2024-01-27 09:08:59','2024-01-27 09:08:59','更新告警页面','/api/v1/alarm_page/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(10,'2024-01-27 09:08:59','2024-01-27 09:08:59','创建通知群组','/api/v1/chat/group/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(11,'2024-01-27 09:08:59','2024-01-27 09:08:59','删除通知群组','/api/v1/chat/group/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(12,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取通知群组','/api/v1/chat/group/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(13,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取通知群组列表','/api/v1/chat/group/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(14,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取通知群组列表(下拉选择)','/api/v1/chat/group/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(15,'2024-01-27 09:08:59','2024-01-27 09:08:59','更新通知群组','/api/v1/chat/group/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(16,'2024-01-27 09:08:59','2024-01-27 09:08:59','添加监控端点','/api/v1/endpoint/append','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(17,'2024-01-27 09:08:59','2024-01-27 09:08:59','批量修改端点状态','/api/v1/endpoint/batch/status','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(18,'2024-01-27 09:08:59','2024-01-27 09:08:59','删除监控端点','/api/v1/endpoint/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(19,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取监控端点详情','/api/v1/endpoint/detail','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(20,'2024-01-27 09:08:59','2024-01-27 09:08:59','编辑端点信息','/api/v1/endpoint/edit','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(21,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取监控端点列表','/api/v1/endpoint/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(22,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取监控端点下拉列表','/api/v1/endpoint/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(23,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取策略组列表明细','/api/v1/strategy/group/all/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(24,'2024-01-27 09:08:59','2024-01-27 09:08:59','批量删除策略组','/api/v1/strategy/group/batch/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(25,'2024-01-27 09:08:59','2024-01-27 09:08:59','创建策略组','/api/v1/strategy/group/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(26,'2024-01-27 09:08:59','2024-01-27 09:08:59','删除策略组','/api/v1/strategy/group/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(27,'2024-01-27 09:08:59','2024-01-27 09:08:59','导出策略组','/api/v1/strategy/group/export','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(28,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取策略组','/api/v1/strategy/group/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(29,'2024-01-27 09:08:59','2024-01-27 09:08:59','导入策略组','/api/v1/strategy/group/import','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(30,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取策略组列表','/api/v1/strategy/group/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(31,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取策略组下拉列表','/api/v1/strategy/group/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(32,'2024-01-27 09:08:59','2024-01-27 09:08:59','批量更新策略组状态','/api/v1/strategy/group/status/batch/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(33,'2024-01-27 09:08:59','2024-01-27 09:08:59','更新策略组','/api/v1/strategy/group/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(34,'2024-01-27 09:08:59','2024-01-27 09:08:59','创建通知对象','/api/v1/prom/notify/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(35,'2024-01-27 09:08:59','2024-01-27 09:08:59','删除通知对象','/api/v1/prom/notify/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(36,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取通知对象详情','/api/v1/prom/notify/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(37,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取通知对象列表','/api/v1/prom/notify/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(38,'2024-01-27 09:08:59','2024-01-27 09:08:59','更新通知对象','/api/v1/prom/notify/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(39,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取通知对象列表(用于下拉选择)','/api/v1/prom/notify/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(40,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取实时告警数据详情','/api/v1/alarm/realtime/detail','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(41,'2024-01-27 09:08:59','2024-01-27 09:08:59','告警认领/介入','/api/v1/alarm/realtime/intervene','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(42,'2024-01-27 09:08:59','2024-01-27 09:08:59','告警抑制','/api/v1/alarm/realtime/suppress','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(43,'2024-01-27 09:08:59','2024-01-27 09:08:59','告警升级','/api/v1/alarm/realtime/upgrade','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(44,'2024-01-27 09:08:59','2024-01-27 09:08:59','创建用户','/api/v1/user/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(45,'2024-01-27 09:08:59','2024-01-27 09:08:59','删除用户','/api/v1/user/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(46,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取用户列表','/api/v1/user/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(47,'2024-01-27 09:08:59','2024-01-27 09:08:59','用户关联角色','/api/v1/user/roles/relate','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(48,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取用户下拉列表','/api/v1/user/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(49,'2024-01-27 09:08:59','2024-01-27 09:08:59','批量修改用户状态','/api/v1/user/status/edit','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(50,'2024-01-27 09:08:59','2024-01-27 09:08:59','更新用户基础信息','/api/v1/user/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(51,'2024-01-27 09:08:59','2024-01-27 09:08:59','修改密码','/api/v1/user/password/edit','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(52,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取用户详情','/api/v1/user/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(53,'2024-01-27 09:08:59','2024-01-27 09:08:59','创建角色','/api/v1/role/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(54,'2024-01-27 09:08:59','2024-01-27 09:08:59','删除角色','/api/v1/role/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(55,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取角色详情','/api/v1/role/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(56,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取角色列表','/api/v1/role/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(57,'2024-01-27 09:08:59','2024-01-27 09:08:59','角色关联API','/api/v1/role/relate/api','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(58,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取角色下拉列表','/api/v1/role/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(59,'2024-01-27 09:08:59','2024-01-27 09:08:59','更新角色','/api/v1/role/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(60,'2024-01-27 09:08:59','2024-01-27 09:08:59','创建API','/api/v1/system/api/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(61,'2024-01-27 09:08:59','2024-01-27 09:08:59','删除API','/api/v1/system/api/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(62,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取API详情','/api/v1/system/api/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(63,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取API列表','/api/v1/system/api/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(64,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取API下拉列表','/api/v1/system/api/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(65,'2024-01-27 09:08:59','2024-01-27 09:08:59','更新API数据','/api/v1/system/api/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(66,'2024-01-27 09:08:59','2024-01-27 09:08:59','测试','/test','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(67,'2024-01-27 09:08:59','2024-01-27 09:08:59','创建字典','/api/v1/dict/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(68,'2024-01-27 09:08:59','2024-01-27 09:08:59','字典列表','/api/v1/dict/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(69,'2024-01-27 09:08:59','2024-01-27 09:08:59','更新字典状态','/api/v1/dict/status/update/batch','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(70,'2024-01-27 09:08:59','2024-01-27 09:08:59','删除字典列表','/api/v1/dict/batch/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(71,'2024-01-27 09:08:59','2024-01-27 09:08:59','删除字典','/api/v1/dict/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(72,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取字典详情','/api/v1/dict/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(73,'2024-01-27 09:08:59','2024-01-27 09:08:59','获取字典下拉列表','/api/v1/dict/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(74,'2024-01-27 09:08:59','2024-01-27 09:08:59','更新字典','/api/v1/dict/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(75,'2024-01-27 09:08:59','2024-01-27 09:08:59','实时告警列表','/api/v1/alarm/realtime/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(76,'2024-01-27 09:08:59','2024-01-27 09:08:59','报警历史','/api/v1/alarm/history/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(77,'2024-01-27 09:08:59','2024-01-27 09:08:59','报警历史详情','/api/v1/alarm/history/get','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(78,'2024-01-27 09:08:59','2024-01-27 09:08:59','策略列表','/api/v1/strategy/list','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(79,'2024-01-27 09:08:59','2024-01-27 09:08:59','策略详情','/api/v1/strategy/detail','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(80,'2024-01-27 09:08:59','2024-01-27 09:08:59','创建策略','/api/v1/strategy/create','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(81,'2024-01-27 09:08:59','2024-01-27 09:08:59','更新策略','/api/v1/strategy/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(82,'2024-01-27 09:08:59','2024-01-27 09:08:59','删除策略','/api/v1/strategy/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(83,'2024-01-27 09:08:59','2024-01-27 09:08:59','批量启用策略','/api/v1/strategy/status/batch/update','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(84,'2024-01-27 09:08:59','2024-01-27 09:08:59','批量删除策略','/api/v1/strategy/batch/delete','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(85,'2024-01-27 09:08:59','2024-01-27 09:08:59','策略列表(下拉)','/api/v1/strategy/select','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0),
	(86,'2024-01-27 09:08:59','2024-01-27 09:08:59','绑定通知对象','/api/v1/strategy/notify/object/bind','POST',1,'这个API没有说明, 赶紧补充吧',0,0,0);

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
	(3,1),
	(3,3),
	(3,5),
	(3,6),
	(3,7),
	(3,9),
	(3,10),
	(3,12),
	(3,13),
	(3,14),
	(3,15),
	(3,16),
	(3,19),
	(3,20),
	(3,21),
	(3,22),
	(3,23),
	(3,25),
	(3,27),
	(3,28),
	(3,29),
	(3,30),
	(3,31),
	(3,32),
	(3,33),
	(3,34),
	(3,36),
	(3,37),
	(3,38),
	(3,39),
	(3,40),
	(3,41),
	(3,42),
	(3,43),
	(3,51),
	(3,55),
	(3,56),
	(3,58),
	(3,62),
	(3,63),
	(3,64),
	(3,67),
	(3,68),
	(3,72),
	(3,73),
	(3,75),
	(3,76),
	(3,77),
	(3,78),
	(3,79),
	(3,80),
	(3,81),
	(3,83),
	(3,85),
	(3,86);

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
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `sys_roles` WRITE;
/*!40000 ALTER TABLE `sys_roles` DISABLE KEYS */;

INSERT INTO `sys_roles` (`id`, `created_at`, `updated_at`, `remark`, `name`, `status`, `deleted_at`)
VALUES
	(1,'2023-12-02 05:52:02','2023-12-03 11:59:59','超级管理员角色','超级管理员',1,0),
	(2,'2023-12-03 03:47:28','2023-12-04 22:48:24','普通管理员角色','普通管理员',1,0),
	(3,'2023-12-03 07:47:23','2024-01-31 18:29:48','普通用户角色','普通用户',1,0);

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
	(1,2),
	(3,3),
	(3,6),
	(3,7);

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
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `sys_users` WRITE;
/*!40000 ALTER TABLE `sys_users` DISABLE KEYS */;

INSERT INTO `sys_users` (`id`, `created_at`, `updated_at`, `username`, `nickname`, `password`, `email`, `phone`, `status`, `remark`, `avatar`, `salt`, `deleted_at`, `gender`)
VALUES
	(1,'1970-01-01 08:00:00','2024-01-27 06:32:35','lili1997','梧桐','zUnQGjDhAd7ChhVLa9R+Dw==','aidedevops@163.com','18275111234',1,'','http://img.touxiangwu.com/uploads/allimg/2022010212/0k31jrfuezv.jpg','2441c436f3fa3613',0,1),
	(2,'1970-01-01 08:00:00','2024-01-23 13:17:46','lier','莉莉','xen5WnrlweLWODeArERA+g==','aidelier@163.com','18275111235',1,'','https://img1.baidu.com/it/u=1090762042,1225374156&fm=253&fmt=auto&app=138&f=JPEG?w=500&h=500','5942d93b2c07ff52',0,2),
	(3,'2023-12-02 06:00:20','2024-01-27 10:29:46','testaxs','测试号','53ZnDtGGF4JyC+fCFJsKHA==','aidetest@163.com','15585310001',1,'','https://img1.baidu.com/it/u=1285375996,3783960243&fm=253&fmt=auto&app=138&f=JPEG?w=500&h=500','0addd2e0f4145e3b',0,2),
	(4,'2023-12-02 09:20:16','2023-12-06 21:28:48','sss','','C6xWQLqxBnZOv813W7RHuw==','123@163.com','13355777441',1,'','','8c12d009db56e0ea',0,1),
	(5,'2023-12-02 09:29:02','2023-12-02 12:34:32','sfasfas','','Bzlv+7/jQnmoNslxPHq2TQ==','12312@qq.com','18812312311',2,'','','96b530aedc68d976',1701520472,0),
	(6,'2024-01-27 10:23:12','2024-01-27 10:28:16','num1','一号体验用户','cEF0InjT0r+VneKN+rMutR8d4BKOxSnPiDOtkDyacYbqV4kSuZuGE6LsZOAbUGCB','123@qq.com','18888888888',1,'','https://img1.baidu.com/it/u=2879608591,2832950529&fm=253&app=138&size=w931&n=0&f=JPEG&fmt=auto?sec=1706461200&t=4dbdef4e024555f7bf77ba4951d36ecd','bd3e6b502c29eee2',0,1),
	(7,'2024-01-27 12:41:23','2024-01-27 12:44:01','prometheus','体验用户','vhfWajJFDIfNAc0AjV/9kw==','1233@qq.com','15555555555',1,'','https://github.com/aide-cloud/prometheus-manager/raw/main/doc/img/prometheus-logo.svg','d0107166c919f1ed',0,1);

/*!40000 ALTER TABLE `sys_users` ENABLE KEYS */;
UNLOCK TABLES;



/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
