-- =====================================================
-- 典型案例库与知识沉淀 - 数据库迁移脚本
-- 创建时间: 2026-06-21
-- =====================================================

USE dispute_resolve;

-- 典型案例库表
CREATE TABLE IF NOT EXISTS case_library (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    case_no VARCHAR(64) NOT NULL UNIQUE COMMENT '案例编号',
    title VARCHAR(256) NOT NULL COMMENT '案例标题(脱敏)',
    description TEXT COMMENT '案例描述(脱敏)',
    dispute_type VARCHAR(128) DEFAULT '' COMMENT '纠纷类型',
    type_id BIGINT DEFAULT 0 COMMENT '纠纷类型ID',
    mediation_tactics TEXT COMMENT '调解话术/策略',
    key_points TEXT COMMENT '调解要点/关键经验',
    result_summary TEXT COMMENT '调解结果摘要',
    difficulty_level TINYINT DEFAULT 1 COMMENT '难度等级: 1-简单 2-一般 3-复杂 4-疑难',
    is_success TINYINT DEFAULT 1 COMMENT '是否调解成功: 0-否 1-是',
    mediator_name VARCHAR(64) DEFAULT '' COMMENT '调解员(脱敏)',
    mediator_id BIGINT DEFAULT 0 COMMENT '调解员ID',
    org_name VARCHAR(128) DEFAULT '' COMMENT '调解组织(脱敏)',
    org_id BIGINT DEFAULT 0 COMMENT '组织ID',
    source_case_id BIGINT DEFAULT 0 COMMENT '来源案件ID',
    keywords VARCHAR(500) DEFAULT '' COMMENT '关键词标签',
    tags VARCHAR(500) DEFAULT '' COMMENT '分类标签',
    vector_id VARCHAR(100) DEFAULT '' COMMENT 'Milvus向量ID',
    vector_status TINYINT DEFAULT 0 COMMENT '向量化状态: 0-未处理 1-处理中 2-已完成 3-失败',
    reference_count INT DEFAULT 0 COMMENT '引用次数',
    avg_score DECIMAL(3,2) DEFAULT 0.00 COMMENT '平均有用性评分(1-5)',
    score_count INT DEFAULT 0 COMMENT '评分总次数',
    total_score DECIMAL(10,2) DEFAULT 0.00 COMMENT '评分总分',
    last_used_at DATETIME DEFAULT NULL COMMENT '最后使用时间',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用 2-已归档',
    archived_at DATETIME DEFAULT NULL COMMENT '归档时间',
    created_by BIGINT DEFAULT 0 COMMENT '创建人ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_case_no(case_no),
    INDEX idx_type_id(type_id),
    INDEX idx_dispute_type(dispute_type),
    INDEX idx_difficulty(difficulty_level),
    INDEX idx_status(status),
    INDEX idx_vector_status(vector_status),
    INDEX idx_avg_score(avg_score),
    INDEX idx_last_used(last_used_at),
    INDEX idx_created_at(created_at),
    INDEX idx_deleted_at(deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='典型案例库';

-- 案例评分记录表
CREATE TABLE IF NOT EXISTS case_library_score (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    case_id BIGINT NOT NULL COMMENT '案例ID',
    case_no VARCHAR(64) DEFAULT '' COMMENT '案例编号',
    user_id BIGINT NOT NULL COMMENT '评分人ID',
    user_name VARCHAR(64) DEFAULT '' COMMENT '评分人姓名',
    score TINYINT NOT NULL COMMENT '有用性评分: 1-5分',
    source_case_id BIGINT DEFAULT 0 COMMENT '使用场景案件ID',
    comment VARCHAR(500) DEFAULT '' COMMENT '评分备注',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE KEY uk_case_user(case_id, user_id, source_case_id),
    INDEX idx_case_id(case_id),
    INDEX idx_user_id(user_id),
    INDEX idx_score(score),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='案例评分记录';

-- 案例引用记录表
CREATE TABLE IF NOT EXISTS case_library_quote (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    source_case_id BIGINT NOT NULL COMMENT '引用的目标案例ID',
    library_case_id BIGINT NOT NULL COMMENT '被引用的案例库ID',
    library_case_no VARCHAR(64) DEFAULT '' COMMENT '被引用的案例编号',
    quote_type TINYINT DEFAULT 1 COMMENT '引用类型: 1-话术引用 2-策略引用 3-全文引用',
    quote_content TEXT COMMENT '引用的具体内容',
    user_id BIGINT NOT NULL COMMENT '操作人ID',
    user_name VARCHAR(64) DEFAULT '' COMMENT '操作人姓名',
    mediation_record_id BIGINT DEFAULT 0 COMMENT '关联的调解记录ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_source_case(source_case_id),
    INDEX idx_library_case(library_case_id),
    INDEX idx_user_id(user_id),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='案例引用记录';

-- 案例历史归档表
CREATE TABLE IF NOT EXISTS case_library_archive (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    original_id BIGINT NOT NULL COMMENT '原案例ID',
    case_no VARCHAR(64) DEFAULT '' COMMENT '案例编号',
    title VARCHAR(256) DEFAULT '' COMMENT '案例标题',
    archive_reason TINYINT DEFAULT 1 COMMENT '归档原因: 1-超1年未使用 2-手动归档 3-评分过低',
    avg_score DECIMAL(3,2) DEFAULT 0.00 COMMENT '归档时平均评分',
    reference_count INT DEFAULT 0 COMMENT '归档时引用次数',
    last_used_at DATETIME DEFAULT NULL COMMENT '最后使用时间',
    archived_by BIGINT DEFAULT 0 COMMENT '归档操作人ID(0为系统自动)',
    case_data JSON COMMENT '归档时案例完整数据快照',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_original_id(original_id),
    INDEX idx_case_no(case_no),
    INDEX idx_archive_reason(archive_reason),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='案例历史归档';
