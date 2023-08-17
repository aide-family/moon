# ************************************************************
# Sequel Ace SQL dump
# 版本号： 20016
#
# https://sequel-ace.com/
# https://github.com/Sequel-Ace/Sequel-Ace
#
# 主机: localhost (MySQL 8.0.33)
# 数据库: prom
# 生成时间: 2023-08-15 07:16:41 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT = @@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS = @@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION = @@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS = @@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS = 0 */;
/*!40101 SET @OLD_SQL_MODE = 'NO_AUTO_VALUE_ON_ZERO', SQL_MODE = 'NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES = @@SQL_NOTES, SQL_NOTES = 0 */;


# 转储表 prom_alarm_pages
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_alarm_pages`;

CREATE TABLE `prom_alarm_pages`
(
    `id`         int unsigned                             NOT NULL AUTO_INCREMENT,
    `name`       varchar(64) COLLATE utf8mb4_general_ci   NOT NULL DEFAULT '' COMMENT '报警页面名称',
    `remark`     varchar(255) COLLATE utf8mb4_general_ci  NOT NULL DEFAULT '' COMMENT '描述信息',
    `icon`       varchar(1024) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '图表',
    `color`      varchar(64) COLLATE utf8mb4_general_ci   NOT NULL DEFAULT '' COMMENT 'tab颜色',
    `status`     tinyint                                  NOT NULL DEFAULT '1' COMMENT '启用状态,1启用;2禁用',
    `created_at` timestamp                                NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp                                NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` timestamp                                NULL     DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx__name` (`name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

# 转储表 prom_dict
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_dict`;

CREATE TABLE `prom_dict`
(
    `id`         int unsigned                                                 NOT NULL AUTO_INCREMENT,
    `name`       varchar(64) COLLATE utf8mb4_general_ci                       NOT NULL DEFAULT '' COMMENT '字典名称',
    `category`   tinyint                                                      NOT NULL DEFAULT '0' COMMENT '字典类型',
    `color`      varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '#165DFF' COMMENT '字典tag颜色',
    `status`     tinyint                                                      NOT NULL DEFAULT '1' COMMENT '状态',
    `remark`     varchar(255) COLLATE utf8mb4_general_ci                      NOT NULL DEFAULT '' COMMENT '字典备注',
    `created_at` timestamp                                                    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp                                                    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` timestamp                                                    NULL     DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx__name__category` (`name`, `category`),
    KEY `idx__category` (`category`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

# 转储表 prom_group_categories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_group_categories`;

CREATE TABLE `prom_group_categories`
(
    `dict_id`       int unsigned NOT NULL,
    `prom_group_id` int unsigned NOT NULL,
    UNIQUE KEY `idx__dict_id__group_id` (`dict_id`, `prom_group_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

# 转储表 prom_groups
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_groups`;

CREATE TABLE `prom_groups`
(
    `id`             int unsigned                                                  NOT NULL AUTO_INCREMENT,
    `name`           varchar(64) COLLATE utf8mb4_general_ci                        NOT NULL DEFAULT '' COMMENT '规则组名称',
    `strategy_count` bigint                                                        NOT NULL DEFAULT '0' COMMENT '规则数量',
    `status`         tinyint                                                       NOT NULL DEFAULT '1' COMMENT '启用状态1:启用;2禁用',
    `remark`         varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '描述信息',
    `created_at`     timestamp                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`     timestamp                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`     timestamp                                                     NULL     DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;


# 转储表 prom_strategies
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategies`;

CREATE TABLE `prom_strategies`
(
    `id`             int unsigned                           NOT NULL AUTO_INCREMENT,
    `group_id`       int unsigned                           NOT NULL DEFAULT '0' COMMENT '所属规则组ID',
    `alert`          varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '规则名称',
    `expr`           text COLLATE utf8mb4_general_ci        NOT NULL COMMENT 'prom ql',
    `for`            varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '10s' COMMENT '持续时间',
    `labels`         json                                   NOT NULL COMMENT '标签',
    `annotations`    json                                   NOT NULL COMMENT '告警文案',
    `alert_level_id` int                                    NOT NULL DEFAULT '0' COMMENT '告警等级dict ID ',
    `status`         tinyint                                NOT NULL DEFAULT '1' COMMENT '启用状态: 1启用;2禁用',
    `created_at`     timestamp                              NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`     timestamp                              NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`     timestamp                              NULL     DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY `idx__alart_level_id` (`alert_level_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

# 转储表 prom_strategy_alarm_pages
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_alarm_pages`;

CREATE TABLE `prom_strategy_alarm_pages`
(
    `alarm_page_id`    int unsigned NOT NULL,
    `prom_strategy_id` int unsigned NOT NULL,
    UNIQUE KEY `idx__prom_strategy_id__dict_id` (`alarm_page_id`, `prom_strategy_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;


# 转储表 prom_strategy_categories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_strategy_categories`;

CREATE TABLE `prom_strategy_categories`
(
    `dict_id`          int unsigned NOT NULL,
    `prom_strategy_id` int unsigned NOT NULL,
    UNIQUE KEY `idx__prom_strategy_id__dict_id` (`dict_id`, `prom_strategy_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

/*!40111 SET SQL_NOTES = @OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE = @OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS = @OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT = @OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS = @OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION = @OLD_COLLATION_CONNECTION */;
