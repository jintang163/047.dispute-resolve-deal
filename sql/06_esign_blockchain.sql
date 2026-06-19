-- =============================================
-- 电子签章与区块链存证功能 (v2 - 统一表结构)
-- esign_flow 完整建表 + esign_signer 完整建表 + blockchain_certificate 建表
-- =============================================

-- 1. esign_flow 完整建表（统一表名和字段）
CREATE TABLE IF NOT EXISTS `esign_flow` (
    `id` BIGINT(20) NOT NULL COMMENT '主键ID',
    `flow_no` VARCHAR(50) NOT NULL COMMENT '签署流程编号',
    `case_id` BIGINT(20) NOT NULL COMMENT '案件ID',
    `case_no` VARCHAR(50) NULL DEFAULT NULL COMMENT '案件编号',
    `doc_type` TINYINT(4) NOT NULL DEFAULT 0 COMMENT '文档类型',
    `doc_title` VARCHAR(200) NOT NULL COMMENT '文档标题',
    `doc_content` TEXT NULL COMMENT '文档内容',
    `doc_url` VARCHAR(500) NULL DEFAULT NULL COMMENT '原文档URL',
    `template_id` BIGINT(20) NULL DEFAULT NULL COMMENT '模板ID',
    `signed_document_url` VARCHAR(500) NULL DEFAULT NULL COMMENT '已签署文档URL',
    `status` TINYINT(4) NOT NULL DEFAULT 0 COMMENT '签署状态 0草稿 10待签署 20签署中 30已完成 40已过期 50已撤销',
    `current_sign_index` INT(11) NOT NULL DEFAULT 0 COMMENT '当前签署序号',
    `total_sign_count` INT(11) NOT NULL DEFAULT 0 COMMENT '总签署人数',
    `signed_count` INT(11) NOT NULL DEFAULT 0 COMMENT '已签署人数',
    `expire_time` DATETIME NULL DEFAULT NULL COMMENT '过期时间',
    `fadada_flow_id` VARCHAR(100) NULL DEFAULT NULL COMMENT '法大大签署流程ID',
    `cross_page_seal` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否骑缝章 0否 1是',
    `bc_cert_no` VARCHAR(50) NULL DEFAULT NULL COMMENT '区块链存证证书编号',
    `bc_tx_id` VARCHAR(100) NULL DEFAULT NULL COMMENT '区块链交易ID',
    `bc_on_chain_time` DATETIME NULL DEFAULT NULL COMMENT '上链时间',
    `bc_status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '区块链存证状态 0待存证 1已存证 2失败 3已验证',
    `creator_id` BIGINT(20) NULL DEFAULT NULL COMMENT '创建人ID',
    `creator_name` VARCHAR(50) NULL DEFAULT NULL COMMENT '创建人姓名',
    `organization_id` BIGINT(20) NULL DEFAULT NULL COMMENT '组织ID',
    `revoke_reason` VARCHAR(500) NULL DEFAULT NULL COMMENT '撤销原因',
    `revoke_time` DATETIME NULL DEFAULT NULL COMMENT '撤销时间',
    `revoke_by` BIGINT(20) NULL DEFAULT NULL COMMENT '撤销人ID',
    `last_sign_time` DATETIME NULL DEFAULT NULL COMMENT '最后签署时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_flow_no` (`flow_no`),
    INDEX `idx_case_id` (`case_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_creator_id` (`creator_id`),
    INDEX `idx_organization_id` (`organization_id`),
    INDEX `idx_fadada_flow_id` (`fadada_flow_id`),
    INDEX `idx_bc_cert_no` (`bc_cert_no`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='电子签章流程表';

-- 如果 esign_flow 表已存在但缺少字段，增量补齐（幂等）
SET @existCol = (SELECT COUNT(*) FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'esign_flow' AND COLUMN_NAME = 'doc_type');
SET @sql = IF(@existCol = 0,
    'ALTER TABLE `esign_flow`
        ADD COLUMN `doc_type` TINYINT(4) NOT NULL DEFAULT 0 COMMENT ''文档类型'' AFTER `case_no`,
        ADD COLUMN `doc_title` VARCHAR(200) NOT NULL DEFAULT '''' COMMENT ''文档标题'' AFTER `doc_type`,
        ADD COLUMN `doc_content` TEXT NULL COMMENT ''文档内容'' AFTER `doc_title`,
        ADD COLUMN `doc_url` VARCHAR(500) NULL DEFAULT NULL COMMENT ''原文档URL'' AFTER `doc_content`,
        ADD COLUMN `creator_name` VARCHAR(50) NULL DEFAULT NULL COMMENT ''创建人姓名'' AFTER `bc_status`,
        ADD COLUMN `organization_id` BIGINT(20) NULL DEFAULT NULL COMMENT ''组织ID'' AFTER `creator_name`,
        ADD COLUMN `revoke_reason` VARCHAR(500) NULL DEFAULT NULL COMMENT ''撤销原因'' AFTER `organization_id`,
        ADD COLUMN `revoke_time` DATETIME NULL DEFAULT NULL COMMENT ''撤销时间'' AFTER `revoke_reason`,
        ADD COLUMN `revoke_by` BIGINT(20) NULL DEFAULT NULL COMMENT ''撤销人ID'' AFTER `revoke_time`,
        ADD COLUMN `last_sign_time` DATETIME NULL DEFAULT NULL COMMENT ''最后签署时间'' AFTER `revoke_by`,
        ADD INDEX `idx_creator_id` (`creator_id`),
        ADD INDEX `idx_organization_id` (`organization_id`)',
    'SELECT ''esign_flow columns already exist'' AS msg');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 如果旧表 esign_record 存在，迁移数据到 esign_flow（幂等）
-- 注意：仅迁移 esign_flow 中不存在的 flow_no
INSERT IGNORE INTO `esign_flow` (`id`, `flow_no`, `case_id`, `case_no`, `doc_title`, `doc_url`, `signed_document_url`, `status`, `total_sign_count`, `signed_count`, `expire_time`, `fadada_flow_id`, `cross_page_seal`, `bc_cert_no`, `bc_tx_id`, `bc_on_chain_time`, `bc_status`, `creator_id`, `creator_name`, `organization_id`, `revoke_reason`, `revoke_time`, `revoke_by`, `last_sign_time`, `created_at`, `updated_at`, `deleted_at`)
SELECT `id`, `flow_id`, `case_id`, `case_no`, `doc_title`, `doc_url`, `signed_document_url`, `status`, `signer_count`, `signed_count`, `expire_time`, `fadada_flow_id`, `cross_page_seal`, `bc_cert_no`, `bc_tx_id`, `bc_on_chain_time`, `bc_status`, `creator_id`, `creator_name`, `organization_id`, `revoke_reason`, `revoke_time`, `revoke_by`, `last_sign_time`, `created_at`, `updated_at`, `deleted_at`
FROM `esign_record`
WHERE NOT EXISTS (SELECT 1 FROM `esign_flow` WHERE `esign_flow`.`flow_no` = `esign_record`.`flow_id`);

-- 2. esign_signer 完整建表（统一字段：sign_status/sign_time 替代 status/signed_at）
CREATE TABLE IF NOT EXISTS `esign_signer` (
    `id` BIGINT(20) NOT NULL COMMENT '主键ID',
    `flow_id` BIGINT(20) NOT NULL COMMENT '签署流程ID(esign_flow.id)',
    `user_id` BIGINT(20) NOT NULL COMMENT '签署人ID',
    `user_name` VARCHAR(50) NOT NULL COMMENT '签署人姓名',
    `user_phone` VARCHAR(20) NULL DEFAULT NULL COMMENT '签署人手机号',
    `id_card` VARCHAR(20) NULL DEFAULT NULL COMMENT '签署人身份证号',
    `sign_order` INT(11) NOT NULL DEFAULT 0 COMMENT '签署顺序',
    `sign_status` TINYINT(4) NOT NULL DEFAULT 0 COMMENT '签署状态 0待签署 1已签署 2已拒绝',
    `sign_time` DATETIME NULL DEFAULT NULL COMMENT '签署时间',
    `signature_url` VARCHAR(500) NULL DEFAULT NULL COMMENT '签名图片URL',
    `verify_code` VARCHAR(10) NULL DEFAULT NULL COMMENT '验证码',
    `sign_ip` VARCHAR(50) NULL DEFAULT NULL COMMENT '签署IP',
    `remark` VARCHAR(500) NULL DEFAULT NULL COMMENT '备注',
    `fadada_sign_url` VARCHAR(500) NULL DEFAULT NULL COMMENT '法大大签署URL',
    `notify_status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '通知状态 0未通知 1短信 2微信 3全部',
    `notify_sent_at` DATETIME NULL DEFAULT NULL COMMENT '通知发送时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_flow_id` (`flow_id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_sign_status` (`sign_status`),
    INDEX `idx_flow_user` (`flow_id`, `user_id`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='签署人表';

-- 如果 esign_signer 已存在但使用旧字段名 status/signed_at，迁移列名
SET @existCol = (SELECT COUNT(*) FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'esign_signer' AND COLUMN_NAME = 'sign_status');
SET @sql = IF(@existCol = 0,
    'ALTER TABLE `esign_signer`
        CHANGE COLUMN `status` `sign_status` TINYINT(4) NOT NULL DEFAULT 0 COMMENT ''签署状态 0待签署 1已签署 2已拒绝'',
        CHANGE COLUMN `signed_at` `sign_time` DATETIME NULL DEFAULT NULL COMMENT ''签署时间''',
    'SELECT ''esign_signer sign_status already exists'' AS msg');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- esign_signer 增量字段补齐（幂等）
SET @existCol = (SELECT COUNT(*) FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'esign_signer' AND COLUMN_NAME = 'user_phone');
SET @sql = IF(@existCol = 0,
    'ALTER TABLE `esign_signer`
        ADD COLUMN `user_phone` VARCHAR(20) NULL DEFAULT NULL COMMENT ''签署人手机号'' AFTER `user_name`,
        ADD COLUMN `id_card` VARCHAR(20) NULL DEFAULT NULL COMMENT ''签署人身份证号'' AFTER `user_phone`,
        ADD COLUMN `fadada_sign_url` VARCHAR(500) NULL DEFAULT NULL COMMENT ''法大大签署URL'' AFTER `remark`,
        ADD COLUMN `notify_status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''通知状态 0未通知 1短信 2微信 3全部'' AFTER `fadada_sign_url`,
        ADD COLUMN `notify_sent_at` DATETIME NULL DEFAULT NULL COMMENT ''通知发送时间'' AFTER `notify_status`',
    'SELECT ''esign_signer columns already exist'' AS msg');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 3. blockchain_certificate 完整建表
CREATE TABLE IF NOT EXISTS `blockchain_certificate` (
    `id` BIGINT(20) NOT NULL COMMENT '主键ID',
    `cert_no` VARCHAR(50) NOT NULL COMMENT '存证证书编号',
    `evidence_id` VARCHAR(50) NOT NULL COMMENT '存证业务ID',
    `evidence_type` VARCHAR(30) NOT NULL COMMENT '存证类型: mediation_protocol/esign_document/evidence',
    `evidence_name` VARCHAR(200) NOT NULL COMMENT '存证名称',
    `evidence_hash` VARCHAR(64) NOT NULL COMMENT '证据哈希值(SHA256)',
    `case_id` BIGINT(20) NOT NULL COMMENT '案件ID',
    `flow_id` VARCHAR(50) NULL DEFAULT NULL COMMENT '签署流程编号(esign_flow.flow_no)',
    `tx_id` VARCHAR(100) NOT NULL COMMENT '区块链交易ID',
    `block_height` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '区块高度',
    `on_chain_time` DATETIME NULL DEFAULT NULL COMMENT '上链时间',
    `cert_url` VARCHAR(500) NULL DEFAULT NULL COMMENT '证书下载URL',
    `qrcode_url` VARCHAR(500) NULL DEFAULT NULL COMMENT '核验二维码URL',
    `verify_url` VARCHAR(500) NULL DEFAULT NULL COMMENT '核验页面URL',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态 0待存证 1已存证 2失败 3已验证',
    `metadata` TEXT NULL COMMENT '附加元数据JSON',
    `created_by` BIGINT(20) NULL DEFAULT NULL COMMENT '创建人ID',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_cert_no` (`cert_no`),
    INDEX `idx_evidence_id` (`evidence_id`),
    INDEX `idx_case_id` (`case_id`),
    INDEX `idx_flow_id` (`flow_id`),
    INDEX `idx_evidence_type` (`evidence_type`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='区块链存证证书表';
