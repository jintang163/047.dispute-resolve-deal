-- =====================================================
-- 视频调解室增强表 (TRTC + 云端录制 + 排队 + 会议纪要)
-- =====================================================

-- 28. 视频录制分段表
CREATE TABLE IF NOT EXISTS video_record_segment (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    room_id BIGINT NOT NULL COMMENT '房间ID',
    case_id BIGINT DEFAULT 0 COMMENT '案件ID',
    case_no VARCHAR(32) DEFAULT '' COMMENT '案件编号',
    task_id VARCHAR(128) DEFAULT '' COMMENT 'TRTC录制任务ID',
    segment_index INT DEFAULT 0 COMMENT '分段序号',
    segment_sec INT DEFAULT 600 COMMENT '分段时长(秒)',
    status TINYINT DEFAULT 1 COMMENT '状态: 1-录制中 2-已结束 3-失败',
    file_url VARCHAR(512) DEFAULT '' COMMENT '录制文件URL',
    file_size BIGINT DEFAULT 0 COMMENT '文件大小(字节)',
    start_time DATETIME DEFAULT NULL COMMENT '开始时间',
    end_time DATETIME DEFAULT NULL COMMENT '结束时间',
    start_time_ms BIGINT DEFAULT 0 COMMENT '开始时间(毫秒)',
    end_time_ms BIGINT DEFAULT 0 COMMENT '结束时间(毫秒)',
    duration_sec INT DEFAULT 0 COMMENT '实际时长(秒)',
    storage_path VARCHAR(512) DEFAULT '' COMMENT '存储路径',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_room_id(room_id),
    INDEX idx_task_id(task_id),
    INDEX idx_status(status),
    INDEX idx_case_id(case_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='视频录制分段表';

-- 29. 视频调解排队表
CREATE TABLE IF NOT EXISTS video_queue (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) DEFAULT '' COMMENT '案件编号',
    mediator_id BIGINT NOT NULL COMMENT '调解员ID',
    mediator_name VARCHAR(64) DEFAULT '' COMMENT '调解员姓名',
    party_name VARCHAR(64) DEFAULT '' COMMENT '排队人姓名',
    party_phone VARCHAR(20) DEFAULT '' COMMENT '排队人电话',
    party_user_id BIGINT DEFAULT 0 COMMENT '排队人用户ID',
    priority INT DEFAULT 3 COMMENT '优先级: 1-特急 2-紧急 3-一般',
    status TINYINT DEFAULT 1 COMMENT '状态: 1-排队中 2-已进入 3-已取消 4-超时',
    enqueue_time DATETIME DEFAULT NULL COMMENT '入队时间',
    dequeue_time DATETIME DEFAULT NULL COMMENT '出队时间',
    notify_count INT DEFAULT 0 COMMENT '通知次数',
    last_notify_time DATETIME DEFAULT NULL COMMENT '最后通知时间',
    remark VARCHAR(512) DEFAULT '' COMMENT '备注',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_case_id(case_id),
    INDEX idx_mediator_id(mediator_id),
    INDEX idx_status(status),
    INDEX idx_enqueue_time(enqueue_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='视频调解排队表';

-- 30. 视频会议纪要表
CREATE TABLE IF NOT EXISTS video_meeting_minutes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    room_id BIGINT NOT NULL COMMENT '房间ID',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) DEFAULT '' COMMENT '案件编号',
    meeting_title VARCHAR(256) DEFAULT '' COMMENT '会议标题',
    meeting_time DATETIME DEFAULT NULL COMMENT '会议时间',
    duration VARCHAR(64) DEFAULT '' COMMENT '会议时长',
    participants JSON COMMENT '参与人列表',
    summary TEXT COMMENT '会议概要',
    key_points JSON COMMENT '要点列表',
    dispute_focus JSON COMMENT '争议焦点',
    mediation_process TEXT COMMENT '调解过程',
    evidence_discussed JSON COMMENT '讨论的证据',
    agreement TEXT COMMENT '达成协议',
    next_steps JSON COMMENT '下一步行动',
    risk_points JSON COMMENT '风险提示',
    emotional_state TEXT COMMENT '情绪状态',
    mediator_advice TEXT COMMENT '调解员建议',
    transcript TEXT COMMENT '原始转录文本',
    ai_model VARCHAR(64) DEFAULT 'deepseek' COMMENT 'AI模型',
    tokens_used INT DEFAULT 0 COMMENT '消耗Token数',
    cost_time INT DEFAULT 0 COMMENT '耗时(ms)',
    is_approved TINYINT DEFAULT 0 COMMENT '是否审核: 0-未审核 1-已审核 2-已修改',
    approved_by BIGINT DEFAULT 0 COMMENT '审核人ID',
    approved_at DATETIME DEFAULT NULL COMMENT '审核时间',
    status TINYINT DEFAULT 1 COMMENT '状态: 1-已生成 2-已审核 3-已作废',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_room_id(room_id),
    INDEX idx_case_id(case_id),
    INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='视频会议纪要表';

-- 31. 视频房间表增加TRTC相关字段
ALTER TABLE video_room ADD COLUMN record_task_id VARCHAR(128) DEFAULT '' COMMENT 'TRTC录制任务ID' AFTER record_url;
ALTER TABLE video_room ADD COLUMN record_status TINYINT DEFAULT 0 COMMENT '录制状态: 0-未录制 1-录制中 2-已结束 3-失败' AFTER record_task_id;
ALTER TABLE video_room ADD COLUMN has_meeting_minutes TINYINT DEFAULT 0 COMMENT '是否有会议纪要: 0-否 1-是' AFTER record_status;
ALTER TABLE video_room ADD COLUMN trtc_room_id INT DEFAULT 0 COMMENT 'TRTC数字房间号' AFTER has_meeting_minutes;
ALTER TABLE video_room ADD COLUMN screen_share_user_id BIGINT DEFAULT 0 COMMENT '屏幕共享用户ID' AFTER trtc_room_id;
ALTER TABLE video_room ADD COLUMN virtual_bg_enabled TINYINT DEFAULT 0 COMMENT '虚拟背景是否启用' AFTER screen_share_user_id;
ALTER TABLE video_room ADD COLUMN beauty_enabled TINYINT DEFAULT 0 COMMENT '美颜是否启用' AFTER virtual_bg_enabled;
