-- 案件批量导出（加密）功能
-- 2026-06-21

-- 数据导出记录表
CREATE TABLE IF NOT EXISTS `data_export_log` (
  `id` bigint(20) NOT NULL COMMENT '主键ID',
  `export_no` varchar(50) NOT NULL COMMENT '导出单号',
  `export_type` tinyint(4) NOT NULL DEFAULT 1 COMMENT '导出类型：1-案件数据，2-考核数据，3-证据数据，4-其他',
  `export_name` varchar(200) NOT NULL COMMENT '导出名称',
  `filter_conditions` text COMMENT '筛选条件JSON',
  `record_count` int(11) DEFAULT 0 COMMENT '导出记录数',
  `file_name` varchar(255) DEFAULT NULL COMMENT '文件名',
  `file_path` varchar(500) DEFAULT NULL COMMENT '文件存储路径',
  `file_size` bigint(20) DEFAULT 0 COMMENT '文件大小（字节）',
  `encryption_algorithm` varchar(50) DEFAULT 'AES-256-GCM' COMMENT '加密算法',
  `password_sms_sent` tinyint(4) DEFAULT 0 COMMENT '密码短信是否已发送：0-未发送，1-已发送，2-发送失败',
  `password_sms_time` datetime DEFAULT NULL COMMENT '密码短信发送时间',
  `export_status` tinyint(4) NOT NULL DEFAULT 10 COMMENT '导出状态：10-导出中，20-导出成功，30-导出失败',
  `error_message` text COMMENT '错误信息',
  `operator_id` bigint(20) NOT NULL COMMENT '操作人ID',
  `operator_name` varchar(50) NOT NULL COMMENT '操作人姓名',
  `operator_phone` varchar(20) DEFAULT NULL COMMENT '操作人手机号',
  `org_id` bigint(20) DEFAULT NULL COMMENT '所属机构ID',
  `ip_address` varchar(50) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` varchar(500) DEFAULT NULL COMMENT 'UserAgent',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `expired_at` datetime NOT NULL COMMENT '过期时间（7天后自动删除）',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_export_no` (`export_no`),
  KEY `idx_operator_id` (`operator_id`),
  KEY `idx_export_status` (`export_status`),
  KEY `idx_export_type` (`export_type`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_expired_at` (`expired_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据导出记录表';

-- 定时清理：过期7天的导出文件（由cron任务执行）
