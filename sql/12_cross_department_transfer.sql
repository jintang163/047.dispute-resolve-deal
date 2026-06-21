-- =====================================================
-- 跨部门协同转办 - 数据库初始化脚本
-- 创建时间: 2026-06-21
-- =====================================================

USE dispute_resolve;

-- =====================================================
-- 1. 转办模板配置表 - 预置常用转办部门
-- =====================================================
CREATE TABLE IF NOT EXISTS dispute_transfer_template (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    template_name VARCHAR(128) NOT NULL COMMENT '模板名称',
    dept_code VARCHAR(64) NOT NULL UNIQUE COMMENT '部门编码',
    dept_name VARCHAR(128) NOT NULL COMMENT '部门名称',
    dept_type VARCHAR(32) NOT NULL COMMENT '部门类型: HR-人社局 MS-市监局 PS-公安局 CT-法院 OT-其他',
    contact_person VARCHAR(64) DEFAULT '' COMMENT '联系人',
    contact_phone VARCHAR(20) DEFAULT '' COMMENT '联系电话',
    contact_email VARCHAR(128) DEFAULT '' COMMENT '联系邮箱',
    description VARCHAR(512) DEFAULT '' COMMENT '部门职责描述',
    applicable_types VARCHAR(512) DEFAULT '' COMMENT '适用纠纷类型(ID列表,逗号分隔)',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    INDEX idx_dept_type(dept_type),
    INDEX idx_status(status),
    INDEX idx_sort_order(sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='转办模板配置表';

-- =====================================================
-- 2. 纠纷转办记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS dispute_transfer (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    transfer_no VARCHAR(32) NOT NULL UNIQUE COMMENT '转办编号',
    case_id BIGINT NOT NULL COMMENT '案件ID',
    case_no VARCHAR(32) NOT NULL COMMENT '案件编号',
    case_title VARCHAR(256) DEFAULT '' COMMENT '案件标题',
    template_id BIGINT DEFAULT 0 COMMENT '转办模板ID',
    from_dept_id BIGINT NOT NULL COMMENT '转出部门ID',
    from_dept_name VARCHAR(128) NOT NULL COMMENT '转出部门名称',
    from_user_id BIGINT NOT NULL COMMENT '转出人ID',
    from_user_name VARCHAR(64) NOT NULL COMMENT '转出人姓名',
    to_dept_code VARCHAR(64) NOT NULL COMMENT '转入部门编码',
    to_dept_name VARCHAR(128) NOT NULL COMMENT '转入部门名称',
    to_dept_type VARCHAR(32) NOT NULL COMMENT '转入部门类型',
    to_contact_person VARCHAR(64) DEFAULT '' COMMENT '转入方联系人',
    to_contact_phone VARCHAR(20) DEFAULT '' COMMENT '转入方联系电话',
    transfer_reason VARCHAR(1000) NOT NULL COMMENT '转办原因',
    transfer_remark VARCHAR(500) DEFAULT '' COMMENT '转办备注',
    attach_ids VARCHAR(512) DEFAULT '' COMMENT '附件ID列表(逗号分隔)',
    status TINYINT DEFAULT 10 COMMENT '状态: 10-待接收 20-已接收 30-处理中 40-已办结 50-已驳回 99-已取消',
    receive_time DATETIME DEFAULT NULL COMMENT '接收时间',
    receive_user_id BIGINT DEFAULT 0 COMMENT '接收人ID',
    receive_user_name VARCHAR(64) DEFAULT '' COMMENT '接收人姓名',
    receive_remark VARCHAR(500) DEFAULT '' COMMENT '接收备注',
    reject_reason VARCHAR(500) DEFAULT '' COMMENT '驳回原因',
    reject_time DATETIME DEFAULT NULL COMMENT '驳回时间',
    process_start_time DATETIME DEFAULT NULL COMMENT '处理开始时间',
    process_end_time DATETIME DEFAULT NULL COMMENT '处理结束时间',
    process_result VARCHAR(1000) DEFAULT '' COMMENT '处理结果',
    process_duration INT DEFAULT 0 COMMENT '处理时长(小时)',
    urge_count INT DEFAULT 0 COMMENT '催办次数',
    last_urge_time DATETIME DEFAULT NULL COMMENT '最后催办时间',
    first_urge_time DATETIME DEFAULT NULL COMMENT '首次催办时间',
    timeout_hours INT DEFAULT 72 COMMENT '超时时间(小时,默认72小时=3天)',
    is_timeout TINYINT DEFAULT 0 COMMENT '是否超时: 0-否 1-是',
    closed_at DATETIME DEFAULT NULL COMMENT '关闭时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    INDEX idx_case_id(case_id),
    INDEX idx_case_no(case_no),
    INDEX idx_status(status),
    INDEX idx_to_dept_code(to_dept_code),
    INDEX idx_from_dept_id(from_dept_id),
    INDEX idx_created_at(created_at),
    INDEX idx_is_timeout(is_timeout)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='纠纷转办记录表';

-- =====================================================
-- 3. 转办催办记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS dispute_transfer_urge (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    transfer_id BIGINT NOT NULL COMMENT '转办记录ID',
    transfer_no VARCHAR(32) NOT NULL COMMENT '转办编号',
    urge_type TINYINT NOT NULL COMMENT '催办类型: 1-用户催办 2-领导催办 3-系统超时催办',
    urge_source INT DEFAULT 2 COMMENT '催办来源: 1-系统 2-人工',
    operator_id BIGINT DEFAULT 0 COMMENT '操作人ID',
    operator_name VARCHAR(64) DEFAULT '' COMMENT '操作人姓名',
    urgency_level TINYINT DEFAULT 2 COMMENT '紧急程度: 1-特急 2-紧急 3-一般',
    urge_content VARCHAR(500) NOT NULL COMMENT '催办内容',
    notify_type VARCHAR(32) DEFAULT 'app,sms' COMMENT '通知方式',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_transfer_id(transfer_id),
    INDEX idx_urge_type(urge_type),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='转办催办记录表';

-- =====================================================
-- 4. 初始化预置转办部门数据
-- =====================================================
INSERT INTO dispute_transfer_template (template_name, dept_code, dept_name, dept_type, contact_person, contact_phone, description, applicable_types, sort_order, status) VALUES
('人力资源和社会保障局', 'DEPT_HR', '人力资源和社会保障局', 'HR', '张主任', '12333', '负责处理劳动争议、社保纠纷、工资拖欠等问题', '3,4,5', 1, 1),
('市场监督管理局', 'DEPT_MS', '市场监督管理局', 'MS', '李科长', '12315', '负责处理消费纠纷、产品质量、不正当竞争等问题', '6,7,8', 2, 1),
('公安局', 'DEPT_PS', '公安局', 'PS', '王警官', '110', '负责处理治安纠纷、刑事案件移送、户籍纠纷等问题', '9,10,11', 3, 1),
('人民法院', 'DEPT_CT', '人民法院', 'CT', '赵法官', '12368', '负责诉讼对接、司法确认、强制执行等', '12,13', 4, 1),
('卫生健康委员会', 'DEPT_HC', '卫生健康委员会', 'OT', '刘主任', '12320', '负责处理医疗纠纷、医患矛盾等问题', '14,15', 5, 1),
('住房和城乡建设局', 'DEPT_HS', '住房和城乡建设局', 'OT', '陈科长', '12345', '负责处理物业纠纷、房屋质量、装修纠纷等问题', '16,17,18', 6, 1),
('教育局', 'DEPT_ED', '教育局', 'OT', '孙主任', '12391', '负责处理校园纠纷、教育收费等问题', '19,20', 7, 1),
('交通运输局', 'DEPT_TT', '交通运输局', 'OT', '周科长', '12328', '负责处理交通事故赔偿、运输纠纷等问题', '21,22', 8, 1),
('生态环境局', 'DEPT_EP', '生态环境局', 'OT', '吴局长', '12369', '负责处理环境污染、噪声扰民等问题', '23,24', 9, 1),
('民政局', 'DEPT_MA', '民政局', 'OT', '郑科长', '12349', '负责处理婚姻家庭、养老服务、低保等纠纷', '25,26,27', 10, 1);
