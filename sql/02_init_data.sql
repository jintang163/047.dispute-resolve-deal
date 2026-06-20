-- =====================================================
-- 初始化数据脚本
-- =====================================================

USE dispute_resolve;

-- =====================================================
-- 一、初始化纠纷类型数据 (三级分类)
-- =====================================================

-- 一级分类
INSERT INTO dispute_type (type_code, type_name, parent_id, level, level_path, icon, sort_order, description) VALUES
('MARRIAGE_FAMILY', '婚姻家庭', 0, 1, '/1/', 'icon-family', 1, '婚姻家庭类纠纷'),
('NEIGHBORHOOD', '邻里纠纷', 0, 1, '/2/', 'icon-neighbor', 2, '邻里之间的纠纷'),
('PROPERTY', '财产纠纷', 0, 1, '/3/', 'icon-property', 3, '财产权益类纠纷'),
('LABOR', '劳务纠纷', 0, 1, '/4/', 'icon-labor', 4, '劳动劳务类纠纷'),
('CONSUMER', '消费纠纷', 0, 1, '/5/', 'icon-consumer', 5, '消费者权益类纠纷'),
('TRAFFIC', '交通纠纷', 0, 1, '/6/', 'icon-traffic', 6, '交通事故类纠纷'),
('MEDICAL', '医疗纠纷', 0, 1, '/7/', 'icon-medical', 7, '医疗医患类纠纷'),
('ADMINISTRATIVE', '行政纠纷', 0, 1, '/8/', 'icon-admin', 8, '行政机关类纠纷'),
('OTHER', '其他纠纷', 0, 1, '/9/', 'icon-other', 99, '其他类型纠纷');

-- 二级分类
INSERT INTO dispute_type (type_code, type_name, parent_id, level, level_path, sort_order) VALUES
('DIVORCE', '离婚纠纷', 1, 2, '/1/1/', 1),
('CUSTODY', '子女抚养', 1, 2, '/1/2/', 2),
('PROPERTY_DIVISION', '财产分割', 1, 2, '/1/3/', 3),
('ALIMONY', '赡养抚养', 1, 2, '/1/4/', 4),
('ADOPTION', '收养继承', 1, 2, '/1/5/', 5),

('NOISE', '噪音扰民', 2, 2, '/2/1/', 1),
('WATER_LEAK', '漏水渗水', 2, 2, '/2/2/', 2),
('PARKING', '停车位纠纷', 2, 2, '/2/3/', 3),
('ADJOINING', '相邻权纠纷', 2, 2, '/2/4/', 4),

('REAL_ESTATE', '房产纠纷', 3, 2, '/3/1/', 1),
('DEBT', '债权债务', 3, 2, '/3/2/', 2),
('CONTRACT', '合同纠纷', 3, 2, '/3/3/', 3),
('INTELLECTUAL', '知识产权', 3, 2, '/3/4/', 4),

('SALARY', '工资报酬', 4, 2, '/4/1/', 1),
('WORK_INJURY', '工伤赔偿', 4, 2, '/4/2/', 2),
('TERMINATION', '劳动合同解除', 4, 2, '/4/3/', 3),
('SOCIAL_INSURANCE', '社会保险', 4, 2, '/4/4/', 4),

('QUALITY', '商品质量', 5, 2, '/5/1/', 1),
('PRICE', '价格欺诈', 5, 2, '/5/2/', 2),
('AFTER_SALE', '售后服务', 5, 2, '/5/3/', 3);

-- 三级分类
INSERT INTO dispute_type (type_code, type_name, parent_id, level, level_path, sort_order) VALUES
('DIVORCE_AGREEMENT', '协议离婚', 6, 3, '/1/1/1/', 1),
('DIVORCE_LAWSUIT', '诉讼离婚', 6, 3, '/1/1/2/', 2),
('DIVORCE_PROPERTY', '离婚财产', 6, 3, '/1/1/3/', 3),

('CUSTODY_DISPUTE', '抚养权争议', 7, 3, '/1/2/1/', 1),
('VISITATION_RIGHT', '探望权纠纷', 7, 3, '/1/2/2/', 2),
('SUPPORT_FEE', '抚养费争议', 7, 3, '/1/2/3/', 3),

('MARITAL_PROPERTY', '夫妻共同财产', 8, 3, '/1/3/1/', 1),
('PREMARITAL_PROPERTY', '婚前财产', 8, 3, '/1/3/2/', 2),
('DEBT_BURDEN', '债务承担', 8, 3, '/1/3/3/', 3),

('ELDERLY_SUPPORT', '老人赡养', 9, 3, '/1/4/1/', 1),
('CHILD_SUPPORT', '子女抚养', 9, 3, '/1/4/2/', 2),
('SPOUSAL_SUPPORT', '夫妻扶养', 9, 3, '/1/4/3/', 3),

('NOISE_LIFE', '生活噪音', 11, 3, '/2/1/1/', 1),
('NOISE_CONSTRUCTION', '施工噪音', 11, 3, '/2/1/2/', 2),
('NOISE_COMMERCIAL', '商业噪音', 11, 3, '/2/1/3/', 3),

('WATER_UPPER', '楼上漏水', 12, 3, '/2/2/1/', 1),
('WATER_LOWER', '楼下渗水', 12, 3, '/2/2/2/', 2),
('WATER_PUBLIC', '公共区域漏水', 12, 3, '/2/2/3/', 3),

('PARKING_SPACE', '车位归属', 13, 3, '/2/3/1/', 1),
('PARKING_OCCUPY', '车位占用', 13, 3, '/2/3/2/', 2),
('PARKING_FEE', '停车费争议', 13, 3, '/2/3/3/', 3);

-- =====================================================
-- 二、初始化组织架构数据
-- =====================================================

INSERT INTO sys_organization (org_code, org_name, org_type, parent_id, level_path, address, longitude, latitude, sort_order) VALUES
('ZZZX_001', '市综治中心', 1, 0, '/1/', '市政中心A座', 116.4074, 39.9042, 1),
('JDB_001', '朝阳区街道办', 2, 1, '/1/2/', '朝阳区政府大楼', 116.4551, 39.9244, 1),
('JDB_002', '海淀区街道办', 2, 1, '/1/3/', '海淀区政府大楼', 116.3176, 39.9599, 2),
('SQ_001', '朝阳门社区', 3, 2, '/1/2/4/', '朝阳门南小街', 116.4280, 39.9196, 1),
('SQ_002', '建国门社区', 3, 2, '/1/2/5/', '建国门内大街', 116.4357, 39.9096, 2),
('SQ_003', '中关村社区', 3, 3, '/1/3/6/', '中关村大街', 116.3160, 39.9847, 1),
('CWH_001', '东升村委会', 4, 3, '/1/3/7/', '东升镇政府', 116.3500, 40.0200, 1);

-- =====================================================
-- 三、初始化用户数据 (密码: 123456, BCrypt加密后的值)
-- =====================================================

-- 密码: 123456 的 BCrypt 哈希
SET @password = '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE8ByOhJIrdAu2';

INSERT INTO sys_user (username, password, real_name, phone, id_card, gender, role, organization_id, position, specialty, longitude, latitude, status) VALUES
('admin', @password, '系统管理员', '13800000001', '110101199001010001', 1, 4, 1, '系统管理员', '系统管理,运维管理', 116.4074, 39.9042, 1),
('director', @password, '张主任', '13800000002', '110101198001010002', 1, 1, 1, '综治中心主任', '全面管理,政策指导', 116.4074, 39.9042, 1),
('leader1', @password, '李组长', '13800000003', '110101198501010003', 1, 2, 2, '调解组组长', '婚姻家庭,邻里纠纷', 116.4551, 39.9244, 1),
('leader2', @password, '王组长', '13800000004', '110101198601010004', 2, 2, 3, '调解组组长', '财产纠纷,劳务纠纷', 116.3176, 39.9599, 1),
('mediator1', @password, '赵调解员', '13800000005', '110101199001010005', 1, 3, 4, '调解员', '婚姻家庭,心理咨询', 116.4280, 39.9196, 1),
('mediator2', @password, '钱调解员', '13800000006', '110101199101010006', 2, 3, 5, '调解员', '邻里纠纷,物业纠纷', 116.4357, 39.9096, 1),
('mediator3', @password, '孙调解员', '13800000007', '110101199201010007', 1, 3, 6, '调解员', '知识产权,合同纠纷', 116.3160, 39.9847, 1),
('mediator4', @password, '周调解员', '13800000008', '110101199301010008', 2, 3, 7, '调解员', '劳务纠纷,工伤赔偿', 116.3500, 40.0200, 1);

-- =====================================================
-- 四、初始化角色权限数据
-- =====================================================

INSERT INTO sys_role_permission (role_code, role_name, permissions, description) VALUES
('ROLE_ADMIN', '系统管理员',
'["user:view","user:add","user:edit","user:delete","org:view","org:add","org:edit","org:delete","role:view","role:add","role:edit","role:delete","case:view","case:add","case:edit","case:delete","case:assign","case:urge","mediation:view","mediation:add","mediation:edit","approval:view","approval:submit","approval:approve","approval:reject","approval:transfer","approval:sign","workflow:view","workflow:add","workflow:edit","workflow:delete","video:view","video:create","video:join","esign:view","esign:create","esign:verify","stats:view","stats:export","dashboard:view","heatmap:view","performance:view","performance:edit","ai:view","ai:config","system:view","system:config","log:view","log:export"]',
'系统管理员，拥有所有权限'),

('ROLE_DIRECTOR', '综治中心主任',
'["user:view","org:view","case:view","case:urge","case:assign","case:escalate","mediation:view","approval:view","approval:approve","approval:reject","approval:transfer","approval:sign","video:view","esign:view","stats:view","stats:export","dashboard:view","heatmap:view","performance:view","ai:view","log:view"]',
'综治中心主任，拥有审批和管理权限'),

('ROLE_LEADER', '调解组组长',
'["user:view","org:view","case:view","case:assign","case:urge","mediation:view","mediation:add","mediation:edit","approval:view","approval:submit","approval:approve","approval:reject","approval:transfer","video:view","video:create","video:join","esign:view","esign:create","stats:view","dashboard:view","performance:view","ai:view"]',
'调解组组长，拥有复核审批权限'),

('ROLE_MEDIATOR', '调解员',
'["user:view","case:view","case:add","mediation:view","mediation:add","mediation:edit","approval:view","approval:submit","video:view","video:join","esign:view","esign:create","stats:view","ai:view"]',
'调解员，负责案件调解和记录');

-- =====================================================
-- 五、初始化消息模板数据
-- =====================================================

INSERT INTO notification_template (template_code, template_name, channel_type, title_template, content_template, params) VALUES
('CASE_CREATE', '案件创建通知', 'app', '新纠纷案件已创建', '您好，您提交的纠纷案件【{caseNo}】已成功受理，我们将尽快为您处理。请保持电话畅通。', '["caseNo","title","createTime"]'),
('CASE_ASSIGN', '案件分派通知', 'app,sms', '您有新的调解案件', '【{caseNo}】{title} 已分派给您，请及时处理。', '["caseNo","title","assignTime","mediatorName"]'),
('CASE_URGE', '案件催办通知', 'app,sms,wechat', '案件催办提醒', '您正在处理的案件【{caseNo}】已被催办，请加快处理进度。催办原因：{urgeContent}', '["caseNo","title","urgeContent","urgeTime"]'),
('CASE_ESCALATE', '案件升级通知', 'app,sms', '案件超时升级通知', '案件【{caseNo}】已超时未处理，已升级至{escalateLevel}。请立即处理。', '["caseNo","title","escalateLevel","pendingHours"]'),
('CASE_STATUS_CHANGE', '案件状态变更通知', 'app,wechat', '案件状态已变更', '您的案件【{caseNo}】状态已变更为：{statusName}。', '["caseNo","title","statusName","updateTime"]'),
('APPROVAL_REMIND', '审批待办提醒', 'app,sms', '您有新的审批待办', '案件【{caseNo}】已提交至您审批，请及时处理。', '["caseNo","title","submitUser","submitTime"]'),
('APPROVAL_RESULT', '审批结果通知', 'app', '审批结果通知', '您提交的案件【{caseNo}】审批{result}。意见：{opinion}', '["caseNo","title","result","opinion","approverName"]'),
('VIDEO_INVITE', '视频调解邀请', 'app,sms,wechat', '视频调解邀请', '请加入视频调解：{roomName}。房间号：{roomId}，密码：{password}，时间：{scheduledTime}', '["caseNo","roomName","roomId","password","scheduledTime"]'),
('SATISFACTION_REMIND', '满意度评价提醒', 'app,wechat', '请对服务进行评价', '案件【{caseNo}】已办结，请对我们的服务进行评价。您的评价对我们很重要。', '["caseNo","title","completeTime"]'),
('ESIGN_VERIFY', '电子签章验证', 'app', '电子签章验证完成', '您的文件{documentName}电子签章验证{result}。', '["documentName","result","verifyTime"]');

-- =====================================================
-- 六、初始化审批流程定义
-- =====================================================

INSERT INTO workflow_approval_definition (def_code, def_name, dispute_type_ids, approval_nodes, timeout_config, version, status, description) VALUES
('DEFAULT_APPROVAL', '默认审批流程', '',
'[
    {"code":"MEDIATOR_SUBMIT","name":"调解员提交","type":"start","approverRole":"mediator","timeout":0},
    {"code":"LEADER_REVIEW","name":"组长复核","type":"approval","approverRole":"leader","timeout":86400, "escalateTo":"director"},
    {"code":"DIRECTOR_APPROVAL","name":"主任审批","type":"approval","approverRole":"director","timeout":172800, "escalateTo":null},
    {"code":"CASE_CLOSE","name":"结案","type":"end","approverRole":"","timeout":0}
]',
'{
    "timeoutPolicy": "escalate",
    "urgeInterval": 43200,
    "maxUrgeCount": 3,
    "escalateLevels": ["leader","director","system"]
}',
1, 1, '默认三级审批流程：调解员→组长→主任'),

('SIMPLE_APPROVAL', '简易审批流程', '',
'[
    {"code":"MEDIATOR_SUBMIT","name":"调解员提交","type":"start","approverRole":"mediator","timeout":0},
    {"code":"LEADER_REVIEW","name":"组长审批","type":"approval","approverRole":"leader","timeout":86400, "escalateTo":"director"},
    {"code":"CASE_CLOSE","name":"结案","type":"end","approverRole":"","timeout":0}
]',
'{
    "timeoutPolicy": "escalate",
    "urgeInterval": 43200,
    "maxUrgeCount": 2,
    "escalateLevels": ["leader","system"]
}',
1, 1, '简易审批流程，适用于小额简单案件');

-- =====================================================
-- 七、初始化自助终端数据
-- =====================================================

INSERT INTO kiosk_device (device_code, device_name, device_model, location, organization_id, longitude, latitude, ip_address, mac_address, status) VALUES
('KIOSK_001', '综治中心大厅1号终端', 'KT-2024-PRO', '市综治中心一楼大厅', 1, 116.4074, 39.9042, '192.168.1.101', '00:1A:2B:3C:4D:01', 1),
('KIOSK_002', '综治中心大厅2号终端', 'KT-2024-PRO', '市综治中心一楼大厅', 1, 116.4075, 39.9043, '192.168.1.102', '00:1A:2B:3C:4D:02', 1),
('KIOSK_003', '朝阳区街道办终端', 'KT-2024-STD', '朝阳区街道办服务大厅', 2, 116.4551, 39.9244, '192.168.2.101', '00:1A:2B:3C:4D:03', 1),
('KIOSK_004', '海淀区街道办终端', 'KT-2024-STD', '海淀区街道办服务大厅', 3, 116.3176, 39.9599, '192.168.3.101', '00:1A:2B:3C:4D:04', 1),
('KIOSK_005', '朝阳门社区终端', 'KT-2024-MINI', '朝阳门社区服务中心', 4, 116.4280, 39.9196, '192.168.4.101', '00:1A:2B:3C:4D:05', 1);

-- =====================================================
-- 八、初始化法条数据 (示例数据)
-- =====================================================

INSERT INTO ai_law_article (law_code, law_name, article_no, article_title, article_content, category, tags, status) VALUES
('CIVIL_CODE', '中华人民共和国民法典', '第一千零七十六条', '协议离婚', '夫妻双方自愿离婚的，应当签订书面离婚协议，并亲自到婚姻登记机关申请离婚登记。离婚协议应当载明双方自愿离婚的意思表示和对子女抚养、财产以及债务处理等事项协商一致的意见。', '婚姻家庭', '离婚,协议离婚', 1),
('CIVIL_CODE', '中华人民共和国民法典', '第一千零七十九条', '诉讼离婚', '夫妻一方要求离婚的，可以由有关组织进行调解或者直接向人民法院提起离婚诉讼。人民法院审理离婚案件，应当进行调解；如果感情确已破裂，调解无效的，应当准予离婚。', '婚姻家庭', '离婚,诉讼离婚', 1),
('CIVIL_CODE', '中华人民共和国民法典', '第一千零八十四条', '离婚后的父母子女关系', '父母与子女间的关系，不因父母离婚而消除。离婚后，子女无论由父或者母直接抚养，仍是父母双方的子女。离婚后，父母对于子女仍有抚养、教育、保护的权利和义务。离婚后，不满两周岁的子女，以由母亲直接抚养为原则。', '婚姻家庭', '抚养权,子女抚养', 1),
('CIVIL_CODE', '中华人民共和国民法典', '第二百八十八条', '处理相邻关系的原则', '不动产的相邻权利人应当按照有利生产、方便生活、团结互助、公平合理的原则，正确处理相邻关系。', '物权', '相邻权,邻里纠纷', 1),
('CIVIL_CODE', '中华人民共和国民法典', '第二百九十四条', '相邻不动产之间不可量物侵害', '不动产权利人不得违反国家规定弃置固体废物，排放大气污染物、水污染物、土壤污染物、噪声、光辐射、电磁辐射等有害物质。', '物权', '噪音,环境污染', 1),
('LABOR_LAW', '中华人民共和国劳动合同法', '第三十条', '劳动报酬', '用人单位应当按照劳动合同约定和国家规定，向劳动者及时足额支付劳动报酬。用人单位拖欠或者未足额支付劳动报酬的，劳动者可以依法向当地人民法院申请支付令，人民法院应当依法发出支付令。', '劳动', '工资,劳动报酬', 1),
('LABOR_LAW', '中华人民共和国社会保险法', '第三十六条', '工伤保险', '职工因工作原因受到事故伤害或者患职业病，且经工伤认定的，享受工伤保险待遇；其中，经劳动能力鉴定丧失劳动能力的，享受伤残待遇。', '劳动', '工伤,工伤保险', 1),
('CONSUMER_LAW', '中华人民共和国消费者权益保护法', '第二十四条', '退货、更换、修理义务', '经营者提供的商品或者服务不符合质量要求的，消费者可以依照国家规定、当事人约定退货，或者要求经营者履行更换、修理等义务。', '消费者权益', '商品质量,退换货', 1),
('TRAFFIC_LAW', '中华人民共和国道路交通安全法', '第七十六条', '交通事故赔偿责任', '机动车发生交通事故造成人身伤亡、财产损失的，由保险公司在机动车第三者责任强制保险责任限额范围内予以赔偿；不足的部分，按照下列规定承担赔偿责任...', '交通', '交通事故,赔偿', 1),
('MEDICAL_LAW', '医疗纠纷预防和处理条例', '第二十二条', '医疗纠纷解决途径', '发生医疗纠纷，医患双方可以通过下列途径解决：（一）双方自愿协商；（二）申请人民调解；（三）申请行政调解；（四）向人民法院提起诉讼；（五）法律、法规规定的其他途径。', '医疗', '医疗纠纷,解决途径', 1);

-- =====================================================
-- 九、初始化绩效指标配置
-- =====================================================

INSERT INTO performance_indicator_config (indicator_code, indicator_name, indicator_type, weight, max_score, calculation_formula, target_value, description) VALUES
('TOTAL_CASE_COUNT', '受理案件数', 1, 20.00, 100.00, '案件数量 / 目标值 * 100', 20.00, '考核周期内受理的案件总数'),
('COMPLETION_RATE', '办结率', 2, 25.00, 100.00, '办结案件数 / 受理案件数 * 100', 90.00, '办结案件占受理案件的比例'),
('SUCCESS_RATE', '调解成功率', 2, 25.00, 100.00, '调解成功案件数 / 办结案件数 * 100', 85.00, '调解成功案件占办结案件的比例'),
('AVG_PROCESS_DAYS', '平均办理天数', 3, 15.00, 100.00, 'MAX(0, 100 - (实际天数 - 目标天数) * 10)', 7.00, '案件从受理到办结的平均天数'),
('SATISFACTION', '满意度评分', 4, 15.00, 100.00, '平均满意度 / 5 * 100', 4.50, '当事人满意度平均评分');

-- =====================================================
-- 十、初始化纠纷关键词词典 (初始种子)
-- =====================================================

INSERT INTO dispute_keyword_dict (keyword, category, frequency, source_type, status) VALUES
('噪音扰民', '纠纷性质', 0, 'manual', 1),
('施工噪音', '纠纷性质', 0, 'manual', 1),
('生活噪音', '纠纷性质', 0, 'manual', 1),
('商业噪音', '纠纷性质', 0, 'manual', 1),
('楼上漏水', '纠纷性质', 0, 'manual', 1),
('楼下渗水', '纠纷性质', 0, 'manual', 1),
('漏水赔偿', '纠纷性质', 0, 'manual', 1),
('公共区域漏水', '纠纷性质', 0, 'manual', 1),
('欠薪3个月', '程度', 0, 'manual', 1),
('拖欠工资', '行为', 0, 'manual', 1),
('工资拖欠', '行为', 0, 'manual', 1),
('欠薪', '行为', 0, 'manual', 1),
('拒付工资', '行为', 0, 'manual', 1),
('工伤赔偿', '纠纷性质', 0, 'manual', 1),
('劳动合同解除', '行为', 0, 'manual', 1),
('违法辞退', '行为', 0, 'manual', 1),
('社保未缴', '行为', 0, 'manual', 1),
('车位占用', '行为', 0, 'manual', 1),
('停车位纠纷', '纠纷性质', 0, 'manual', 1),
('停车费争议', '纠纷性质', 0, 'manual', 1),
('相邻权', '纠纷性质', 0, 'manual', 1),
('围墙争议', '纠纷性质', 0, 'manual', 1),
('通道占用', '行为', 0, 'manual', 1),
('采光遮挡', '纠纷性质', 0, 'manual', 1),
('协议离婚', '纠纷性质', 0, 'manual', 1),
('诉讼离婚', '纠纷性质', 0, 'manual', 1),
('抚养权争议', '纠纷性质', 0, 'manual', 1),
('探望权', '纠纷性质', 0, 'manual', 1),
('抚养费', '纠纷性质', 0, 'manual', 1),
('财产分割', '纠纷性质', 0, 'manual', 1),
('夫妻共同财产', '纠纷性质', 0, 'manual', 1),
('婚前财产', '纠纷性质', 0, 'manual', 1),
('老人赡养', '纠纷性质', 0, 'manual', 1),
('夫妻扶养', '纠纷性质', 0, 'manual', 1),
('房产纠纷', '纠纷性质', 0, 'manual', 1),
('债权债务', '纠纷性质', 0, 'manual', 1),
('合同违约', '行为', 0, 'manual', 1),
('拖欠货款', '行为', 0, 'manual', 1),
('商品质量', '纠纷性质', 0, 'manual', 1),
('价格欺诈', '行为', 0, 'manual', 1),
('售后服务', '纠纷性质', 0, 'manual', 1),
('虚假宣传', '行为', 0, 'manual', 1),
('交通事故', '纠纷性质', 0, 'manual', 1),
('医疗费赔偿', '纠纷性质', 0, 'manual', 1),
('医疗事故', '纠纷性质', 0, 'manual', 1),
('物业', '对象', 0, 'manual', 1),
('房东', '对象', 0, 'manual', 1),
('雇主', '对象', 0, 'manual', 1),
('邻里', '对象', 0, 'manual', 1),
('开发商', '对象', 0, 'manual', 1);
