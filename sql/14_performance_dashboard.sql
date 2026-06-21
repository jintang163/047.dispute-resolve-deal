-- =====================================================
-- 绩效考核看板增强 - 数据库迁移脚本
-- 创建时间: 2026-06-21
-- =====================================================

USE dispute_resolve;

-- 绩效面谈记录表
CREATE TABLE IF NOT EXISTS performance_interview (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    interview_no VARCHAR(64) NOT NULL UNIQUE COMMENT '面谈编号',
    score_id BIGINT DEFAULT 0 COMMENT '关联考核记录ID',
    user_id BIGINT NOT NULL COMMENT '被考核人ID(调解员)',
    user_name VARCHAR(64) DEFAULT '' COMMENT '被考核人姓名',
    org_id BIGINT DEFAULT 0 COMMENT '组织ID',
    period_type TINYINT DEFAULT 1 COMMENT '周期类型: 1-月度 2-季度 3-年度',
    period_value VARCHAR(32) DEFAULT '' COMMENT '周期值',
    total_score DECIMAL(10,2) DEFAULT 0.00 COMMENT '考核得分',
    level VARCHAR(16) DEFAULT '' COMMENT '考核等级',
    interviewer_id BIGINT NOT NULL COMMENT '面谈人ID',
    interviewer_name VARCHAR(64) DEFAULT '' COMMENT '面谈人姓名',
    interview_time DATETIME NOT NULL COMMENT '面谈时间',
    interview_place VARCHAR(256) DEFAULT '' COMMENT '面谈地点',
    interview_type TINYINT DEFAULT 1 COMMENT '面谈类型: 1-绩效反馈 2-改进计划 3-表彰面谈 4-预警面谈',
    strengths TEXT COMMENT '工作亮点',
    weaknesses TEXT COMMENT '待改进方面',
    improvement_plan TEXT COMMENT '改进计划',
    target_next_period TEXT COMMENT '下期目标',
    mediator_comment TEXT COMMENT '调解员自评/反馈',
    status TINYINT DEFAULT 1 COMMENT '状态: 1-待确认 2-已确认 3-已归档',
    confirmed_at DATETIME DEFAULT NULL COMMENT '调解员确认时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_user_id(user_id),
    INDEX idx_interviewer_id(interviewer_id),
    INDEX idx_period(period_type, period_value),
    INDEX idx_status(status),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='绩效面谈记录表';

-- 绩效月度聚合快照表
CREATE TABLE IF NOT EXISTS performance_monthly_snapshot (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    user_id BIGINT NOT NULL COMMENT '调解员ID',
    user_name VARCHAR(64) DEFAULT '' COMMENT '调解员姓名',
    org_id BIGINT DEFAULT 0 COMMENT '组织ID',
    org_name VARCHAR(128) DEFAULT '' COMMENT '组织名称',
    year INT NOT NULL COMMENT '年份',
    month INT NOT NULL COMMENT '月份',
    case_count INT DEFAULT 0 COMMENT '受理数',
    closed_count INT DEFAULT 0 COMMENT '办结数',
    close_rate DECIMAL(5,2) DEFAULT 0.00 COMMENT '办结率(%)',
    success_count INT DEFAULT 0 COMMENT '调解成功数',
    success_rate DECIMAL(5,2) DEFAULT 0.00 COMMENT '成功率(%)',
    avg_days DECIMAL(10,2) DEFAULT 0.00 COMMENT '平均调解天数',
    avg_satisfaction DECIMAL(3,2) DEFAULT 0.00 COMMENT '平均满意度(1-5)',
    urge_count INT DEFAULT 0 COMMENT '被催办次数',
    total_score DECIMAL(10,2) DEFAULT 0.00 COMMENT '综合得分',
    level VARCHAR(16) DEFAULT '' COMMENT '等级',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_user_month(user_id, year, month),
    INDEX idx_org_id(org_id),
    INDEX idx_year_month(year, month),
    INDEX idx_total_score(total_score)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='绩效月度聚合快照表';

-- 初始化绩效指标配置数据
INSERT IGNORE INTO performance_indicator_config (indicator_code, indicator_name, indicator_type, weight, max_score, calculation_formula, target_value, description, sort_order, status)
VALUES
('CASE_COUNT', '受理数', 1, 0.20, 100.00, '受理案件数/目标值*满分', 30.00, '月度受理案件数量', 1, 1),
('SUCCESS_RATE', '调解成功率', 2, 0.25, 100.00, '成功案件数/已结案件数*100', 90.00, '调解成功占已结案比例', 2, 1),
('AVG_DAYS', '平均调解时长', 3, 0.20, 100.00, '逆向指标: 7天内满分，超过30天最低分', 7.00, '从受理到结案的平均天数', 3, 1),
('SATISFACTION', '满意度', 4, 0.20, 100.00, '平均满意度评分/5*100', 4.50, '群众满意度评分均值(1-5)', 4, 1),
('URGE_COUNT', '被催办次数', 3, 0.15, 100.00, '逆向指标: 0次满分，超过5次最低分', 0.00, '被催办的次数(越少越好)', 5, 1);
