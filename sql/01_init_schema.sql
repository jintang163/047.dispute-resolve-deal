-- =====================================================
-- 综治中心矛盾纠纷管理系统 - 数据库初始化脚本
-- 数据库: TiDB 7.x
-- 创建时间: 2026-06-18
-- =====================================================

-- 切换数据库
CREATE DATABASE IF NOT EXISTS dispute_resolve DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE dispute_resolve;

-- =====================================================
-- 一、系统基础表
-- =====================================================

-- 1. 组织架构表
CREATE TABLE IF NOT EXISTS sys_organization (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    org_code VARCHAR(64) NOT NULL UNIQUE COMMENT '组织编码',
    org_name VARCHAR(128) NOT NULL COMMENT '组织名称',
    org_type TINYINT NOT NULL DEFAULT 1 COMMENT '组织类型: 1-综治中心 2-街道办 3-社区 4-村委会',
    parent_id BIGINT DEFAULT 0 COMMENT '父级组织ID',
    level_path VARCHAR(512) DEFAULT '' COMMENT '层级路径',
    leader_id BIGINT DEFAULT 0 COMMENT '负责人ID',
    address VARCHAR(256) DEFAULT '' COMMENT '地址',
    longitude DECIMAL(10,7) DEFAULT NULL COMMENT '经度',
    latitude DECIMAL(10,7) DEFAULT NULL COMMENT '纬度',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    INDEX idx_parent_id(parent_id),
    INDEX idx_status(status),
    INDEX idx_org_type(org_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='组织架构表';

-- 2. 用户表
CREATE TABLE IF NOT EXISTS sys_user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    username VARCHAR(64) NOT NULL UNIQUE COMMENT '用户名',
    password VARCHAR(256) NOT NULL COMMENT '密码(BCrypt加密)',
    real_name VARCHAR(64) NOT NULL COMMENT '真实姓名',
    phone VARCHAR(20) UNIQUE COMMENT '手机号',
    id_card VARCHAR(32) UNIQUE COMMENT '身份证号(加密)',
    avatar VARCHAR(256) DEFAULT '' COMMENT '头像',
    gender TINYINT DEFAULT 0 COMMENT '性别: 0-未知 1-男 2-女',
    email VARCHAR(128) DEFAULT '' COMMENT '邮箱',
    role TINYINT NOT NULL DEFAULT 3 COMMENT '角色: 1-主任 2-组长 3-调解员 4-管理员',
    organization_id BIGINT NOT NULL COMMENT '所属组织ID',
    position VARCHAR(64) DEFAULT '' COMMENT '职位',
    specialty VARCHAR(512) DEFAULT '' COMMENT '专业领域(逗号分隔)',
    longitude DECIMAL(10,7) DEFAULT NULL COMMENT '工作地点经度',
    latitude DECIMAL(10,7) DEFAULT NULL COMMENT '工作地点纬度',
    mediation_count INT DEFAULT 0 COMMENT '调解案件数',
    success_rate DECIMAL(5,2) DEFAULT 0.00 COMMENT '调解成功率',
    last_login_at DATETIME DEFAULT NULL COMMENT '最后登录时间',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    INDEX idx_org_id(organization_id),
    INDEX idx_role(role),
    INDEX idx_status(status),
    INDEX idx_phone(phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 3. 角色权限表
CREATE TABLE IF NOT EXISTS sys_role_permission (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    role_code VARCHAR(64) NOT NULL COMMENT '角色编码',
    role_name VARCHAR(64) NOT NULL COMMENT '角色名称',
    permissions TEXT COMMENT '权限列表(JSON数组)',
    description VARCHAR(256) DEFAULT '' COMMENT '描述',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_role_code(role_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限表';

-- 4. 操作日志表
CREATE TABLE IF NOT EXISTS sys_operation_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    user_id BIGINT NOT NULL COMMENT '操作人ID',
    username VARCHAR(64) DEFAULT '' COMMENT '用户名',
    operation VARCHAR(128) NOT NULL COMMENT '操作类型',
    module VARCHAR(64) DEFAULT '' COMMENT '模块',
    method VARCHAR(16) DEFAULT '' COMMENT '请求方法',
    url VARCHAR(256) DEFAULT '' COMMENT '请求URL',
    ip VARCHAR(64) DEFAULT '' COMMENT 'IP地址',
    params TEXT COMMENT '请求参数',
    result TEXT COMMENT '返回结果',
    cost_time INT DEFAULT 0 COMMENT '耗时(ms)',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-失败 1-成功',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_user_id(user_id),
    INDEX idx_created_at(created_at),
    INDEX idx_module(module)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志表';

-- =====================================================
-- 二、纠纷核心表
-- =====================================================

-- 5. 纠纷类型表 (三级分类)
CREATE TABLE IF NOT EXISTS dispute_type (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    type_code VARCHAR(64) NOT NULL UNIQUE COMMENT '类型编码',
    type_name VARCHAR(128) NOT NULL COMMENT '类型名称',
    parent_id BIGINT DEFAULT 0 COMMENT '父级ID',
    level TINYINT DEFAULT 1 COMMENT '层级: 1-一级 2-二级 3-三级',
    level_path VARCHAR(512) DEFAULT '' COMMENT '层级路径',
    icon VARCHAR(256) DEFAULT '' COMMENT '图标',
    sort_order INT DEFAULT 0 COMMENT '排序',
    description VARCHAR(512) DEFAULT '' COMMENT '描述',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_parent_id(parent_id),
    INDEX idx_level(level),
    INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='纠纷类型表';

-- 6. 纠纷案件主表
CREATE TABLE IF NOT EXISTS dispute_case (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    case_no VARCHAR(32) NOT NULL UNIQUE COMMENT '案件编号',
    title VARCHAR(256) NOT NULL COMMENT '纠纷标题',
    description TEXT COMMENT '纠纷简述',
    type_id BIGINT NOT NULL COMMENT '纠纷类型ID(三级)',
    type_path VARCHAR(512) DEFAULT '' COMMENT '类型路径',
    keywords JSON DEFAULT NULL COMMENT 'AI自动提取的关键词标签(JSON数组)',
    case_level TINYINT DEFAULT 3 COMMENT '紧急程度: 1-特急 2-紧急 3-一般 4-普通',
    case_source TINYINT NOT NULL COMMENT '来源: 1-自助终端 2-小程序 3-电话 4-窗口 5-转办',
    channel_id VARCHAR(64) DEFAULT '' COMMENT '渠道终端ID',
    status TINYINT DEFAULT 10 COMMENT '状态: 10-待分派 20-调解中 30-待审批 40-审批中 50-已结案 99-已取消',
    sub_status TINYINT DEFAULT 0 COMMENT '子状态',
    reporter_name VARCHAR(64) DEFAULT '' COMMENT '报案人姓名',
    reporter_phone VARCHAR(20) DEFAULT '' COMMENT '报案人电话',
    reporter_id_card VARCHAR(32) DEFAULT '' COMMENT '报案人身份证',
    reporter_address VARCHAR(256) DEFAULT '' COMMENT '报案人地址',
    respondent_name VARCHAR(64) DEFAULT '' COMMENT '对方姓名',
    respondent_phone VARCHAR(20) DEFAULT '' COMMENT '对方电话',
    respondent_address VARCHAR(256) DEFAULT '' COMMENT '对方地址',
    occur_address VARCHAR(256) DEFAULT '' COMMENT '发生地点',
    occur_time DATETIME DEFAULT NULL COMMENT '发生时间',
    expectation TEXT COMMENT '期望解决方式',
    longitude DECIMAL(10,7) DEFAULT NULL COMMENT '发生地经度',
    latitude DECIMAL(10,7) DEFAULT NULL COMMENT '发生地纬度',
    organization_id BIGINT NOT NULL COMMENT '归属组织ID',
    mediator_id BIGINT DEFAULT 0 COMMENT '调解员ID',
    mediator_name VARCHAR(64) DEFAULT '' COMMENT '调解员姓名',
    mediator_time DATETIME DEFAULT NULL COMMENT '分派时间',
    mediation_start_time DATETIME DEFAULT NULL COMMENT '调解开始时间',
    mediation_end_time DATETIME DEFAULT NULL COMMENT '调解结束时间',
    mediation_result TINYINT DEFAULT 0 COMMENT '调解结果: 0-未调解 1-达成协议 2-未达成协议 3-部分达成',
    agreement_content TEXT COMMENT '协议内容',
    satisfaction_score TINYINT DEFAULT 0 COMMENT '满意度评分: 1-5星',
    satisfaction_comment TEXT COMMENT '满意度评价',
    urgency_time DATETIME DEFAULT NULL COMMENT '催办时间',
    urgency_count INT DEFAULT 0 COMMENT '催办次数',
    escalate_level TINYINT DEFAULT 0 COMMENT '升级级别: 0-未升级 1-组长 2-主任 3-领导',
    escalate_time DATETIME DEFAULT NULL COMMENT '升级时间',
    is_video_mediation TINYINT DEFAULT 0 COMMENT '是否视频调解: 0-否 1-是',
    video_room_id VARCHAR(64) DEFAULT '' COMMENT '视频房间ID',
    video_start_time DATETIME DEFAULT NULL COMMENT '视频开始时间',
    video_end_time DATETIME DEFAULT NULL COMMENT '视频结束时间',
    has_esign TINYINT DEFAULT 0 COMMENT '是否电子签章: 0-否 1-是',
    esign_time DATETIME DEFAULT NULL COMMENT '签章时间',
    close_reason VARCHAR(512) DEFAULT '' COMMENT '结案原因',
    close_user_id BIGINT DEFAULT 0 COMMENT '结案人ID',
    close_time DATETIME DEFAULT NULL COMMENT '结案时间',
    deadline DATETIME DEFAULT NULL COMMENT '办理时限',
    created_by BIGINT DEFAULT 0 COMMENT '创建人ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    INDEX idx_case_no(case_no),
    INDEX idx_status(status),
    INDEX idx_type_id(type_id),
    INDEX idx_mediator_id(mediator_id),
    INDEX idx_org_id(organization_id),
    INDEX idx_created_at(created_at),
    INDEX idx_case_level(case_level),
    INDEX idx_case_source(case_source),
    INDEX idx_location(longitude, latitude),
    INDEX idx_keywords ((CAST(keywords AS CHAR(512) ARRAY)))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='纠纷案件主表';

-- 6a. 纠纷关键词词典表
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='纠纷关键词词典';

-- 7. 证据材料表
CREATE TABLE IF NOT EXISTS dispute_evidence (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    file_type TINYINT NOT NULL COMMENT '文件类型: 1-图片 2-视频 3-音频 4-文档 5-其他',
    file_name VARCHAR(256) NOT NULL COMMENT '文件名',
    file_url VARCHAR(512) NOT NULL COMMENT '文件URL(MinIO)',
    file_size BIGINT DEFAULT 0 COMMENT '文件大小(字节)',
    file_md5 VARCHAR(64) DEFAULT '' COMMENT '文件MD5',
    upload_source TINYINT DEFAULT 1 COMMENT '上传来源: 1-自助终端 2-小程序 3-管理端',
    uploader_id BIGINT DEFAULT 0 COMMENT '上传人ID',
    description VARCHAR(512) DEFAULT '' COMMENT '证据说明',
    sort_order INT DEFAULT 0 COMMENT '排序',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_case_id(case_id),
    INDEX idx_case_no(case_no),
    INDEX idx_file_type(file_type),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='证据材料表';

-- 8. 调解记录表
CREATE TABLE IF NOT EXISTS dispute_mediation_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    record_type TINYINT DEFAULT 1 COMMENT '记录类型: 1-初次调解 2-再次调解 3-补充调解',
    mediator_id BIGINT NOT NULL COMMENT '调解员ID',
    mediator_name VARCHAR(64) DEFAULT '' COMMENT '调解员姓名',
    participant_names VARCHAR(512) DEFAULT '' COMMENT '参与人员姓名(逗号分隔)',
    mediation_time DATETIME NOT NULL COMMENT '调解时间',
    mediation_place VARCHAR(256) DEFAULT '' COMMENT '调解地点',
    mediation_duration INT DEFAULT 0 COMMENT '调解时长(分钟)',
    process_content TEXT COMMENT '调解过程记录',
    dispute_focus TEXT COMMENT '争议焦点',
    mediation_opinion TEXT COMMENT '调解意见',
    agreement_content TEXT COMMENT '协议内容',
    result TINYINT DEFAULT 0 COMMENT '结果: 0-进行中 1-达成协议 2-未达成 3-转审',
    next_step VARCHAR(512) DEFAULT '' COMMENT '下一步计划',
    assist_mediators VARCHAR(512) DEFAULT '' COMMENT '协助调解员ID列表',
    ai_summary TEXT COMMENT 'AI生成的调解摘要',
    is_key_record TINYINT DEFAULT 0 COMMENT '是否关键记录: 0-否 1-是',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_case_id(case_id),
    INDEX idx_mediator_id(mediator_id),
    INDEX idx_mediation_time(mediation_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='调解记录表';

-- 9. 案件操作历史表
CREATE TABLE IF NOT EXISTS dispute_case_history (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    operation_type VARCHAR(64) NOT NULL COMMENT '操作类型',
    operation_detail TEXT COMMENT '操作详情(JSON)',
    operator_id BIGINT DEFAULT 0 COMMENT '操作人ID',
    operator_name VARCHAR(64) DEFAULT '' COMMENT '操作人姓名',
    operator_role TINYINT DEFAULT 0 COMMENT '操作人角色',
    old_status TINYINT DEFAULT 0 COMMENT '原状态',
    new_status TINYINT DEFAULT 0 COMMENT '新状态',
    remark VARCHAR(512) DEFAULT '' COMMENT '备注',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_case_id(case_id),
    INDEX idx_case_no(case_no),
    INDEX idx_operation_type(operation_type),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='案件操作历史表';

-- =====================================================
-- 三、工作流审批表
-- =====================================================

-- 10. 审批流程定义表
CREATE TABLE IF NOT EXISTS workflow_approval_definition (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    def_code VARCHAR(64) NOT NULL UNIQUE COMMENT '流程编码',
    def_name VARCHAR(128) NOT NULL COMMENT '流程名称',
    dispute_type_ids VARCHAR(512) DEFAULT '' COMMENT '适用纠纷类型ID列表',
    flowable_process_key VARCHAR(64) DEFAULT '' COMMENT 'Flowable流程定义Key',
    flowable_deployment_id VARCHAR(64) DEFAULT '' COMMENT 'Flowable部署ID',
    approval_nodes TEXT COMMENT '审批节点配置(JSON)',
    timeout_config TEXT COMMENT '超时配置(JSON)',
    version INT DEFAULT 1 COMMENT '版本号',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    description VARCHAR(512) DEFAULT '' COMMENT '描述',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审批流程定义表';

-- 11. 审批实例表
CREATE TABLE IF NOT EXISTS workflow_approval_instance (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    instance_no VARCHAR(64) NOT NULL UNIQUE COMMENT '审批实例编号',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    def_id BIGINT NOT NULL COMMENT '流程定义ID',
    def_code VARCHAR(64) DEFAULT '' COMMENT '流程编码',
    flowable_instance_id VARCHAR(64) DEFAULT '' COMMENT 'Flowable流程实例ID',
    current_node_code VARCHAR(64) DEFAULT '' COMMENT '当前节点编码',
    current_node_name VARCHAR(128) DEFAULT '' COMMENT '当前节点名称',
    approver_id BIGINT DEFAULT 0 COMMENT '当前审批人ID',
    approver_name VARCHAR(64) DEFAULT '' COMMENT '当前审批人姓名',
    status TINYINT DEFAULT 10 COMMENT '状态: 10-审批中 20-已通过 30-已驳回 40-已取消',
    submit_user_id BIGINT NOT NULL COMMENT '提交人ID',
    submit_user_name VARCHAR(64) DEFAULT '' COMMENT '提交人姓名',
    submit_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
    end_time DATETIME DEFAULT NULL COMMENT '结束时间',
    total_nodes INT DEFAULT 0 COMMENT '总节点数',
    current_node_index INT DEFAULT 0 COMMENT '当前节点序号',
    timeout_time DATETIME DEFAULT NULL COMMENT '超时时间',
    escalate_level TINYINT DEFAULT 0 COMMENT '升级级别',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_case_id(case_id),
    INDEX idx_status(status),
    INDEX idx_approver_id(approver_id),
    INDEX idx_submit_time(submit_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审批实例表';

-- 12. 审批记录表
CREATE TABLE IF NOT EXISTS workflow_approval_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    instance_id BIGINT NOT NULL COMMENT '审批实例ID',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    node_code VARCHAR(64) NOT NULL COMMENT '节点编码',
    node_name VARCHAR(128) DEFAULT '' COMMENT '节点名称',
    node_type TINYINT DEFAULT 1 COMMENT '节点类型: 1-审批 2-抄送 3-加签',
    approver_id BIGINT NOT NULL COMMENT '审批人ID',
    approver_name VARCHAR(64) DEFAULT '' COMMENT '审批人姓名',
    approval_action TINYINT NOT NULL COMMENT '审批动作: 1-通过 2-驳回 3-退回修改 4-加签 5-转审 6-拒绝',
    approval_opinion TEXT COMMENT '审批意见',
    ai_suggestion TEXT COMMENT 'AI辅助建议',
    sign_url VARCHAR(512) DEFAULT '' COMMENT '电子签章URL',
    approval_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '审批时间',
    next_approver_id BIGINT DEFAULT 0 COMMENT '下一审批人ID(转审/加签时)',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_instance_id(instance_id),
    INDEX idx_case_id(case_id),
    INDEX idx_approver_id(approver_id),
    INDEX idx_approval_time(approval_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审批记录表';

-- 13. 催办记录表
CREATE TABLE IF NOT EXISTS workflow_urge (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    urge_type TINYINT NOT NULL COMMENT '催办类型: 1-用户催办 2-领导催办 3-系统自动催办 4-超时升级',
    urge_source TINYINT DEFAULT 1 COMMENT '催办来源: 1-小程序 2-管理端 3-系统',
    operator_id BIGINT DEFAULT 0 COMMENT '催办人ID',
    operator_name VARCHAR(64) DEFAULT '' COMMENT '催办人姓名',
    operator_phone VARCHAR(20) DEFAULT '' COMMENT '催办人电话',
    current_handler_id BIGINT NOT NULL COMMENT '当前处理人ID',
    current_handler_name VARCHAR(64) DEFAULT '' COMMENT '当前处理人姓名',
    current_node VARCHAR(128) DEFAULT '' COMMENT '当前节点',
    urgency_level TINYINT DEFAULT 2 COMMENT '紧急程度: 1-特急 2-紧急 3-一般',
    urge_content VARCHAR(512) DEFAULT '' COMMENT '催办内容',
    notify_type VARCHAR(128) DEFAULT '' COMMENT '通知方式: sms,app,wechat,email',
    is_read TINYINT DEFAULT 0 COMMENT '是否已读: 0-否 1-是',
    read_time DATETIME DEFAULT NULL COMMENT '阅读时间',
    escalate_level TINYINT DEFAULT 0 COMMENT '升级级别',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_case_id(case_id),
    INDEX idx_handler_id(current_handler_id),
    INDEX idx_urge_type(urge_type),
    INDEX idx_created_at(created_at),
    INDEX idx_is_read(is_read)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='催办记录表';

-- =====================================================
-- 四、消息通知表
-- =====================================================

-- 14. 消息模板表
CREATE TABLE IF NOT EXISTS notification_template (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    template_code VARCHAR(64) NOT NULL UNIQUE COMMENT '模板编码',
    template_name VARCHAR(128) NOT NULL COMMENT '模板名称',
    channel_type VARCHAR(32) NOT NULL COMMENT '渠道类型: sms,wechat,app,email',
    title_template VARCHAR(256) DEFAULT '' COMMENT '标题模板',
    content_template TEXT NOT NULL COMMENT '内容模板',
    params VARCHAR(512) DEFAULT '' COMMENT '模板参数列表(JSON)',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息模板表';

-- 15. 消息通知记录表
CREATE TABLE IF NOT EXISTS notification_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    msg_no VARCHAR(64) NOT NULL UNIQUE COMMENT '消息编号',
    template_code VARCHAR(64) DEFAULT '' COMMENT '模板编码',
    channel_type VARCHAR(32) NOT NULL COMMENT '渠道类型',
    receiver_id BIGINT DEFAULT 0 COMMENT '接收人ID',
    receiver_type TINYINT DEFAULT 1 COMMENT '接收人类型: 1-用户 2-调解员 3-管理员',
    receiver_phone VARCHAR(20) DEFAULT '' COMMENT '接收人手机号',
    receiver_openid VARCHAR(128) DEFAULT '' COMMENT '接收人OpenID',
    title VARCHAR(256) DEFAULT '' COMMENT '消息标题',
    content TEXT COMMENT '消息内容',
    params TEXT COMMENT '消息参数(JSON)',
    biz_type VARCHAR(64) DEFAULT '' COMMENT '业务类型',
    biz_id VARCHAR(64) DEFAULT '' COMMENT '业务ID',
    case_id BIGINT DEFAULT 0 COMMENT '关联案件ID',
    case_no VARCHAR(32) DEFAULT '' COMMENT '关联案件编号',
    send_status TINYINT DEFAULT 0 COMMENT '发送状态: 0-待发送 1-发送中 2-已发送 3-发送失败',
    send_time DATETIME DEFAULT NULL COMMENT '发送时间',
    fail_reason VARCHAR(512) DEFAULT '' COMMENT '失败原因',
    retry_count INT DEFAULT 0 COMMENT '重试次数',
    is_read TINYINT DEFAULT 0 COMMENT '是否已读',
    read_time DATETIME DEFAULT NULL COMMENT '阅读时间',
    mq_message_id VARCHAR(128) DEFAULT '' COMMENT 'MQ消息ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_receiver_id(receiver_id),
    INDEX idx_case_id(case_id),
    INDEX idx_send_status(send_status),
    INDEX idx_channel_type(channel_type),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息通知记录表';

-- =====================================================
-- 五、AI相关表
-- =====================================================

-- 16. 法条知识库表
CREATE TABLE IF NOT EXISTS ai_law_article (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    law_code VARCHAR(64) DEFAULT '' COMMENT '法律编码',
    law_name VARCHAR(256) NOT NULL COMMENT '法律名称',
    article_no VARCHAR(32) DEFAULT '' COMMENT '法条编号',
    article_title VARCHAR(256) DEFAULT '' COMMENT '法条标题',
    article_content TEXT NOT NULL COMMENT '法条内容',
    category VARCHAR(128) DEFAULT '' COMMENT '分类',
    tags VARCHAR(512) DEFAULT '' COMMENT '标签',
    vector_id BIGINT DEFAULT 0 COMMENT '向量ID(Milvus)',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_law_code(law_code),
    INDEX idx_category(category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='法条知识库表';

-- 17. 法律咨询记录表
CREATE TABLE IF NOT EXISTS ai_law_consult (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    consult_no VARCHAR(64) NOT NULL UNIQUE COMMENT '咨询编号',
    user_type TINYINT DEFAULT 1 COMMENT '用户类型: 1-群众 2-调解员 3-管理员',
    user_id BIGINT DEFAULT 0 COMMENT '用户ID',
    user_name VARCHAR(64) DEFAULT '' COMMENT '用户姓名',
    question TEXT NOT NULL COMMENT '用户问题',
    question_type VARCHAR(64) DEFAULT '' COMMENT '问题类型',
    related_law_articles VARCHAR(1024) DEFAULT '' COMMENT '关联法条ID列表',
    ai_answer TEXT COMMENT 'AI回答',
    ai_model VARCHAR(64) DEFAULT 'deepseek' COMMENT 'AI模型',
    reference_cases VARCHAR(1024) DEFAULT '' COMMENT '参考案例ID列表',
    satisfaction TINYINT DEFAULT 0 COMMENT '满意度: 1-5星',
    is_helpful TINYINT DEFAULT 0 COMMENT '是否有帮助: 0-否 1-是',
    feedback VARCHAR(512) DEFAULT '' COMMENT '用户反馈',
    tokens_used INT DEFAULT 0 COMMENT '消耗Token数',
    cost_time INT DEFAULT 0 COMMENT '耗时(ms)',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_user_id(user_id),
    INDEX idx_question_type(question_type),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='法律咨询记录表';

-- 18. AI调解摘要表
CREATE TABLE IF NOT EXISTS ai_mediation_summary (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    record_id BIGINT DEFAULT 0 COMMENT '调解记录ID',
    summary_type TINYINT DEFAULT 1 COMMENT '摘要类型: 1-调解摘要 2-审批建议 3-风险评估',
    original_content LONGTEXT COMMENT '原始内容',
    ai_summary TEXT COMMENT 'AI生成摘要',
    ai_suggestion TEXT COMMENT 'AI建议',
    risk_level TINYINT DEFAULT 0 COMMENT '风险等级: 0-无 1-低 2-中 3-高',
    risk_points TEXT COMMENT '风险点分析',
    ai_model VARCHAR(64) DEFAULT 'deepseek' COMMENT 'AI模型',
    tokens_used INT DEFAULT 0 COMMENT '消耗Token数',
    cost_time INT DEFAULT 0 COMMENT '耗时(ms)',
    is_approved TINYINT DEFAULT 0 COMMENT '是否被采纳: 0-否 1-是',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_case_id(case_id),
    INDEX idx_summary_type(summary_type),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI调解摘要表';

-- =====================================================
-- 六、绩效考核表
-- =====================================================

-- 19. 绩效考核表
CREATE TABLE IF NOT EXISTS performance_score (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    user_name VARCHAR(64) DEFAULT '' COMMENT '用户姓名',
    organization_id BIGINT DEFAULT 0 COMMENT '组织ID',
    period_type TINYINT NOT NULL COMMENT '周期类型: 1-月度 2-季度 3-年度',
    period_value VARCHAR(32) NOT NULL COMMENT '周期值: 2026-06, 2026-Q2, 2026',
    total_case_count INT DEFAULT 0 COMMENT '受理案件数',
    completed_case_count INT DEFAULT 0 COMMENT '办结案件数',
    mediation_success_count INT DEFAULT 0 COMMENT '调解成功数',
    mediation_success_rate DECIMAL(5,2) DEFAULT 0.00 COMMENT '调解成功率',
    avg_mediation_days DECIMAL(10,2) DEFAULT 0.00 COMMENT '平均调解天数',
    satisfaction_score_avg DECIMAL(3,2) DEFAULT 0.00 COMMENT '平均满意度',
    overdue_count INT DEFAULT 0 COMMENT '超期案件数',
    urge_count INT DEFAULT 0 COMMENT '被催办次数',
    video_mediation_count INT DEFAULT 0 COMMENT '视频调解次数',
    approval_count INT DEFAULT 0 COMMENT '审批案件数',
    ai_usage_count INT DEFAULT 0 COMMENT 'AI工具使用次数',
    score_quantity DECIMAL(10,2) DEFAULT 0.00 COMMENT '工作量得分',
    score_quality DECIMAL(10,2) DEFAULT 0.00 COMMENT '质量得分',
    score_efficiency DECIMAL(10,2) DEFAULT 0.00 COMMENT '效率得分',
    score_satisfaction DECIMAL(10,2) DEFAULT 0.00 COMMENT '满意度得分',
    score_total DECIMAL(10,2) DEFAULT 0.00 COMMENT '总得分',
    rank_org INT DEFAULT 0 COMMENT '组织内排名',
    rank_area INT DEFAULT 0 COMMENT '区域内排名',
    level VARCHAR(16) DEFAULT '' COMMENT '等级: A/B/C/D',
    comment VARCHAR(512) DEFAULT '' COMMENT '评语',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_user_period(user_id, period_type, period_value),
    INDEX idx_org_id(organization_id),
    INDEX idx_period(period_type, period_value),
    INDEX idx_score_total(score_total)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='绩效考核表';

-- 20. 绩效指标配置表
CREATE TABLE IF NOT EXISTS performance_indicator_config (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    indicator_code VARCHAR(64) NOT NULL UNIQUE COMMENT '指标编码',
    indicator_name VARCHAR(128) NOT NULL COMMENT '指标名称',
    indicator_type TINYINT DEFAULT 1 COMMENT '指标类型: 1-工作量 2-质量 3-效率 4-满意度',
    weight DECIMAL(5,2) DEFAULT 0.00 COMMENT '权重',
    max_score DECIMAL(10,2) DEFAULT 100.00 COMMENT '最高分',
    calculation_formula VARCHAR(512) DEFAULT '' COMMENT '计算公式',
    target_value DECIMAL(10,2) DEFAULT 0.00 COMMENT '目标值',
    description VARCHAR(512) DEFAULT '' COMMENT '描述',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='绩效指标配置表';

-- =====================================================
-- 七、自助终端表
-- =====================================================

-- 21. 自助终端表
CREATE TABLE IF NOT EXISTS kiosk_device (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    device_code VARCHAR(64) NOT NULL UNIQUE COMMENT '终端编码',
    device_name VARCHAR(128) DEFAULT '' COMMENT '终端名称',
    device_model VARCHAR(64) DEFAULT '' COMMENT '终端型号',
    location VARCHAR(256) DEFAULT '' COMMENT '放置位置',
    organization_id BIGINT DEFAULT 0 COMMENT '所属组织ID',
    longitude DECIMAL(10,7) DEFAULT NULL COMMENT '经度',
    latitude DECIMAL(10,7) DEFAULT NULL COMMENT '纬度',
    ip_address VARCHAR(64) DEFAULT '' COMMENT 'IP地址',
    mac_address VARCHAR(64) DEFAULT '' COMMENT 'MAC地址',
    last_heartbeat DATETIME DEFAULT NULL COMMENT '最后心跳时间',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-离线 1-在线 2-故障',
    status_detail VARCHAR(256) DEFAULT '' COMMENT '状态详情',
    total_register_count INT DEFAULT 0 COMMENT '总登记数',
    today_register_count INT DEFAULT 0 COMMENT '今日登记数',
    id_card_reader_status TINYINT DEFAULT 1 COMMENT '读卡器状态',
    printer_status TINYINT DEFAULT 1 COMMENT '打印机状态',
    camera_status TINYINT DEFAULT 1 COMMENT '摄像头状态',
    remark VARCHAR(512) DEFAULT '' COMMENT '备注',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_org_id(organization_id),
    INDEX idx_status(status),
    INDEX idx_location(longitude, latitude)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自助终端表';

-- 22. 终端离线队列表
CREATE TABLE IF NOT EXISTS kiosk_offline_queue (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    device_code VARCHAR(64) NOT NULL COMMENT '终端编码',
    queue_no VARCHAR(64) NOT NULL UNIQUE COMMENT '队列编号',
    data_type TINYINT NOT NULL COMMENT '数据类型: 1-纠纷登记 2-证据上传',
    data_content LONGTEXT NOT NULL COMMENT '数据内容(JSON)',
    file_count INT DEFAULT 0 COMMENT '附件数量',
    synced TINYINT DEFAULT 0 COMMENT '是否已同步: 0-否 1-是',
    sync_time DATETIME DEFAULT NULL COMMENT '同步时间',
    sync_error VARCHAR(512) DEFAULT '' COMMENT '同步错误信息',
    retry_count INT DEFAULT 0 COMMENT '重试次数',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_device_code(device_code),
    INDEX idx_synced(synced),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='终端离线队列表';

-- =====================================================
-- 八、热力图/统计表
-- =====================================================

-- 23. 案件区域统计表 (用于热力图)
CREATE TABLE IF NOT EXISTS stats_case_area (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    stat_date DATE NOT NULL COMMENT '统计日期',
    area_code VARCHAR(64) NOT NULL COMMENT '区域编码',
    area_name VARCHAR(128) DEFAULT '' COMMENT '区域名称',
    organization_id BIGINT DEFAULT 0 COMMENT '组织ID',
    longitude DECIMAL(10,7) DEFAULT NULL COMMENT '区域中心经度',
    latitude DECIMAL(10,7) DEFAULT NULL COMMENT '区域中心纬度',
    total_count INT DEFAULT 0 COMMENT '案件总数',
    pending_count INT DEFAULT 0 COMMENT '待处理数',
    processing_count INT DEFAULT 0 COMMENT '处理中数',
    completed_count INT DEFAULT 0 COMMENT '已完成数',
    success_count INT DEFAULT 0 COMMENT '调解成功数',
    type1_count INT DEFAULT 0 COMMENT '婚姻家庭类',
    type2_count INT DEFAULT 0 COMMENT '邻里纠纷类',
    type3_count INT DEFAULT 0 COMMENT '财产纠纷类',
    type4_count INT DEFAULT 0 COMMENT '劳务纠纷类',
    type5_count INT DEFAULT 0 COMMENT '其他类型',
    urgent_count INT DEFAULT 0 COMMENT '紧急案件数',
    heat_value DECIMAL(10,2) DEFAULT 0.00 COMMENT '热度值',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_date_area(stat_date, area_code),
    INDEX idx_org_id(organization_id),
    INDEX idx_stat_date(stat_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='案件区域统计表';

-- 24. 数据大屏统计表
CREATE TABLE IF NOT EXISTS stats_dashboard (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    stat_type VARCHAR(32) NOT NULL COMMENT '统计类型: today/week/month/quarter/year',
    stat_date DATE NOT NULL COMMENT '统计日期',
    total_case_count INT DEFAULT 0 COMMENT '案件总数',
    new_case_count INT DEFAULT 0 COMMENT '新增案件数',
    completed_case_count INT DEFAULT 0 COMMENT '办结案件数',
    mediation_success_rate DECIMAL(5,2) DEFAULT 0.00 COMMENT '调解成功率',
    avg_processing_days DECIMAL(10,2) DEFAULT 0.00 COMMENT '平均办理天数',
    pending_over_24h INT DEFAULT 0 COMMENT '24h未处理数',
    pending_over_72h INT DEFAULT 0 COMMENT '72h未处理数',
    satisfaction_avg DECIMAL(3,2) DEFAULT 0.00 COMMENT '平均满意度',
    video_mediation_count INT DEFAULT 0 COMMENT '视频调解数',
    online_kiosk_count INT DEFAULT 0 COMMENT '在线终端数',
    active_mediator_count INT DEFAULT 0 COMMENT '活跃调解员数',
    ai_consult_count INT DEFAULT 0 COMMENT 'AI咨询数',
    urge_count INT DEFAULT 0 COMMENT '催办数',
    escalate_count INT DEFAULT 0 COMMENT '升级数',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_type_date(stat_type, stat_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据大屏统计表';

-- =====================================================
-- 九、电子签章表
-- =====================================================

-- 25. 电子签章记录表
CREATE TABLE IF NOT EXISTS esign_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    esign_no VARCHAR(64) NOT NULL UNIQUE COMMENT '签章编号',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    document_type TINYINT NOT NULL COMMENT '文件类型: 1-调解协议 2-审批意见 3-送达回执',
    document_name VARCHAR(256) DEFAULT '' COMMENT '文件名称',
    document_url VARCHAR(512) DEFAULT '' COMMENT '原文件URL',
    signed_document_url VARCHAR(512) DEFAULT '' COMMENT '已签章文件URL',
    signer_id BIGINT NOT NULL COMMENT '签章人ID',
    signer_name VARCHAR(64) DEFAULT '' COMMENT '签章人姓名',
    signer_id_card VARCHAR(32) DEFAULT '' COMMENT '签章人身份证',
    signer_role VARCHAR(64) DEFAULT '' COMMENT '签章人角色',
    seal_type TINYINT DEFAULT 1 COMMENT '印章类型: 1-个人章 2-单位章 3-调解专用章',
    seal_url VARCHAR(512) DEFAULT '' COMMENT '印章图片URL',
    certificate_sn VARCHAR(256) DEFAULT '' COMMENT '数字证书序列号',
    sign_time DATETIME DEFAULT NULL COMMENT '签章时间',
    sign_ip VARCHAR(64) DEFAULT '' COMMENT '签章IP',
    hash_value VARCHAR(128) DEFAULT '' COMMENT '文件哈希值',
    timestamp_token VARCHAR(512) DEFAULT '' COMMENT '时间戳令牌',
    verify_result TINYINT DEFAULT 0 COMMENT '验证结果: 0-未验证 1-验证通过 2-验证失败',
    verify_time DATETIME DEFAULT NULL COMMENT '验证时间',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-作废 1-有效',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_case_id(case_id),
    INDEX idx_signer_id(signer_id),
    INDEX idx_sign_time(sign_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='电子签章记录表';

-- =====================================================
-- 十、视频调解表
-- =====================================================

-- 26. 视频调解房间表
CREATE TABLE IF NOT EXISTS video_room (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    room_id VARCHAR(64) NOT NULL UNIQUE COMMENT '房间ID',
    room_name VARCHAR(128) DEFAULT '' COMMENT '房间名称',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    host_user_id BIGINT NOT NULL COMMENT '主持人ID(调解员)',
    host_user_name VARCHAR(64) DEFAULT '' COMMENT '主持人姓名',
    participant_ids VARCHAR(1024) DEFAULT '' COMMENT '参与人ID列表(JSON)',
    participant_names VARCHAR(1024) DEFAULT '' COMMENT '参与人姓名列表(逗号分隔)',
    scheduled_time DATETIME DEFAULT NULL COMMENT '预约时间',
    start_time DATETIME DEFAULT NULL COMMENT '开始时间',
    end_time DATETIME DEFAULT NULL COMMENT '结束时间',
    duration INT DEFAULT 0 COMMENT '时长(分钟)',
    room_password VARCHAR(64) DEFAULT '' COMMENT '房间密码',
    max_participants INT DEFAULT 10 COMMENT '最大参与人数',
    status TINYINT DEFAULT 10 COMMENT '状态: 10-未开始 20-进行中 30-已结束 40-已取消',
    record_url VARCHAR(512) DEFAULT '' COMMENT '录制文件URL',
    record_size BIGINT DEFAULT 0 COMMENT '录制文件大小',
    platform VARCHAR(32) DEFAULT 'web' COMMENT '平台: web/android/ios',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_case_id(case_id),
    INDEX idx_host_id(host_user_id),
    INDEX idx_status(status),
    INDEX idx_scheduled_time(scheduled_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='视频调解房间表';

-- 27. 视频参与记录表
CREATE TABLE IF NOT EXISTS video_participant (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    room_id VARCHAR(64) NOT NULL COMMENT '房间ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    user_name VARCHAR(64) DEFAULT '' COMMENT '用户姓名',
    user_role VARCHAR(32) DEFAULT '' COMMENT '用户角色',
    join_time DATETIME DEFAULT NULL COMMENT '加入时间',
    leave_time DATETIME DEFAULT NULL COMMENT '离开时间',
    duration INT DEFAULT 0 COMMENT '在线时长(分钟)',
    ip_address VARCHAR(64) DEFAULT '' COMMENT 'IP地址',
    device_info VARCHAR(256) DEFAULT '' COMMENT '设备信息',
    network_quality TINYINT DEFAULT 3 COMMENT '网络质量: 1-优 2-良 3-中 4-差',
    is_host TINYINT DEFAULT 0 COMMENT '是否主持人',
    is_muted TINYINT DEFAULT 0 COMMENT '是否静音',
    is_video_enabled TINYINT DEFAULT 1 COMMENT '是否开启视频',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_room_id(room_id),
    INDEX idx_user_id(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='视频参与记录表';
