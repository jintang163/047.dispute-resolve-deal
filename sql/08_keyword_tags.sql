USE dispute_resolve;

ALTER TABLE dispute_case
    ADD COLUMN keywords JSON DEFAULT NULL COMMENT 'AI自动提取的关键词标签(JSON数组)'
    AFTER description;

ALTER TABLE dispute_case
    ADD INDEX idx_keywords ((CAST(keywords AS CHAR(512) ARRAY)));

CREATE TABLE IF NOT EXISTS dispute_keyword_dict (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    keyword VARCHAR(64) NOT NULL COMMENT '关键词',
    category VARCHAR(32) DEFAULT NULL COMMENT '关键词分类(纠纷性质/行为/对象/程度)',
    frequency INT NOT NULL DEFAULT 1 COMMENT '出现频次',
    source_type VARCHAR(16) NOT NULL DEFAULT 'ai' COMMENT '来源类型 ai/manual',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_keyword (keyword),
    INDEX idx_category (category),
    INDEX idx_frequency (frequency DESC),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='纠纷关键词词典(供筛选和统计)';
