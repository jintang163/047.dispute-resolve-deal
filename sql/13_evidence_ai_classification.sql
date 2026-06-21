-- 证据材料AI智能分类字段
ALTER TABLE dispute_evidence
ADD COLUMN IF NOT EXISTS evidence_category TINYINT DEFAULT 0 COMMENT '证据类别: 0-未分类 1-身份证 2-合同协议 3-收据凭证 4-现场照片 5-聊天记录 6-录音录像 7-发票票据 8-证件证明 9-其他',
ADD COLUMN IF NOT EXISTS ai_category TINYINT DEFAULT 0 COMMENT 'AI识别类别',
ADD COLUMN IF NOT EXISTS ai_confidence DECIMAL(5,4) DEFAULT 0 COMMENT 'AI识别置信度 0-1',
ADD COLUMN IF NOT EXISTS ai_keywords VARCHAR(256) DEFAULT '' COMMENT 'AI识别关键词',
ADD COLUMN IF NOT EXISTS ai_processed TINYINT DEFAULT 0 COMMENT 'AI是否已处理: 0-未处理 1-处理中 2-处理完成 3-处理失败',
ADD COLUMN IF NOT EXISTS ai_processed_at DATETIME DEFAULT NULL COMMENT 'AI处理时间',
ADD COLUMN IF NOT EXISTS manual_category TINYINT DEFAULT 0 COMMENT '手动修正类别',
ADD COLUMN IF NOT EXISTS manual_updated_at DATETIME DEFAULT NULL COMMENT '手动修正时间',
ADD COLUMN IF NOT EXISTS manual_updated_by BIGINT DEFAULT 0 COMMENT '手动修正人ID',
ADD INDEX IF NOT EXISTS idx_evidence_category(evidence_category),
ADD INDEX IF NOT EXISTS idx_ai_processed(ai_processed);
