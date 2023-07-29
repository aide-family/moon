# ************************************************************
# Sequel Ace SQL dump
# 版本号： 20016
#
# https://sequel-ace.com/
# https://github.com/Sequel-Ace/Sequel-Ace
#
# 主机: localhost (MySQL 8.0.33)
# 数据库: prom
# 生成时间: 2023-07-29 10:08:40 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT = @@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS = @@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION = @@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS = @@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS = 0 */;
/*!40101 SET @OLD_SQL_MODE = 'NO_AUTO_VALUE_ON_ZERO', SQL_MODE = 'NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES = @@SQL_NOTES, SQL_NOTES = 0 */;


# 转储表 prom_rules
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_rules`;

CREATE TABLE `prom_rules`
(
    `id`          int unsigned                                                   NOT NULL AUTO_INCREMENT,
    `alert`       varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci   NOT NULL DEFAULT '_rules' COMMENT '策略名称',
    `expr`        varchar(4096) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '0' COMMENT 'prom QL',
    `for`         varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci   NOT NULL DEFAULT '7200h' COMMENT '持续时间',
    `labels`      json                                                           NOT NULL COMMENT '标签',
    `annotations` json                                                           NOT NULL COMMENT '内容',
    `created_at`  timestamp                                                      NOT NULL COMMENT '创建时间',
    `updated_at`  timestamp                                                      NULL     DEFAULT NULL COMMENT '更新时间',
    `deleted_at`  bigint                                                         NOT NULL DEFAULT '0' COMMENT '删除时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;



# 转储表 prom_combo_strategies
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_combo_strategies`;

CREATE TABLE `prom_combo_strategies`
(
    `prom_combo_id` int unsigned NOT NULL,
    `prom_rule_id`  int unsigned NOT NULL,
    UNIQUE KEY `idx_combo_id__strategy_id` (`prom_combo_id`, `prom_rule_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;



# 转储表 prom_combos
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_combos`;

CREATE TABLE `prom_combos`
(
    `id`         int unsigned                             NOT NULL AUTO_INCREMENT,
    `name`       varchar(64) COLLATE utf8mb4_general_ci   NOT NULL DEFAULT '' COMMENT '套餐名称',
    `remark`     varchar(2048) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '套餐说明',
    `created_at` timestamp                                NOT NULL,
    `updated_at` timestamp                                NULL     DEFAULT NULL,
    `deleted_at` bigint unsigned                          NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;



# 转储表 prom_node_dir_file_group_strategies
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_node_dir_file_group_strategies`;

CREATE TABLE `prom_node_dir_file_group_strategies`
(
    `id`          int unsigned                                                   NOT NULL AUTO_INCREMENT,
    `group_id`    int unsigned                                                   NOT NULL DEFAULT '0' COMMENT '策略组 ID',
    `alert`       varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci   NOT NULL DEFAULT '_rules' COMMENT '策略名称',
    `expr`        varchar(4096) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '0' COMMENT 'prom QL',
    `for`         varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci   NOT NULL DEFAULT '7200h' COMMENT '持续时间',
    `labels`      json                                                           NOT NULL COMMENT '标签',
    `annotations` json                                                           NOT NULL COMMENT '内容',
    `created_at`  timestamp                                                      NOT NULL COMMENT '创建时间',
    `updated_at`  timestamp                                                      NULL     DEFAULT NULL COMMENT '更新时间',
    `deleted_at`  bigint                                                         NOT NULL DEFAULT '0' COMMENT '删除时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;



# 转储表 prom_node_dir_file_groups
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_node_dir_file_groups`;

CREATE TABLE `prom_node_dir_file_groups`
(
    `id`         int unsigned                                                  NOT NULL AUTO_INCREMENT,
    `name`       varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL DEFAULT 'alert_' COMMENT '策略组名称',
    `remark`     varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '说明',
    `created_at` timestamp                                                     NOT NULL,
    `updated_at` timestamp                                                     NULL     DEFAULT NULL,
    `deleted_at` bigint unsigned                                               NOT NULL DEFAULT '0',
    `file_id`    int unsigned                                                  NOT NULL DEFAULT '0' COMMENT 'files id',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;



# 转储表 prom_node_dir_files
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_node_dir_files`;

CREATE TABLE `prom_node_dir_files`
(
    `id`         int unsigned                           NOT NULL AUTO_INCREMENT,
    `filename`   varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'yaml file',
    `created_at` timestamp                              NOT NULL,
    `updated_at` timestamp                              NULL     DEFAULT NULL,
    `deleted_at` bigint unsigned                        NOT NULL DEFAULT '0',
    `dir_id`     int unsigned                           NOT NULL DEFAULT '0' COMMENT '目录 ID',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;



# 转储表 prom_node_dirs
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_node_dirs`;

CREATE TABLE `prom_node_dirs`
(
    `id`         int unsigned    NOT NULL AUTO_INCREMENT,
    `node_id`    int unsigned    NOT NULL                DEFAULT '0' COMMENT '节点 ID',
    `path`       varchar(255) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '目录地址',
    `created_at` timestamp       NOT NULL,
    `updated_at` timestamp       NULL                    DEFAULT NULL,
    `deleted_at` bigint unsigned NOT NULL                DEFAULT '0',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;



# 转储表 prom_nodes
# ------------------------------------------------------------

DROP TABLE IF EXISTS `prom_nodes`;

CREATE TABLE `prom_nodes`
(
    `id`         int unsigned                             NOT NULL AUTO_INCREMENT,
    `created_at` timestamp                                NOT NULL,
    `updated_at` timestamp                                NULL     DEFAULT NULL,
    `deleted_at` bigint unsigned                          NOT NULL DEFAULT '0',
    `en_name`    varchar(64) COLLATE utf8mb4_general_ci   NOT NULL DEFAULT '' COMMENT '节点英文名称',
    `ch_name`    varchar(64) COLLATE utf8mb4_general_ci   NOT NULL DEFAULT '' COMMENT '节点中文名称',
    `datasource` varchar(1024) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'prom数据源地址',
    `remark`     varchar(255) COLLATE utf8mb4_general_ci  NOT NULL DEFAULT '' COMMENT '备注',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;



/*!40111 SET SQL_NOTES = @OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE = @OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS = @OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT = @OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS = @OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION = @OLD_COLLATION_CONNECTION */;
