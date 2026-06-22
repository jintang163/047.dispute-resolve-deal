-- =====================================================
-- 排查走访管理系统 - 数据库脚本
-- 创建时间: 2026-06-22
-- 包含: 排查任务、走访记录、签到打卡、积分系统、礼品管理
-- =====================================================

USE dispute_resolve;

-- =====================================================
-- 一、网格员信息表
-- =====================================================
CREATE TABLE IF NOT EXISTS grid_member (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    user_id BIGINT NOT NULL COMMENT '关联用户ID',
    member_no VARCHAR(32) NOT NULL UNIQUE COMMENT '网格员编号',
    real_name VARCHAR(64) NOT NULL COMMENT '姓名',
    phone VARCHAR(20) NOT NULL COMMENT '手机号',
    id_card VARCHAR(32) COMMENT '身份证号',
    gender TINYINT DEFAULT 0 COMMENT '性别: 0-未知 1-男 2-女',
    avatar VARCHAR(256) COMMENT '头像',
    organization_id BIGINT NOT NULL COMMENT '所属组织ID',
    organization_name VARCHAR(128) COMMENT '所属组织名称',
    grid_area VARCHAR(256) COMMENT '负责网格区域',
    grid_code VARCHAR(64) COMMENT '网格编码',
    work_address VARCHAR(256) COMMENT '工作地址',
    longitude DECIMAL(10,7) COMMENT '工作地点经度',
    latitude DECIMAL(10,7) COMMENT '工作地点纬度',
    entry_date DATE COMMENT '入职日期',
    points INT DEFAULT 0 COMMENT '当前积分',
    total_points INT DEFAULT 0 COMMENT '累计积分',
    task_count INT DEFAULT 0 COMMENT '完成任务数',
    visit_count INT DEFAULT 0 COMMENT '走访记录数',
    danger_count INT DEFAULT 0 COMMENT '上报隐患数',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-离职 1-在职 2-休假',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME COMMENT '删除时间',
    INDEX idx_user_id(user_id),
    INDEX idx_org_id(organization_id),
    INDEX idx_member_no(member_no),
    INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网格员信息表';

-- =====================================================
-- 二、排查任务表
-- =====================================================
CREATE TABLE IF NOT EXISTS patrol_task (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    task_no VARCHAR(32) NOT NULL UNIQUE COMMENT '任务编号',
    title VARCHAR(256) NOT NULL COMMENT '任务标题',
    description TEXT COMMENT '任务描述',
    task_type TINYINT NOT NULL DEFAULT 1 COMMENT '任务类型: 1-日常排查 2-专项排查 3-紧急排查 4-定点走访',
    priority TINYINT DEFAULT 3 COMMENT '优先级: 1-特急 2-紧急 3-一般 4-普通',
    assigner_id BIGINT NOT NULL COMMENT '下发人ID',
    assigner_name VARCHAR(64) COMMENT '下发人姓名',
    assignee_id BIGINT NOT NULL COMMENT '网格员ID',
    assignee_name VARCHAR(64) COMMENT '网格员姓名',
    organization_id BIGINT NOT NULL COMMENT '所属组织ID',
    plan_start_time DATETIME COMMENT '计划开始时间',
    plan_end_time DATETIME COMMENT '计划结束时间',
    actual_start_time DATETIME COMMENT '实际开始时间',
    actual_end_time DATETIME COMMENT '实际结束时间',
    point_count INT DEFAULT 0 COMMENT '排查点位数',
    completed_point_count INT DEFAULT 0 COMMENT '已完成点位数',
    status TINYINT DEFAULT 10 COMMENT '状态: 10-待执行 20-进行中 30-已完成 40-已逾期 99-已取消',
    points_reward INT DEFAULT 0 COMMENT '积分奖励',
    is_deleted TINYINT DEFAULT 0 COMMENT '是否删除: 0-否 1-是',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME COMMENT '删除时间',
    INDEX idx_task_no(task_no),
    INDEX idx_assignee_id(assignee_id),
    INDEX idx_assigner_id(assigner_id),
    INDEX idx_org_id(organization_id),
    INDEX idx_status(status),
    INDEX idx_priority(priority),
    INDEX idx_plan_time(plan_start_time, plan_end_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排查任务表';

-- =====================================================
-- 三、排查任务点位表
-- =====================================================
CREATE TABLE IF NOT EXISTS patrol_task_point (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    task_id BIGINT NOT NULL COMMENT '任务ID',
    point_name VARCHAR(128) NOT NULL COMMENT '点位名称',
    point_type VARCHAR(64) COMMENT '点位类型: 社区-学校-企业-商铺-重点人员',
    address VARCHAR(256) COMMENT '点位地址',
    longitude DECIMAL(10,7) NOT NULL COMMENT '经度',
    latitude DECIMAL(10,7) NOT NULL COMMENT '纬度',
    contact_person VARCHAR(64) COMMENT '联系人',
    contact_phone VARCHAR(20) COMMENT '联系电话',
    checkin_radius INT DEFAULT 100 COMMENT '签到半径(米)',
    sort_order INT DEFAULT 0 COMMENT '排序(路线顺序)',
    is_checked TINYINT DEFAULT 0 COMMENT '是否已签到: 0-否 1-是',
    checkin_time DATETIME COMMENT '签到时间',
    remark VARCHAR(512) COMMENT '备注',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_task_id(task_id),
    INDEX idx_is_checked(is_checked),
    INDEX idx_sort_order(sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排查任务点位表';

-- =====================================================
-- 四、签到打卡表
-- =====================================================
CREATE TABLE IF NOT EXISTS patrol_checkin (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    checkin_no VARCHAR(32) NOT NULL UNIQUE COMMENT '签到编号',
    task_id BIGINT COMMENT '任务ID',
    point_id BIGINT COMMENT '点位ID',
    member_id BIGINT NOT NULL COMMENT '网格员ID',
    member_name VARCHAR(64) COMMENT '网格员姓名',
    checkin_type TINYINT NOT NULL DEFAULT 1 COMMENT '签到类型: 1-任务签到 2-日常打卡 3-离岗签到',
    checkin_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '签到时间',
    longitude DECIMAL(10,7) NOT NULL COMMENT '签到经度',
    latitude DECIMAL(10,7) NOT NULL COMMENT '签到纬度',
    address VARCHAR(256) COMMENT '签到地址',
    location_accuracy DECIMAL(8,2) COMMENT '定位精度(米)',
    photo_url VARCHAR(512) COMMENT '签到照片URL',
    live_photo_url VARCHAR(512) COMMENT '活体检测照片URL',
    is_live_verified TINYINT DEFAULT 0 COMMENT '是否活体验证: 0-否 1-是',
    live_verify_score DECIMAL(5,2) COMMENT '活体检测分数',
    checkin_distance DECIMAL(8,2) COMMENT '与目标点位距离(米)',
    is_valid TINYINT DEFAULT 1 COMMENT '是否有效签到: 0-无效 1-有效',
    device_info VARCHAR(512) COMMENT '设备信息(JSON)',
    ip_address VARCHAR(64) COMMENT 'IP地址',
    remark VARCHAR(256) COMMENT '备注',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_checkin_no(checkin_no),
    INDEX idx_task_id(task_id),
    INDEX idx_point_id(point_id),
    INDEX idx_member_id(member_id),
    INDEX idx_checkin_time(checkin_time),
    INDEX idx_is_valid(is_valid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='签到打卡表';

-- =====================================================
-- 五、走访记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS patrol_visit_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    record_no VARCHAR(32) NOT NULL UNIQUE COMMENT '记录编号',
    task_id BIGINT COMMENT '关联任务ID',
    point_id BIGINT COMMENT '关联点位ID',
    checkin_id BIGINT COMMENT '关联签到ID',
    member_id BIGINT NOT NULL COMMENT '网格员ID',
    member_name VARCHAR(64) COMMENT '网格员姓名',
    organization_id BIGINT COMMENT '所属组织ID',
    visit_type TINYINT NOT NULL DEFAULT 1 COMMENT '走访类型: 1-日常走访 2-隐患排查 3-纠纷调解 4-重点人员走访 5-政策宣传',
    visit_object VARCHAR(128) COMMENT '走访对象',
    visit_object_phone VARCHAR(20) COMMENT '走访对象电话',
    visit_address VARCHAR(256) COMMENT '走访地址',
    longitude DECIMAL(10,7) COMMENT '走访经度',
    latitude DECIMAL(10,7) COMMENT '走访纬度',
    visit_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '走访时间',
    visit_duration INT DEFAULT 0 COMMENT '走访时长(分钟)',
    content TEXT COMMENT '走访内容',
    situation TEXT COMMENT '现场情况',
    problem_desc TEXT COMMENT '发现问题',
    handle_situation TEXT COMMENT '处理情况',
    next_plan TEXT COMMENT '下一步计划',
    photo_urls TEXT COMMENT '现场照片URL(JSON数组)',
    video_url VARCHAR(512) COMMENT '视频URL',
    audio_url VARCHAR(512) COMMENT '音频URL',
    has_danger TINYINT DEFAULT 0 COMMENT '是否发现隐患: 0-否 1-是',
    danger_id BIGINT COMMENT '关联隐患ID',
    has_dispute TINYINT DEFAULT 0 COMMENT '是否发现纠纷: 0-否 1-是',
    dispute_case_id BIGINT COMMENT '关联纠纷案件ID',
    status TINYINT DEFAULT 1 COMMENT '状态: 1-草稿 2-已提交 3-已审核',
    auditor_id BIGINT COMMENT '审核人ID',
    audit_time DATETIME COMMENT '审核时间',
    audit_remark VARCHAR(512) COMMENT '审核备注',
    points_reward INT DEFAULT 0 COMMENT '积分奖励',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME COMMENT '删除时间',
    INDEX idx_record_no(record_no),
    INDEX idx_task_id(task_id),
    INDEX idx_member_id(member_id),
    INDEX idx_org_id(organization_id),
    INDEX idx_visit_time(visit_time),
    INDEX idx_visit_type(visit_type),
    INDEX idx_status(status),
    INDEX idx_has_danger(has_danger),
    INDEX idx_has_dispute(has_dispute)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='走访记录表';

-- =====================================================
-- 六、隐患纠纷上报
-- =====================================================
CREATE TABLE IF NOT EXISTS hidden_danger (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    danger_no VARCHAR(32) NOT NULL UNIQUE COMMENT '隐患编号',
    reporter_id BIGINT NOT NULL COMMENT '上报人ID(网格员)',
    reporter_name VARCHAR(64) COMMENT '上报人姓名',
    organization_id BIGINT COMMENT '所属组织ID',
    danger_type TINYINT NOT NULL DEFAULT 1 COMMENT '隐患类型: 1-安全隐患 2-矛盾纠纷 3-环境卫生 4-治安问题 5-其他',
    danger_level TINYINT DEFAULT 3 COMMENT '隐患等级: 1-重大 2-较大 3-一般 4-轻微',
    title VARCHAR(256) NOT NULL COMMENT '隐患标题',
    description TEXT COMMENT '隐患描述',
    address VARCHAR(256) COMMENT '发生地址',
    longitude DECIMAL(10,7) COMMENT '经度',
    latitude DECIMAL(10,7) COMMENT '纬度',
    happen_time DATETIME COMMENT '发现时间',
    involved_person VARCHAR(128) COMMENT '涉及人员',
    involved_person_phone VARCHAR(20) COMMENT '涉及人员电话',
    photo_urls TEXT COMMENT '照片URL(JSON数组)',
    video_url VARCHAR(512) COMMENT '视频URL',
    is_dispute TINYINT DEFAULT 0 COMMENT '是否纠纷: 0-否 1-是',
    dispute_case_id BIGINT COMMENT '关联纠纷案件ID',
    handle_status TINYINT DEFAULT 10 COMMENT '处理状态: 10-待处理 20-处理中 30-已处理 40-已结案',
    handler_id BIGINT COMMENT '处理人ID',
    handler_name VARCHAR(64) COMMENT '处理人姓名',
    handle_result TEXT COMMENT '处理结果',
    handle_time DATETIME COMMENT '处理时间',
    points_reward INT DEFAULT 0 COMMENT '积分奖励',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME COMMENT '删除时间',
    INDEX idx_danger_no(danger_no),
    INDEX idx_reporter_id(reporter_id),
    INDEX idx_org_id(organization_id),
    INDEX idx_danger_type(danger_type),
    INDEX idx_danger_level(danger_level),
    INDEX idx_handle_status(handle_status),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='隐患纠纷上报表';

-- =====================================================
-- 七、网格员积分账户表
-- =====================================================
CREATE TABLE IF NOT EXISTS grid_member_points_account (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    member_id BIGINT NOT NULL UNIQUE COMMENT '网格员ID',
    member_name VARCHAR(64) COMMENT '网格员姓名',
    member_no VARCHAR(32) COMMENT '网格员编号',
    organization_id BIGINT COMMENT '所属组织ID',
    balance INT DEFAULT 0 COMMENT '当前积分余额',
    total_earned INT DEFAULT 0 COMMENT '累计获得积分',
    total_spent INT DEFAULT 0 COMMENT '累计消费积分',
    total_expired INT DEFAULT 0 COMMENT '累计过期积分',
    level INT DEFAULT 1 COMMENT '等级: 1-初级 2-中级 3-高级 4-专家',
    level_name VARCHAR(32) DEFAULT '初级网格员' COMMENT '等级名称',
    next_level_points INT DEFAULT 1000 COMMENT '下一等级所需积分',
    checkin_days INT DEFAULT 0 COMMENT '连续签到天数',
    total_checkin_days INT DEFAULT 0 COMMENT '累计签到天数',
    last_checkin_date DATE COMMENT '最后签到日期',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_member_id(member_id),
    INDEX idx_org_id(organization_id),
    INDEX idx_level(level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网格员积分账户表';

-- =====================================================
-- 八、积分流水表
-- =====================================================
CREATE TABLE IF NOT EXISTS points_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    record_no VARCHAR(32) NOT NULL UNIQUE COMMENT '流水编号',
    member_id BIGINT NOT NULL COMMENT '网格员ID',
    member_name VARCHAR(64) COMMENT '网格员姓名',
    organization_id BIGINT COMMENT '所属组织ID',
    type TINYINT NOT NULL COMMENT '类型: 1-获得 2-消费 3-过期 4-退还',
    type_name VARCHAR(32) COMMENT '类型名称',
    business_type VARCHAR(64) NOT NULL COMMENT '业务类型: task_completed-任务完成 checkin-签到 visit_record-走访 danger_report-隐患上报 exchange-礼品兑换',
    business_id BIGINT COMMENT '业务ID',
    business_no VARCHAR(64) COMMENT '业务编号',
    points INT NOT NULL COMMENT '变动积分(正为加,负为减)',
    balance_before INT DEFAULT 0 COMMENT '变动前余额',
    balance_after INT DEFAULT 0 COMMENT '变动后余额',
    description VARCHAR(256) COMMENT '变动描述',
    operator_id BIGINT COMMENT '操作人ID',
    operator_name VARCHAR(64) COMMENT '操作人姓名',
    expire_date DATE COMMENT '过期日期(获得时)',
    is_expired TINYINT DEFAULT 0 COMMENT '是否已过期',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_record_no(record_no),
    INDEX idx_member_id(member_id),
    INDEX idx_org_id(organization_id),
    INDEX idx_type(type),
    INDEX idx_business_type(business_type),
    INDEX idx_business_id(business_id),
    INDEX idx_created_at(created_at),
    INDEX idx_expire_date(expire_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分流水表';

-- =====================================================
-- 九、礼品表
-- =====================================================
CREATE TABLE IF NOT EXISTS gift (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    gift_no VARCHAR(32) NOT NULL UNIQUE COMMENT '礼品编号',
    name VARCHAR(128) NOT NULL COMMENT '礼品名称',
    category_id BIGINT COMMENT '分类ID',
    category_name VARCHAR(64) COMMENT '分类名称',
    description TEXT COMMENT '礼品描述',
    image_url VARCHAR(512) COMMENT '礼品图片URL',
    images TEXT COMMENT '多图URL(JSON数组)',
    points_required INT NOT NULL DEFAULT 0 COMMENT '所需积分',
    market_price DECIMAL(10,2) DEFAULT 0.00 COMMENT '市场价',
    stock INT DEFAULT 0 COMMENT '库存',
    sold_count INT DEFAULT 0 COMMENT '已兑换数量',
    sort_order INT DEFAULT 0 COMMENT '排序',
    is_hot TINYINT DEFAULT 0 COMMENT '是否热门: 0-否 1-是',
    is_new TINYINT DEFAULT 0 COMMENT '是否新品: 0-否 1-是',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-下架 1-上架',
    exchange_limit INT DEFAULT 0 COMMENT '每人限兑数量(0为不限)',
    valid_start_date DATE COMMENT '有效开始日期',
    valid_end_date DATE COMMENT '有效结束日期',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME COMMENT '删除时间',
    INDEX idx_gift_no(gift_no),
    INDEX idx_category_id(category_id),
    INDEX idx_status(status),
    INDEX idx_sort_order(sort_order),
    INDEX idx_is_hot(is_hot),
    INDEX idx_points_required(points_required)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='礼品表';

-- =====================================================
-- 十、礼品分类表
-- =====================================================
CREATE TABLE IF NOT EXISTS gift_category (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    category_code VARCHAR(64) NOT NULL UNIQUE COMMENT '分类编码',
    category_name VARCHAR(64) NOT NULL COMMENT '分类名称',
    icon VARCHAR(256) COMMENT '分类图标',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_category_code(category_code),
    INDEX idx_sort_order(sort_order),
    INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='礼品分类表';

-- =====================================================
-- 十一、礼品兑换记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS gift_exchange (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    exchange_no VARCHAR(32) NOT NULL UNIQUE COMMENT '兑换编号',
    member_id BIGINT NOT NULL COMMENT '网格员ID',
    member_name VARCHAR(64) COMMENT '网格员姓名',
    member_phone VARCHAR(20) COMMENT '网格员电话',
    organization_id BIGINT COMMENT '所属组织ID',
    gift_id BIGINT NOT NULL COMMENT '礼品ID',
    gift_name VARCHAR(128) COMMENT '礼品名称',
    gift_image VARCHAR(512) COMMENT '礼品图片',
    gift_points INT NOT NULL COMMENT '礼品所需积分',
    quantity INT DEFAULT 1 COMMENT '兑换数量',
    total_points INT NOT NULL COMMENT '消耗总积分',
    receiver_name VARCHAR(64) COMMENT '收货人姓名',
    receiver_phone VARCHAR(20) COMMENT '收货人电话',
    receiver_address VARCHAR(512) COMMENT '收货地址',
    express_company VARCHAR(64) COMMENT '快递公司',
    express_no VARCHAR(64) COMMENT '快递单号',
    status TINYINT DEFAULT 10 COMMENT '状态: 10-待审核 20-待发货 30-已发货 40-已完成 50-已取消',
    audit_id BIGINT COMMENT '审核人ID',
    audit_time DATETIME COMMENT '审核时间',
    audit_remark VARCHAR(512) COMMENT '审核备注',
    ship_time DATETIME COMMENT '发货时间',
    receive_time DATETIME COMMENT '收货时间',
    cancel_reason VARCHAR(512) COMMENT '取消原因',
    remark VARCHAR(512) COMMENT '备注',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_exchange_no(exchange_no),
    INDEX idx_member_id(member_id),
    INDEX idx_org_id(organization_id),
    INDEX idx_gift_id(gift_id),
    INDEX idx_status(status),
    INDEX idx_created_at(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='礼品兑换记录表';

-- =====================================================
-- 十二、积分规则配置表
-- =====================================================
CREATE TABLE IF NOT EXISTS points_rule (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    rule_code VARCHAR(64) NOT NULL UNIQUE COMMENT '规则编码',
    rule_name VARCHAR(128) NOT NULL COMMENT '规则名称',
    rule_type VARCHAR(64) NOT NULL COMMENT '规则类型: checkin-签到 task-任务 visit-走访 danger-隐患 other-其他',
    points INT NOT NULL DEFAULT 0 COMMENT '奖励积分',
    max_points_per_day INT DEFAULT 0 COMMENT '每日上限(0为不限)',
    max_points_per_month INT DEFAULT 0 COMMENT '每月上限(0为不限)',
    is_active TINYINT DEFAULT 1 COMMENT '是否启用: 0-否 1-是',
    description VARCHAR(512) COMMENT '规则描述',
    expire_days INT DEFAULT 365 COMMENT '积分有效期(天)',
    sort_order INT DEFAULT 0 COMMENT '排序',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_rule_code(rule_code),
    INDEX idx_rule_type(rule_type),
    INDEX idx_is_active(is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分规则配置表';

-- =====================================================
-- 插入初始数据 - 积分规则
-- =====================================================
INSERT INTO points_rule (rule_code, rule_name, rule_type, points, max_points_per_day, description, expire_days, sort_order) VALUES
('checkin_daily', '每日签到', 'checkin', 5, 5, '每日首次签到获得5积分', 365, 1),
('checkin_continuous_7', '连续签到7天', 'checkin', 50, 50, '连续签到7天额外奖励50积分', 365, 2),
('checkin_continuous_30', '连续签到30天', 'checkin', 300, 300, '连续签到30天额外奖励300积分', 365, 3),
('task_complete_normal', '完成普通任务', 'task', 20, 200, '完成普通排查任务获得20积分', 365, 4),
('task_complete_urgent', '完成紧急任务', 'task', 50, 500, '完成紧急排查任务获得50积分', 365, 5),
('task_complete_special', '完成专项任务', 'task', 30, 300, '完成专项排查任务获得30积分', 365, 6),
('visit_record_normal', '提交走访记录', 'visit', 10, 50, '提交一条走访记录获得10积分', 365, 7),
('visit_record_quality', '优质走访记录', 'visit', 30, 150, '走访记录被评为优质额外奖励20积分', 365, 8),
('danger_report_normal', '上报一般隐患', 'danger', 30, 300, '上报一般隐患获得30积分', 365, 9),
('danger_report_serious', '上报较大隐患', 'danger', 100, 1000, '上报较大隐患获得100积分', 365, 10),
('danger_report_critical', '上报重大隐患', 'danger', 300, 3000, '上报重大隐患获得300积分', 365, 11),
('danger_report_dispute', '上报矛盾纠纷', 'danger', 50, 500, '上报矛盾纠纷获得50积分', 365, 12);

-- =====================================================
-- 插入初始数据 - 礼品分类
-- =====================================================
INSERT INTO gift_category (category_code, category_name, sort_order) VALUES
('daily', '日用品', 1),
('digital', '数码产品', 2),
('food', '食品饮料', 3),
('home', '家居用品', 4),
('outdoor', '户外用品', 5),
('stationery', '文具办公', 6);

-- =====================================================
-- 插入初始数据 - 礼品
-- =====================================================
INSERT INTO gift (gift_no, name, category_id, category_name, description, points_required, market_price, stock, sort_order, is_hot, exchange_limit) VALUES
('GIFT001', '定制保温杯', 1, '日用品', '304不锈钢保温杯，500ml容量，印有综治中心Logo', 100, 68.00, 100, 1, 1, 2),
('GIFT002', '晴雨两用伞', 1, '日用品', '黑胶防晒晴雨伞，可折叠', 80, 45.00, 200, 2, 1, 3),
('GIFT003', '笔记本套装', 6, '文具办公', '优质皮面笔记本+签字笔套装', 50, 35.00, 300, 3, 0, 5),
('GIFT004', '蓝牙音箱', 2, '数码产品', '便携蓝牙音箱，音质清晰', 300, 199.00, 50, 4, 1, 1),
('GIFT005', '运动手环', 2, '数码产品', '智能运动手环，心率监测', 500, 299.00, 30, 5, 1, 1),
('GIFT006', '大米5kg', 3, '食品饮料', '优质东北大米，5kg装', 150, 88.00, 100, 6, 0, 2),
('GIFT007', '食用油5L', 3, '食品饮料', '压榨一级花生油，5L装', 200, 128.00, 80, 7, 0, 2),
('GIFT008', '抱枕被', 4, '家居用品', '多功能抱枕被，可展开做毯子', 120, 78.00, 150, 8, 0, 2),
('GIFT009', '运动背包', 5, '户外用品', '大容量运动双肩背包，防水面料', 250, 158.00, 60, 9, 1, 1),
('GIFT010', '电热水壶', 4, '家居用品', '食品级304不锈钢电热水壶，1.7L', 350, 228.00, 40, 10, 0, 1);
