-- =====================================================
-- 心理咨询服务预约系统 - 数据库迁移脚本
-- 创建时间: 2026-06-21
-- =====================================================

USE dispute_resolve;

-- 心理咨询师表
CREATE TABLE IF NOT EXISTS counselor (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    counselor_no VARCHAR(50) NOT NULL UNIQUE COMMENT '心理咨询师编号',
    user_id BIGINT DEFAULT 0 COMMENT '关联系统用户ID',
    real_name VARCHAR(50) NOT NULL COMMENT '真实姓名',
    gender TINYINT DEFAULT 0 COMMENT '性别: 0-未知 1-男 2-女',
    phone VARCHAR(20) COMMENT '联系电话',
    email VARCHAR(100) COMMENT '邮箱',
    avatar VARCHAR(255) COMMENT '头像URL',
    title VARCHAR(100) DEFAULT '' COMMENT '职称: 初级心理咨询师/中级/高级',
    license_no VARCHAR(100) DEFAULT '' COMMENT '执业证书编号',
    specialty VARCHAR(500) DEFAULT '' COMMENT '专业领域: 婚姻家庭/创伤后应激/青少年心理/焦虑抑郁等(逗号分隔)',
    specialty_tags VARCHAR(500) DEFAULT '' COMMENT '擅长标签(逗号分隔)',
    years_of_experience INT DEFAULT 0 COMMENT '从业年限',
    education VARCHAR(100) DEFAULT '' COMMENT '学历',
    introduction TEXT COMMENT '个人简介',
    consultation_types VARCHAR(200) DEFAULT '' COMMENT '咨询方式: 1-线上视频 2-线上语音 3-线下面谈(逗号分隔)',
    work_days VARCHAR(100) DEFAULT '1,2,3,4,5' COMMENT '可预约工作日: 1-周一 2-周二 ... 7-周日',
    work_start_time TIME DEFAULT '09:00:00' COMMENT '每日可预约开始时间',
    work_end_time TIME DEFAULT '18:00:00' COMMENT '每日可预约结束时间',
    session_duration INT DEFAULT 50 COMMENT '单次咨询时长(分钟)',
    price DECIMAL(10,2) DEFAULT 0.00 COMMENT '咨询费用(元/次)',
    org_id BIGINT DEFAULT 0 COMMENT '所属机构ID',
    org_name VARCHAR(100) DEFAULT '' COMMENT '所属机构名称',
    rating_avg DECIMAL(3,2) DEFAULT 0.00 COMMENT '平均评分',
    rating_count INT DEFAULT 0 COMMENT '评价次数',
    appointment_count INT DEFAULT 0 COMMENT '预约次数',
    completed_count INT DEFAULT 0 COMMENT '完成咨询次数',
    is_emergency_available TINYINT DEFAULT 0 COMMENT '是否接受紧急干预: 0-否 1-是',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-停用 1-启用',
    sort_order INT DEFAULT 0 COMMENT '排序',
    created_by BIGINT DEFAULT 0 COMMENT '创建人ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_user_id(user_id),
    INDEX idx_status(status),
    INDEX idx_org_id(org_id),
    INDEX idx_rating_avg(rating_avg),
    INDEX idx_is_emergency(is_emergency_available),
    INDEX idx_deleted_at(deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='心理咨询师表';

-- 心理咨询预约表
CREATE TABLE IF NOT EXISTS counselor_appointment (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    appointment_no VARCHAR(50) NOT NULL UNIQUE COMMENT '预约编号',
    counselor_id BIGINT NOT NULL COMMENT '心理咨询师ID',
    counselor_name VARCHAR(50) DEFAULT '' COMMENT '心理咨询师姓名',
    case_id BIGINT DEFAULT 0 COMMENT '关联纠纷案件ID',
    party_id BIGINT DEFAULT 0 COMMENT '当事人ID(关联dispute_case_party)',
    party_name VARCHAR(50) DEFAULT '' COMMENT '当事人姓名',
    party_phone VARCHAR(20) DEFAULT '' COMMENT '当事人电话',
    party_id_card VARCHAR(18) DEFAULT '' COMMENT '当事人身份证号',
    is_anonymous TINYINT DEFAULT 0 COMMENT '是否匿名模式: 0-否 1-是',
    anonymous_code VARCHAR(50) DEFAULT '' COMMENT '匿名编号(匿名模式下显示)',
    appointment_date DATE NOT NULL COMMENT '预约日期',
    start_time TIME NOT NULL COMMENT '预约开始时间',
    end_time TIME NOT NULL COMMENT '预约结束时间',
    consultation_type TINYINT DEFAULT 1 COMMENT '咨询方式: 1-线上视频 2-线上语音 3-线下面谈',
    appointment_source TINYINT DEFAULT 1 COMMENT '预约来源: 1-管理员创建 2-自助预约 3-系统推荐',
    is_emergency TINYINT DEFAULT 0 COMMENT '是否紧急预约: 0-否 1-是(自杀/自残等风险)',
    emergency_trigger_words VARCHAR(500) DEFAULT '' COMMENT '触发紧急的关键词',
    emergency_level TINYINT DEFAULT 0 COMMENT '紧急级别: 0-普通 1-关注 2-紧急 3-高危',
    concern_type VARCHAR(100) DEFAULT '' COMMENT '咨询问题类型: 家暴/心理创伤/焦虑抑郁/家庭关系/其他',
    concern_description TEXT COMMENT '问题描述',
    status TINYINT DEFAULT 10 COMMENT '预约状态: 10-待确认 20-已确认 30-咨询中 40-已完成 50-已取消 60-已过期',
    cancel_reason VARCHAR(500) DEFAULT '' COMMENT '取消原因',
    cancelled_by BIGINT DEFAULT 0 COMMENT '取消人ID',
    cancelled_at DATETIME DEFAULT NULL COMMENT '取消时间',
    confirmed_by BIGINT DEFAULT 0 COMMENT '确认人ID',
    confirmed_at DATETIME DEFAULT NULL COMMENT '确认时间',
    started_at DATETIME DEFAULT NULL COMMENT '咨询开始时间',
    completed_at DATETIME DEFAULT NULL COMMENT '咨询完成时间',
    consultation_summary TEXT COMMENT '咨询摘要',
    follow_up_suggestion TEXT COMMENT '后续建议',
    next_appointment_suggestion VARCHAR(200) DEFAULT '' COMMENT '建议下次预约时间',
    room_id VARCHAR(100) DEFAULT '' COMMENT '线上咨询房间号',
    room_url VARCHAR(255) DEFAULT '' COMMENT '线上咨询链接',
    location VARCHAR(255) DEFAULT '' COMMENT '线下咨询地点',
    reminder_sent TINYINT DEFAULT 0 COMMENT '是否已发送提醒: 0-否 1-是',
    rating_submitted TINYINT DEFAULT 0 COMMENT '是否已提交评价: 0-否 1-是',
    created_by BIGINT DEFAULT 0 COMMENT '创建人ID',
    created_by_name VARCHAR(50) DEFAULT '' COMMENT '创建人姓名',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_counselor_id(counselor_id),
    INDEX idx_case_id(case_id),
    INDEX idx_party_id(party_id),
    INDEX idx_status(status),
    INDEX idx_appointment_date(appointment_date),
    INDEX idx_is_emergency(is_emergency),
    INDEX idx_is_anonymous(is_anonymous),
    INDEX idx_created_at(created_at),
    INDEX idx_deleted_at(deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='心理咨询预约表';

-- 心理咨询评价表
CREATE TABLE IF NOT EXISTS counselor_rating (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    appointment_id BIGINT NOT NULL COMMENT '预约ID',
    counselor_id BIGINT NOT NULL COMMENT '心理咨询师ID',
    rater_id BIGINT DEFAULT 0 COMMENT '评价人ID',
    rater_name VARCHAR(50) DEFAULT '' COMMENT '评价人姓名(匿名时留空)',
    is_anonymous_rating TINYINT DEFAULT 0 COMMENT '是否匿名评价: 0-否 1-是',
    overall_score TINYINT NOT NULL COMMENT '总体评分: 1-5星',
    professional_score TINYINT DEFAULT 0 COMMENT '专业度评分: 1-5',
    attitude_score TINYINT DEFAULT 0 COMMENT '态度评分: 1-5',
    empathy_score TINYINT DEFAULT 0 COMMENT '共情能力评分: 1-5',
    helpful_score TINYINT DEFAULT 0 COMMENT '帮助程度评分: 1-5',
    content TEXT COMMENT '评价内容',
    tags VARCHAR(500) DEFAULT '' COMMENT '评价标签: 专业/耐心/温暖/有帮助等(逗号分隔)',
    counselor_reply TEXT COMMENT '咨询师回复',
    counselor_reply_at DATETIME DEFAULT NULL COMMENT '咨询师回复时间',
    is_helpful INT DEFAULT 0 COMMENT '认为有帮助的人数',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-隐藏 1-显示',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_appointment_id(appointment_id),
    INDEX idx_counselor_id(counselor_id),
    INDEX idx_rater_id(rater_id),
    INDEX idx_overall_score(overall_score),
    INDEX idx_created_at(created_at),
    INDEX idx_deleted_at(deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='心理咨询评价表';

-- 咨询师日程表(记录不可预约时间段)
CREATE TABLE IF NOT EXISTS counselor_schedule (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    counselor_id BIGINT NOT NULL COMMENT '心理咨询师ID',
    schedule_date DATE NOT NULL COMMENT '日期',
    start_time TIME NOT NULL COMMENT '开始时间',
    end_time TIME NOT NULL COMMENT '结束时间',
    schedule_type TINYINT DEFAULT 1 COMMENT '类型: 1-休息 2-已有预约 3-其他事务',
    title VARCHAR(200) DEFAULT '' COMMENT '日程标题',
    remark VARCHAR(500) DEFAULT '' COMMENT '备注',
    appointment_id BIGINT DEFAULT 0 COMMENT '关联预约ID',
    created_by BIGINT DEFAULT 0 COMMENT '创建人ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    UNIQUE KEY uk_counselor_datetime(counselor_id, schedule_date, start_time, end_time),
    INDEX idx_counselor_date(counselor_id, schedule_date),
    INDEX idx_deleted_at(deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='咨询师日程表';

-- =====================================================
-- 初始化心理咨询师数据
-- =====================================================

INSERT INTO counselor (counselor_no, real_name, gender, phone, title, license_no, specialty, specialty_tags,
    years_of_experience, education, introduction, consultation_types, work_days, session_duration, price,
    org_name, is_emergency_available, status, sort_order) VALUES
('CS-2026-001', '李明心', 2, '13800138001', '高级心理咨询师', 'XL-2015-00123',
 '婚姻家庭,创伤后应激障碍,焦虑抑郁', '婚姻家庭,创伤疗愈,焦虑抑郁,情绪管理',
 12, '心理学硕士', '国家二级心理咨询师，从事心理咨询工作12年，擅长婚姻家庭问题、创伤后心理干预、焦虑抑郁情绪疏导。累计个案咨询时长3000+小时。',
 '1,2,3', '1,2,3,4,5,6', 50, 300.00, '市心理卫生中心', 1, 1, 1),

('CS-2026-002', '王安宁', 1, '13800138002', '中级心理咨询师', 'XL-2018-00456',
 '家庭暴力,心理创伤,青少年心理', '家暴干预,创伤疗愈,青少年心理,家庭治疗',
 8, '心理学学士', '国家三级心理咨询师，专注于家庭暴力受害者心理援助、创伤后应激障碍(PTSD)干预，具有丰富的危机干预经验。',
 '1,2,3', '1,2,3,4,5,6,7', 50, 200.00, '市心理卫生中心', 1, 1, 2),

('CS-2026-003', '张晓暖', 2, '13800138003', '高级心理咨询师', 'XL-2012-00078',
 '焦虑抑郁,情绪管理,职场压力', '焦虑抑郁,情绪疏导,职场减压,正念疗法',
 15, '临床心理学博士', '临床心理学博士，副教授，擅长运用认知行为疗法(CBT)、正念减压疗法(MBSR)治疗焦虑、抑郁等情绪障碍。',
 '1,2', '2,3,4,5,6', 50, 500.00, '大学心理咨询中心', 0, 1, 3),

('CS-2026-004', '陈志强', 1, '13800138004', '中级心理咨询师', 'XL-2019-00789',
 '家庭关系,亲子沟通,婚姻情感', '家庭治疗,亲子关系,婚姻情感,沟通技巧',
 6, '应用心理学硕士', '系统式家庭治疗师，擅长家庭关系调解、亲子沟通指导、婚姻情感咨询。',
 '1,2,3', '1,3,5,6,7', 50, 250.00, '市心理卫生中心', 0, 1, 4),

('CS-2026-005', '刘思雨', 2, '13800138005', '初级心理咨询师', 'XL-2022-00234',
 '青少年心理,学业压力,人际关系', '青少年成长,学业压力,人际关系,自我认知',
 3, '应用心理学学士', '专注于青少年心理健康领域，擅长处理学业压力、同伴关系、青春期心理困惑等问题。',
 '1,2,3', '1,2,3,4,5,6,7', 50, 150.00, '青年心理服务中心', 1, 1, 5);
