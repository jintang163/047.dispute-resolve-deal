-- =============================================
-- 调解协议AI智能生成功能增强
-- 包含：mediation_protocol 建表 + AI字段扩展 + ai_generation_log
-- =============================================

-- 1. mediation_protocol 完整建表
CREATE TABLE IF NOT EXISTS `mediation_protocol` (
    `id` BIGINT(20) NOT NULL COMMENT '主键ID',
    `case_id` BIGINT(20) NOT NULL COMMENT '案件ID',
    `protocol_no` VARCHAR(50) NOT NULL COMMENT '协议编号',
    `title` VARCHAR(200) NOT NULL COMMENT '协议标题',
    `content` TEXT NULL COMMENT '协议完整内容（Markdown格式）',
    `party_a_name` VARCHAR(50) NULL DEFAULT NULL COMMENT '甲方姓名',
    `party_b_name` VARCHAR(50) NULL DEFAULT NULL COMMENT '乙方姓名',
    `mediator_name` VARCHAR(50) NULL DEFAULT NULL COMMENT '调解员姓名',
    `agreement_items` TEXT NULL COMMENT '协议事项摘要',
    `breach_clause` VARCHAR(1000) NULL DEFAULT NULL COMMENT '违约条款',
    `effective_date` DATETIME NULL DEFAULT NULL COMMENT '生效日期',
    `is_signed` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否签署 0否 1是',
    `signed_at` DATETIME NULL DEFAULT NULL COMMENT '签署时间',
    `file_url` VARCHAR(500) NULL DEFAULT NULL COMMENT '协议文件URL',
    `created_by` BIGINT(20) NULL DEFAULT NULL COMMENT '创建人ID',
    `is_ai_generated` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否AI生成 0否 1是',
    `ai_generated_at` DATETIME NULL DEFAULT NULL COMMENT 'AI生成时间',
    `is_adopted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否采用 0否 1是',
    `adopted_by` BIGINT(20) NULL DEFAULT NULL COMMENT '采用人ID',
    `adopted_at` DATETIME NULL DEFAULT NULL COMMENT '采用时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_protocol_no` (`protocol_no`),
    INDEX `idx_case_id` (`case_id`),
    INDEX `idx_is_signed` (`is_signed`),
    INDEX `idx_ai_generated` (`is_ai_generated`),
    INDEX `idx_is_adopted` (`is_adopted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='调解协议表';

-- 2. 若表已存在则补齐AI相关字段（增量迁移，幂等）
-- 检测 is_ai_generated 字段是否存在，不存在则添加
SET @existCol = (SELECT COUNT(*) FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'mediation_protocol' AND COLUMN_NAME = 'is_ai_generated');
SET @sql = IF(@existCol = 0,
    'ALTER TABLE `mediation_protocol`
        ADD COLUMN `is_ai_generated` TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''是否AI生成 0否 1是'' AFTER `created_by`,
        ADD COLUMN `ai_generated_at` DATETIME NULL DEFAULT NULL COMMENT ''AI生成时间'' AFTER `is_ai_generated`,
        ADD COLUMN `is_adopted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''是否采用 0否 1是'' AFTER `ai_generated_at`,
        ADD COLUMN `adopted_by` BIGINT(20) NULL DEFAULT NULL COMMENT ''采用人ID'' AFTER `is_adopted`,
        ADD COLUMN `adopted_at` DATETIME NULL DEFAULT NULL COMMENT ''采用时间'' AFTER `adopted_by`,
        ADD INDEX `idx_ai_generated` (`is_ai_generated`),
        ADD INDEX `idx_is_adopted` (`is_adopted`)',
    'SELECT ''AI columns already exist'' AS msg');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 3. AI生成记录表（已有则跳过）
CREATE TABLE IF NOT EXISTS `ai_generation_log` (
    `id` BIGINT(20) NOT NULL COMMENT '主键ID',
    `case_id` BIGINT(20) NOT NULL COMMENT '案件ID',
    `record_type` VARCHAR(50) NOT NULL COMMENT '记录类型: mediation_protocol/meeting_minutes/summary',
    `ref_id` BIGINT(20) NULL DEFAULT NULL COMMENT '关联业务ID',
    `original_params` TEXT NULL COMMENT '原始参数JSON',
    `ai_content` LONGTEXT NULL COMMENT 'AI生成内容',
    `legal_basis` TEXT NULL COMMENT '引用法律依据JSON',
    `tokens_used` INT(11) NOT NULL DEFAULT 0 COMMENT '消耗token数',
    `cost_time` INT(11) NOT NULL DEFAULT 0 COMMENT '耗时(毫秒)',
    `created_by` BIGINT(20) NULL DEFAULT NULL COMMENT '创建人ID',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_case_id` (`case_id`),
    INDEX `idx_record_type` (`record_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI生成日志表';
