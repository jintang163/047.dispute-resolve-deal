-- =====================================================
-- 法律援助转介 - 数据库迁移脚本
-- 创建时间: 2026-06-21
-- =====================================================

USE dispute_resolve;

-- 1. 法律援助机构表
CREATE TABLE IF NOT EXISTS legal_aid_org (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    org_code VARCHAR(64) NOT NULL UNIQUE COMMENT '机构编码',
    org_name VARCHAR(128) NOT NULL COMMENT '机构名称',
    org_type TINYINT NOT NULL DEFAULT 1 COMMENT '机构类型: 1-法律援助中心 2-律师事务所 3-公证处 4-司法鉴定所',
    level TINYINT DEFAULT 1 COMMENT '行政级别: 1-市级 2-区级 3-街道级',
    address VARCHAR(256) DEFAULT '' COMMENT '地址',
    longitude DECIMAL(10,7) DEFAULT NULL COMMENT '经度',
    latitude DECIMAL(10,7) DEFAULT NULL COMMENT '纬度',
    contact_person VARCHAR(64) DEFAULT '' COMMENT '联系人',
    contact_phone VARCHAR(20) DEFAULT '' COMMENT '联系电话',
    contact_email VARCHAR(128) DEFAULT '' COMMENT '联系邮箱',
    service_scope TEXT COMMENT '服务范围(纠纷类型ID列表，逗号分隔)',
    work_hours VARCHAR(256) DEFAULT '' COMMENT '工作时间',
    description TEXT COMMENT '机构简介',
    lawyer_count INT DEFAULT 0 COMMENT '律师人数',
    case_capacity INT DEFAULT 0 COMMENT '月办案容量',
    accept_count INT DEFAULT 0 COMMENT '已受理案件数',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_org_type(org_type),
    INDEX idx_level(level),
    INDEX idx_status(status),
    INDEX idx_location(longitude, latitude),
    INDEX idx_deleted_at(deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='法律援助机构表';

-- 2. 法援律师表
CREATE TABLE IF NOT EXISTS legal_aid_lawyer (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    org_id BIGINT NOT NULL COMMENT '所属机构ID',
    org_name VARCHAR(128) DEFAULT '' COMMENT '所属机构名称',
    lawyer_name VARCHAR(64) NOT NULL COMMENT '律师姓名',
    license_no VARCHAR(64) DEFAULT '' COMMENT '执业证号',
    phone VARCHAR(20) DEFAULT '' COMMENT '手机号',
    email VARCHAR(128) DEFAULT '' COMMENT '邮箱',
    avatar VARCHAR(256) DEFAULT '' COMMENT '头像',
    gender TINYINT DEFAULT 0 COMMENT '性别: 0-未知 1-男 2-女',
    specialty VARCHAR(512) DEFAULT '' COMMENT '专业领域(逗号分隔)',
    years_of_experience INT DEFAULT 0 COMMENT '从业年限',
    title VARCHAR(64) DEFAULT '' COMMENT '职称',
    intro TEXT COMMENT '个人简介',
    consult_count INT DEFAULT 0 COMMENT '咨询次数',
    consult_rating DECIMAL(3,2) DEFAULT 0.00 COMMENT '咨询评分(1-5)',
    is_online TINYINT DEFAULT 0 COMMENT '是否在线: 0-离线 1-在线',
    last_online_at DATETIME DEFAULT NULL COMMENT '最后在线时间',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_org_id(org_id),
    INDEX idx_status(status),
    INDEX idx_is_online(is_online),
    INDEX idx_deleted_at(deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='法援律师表';

-- 3. 法律援助申请表(资质审核)
CREATE TABLE IF NOT EXISTS legal_aid_application (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    apply_no VARCHAR(64) NOT NULL UNIQUE COMMENT '申请编号',
    case_id BIGINT NOT NULL COMMENT '纠纷案件ID',
    case_no VARCHAR(32) DEFAULT '' COMMENT '案件编号',
    applicant_name VARCHAR(64) NOT NULL COMMENT '申请人姓名',
    applicant_phone VARCHAR(20) DEFAULT '' COMMENT '申请人电话',
    applicant_id_card VARCHAR(32) DEFAULT '' COMMENT '申请人身份证号',
    applicant_address VARCHAR(256) DEFAULT '' COMMENT '申请人地址',
    income_level TINYINT DEFAULT 3 COMMENT '收入水平: 1-低保 2-低收入 3-普通',
    family_size INT DEFAULT 0 COMMENT '家庭人口数',
    monthly_income DECIMAL(10,2) DEFAULT 0.00 COMMENT '月均收入',
    aid_reason TEXT COMMENT '申请法律援助理由',
    dispute_type VARCHAR(128) DEFAULT '' COMMENT '纠纷类型',
    evidence_summary TEXT COMMENT '证据情况简述',
    material_urls TEXT COMMENT '证明材料URL列表(JSON数组)',
    status TINYINT DEFAULT 10 COMMENT '状态: 10-待审核 20-审核通过 30-审核驳回 40-已撤销',
    auditor_id BIGINT DEFAULT 0 COMMENT '审核人ID',
    auditor_name VARCHAR(64) DEFAULT '' COMMENT '审核人姓名',
    audit_time DATETIME DEFAULT NULL COMMENT '审核时间',
    audit_opinion VARCHAR(500) DEFAULT '' COMMENT '审核意见',
    reject_reason VARCHAR(500) DEFAULT '' COMMENT '驳回原因',
    submitter_id BIGINT DEFAULT 0 COMMENT '提交人ID',
    submitter_name VARCHAR(64) DEFAULT '' COMMENT '提交人姓名',
    submit_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_case_id(case_id),
    INDEX idx_case_no(case_no),
    INDEX idx_status(status),
    INDEX idx_applicant_phone(applicant_phone),
    INDEX idx_submit_time(submit_time),
    INDEX idx_deleted_at(deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='法律援助申请表';

-- 4. 法律援助转介记录表
CREATE TABLE IF NOT EXISTS legal_aid_transfer (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    transfer_no VARCHAR(64) NOT NULL UNIQUE COMMENT '转介编号',
    case_id BIGINT NOT NULL COMMENT '纠纷案件ID',
    case_no VARCHAR(32) DEFAULT '' COMMENT '案件编号',
    case_title VARCHAR(256) DEFAULT '' COMMENT '案件标题',
    dispute_type VARCHAR(128) DEFAULT '' COMMENT '纠纷类型',
    from_org_id BIGINT DEFAULT 0 COMMENT '转出机构ID(调解组织)',
    from_org_name VARCHAR(128) DEFAULT '' COMMENT '转出机构名称',
    from_user_id BIGINT DEFAULT 0 COMMENT '转出操作人ID',
    from_user_name VARCHAR(64) DEFAULT '' COMMENT '转出操作人姓名',
    to_org_id BIGINT NOT NULL COMMENT '转入机构ID(法援机构)',
    to_org_name VARCHAR(128) DEFAULT '' COMMENT '转入机构名称',
    to_lawyer_id BIGINT DEFAULT 0 COMMENT '分配律师ID',
    to_lawyer_name VARCHAR(64) DEFAULT '' COMMENT '分配律师姓名',
    transfer_reason TEXT COMMENT '转介原因',
    case_summary TEXT COMMENT '案件摘要',
    attach_ids VARCHAR(512) DEFAULT '' COMMENT '附件材料ID列表(逗号分隔)',
    accept_status TINYINT DEFAULT 10 COMMENT '受理状态: 10-待受理 20-已受理 30-已驳回 40-已办结',
    accept_time DATETIME DEFAULT NULL COMMENT '受理时间',
    legal_case_no VARCHAR(64) DEFAULT '' COMMENT '法援案件编号',
    reject_reason VARCHAR(500) DEFAULT '' COMMENT '驳回原因',
    close_result TEXT COMMENT '办理结果',
    close_time DATETIME DEFAULT NULL COMMENT '办结时间',
    transfer_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '转介时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_case_id(case_id),
    INDEX idx_case_no(case_no),
    INDEX idx_to_org_id(to_org_id),
    INDEX idx_from_org_id(from_org_id),
    INDEX idx_accept_status(accept_status),
    INDEX idx_transfer_time(transfer_time),
    INDEX idx_legal_case_no(legal_case_no),
    INDEX idx_deleted_at(deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='法律援助转介记录表';

-- 5. 法援咨询记录表
CREATE TABLE IF NOT EXISTS legal_aid_consult (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    consult_no VARCHAR(64) NOT NULL UNIQUE COMMENT '咨询编号',
    transfer_id BIGINT DEFAULT 0 COMMENT '转介记录ID',
    case_id BIGINT DEFAULT 0 COMMENT '案件ID',
    case_no VARCHAR(32) DEFAULT '' COMMENT '案件编号',
    applicant_id BIGINT DEFAULT 0 COMMENT '咨询人ID',
    applicant_name VARCHAR(64) DEFAULT '' COMMENT '咨询人姓名',
    applicant_phone VARCHAR(20) DEFAULT '' COMMENT '咨询人电话',
    lawyer_id BIGINT NOT NULL COMMENT '律师ID',
    lawyer_name VARCHAR(64) DEFAULT '' COMMENT '律师姓名',
    org_id BIGINT DEFAULT 0 COMMENT '法援机构ID',
    org_name VARCHAR(128) DEFAULT '' COMMENT '法援机构名称',
    consult_type TINYINT DEFAULT 1 COMMENT '咨询类型: 1-文字 2-语音 3-视频',
    consult_status TINYINT DEFAULT 10 COMMENT '咨询状态: 10-待开始 20-进行中 30-已完成 40-已取消 50-超时',
    question_title VARCHAR(256) DEFAULT '' COMMENT '咨询问题标题',
    question_content TEXT COMMENT '咨询问题详情',
    answer_content TEXT COMMENT '律师回复内容',
    total_duration INT DEFAULT 0 COMMENT '总咨询时长(秒)',
    free_duration INT DEFAULT 1800 COMMENT '免费时长(秒，默认30分钟)',
    used_duration INT DEFAULT 0 COMMENT '已用时长(秒)',
    is_free TINYINT DEFAULT 1 COMMENT '是否免费: 0-否 1-是',
    rating TINYINT DEFAULT 0 COMMENT '评分: 1-5星',
    comment VARCHAR(500) DEFAULT '' COMMENT '评价内容',
    start_time DATETIME DEFAULT NULL COMMENT '开始时间',
    end_time DATETIME DEFAULT NULL COMMENT '结束时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_transfer_id(transfer_id),
    INDEX idx_case_id(case_id),
    INDEX idx_lawyer_id(lawyer_id),
    INDEX idx_applicant_id(applicant_id),
    INDEX idx_consult_status(consult_status),
    INDEX idx_created_at(created_at),
    INDEX idx_deleted_at(deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='法援咨询记录表';

-- 6. 法援咨询消息表
CREATE TABLE IF NOT EXISTS legal_aid_consult_message (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    consult_id BIGINT NOT NULL COMMENT '咨询ID',
    sender_id BIGINT NOT NULL COMMENT '发送人ID',
    sender_name VARCHAR(64) DEFAULT '' COMMENT '发送人姓名',
    sender_type TINYINT NOT NULL COMMENT '发送人类型: 1-用户 2-律师',
    message_type TINYINT DEFAULT 1 COMMENT '消息类型: 1-文字 2-图片 3-语音 4-视频 5-文件',
    content TEXT COMMENT '消息内容',
    file_url VARCHAR(512) DEFAULT '' COMMENT '文件URL',
    file_name VARCHAR(256) DEFAULT '' COMMENT '文件名',
    file_size BIGINT DEFAULT 0 COMMENT '文件大小(字节)',
    duration INT DEFAULT 0 COMMENT '语音/视频时长(秒)',
    is_read TINYINT DEFAULT 0 COMMENT '是否已读: 0-未读 1-已读',
    read_time DATETIME DEFAULT NULL COMMENT '阅读时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_consult_id(consult_id),
    INDEX idx_sender_id(sender_id),
    INDEX idx_is_read(is_read),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='法援咨询消息表';

-- =====================================================
-- 初始化数据 - 法律援助机构示例
-- =====================================================

INSERT INTO legal_aid_org (org_code, org_name, org_type, level, address, longitude, latitude, 
    contact_person, contact_phone, service_scope, work_hours, description, lawyer_count, case_capacity, sort_order) VALUES
('LAZ-001', '市法律援助中心', 1, 1, '市中心区法律援助大厦1楼', 116.4074000, 39.9042000, 
 '张主任', '010-12345678', '1,2,3,4,5', '周一至周五 9:00-17:00', '市级综合性法律援助服务机构', 20, 100, 1),
('LAZ-002', '东城区法律援助中心', 1, 2, '东城区政府服务中心B座', 116.4150000, 39.9280000, 
 '李主任', '010-23456789', '1,2,3', '周一至周五 9:00-17:00', '东城区法律援助服务机构', 10, 50, 2),
('LAZ-003', '西城区法律援助中心', 1, 2, '西城区政务服务中心3层', 116.3660000, 39.9150000, 
 '王主任', '010-34567890', '1,2,4,5', '周一至周五 9:00-17:00', '西城区法律援助服务机构', 12, 60, 3),
('LAZ-004', '朝阳区法律援助中心', 1, 2, '朝阳区公共法律服务中心', 116.4550000, 39.9210000, 
 '赵主任', '010-45678901', '1,2,3,5', '周一至周五 9:00-17:00', '朝阳区法律援助服务机构', 15, 80, 4),
('LAZ-005', '海淀区法律援助中心', 1, 2, '海淀区司法局一楼大厅', 116.3100000, 39.9590000, 
 '刘主任', '010-56789012', '1,2,3,4,5', '周一至周五 9:00-17:00', '海淀区法律援助服务机构', 18, 90, 5),
('LAW-001', '正义律师事务所', 2, 2, '朝阳区建国路88号SOHO现代城', 116.4650000, 39.9080000, 
 '陈律师', '010-67890123', '2,3,4', '周一至周五 9:00-18:00', '综合性合伙制律师事务所', 30, 200, 6),
('LAW-002', '明德律师事务所', 2, 2, '海淀区中关村大街1号', 116.3250000, 39.9850000, 
 '周律师', '010-78901234', '1,3,5', '周一至周五 9:00-18:00', '知识产权专业律师事务所', 25, 150, 7);

-- 初始化法援律师示例数据
INSERT INTO legal_aid_lawyer (org_id, org_name, lawyer_name, license_no, phone, gender, 
    specialty, years_of_experience, title, intro, consult_count, consult_rating, is_online, status) VALUES
(1, '市法律援助中心', '张伟', '1101012020001', '13800001001', 1, '婚姻家庭,劳动争议,合同纠纷', 10, '三级律师', '资深法律援助律师，擅长民事纠纷调解', 150, 4.80, 1, 1),
(1, '市法律援助中心', '李娜', '1101012020002', '13800001002', 2, '婚姻家庭,未成年人保护,侵权责任', 8, '四级律师', '专注妇女儿童权益保护', 120, 4.90, 1, 1),
(2, '东城区法律援助中心', '王强', '1101012020003', '13800001003', 1, '劳动争议,工伤赔偿,社会保险', 6, '四级律师', '劳动法律援助专业律师', 90, 4.70, 0, 1),
(3, '西城区法律援助中心', '刘芳', '1101012020004', '13800001004', 2, '合同纠纷,债务纠纷,房产纠纷', 12, '三级律师', '民商事法律援助资深律师', 200, 4.85, 1, 1),
(4, '朝阳区法律援助中心', '陈明', '1101012020005', '13800001005', 1, '交通事故,人身损害,保险理赔', 7, '四级律师', '交通事故专业援助律师', 100, 4.60, 1, 1),
(5, '海淀区法律援助中心', '杨丽', '1101012020006', '13800001006', 2, '知识产权,合同纠纷,公司事务', 15, '二级律师', '知识产权专业律师', 180, 4.95, 0, 1),
(6, '正义律师事务所', '赵刚', '1101012010007', '13800001007', 1, '刑事辩护,取保候审,刑民交叉', 20, '一级律师', '资深刑事辩护律师', 300, 4.75, 1, 1),
(7, '明德律师事务所', '黄磊', '1101012015008', '13800001008', 1, '专利侵权,商标纠纷,著作权', 10, '三级律师', '知识产权专家律师', 160, 4.90, 1, 1);
