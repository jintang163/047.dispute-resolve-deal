-- =============================================
-- 调解协议AI智能生成功能增强
-- =============================================

-- 1. 为现有 mediation_protocol 表新增AI生成与采用相关字段
ALTER TABLE `mediation_protocol`
    ADD COLUMN `is_ai_generated` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否AI生成 0否 1是' AFTER `created_by`,
    ADD COLUMN `ai_generated_at` DATETIME NULL DEFAULT NULL COMMENT 'AI生成时间' AFTER `is_ai_generated`,
    ADD COLUMN `is_adopted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否采用 0否 1是' AFTER `ai_generated_at`,
    ADD COLUMN `adopted_by` BIGINT(20) NULL DEFAULT NULL COMMENT '采用人ID' AFTER `is_adopted`,
    ADD COLUMN `adopted_at` DATETIME NULL DEFAULT NULL COMMENT '采用时间' AFTER `adopted_by`,
    ADD INDEX `idx_ai_generated` (`is_ai_generated`),
    ADD INDEX `idx_is_adopted` (`is_adopted`);

-- 2. AI生成记录表（已有则跳过）
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
