-- =============================================
-- 电子签章与区块链存证功能
-- 包含：esign_flow 增强 + esign_signer 增强 + blockchain_certificate 建表
-- =============================================

-- 1. esign_flow 增加法大大和区块链字段（增量迁移，幂等）
SET @existCol = (SELECT COUNT(*) FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'esign_flow' AND COLUMN_NAME = 'fadada_flow_id');
SET @sql = IF(@existCol = 0,
    'ALTER TABLE `esign_flow`
        ADD COLUMN `case_no` VARCHAR(50) NULL DEFAULT NULL COMMENT ''案件编号'' AFTER `case_id`,
        ADD COLUMN `signed_document_url` VARCHAR(500) NULL DEFAULT NULL COMMENT ''已签署文档URL'' AFTER `document_url`,
        ADD COLUMN `fadada_flow_id` VARCHAR(100) NULL DEFAULT NULL COMMENT ''法大大签署流程ID'' AFTER `expire_time`,
        ADD COLUMN `cross_page_seal` TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''是否骑缝章 0否 1是'' AFTER `fadada_flow_id`,
        ADD COLUMN `bc_cert_no` VARCHAR(50) NULL DEFAULT NULL COMMENT ''区块链存证证书编号'' AFTER `cross_page_seal`,
        ADD COLUMN `bc_tx_id` VARCHAR(100) NULL DEFAULT NULL COMMENT ''区块链交易ID'' AFTER `bc_cert_no`,
        ADD COLUMN `bc_on_chain_time` DATETIME NULL DEFAULT NULL COMMENT ''上链时间'' AFTER `bc_tx_id`,
        ADD COLUMN `bc_status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''区块链存证状态 0待存证 1已存证 2失败 3已验证'' AFTER `bc_on_chain_time`,
        ADD INDEX `idx_fadada_flow_id` (`fadada_flow_id`),
        ADD INDEX `idx_bc_cert_no` (`bc_cert_no`)',
    'SELECT ''esign_flow columns already exist'' AS msg');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. esign_signer 增加法大大和通知字段（增量迁移，幂等）
SET @existCol = (SELECT COUNT(*) FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'esign_signer' AND COLUMN_NAME = 'fadada_sign_url');
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
    `flow_id` VARCHAR(50) NULL DEFAULT NULL COMMENT '签署流程编号',
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
