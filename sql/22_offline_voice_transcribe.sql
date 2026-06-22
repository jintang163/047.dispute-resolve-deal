-- =====================================================
-- 离线语音转文字功能 - 数据库脚本
-- 创建时间: 2026-06-22
-- 包含: 语音转写任务表、调解记录表新增字段
-- =====================================================

USE dispute_resolve;

-- =====================================================
-- 一、语音转写任务表
-- =====================================================
CREATE TABLE IF NOT EXISTS voice_transcribe_task (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    task_id VARCHAR(128) NOT NULL UNIQUE COMMENT '通义听悟任务ID或本地任务ID',
    task_type TINYINT DEFAULT 1 COMMENT '任务类型: 1-实时短语音 2-长音频转写',
    status TINYINT DEFAULT 0 COMMENT '状态: 0-待提交 1-转写中 2-成功 3-失败',
    file_name VARCHAR(255) COMMENT '文件名',
    file_url VARCHAR(500) COMMENT 'MinIO/OSS地址',
    file_size BIGINT COMMENT '文件大小(byte)',
    format VARCHAR(20) COMMENT '音频格式: mp3/wav/webm等',
    duration INT COMMENT '音频时长(秒)',
    transcript_text TEXT COMMENT '转写结果文本',
    speaker_count INT DEFAULT 0 COMMENT '说话人数量',
    diarization JSON COMMENT '说话人分离结果(JSON)',
    word_count INT COMMENT '字数统计',
    error_msg VARCHAR(500) COMMENT '错误信息',
    case_id BIGINT COMMENT '关联案件ID',
    record_id BIGINT COMMENT '关联调解记录ID',
    created_by BIGINT COMMENT '创建人ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    completed_at DATETIME COMMENT '完成时间',
    is_deleted TINYINT DEFAULT 0 COMMENT '是否删除: 0-否 1-是',
    INDEX idx_task_id(task_id),
    INDEX idx_status(status),
    INDEX idx_case_id(case_id),
    INDEX idx_record_id(record_id),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='语音转写任务表';

-- =====================================================
-- 二、调解记录表新增字段
-- =====================================================
ALTER TABLE dispute_mediation_record
    ADD COLUMN IF NOT EXISTS audio_url VARCHAR(500) DEFAULT '' COMMENT '录音文件地址' AFTER ai_summary,
    ADD COLUMN IF NOT EXISTS audio_duration INT DEFAULT 0 COMMENT '录音时长(秒)' AFTER audio_url,
    ADD COLUMN IF NOT EXISTS audio_file_size BIGINT DEFAULT 0 COMMENT '录音文件大小(byte)' AFTER audio_duration,
    ADD COLUMN IF NOT EXISTS transcript_text TEXT COMMENT '语音转写文本' AFTER audio_file_size,
    ADD COLUMN IF NOT EXISTS transcribe_status TINYINT DEFAULT 0 COMMENT '转写状态: 0-未转写 1-转写中 2-成功 3-失败' AFTER transcript_text,
    ADD COLUMN IF NOT EXISTS transcribe_task_id VARCHAR(128) DEFAULT '' COMMENT '转写任务ID' AFTER transcribe_status,
    ADD COLUMN IF NOT EXISTS transcribe_at DATETIME COMMENT '转写完成时间' AFTER transcribe_task_id,
    ADD INDEX IF NOT EXISTS idx_transcribe_status(transcribe_status),
    ADD INDEX IF NOT EXISTS idx_transcribe_task_id(transcribe_task_id);
