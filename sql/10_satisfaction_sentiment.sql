-- =====================================================
-- 满意度评价自动分析 - 数据库扩展
-- =====================================================

-- 1. 扩展dispute_case表，增加情感分析字段
ALTER TABLE dispute_case
    ADD COLUMN sentiment_emotion VARCHAR(20) DEFAULT '' COMMENT '情感分类: positive-正面 neutral-中性 negative-负面' AFTER satisfaction_comment,
    ADD COLUMN sentiment_score DECIMAL(5,4) DEFAULT 0.0000 COMMENT '情感评分: -1.0~1.0' AFTER sentiment_emotion,
    ADD COLUMN sentiment_confidence DECIMAL(5,4) DEFAULT 0.0000 COMMENT '分析置信度: 0~1' AFTER sentiment_score,
    ADD COLUMN sentiment_keywords JSON DEFAULT NULL COMMENT '情感关键词(JSON)' AFTER sentiment_confidence,
    ADD COLUMN sentiment_summary VARCHAR(500) DEFAULT '' COMMENT '情感分析摘要' AFTER sentiment_keywords,
    ADD COLUMN sentiment_analyzed_at DATETIME DEFAULT NULL COMMENT '情感分析时间' AFTER sentiment_summary;

-- 2. 创建改进工单表
CREATE TABLE IF NOT EXISTS improvement_order (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    order_no VARCHAR(32) NOT NULL UNIQUE COMMENT '工单编号',
    case_id BIGINT NOT NULL COMMENT '关联案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '关联案件编号',
    case_title VARCHAR(256) DEFAULT '' COMMENT '案件标题',
    applicant_id BIGINT DEFAULT 0 COMMENT '申请人ID',
    applicant_name VARCHAR(64) DEFAULT '' COMMENT '申请人姓名',
    mediator_id BIGINT NOT NULL COMMENT '调解员ID',
    mediator_name VARCHAR(64) DEFAULT '' COMMENT '调解员姓名',
    org_id BIGINT DEFAULT 0 COMMENT '组织ID',
    org_name VARCHAR(100) DEFAULT '' COMMENT '组织名称',
    satisfaction_score INT DEFAULT 0 COMMENT '满意度评分',
    satisfaction_comment TEXT COMMENT '满意度评语',
    sentiment_emotion VARCHAR(20) DEFAULT '' COMMENT '情感分类',
    sentiment_score DECIMAL(5,4) DEFAULT 0.0000 COMMENT '情感评分',
    sentiment_summary VARCHAR(500) DEFAULT '' COMMENT '情感分析摘要',
    issue_type VARCHAR(50) DEFAULT '' COMMENT '问题类型: attitude-态度问题 efficiency-效率问题 professional-专业性问题 result-结果不满意 process-流程问题 other-其他',
    issue_description TEXT COMMENT '问题描述(AI生成)',
    improvement_suggestion TEXT COMMENT '改进建议(AI生成)',
    status INT DEFAULT 10 COMMENT '状态: 10-待整改 20-整改中 30-已整改 40-已审核 99-已关闭',
    priority INT DEFAULT 2 COMMENT '优先级: 1-高 2-中 3-低',
    deadline DATETIME DEFAULT NULL COMMENT '整改截止时间',
    assigned_at DATETIME DEFAULT NULL COMMENT '指派时间',
    rectify_content TEXT COMMENT '整改内容',
    rectify_result TEXT COMMENT '整改结果',
    rectified_at DATETIME DEFAULT NULL COMMENT '整改完成时间',
    review_opinion VARCHAR(500) DEFAULT '' COMMENT '审核意见',
    reviewed_by BIGINT DEFAULT 0 COMMENT '审核人ID',
    reviewed_by_name VARCHAR(64) DEFAULT '' COMMENT '审核人姓名',
    reviewed_at DATETIME DEFAULT NULL COMMENT '审核时间',
    deduction_score DECIMAL(10,2) DEFAULT 0.00 COMMENT '绩效扣分',
    deduction_reason VARCHAR(500) DEFAULT '' COMMENT '扣分原因',
    is_deduction_applied TINYINT DEFAULT 0 COMMENT '是否已扣分: 0-否 1-是',
    deduction_applied_at DATETIME DEFAULT NULL COMMENT '扣分生效时间',
    remark VARCHAR(500) DEFAULT '' COMMENT '备注',
    created_by BIGINT DEFAULT 0 COMMENT '创建人ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    INDEX idx_case_id(case_id),
    INDEX idx_case_no(case_no),
    INDEX idx_mediator_id(mediator_id),
    INDEX idx_status(status),
    INDEX idx_deadline(deadline),
    INDEX idx_sentiment_emotion(sentiment_emotion),
    INDEX idx_is_deduction_applied(is_deduction_applied),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='改进工单表';

-- 3. 扩展performance_stat表，增加扣分相关字段
ALTER TABLE performance_stat
    ADD COLUMN deduction_total DECIMAL(10,2) DEFAULT 0.00 COMMENT '累计扣分' AFTER grade,
    ADD COLUMN deduction_count INT DEFAULT 0 COMMENT '扣分次数' AFTER deduction_total,
    ADD COLUMN final_score DECIMAL(10,2) DEFAULT 0.00 COMMENT '最终得分(原始分-扣分)' AFTER deduction_count;
