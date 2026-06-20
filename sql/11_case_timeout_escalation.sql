-- =====================================================
-- 需求四：案件超时自动催办与升级
-- =====================================================

-- 1. 扩展 dispute_case 表：添加最后进展时间字段（用于判断调解中7天无进展）
ALTER TABLE dispute_case 
ADD COLUMN last_progress_time DATETIME DEFAULT NULL COMMENT '最后进展时间（调解记录、状态变更等操作更新）' 
AFTER escalate_time;

-- 2. 创建案件升级记录表（升级留痕备查）
CREATE TABLE IF NOT EXISTS dispute_escalation (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    escalate_type TINYINT NOT NULL DEFAULT 1 COMMENT '升级类型: 1-待分派超时 2-调解中超时无进展',
    from_level TINYINT DEFAULT 0 COMMENT '原处理级别: 0-未分派 1-调解员 2-组长 3-主任',
    to_level TINYINT NOT NULL COMMENT '升级到级别: 1-组长 2-主任 3-领导',
    from_user_id BIGINT DEFAULT 0 COMMENT '原处理人ID（0表示未分派）',
    from_user_name VARCHAR(64) DEFAULT '' COMMENT '原处理人姓名',
    to_user_id BIGINT DEFAULT 0 COMMENT '升级后处理人ID',
    to_user_name VARCHAR(64) DEFAULT '' COMMENT '升级后处理人姓名',
    to_org_id BIGINT DEFAULT 0 COMMENT '升级后归属组织ID',
    to_org_name VARCHAR(100) DEFAULT '' COMMENT '升级后归属组织名称',
    reason VARCHAR(500) NOT NULL COMMENT '升级原因',
    urge_count INT DEFAULT 0 COMMENT '催办次数（升级前已催办几次）',
    first_urge_time DATETIME DEFAULT NULL COMMENT '首次催办时间',
    timeout_hours INT DEFAULT 0 COMMENT '超时小时数',
    operator_id BIGINT DEFAULT 0 COMMENT '操作人ID（0表示系统自动）',
    operator_name VARCHAR(64) DEFAULT '系统自动' COMMENT '操作人姓名',
    status TINYINT DEFAULT 10 COMMENT '状态: 10-待处理 20-处理中 30-已处理 40-已关闭',
    remark VARCHAR(500) DEFAULT '' COMMENT '备注',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_case_id(case_id),
    INDEX idx_case_no(case_no),
    INDEX idx_to_level(to_level),
    INDEX idx_status(status),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='案件升级记录表';

-- 3. 为 dispute_urge 表增加扩展字段（用于区分催办级别、是否已触发升级）
ALTER TABLE dispute_urge 
ADD COLUMN urge_level TINYINT DEFAULT 0 COMMENT '催办级别: 0-普通催办 1-超时催办' AFTER urge_type,
ADD COLUMN escalate_triggered TINYINT DEFAULT 0 COMMENT '是否已触发升级: 0-否 1-是' AFTER urge_level;
