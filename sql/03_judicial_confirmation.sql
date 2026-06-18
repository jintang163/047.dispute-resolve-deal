-- =====================================================
-- 司法确认模块表结构
-- =====================================================

-- 28. 司法确认申请表
CREATE TABLE IF NOT EXISTS judicial_confirmation (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    confirm_no VARCHAR(32) NOT NULL UNIQUE COMMENT '确认编号',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    case_title VARCHAR(256) DEFAULT '' COMMENT '案件标题',
    mediation_record_id BIGINT DEFAULT 0 COMMENT '调解记录ID',
    protocol_id BIGINT DEFAULT 0 COMMENT '调解协议ID',

    applicant_name VARCHAR(64) NOT NULL COMMENT '申请人姓名',
    applicant_phone VARCHAR(20) NOT NULL COMMENT '申请人电话',
    applicant_id_card VARCHAR(32) DEFAULT '' COMMENT '申请人身份证',
    applicant_address VARCHAR(256) DEFAULT '' COMMENT '申请人地址',

    respondent_name VARCHAR(64) NOT NULL COMMENT '被申请人姓名',
    respondent_phone VARCHAR(20) NOT NULL COMMENT '被申请人电话',
    respondent_id_card VARCHAR(32) DEFAULT '' COMMENT '被申请人身份证',
    respondent_address VARCHAR(256) DEFAULT '' COMMENT '被申请人地址',

    court_id BIGINT DEFAULT 0 COMMENT '推送法院ID',
    court_name VARCHAR(128) DEFAULT '' COMMENT '推送法院名称',
    court_code VARCHAR(64) DEFAULT '' COMMENT '法院编码',

    agreement_content TEXT COMMENT '调解协议内容',
    performance_deadline DATE DEFAULT NULL COMMENT '履行期限',
    confirm_amount DECIMAL(15,2) DEFAULT 0.00 COMMENT '确认金额',

    status TINYINT DEFAULT 10 COMMENT '状态: 10-已提交 20-审核中 30-已确认 40-已驳回 50-已失效',
    sub_status TINYINT DEFAULT 0 COMMENT '子状态',
    review_opinion VARCHAR(512) DEFAULT '' COMMENT '审核意见',
    confirm_date DATE DEFAULT NULL COMMENT '确认日期',
    confirm_court VARCHAR(128) DEFAULT '' COMMENT '确认法院',
    confirm_judge VARCHAR(64) DEFAULT '' COMMENT '承办法官',
    confirm_document_no VARCHAR(64) DEFAULT '' COMMENT '确认文书号',

    document_url VARCHAR(512) DEFAULT '' COMMENT '确认书PDF地址',
    seal_status TINYINT DEFAULT 0 COMMENT '签章状态: 0-未签章 1-已签章',
    seal_time DATETIME DEFAULT NULL COMMENT '签章时间',

    submit_time DATETIME DEFAULT NULL COMMENT '提交时间',
    submit_by BIGINT DEFAULT 0 COMMENT '提交人ID',
    submit_by_name VARCHAR(64) DEFAULT '' COMMENT '提交人姓名',

    review_time DATETIME DEFAULT NULL COMMENT '审核时间',
    review_by BIGINT DEFAULT 0 COMMENT '审核人ID',
    review_by_name VARCHAR(64) DEFAULT '' COMMENT '审核人姓名',

    expiration_remind_sent TINYINT DEFAULT 0 COMMENT '失效提醒是否已发送: 0-否 1-是',
    performance_remind_sent TINYINT DEFAULT 0 COMMENT '履行提醒是否已发送: 0-否 1-是',

    organization_id BIGINT NOT NULL COMMENT '所属组织ID',
    remark VARCHAR(512) DEFAULT '' COMMENT '备注',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',

    INDEX idx_case_id(case_id),
    INDEX idx_case_no(case_no),
    INDEX idx_status(status),
    INDEX idx_court_id(court_id),
    INDEX idx_performance_deadline(performance_deadline),
    INDEX idx_organization_id(organization_id),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司法确认申请表';

-- 29. 司法确认进度轨迹表
CREATE TABLE IF NOT EXISTS judicial_confirm_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    confirm_id BIGINT NOT NULL COMMENT '确认申请ID',
    confirm_no VARCHAR(32) NOT NULL COMMENT '确认编号',
    action_type INT NOT NULL COMMENT '操作类型: 10-提交申请 20-法院受理 30-审核通过 40-审核驳回 50-已签章 60-确认书送达 70-履行提醒 80-失效提醒 90-已履行 99-已失效',
    action_name VARCHAR(64) NOT NULL COMMENT '操作名称',
    action_detail TEXT COMMENT '操作详情',
    operator_id BIGINT DEFAULT 0 COMMENT '操作人ID',
    operator_name VARCHAR(64) DEFAULT '' COMMENT '操作人姓名',
    operator_type TINYINT DEFAULT 1 COMMENT '操作人类型: 1-系统 2-管理员 3-法院 4-当事人',

    court_remark VARCHAR(512) DEFAULT '' COMMENT '法院备注',
    court_operator VARCHAR(64) DEFAULT '' COMMENT '法院经办人',
    court_handle_time DATETIME DEFAULT NULL COMMENT '法院处理时间',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_confirm_id(confirm_id),
    INDEX idx_action_type(action_type),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司法确认进度轨迹表';

-- 30. 法院配置表
CREATE TABLE IF NOT EXISTS court_config (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    court_code VARCHAR(64) NOT NULL UNIQUE COMMENT '法院编码',
    court_name VARCHAR(128) NOT NULL COMMENT '法院名称',
    court_level TINYINT DEFAULT 3 COMMENT '法院级别: 1-最高院 2-高院 3-中院 4-基层法院 5-法庭',
    jurisdiction_area VARCHAR(256) DEFAULT '' COMMENT '管辖区域',
    address VARCHAR(256) DEFAULT '' COMMENT '法院地址',
    contact VARCHAR(64) DEFAULT '' COMMENT '联系人',
    phone VARCHAR(20) DEFAULT '' COMMENT '联系电话',

    api_endpoint VARCHAR(256) DEFAULT '' COMMENT '微法院API地址',
    api_app_id VARCHAR(128) DEFAULT '' COMMENT 'API AppID',
    api_secret VARCHAR(256) DEFAULT '' COMMENT 'API Secret',
    api_public_key TEXT COMMENT 'API公钥',

    seal_cert_no VARCHAR(64) DEFAULT '' COMMENT '电子签章证书编号',
    seal_image_url VARCHAR(512) DEFAULT '' COMMENT '法院印章图片地址',

    organization_id BIGINT DEFAULT 0 COMMENT '关联综治中心ID',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',

    INDEX idx_organization_id(organization_id),
    INDEX idx_court_level(court_level),
    INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='法院配置表';
