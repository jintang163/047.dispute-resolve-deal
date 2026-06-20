-- 自动回访机器人功能表结构
-- 创建时间: 2024-06-20

-- 回访记录表
CREATE TABLE IF NOT EXISTS `callback_record` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `case_id` bigint NOT NULL COMMENT '案件ID',
  `case_no` varchar(50) DEFAULT NULL COMMENT '案件编号',
  `case_title` varchar(200) DEFAULT NULL COMMENT '案件标题',
  `applicant_id` bigint DEFAULT NULL COMMENT '申请人ID',
  `applicant_name` varchar(50) DEFAULT NULL COMMENT '申请人姓名',
  `applicant_phone` varchar(20) DEFAULT NULL COMMENT '申请人手机号',
  `task_id` varchar(100) DEFAULT NULL COMMENT '阿里云语音任务ID',
  `call_id` varchar(100) DEFAULT NULL COMMENT '通话ID',
  `status` int NOT NULL DEFAULT '10' COMMENT '回访状态:10待回访,20回访中,30回访成功,40回访失败,99已取消',
  `call_status` int DEFAULT '0' COMMENT '通话状态:0未呼叫,10振铃中,20已接听,30无人接听,40用户忙,50呼叫失败,60已挂断',
  `call_time` datetime DEFAULT NULL COMMENT '呼叫时间',
  `call_duration` int DEFAULT '0' COMMENT '通话时长(秒)',
  `retry_count` int NOT NULL DEFAULT '0' COMMENT '重试次数',
  `max_retry_count` int NOT NULL DEFAULT '3' COMMENT '最大重试次数',
  `next_retry_time` datetime DEFAULT NULL COMMENT '下次重试时间',
  `scheduled_time` datetime DEFAULT NULL COMMENT '计划回访时间',
  `transcript_text` text COMMENT '语音转文字内容',
  `sentiment_result` text COMMENT '情绪分析结果(JSON)',
  `sentiment_score` double DEFAULT '0' COMMENT '情绪评分(-1到1)',
  `emotion` varchar(20) DEFAULT NULL COMMENT '情绪分类:positive正面,neutral中性,negative负面',
  `performance_score` int DEFAULT '0' COMMENT '履约情况评分(1-5分)',
  `satisfaction_score` int DEFAULT '0' COMMENT '满意度评分(1-5分)',
  `recording_url` varchar(500) DEFAULT NULL COMMENT '录音文件URL',
  `recording_size` bigint DEFAULT '0' COMMENT '录音文件大小(字节)',
  `expire_at` datetime DEFAULT NULL COMMENT '录音过期时间(保留1年)',
  `result_data` json DEFAULT NULL COMMENT '回访结果数据(JSON)',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_case_id` (`case_id`),
  KEY `idx_case_no` (`case_no`),
  KEY `idx_applicant_id` (`applicant_id`),
  KEY `idx_applicant_phone` (`applicant_phone`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_call_id` (`call_id`),
  KEY `idx_status` (`status`),
  KEY `idx_call_status` (`call_status`),
  KEY `idx_scheduled_time` (`scheduled_time`),
  KEY `idx_next_retry_time` (`next_retry_time`),
  KEY `idx_expire_at` (`expire_at`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自动回访记录表';

-- 回访话术模板表
CREATE TABLE IF NOT EXISTS `callback_template` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `name` varchar(100) NOT NULL COMMENT '模板名称',
  `code` varchar(50) NOT NULL COMMENT '模板编码',
  `type` int NOT NULL DEFAULT '1' COMMENT '模板类型:1满意度回访,2履约回访,3通用回访',
  `welcome_text` varchar(500) DEFAULT NULL COMMENT '欢迎语',
  `question_flow` json DEFAULT NULL COMMENT '问题流程配置(JSON)',
  `end_text` varchar(500) DEFAULT NULL COMMENT '结束语',
  `voice_type` varchar(50) DEFAULT 'xiaoyun' COMMENT '语音音色',
  `speed` int DEFAULT '0' COMMENT '语速:-500到500',
  `volume` int DEFAULT '0' COMMENT '音量:-500到500',
  `pause_time` int DEFAULT '800' COMMENT '句间停顿(毫秒)',
  `is_default` tinyint NOT NULL DEFAULT '0' COMMENT '是否默认:0否,1是',
  `status` int NOT NULL DEFAULT '1' COMMENT '状态:0禁用,1启用',
  `created_by` bigint DEFAULT NULL COMMENT '创建人',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='回访话术模板表';

-- 插入默认回访模板
INSERT INTO `callback_template` (`id`, `name`, `code`, `type`, `welcome_text`, `question_flow`, `end_text`, `voice_type`, `speed`, `volume`, `pause_time`, `is_default`, `status`, `created_by`, `created_at`) VALUES
(1, '案件结案回访模板', 'CASE_CLOSE_CALLBACK', 1, '您好，我是XX调解中心的智能回访机器人，耽误您2分钟时间，对您之前的调解案件做一个简单回访，可以吗？', '{\"questions\": [{\"id\": \"q1\", \"text\": \"请问调解协议中约定的内容，对方是否已经履行了呢？\", \"type\": \"choice\", \"options\": [{\"key\": \"1\", \"text\": \"已全部履行\", \"next\": \"q2\"}, {\"key\": \"2\", \"text\": \"部分履行\", \"next\": \"q2\"}, {\"key\": \"3\", \"text\": \"未履行\", \"next\": \"q2\"}]}, {\"id\": \"q2\", \"text\": \"请您对本次调解服务的满意度进行评分，1到5分，5分表示非常满意，请问您打几分？\", \"type\": \"number\", \"min\": 1, \"max\": 5, \"next\": \"end\"}]}', '感谢您的配合，祝您生活愉快，再见！', 'xiaoyun', 0, 0, 800, 1, 1, 1, NOW()),
(2, '履约情况回访模板', 'PERFORMANCE_CALLBACK', 2, '您好，我是XX调解中心的智能回访机器人，关于您之前的调解案件，想了解一下协议履行情况，耽误您1分钟时间可以吗？', '{\"questions\": [{\"id\": \"q1\", \"text\": \"请问调解协议约定的内容，目前履行情况如何？\", \"type\": \"choice\", \"options\": [{\"key\": \"1\", \"text\": \"已全部履行完毕\", \"next\": \"q2\"}, {\"key\": \"2\", \"text\": \"正在履行中\", \"next\": \"q2\"}, {\"key\": \"3\", \"text\": \"对方未按约定履行\", \"next\": \"q3\"}, {\"key\": \"4\", \"text\": \"还到履行时间\", \"next\": \"end\"}]}, {\"id\": \"q2\", \"text\": \"好的，感谢您的反馈，请问还有其他问题需要我们协助吗？\", \"type\": \"choice\", \"options\": [{\"key\": \"1\", \"text\": \"没有了\", \"next\": \"end\"}, {\"key\": \"2\", \"text\": \"需要帮助\", \"next\": \"end\"}]}, {\"id\": \"q3\", \"text\": \"了解了，如果对方未按约定履行，您可以向法院申请司法确认或强制执行，需要我们为您提供相关指引吗？\", \"type\": \"choice\", \"options\": [{\"key\": \"1\", \"text\": \"需要\", \"next\": \"end\"}, {\"key\": \"2\", \"text\": \"暂时不需要\", \"next\": \"end\"}]}]}', '感谢您的配合，如有需要请随时联系我们，再见！', 'xiaoyun', 0, 0, 800, 0, 1, 1, NOW());
