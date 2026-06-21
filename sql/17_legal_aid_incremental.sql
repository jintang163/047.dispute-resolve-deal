-- =====================================================
-- 法律援助转介 - 增量迁移脚本 v2
-- 创建时间: 2026-06-21
-- 说明: 添加法援材料表、申请表扩展字段
-- =====================================================

USE dispute_resolve;

-- =====================================================
-- 1. 法援证明材料表
-- =====================================================
CREATE TABLE IF NOT EXISTS legal_aid_material (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    case_id BIGINT DEFAULT 0 COMMENT '案件ID',
    application_id VARCHAR(64) DEFAULT '' COMMENT '关联的法援申请编号',
    material_type TINYINT DEFAULT 1 COMMENT '材料类型: 1-低保证明 2-收入证明 3-身份证 4-户口本 5-残疾证明 6-病例材料 7-其他证明',
    material_name VARCHAR(64) DEFAULT '' COMMENT '材料名称',
    file_name VARCHAR(256) DEFAULT '' COMMENT '文件原名',
    file_path VARCHAR(512) DEFAULT '' COMMENT '文件存储路径',
    file_url VARCHAR(512) DEFAULT '' COMMENT '文件访问URL',
    file_size BIGINT DEFAULT 0 COMMENT '文件大小(字节)',
    file_ext VARCHAR(16) DEFAULT '' COMMENT '文件扩展名',
    mime_type VARCHAR(128) DEFAULT '' COMMENT '文件MIME类型',
    uploader_id BIGINT DEFAULT 0 COMMENT '上传人ID',
    uploader_name VARCHAR(64) DEFAULT '' COMMENT '上传人姓名',
    organization_id BIGINT DEFAULT 0 COMMENT '所属机构ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_case_id(case_id),
    INDEX idx_application_id(application_id),
    INDEX idx_material_type(material_type),
    INDEX idx_uploader_id(uploader_id),
    INDEX idx_deleted_at(deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='法援证明材料表';

-- =====================================================
-- 2. 为法援申请表添加转介关联字段
-- =====================================================
ALTER TABLE legal_aid_application 
    ADD COLUMN IF NOT EXISTS transfer_id BIGINT DEFAULT 0 COMMENT '转介记录ID' AFTER submit_time,
    ADD COLUMN IF NOT EXISTS transfer_no VARCHAR(64) DEFAULT '' COMMENT '转介编号' AFTER transfer_id,
    ADD COLUMN IF NOT EXISTS transferred TINYINT DEFAULT 0 COMMENT '是否已转介: 0-否 1-是' AFTER transfer_no,
    ADD COLUMN IF NOT EXISTS transferred_at DATETIME DEFAULT NULL COMMENT '转介时间' AFTER transferred,
    ADD INDEX IF NOT EXISTS idx_transfer_id(transfer_id),
    ADD INDEX IF NOT EXISTS idx_transferred(transferred);

-- =====================================================
-- 初始化说明
-- =====================================================
-- 执行本脚本前请确保已执行 16_legal_aid_transfer.sql
