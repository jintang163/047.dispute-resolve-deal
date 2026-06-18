-- =====================================================
-- 综治中心矛盾纠纷管理系统 - 初始化数据脚本
-- 数据库: TiDB 7.x / MySQL 8.x
-- 创建时间: 2026-06-18
-- =====================================================

USE dispute_resolve;

-- =====================================================
-- 一、组织架构数据 (4条记录: 3级架构)
-- =====================================================

INSERT INTO sys_organization (id, org_code, org_name, org_type, parent_id, level_path, leader_id, address, longitude, latitude, sort_order, status, created_at, updated_at) VALUES
(1, 'ZZZX', '区级综治中心', 1, 0, '/1', 2, 'XX区人民政府综合楼A座', 116.4074000, 39.9042000, 1, 1, NOW(), NOW()),
(2, 'JD001', '和平街道办事处', 2, 1, '/1/2', 3, '和平路123号', 116.4100000, 39.9080000, 1, 1, NOW(), NOW()),
(3, 'JD002', '建设街道办事处', 2, 1, '/1/3', 0, '建设路456号', 116.4050000, 39.9020000, 2, 1, NOW(), NOW()),
(4, 'SQ001', '阳光社区居委会', 3, 2, '/1/2/4', 0, '阳光花园小区内', 116.4120000, 39.9100000, 1, 1, NOW(), NOW());

-- =====================================================
-- 二、用户数据 (6条记录)
-- 密码均为 123456 的 BCrypt 哈希值 (cost=10)
-- $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy 是 123456 的哈希
-- =====================================================

INSERT INTO sys_user (id, username, password, real_name, phone, id_card, avatar, gender, email, role, organization_id, position, specialty, longitude, latitude, mediation_count, success_rate, last_login_at, status, created_at, updated_at) VALUES
(1, 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '系统管理员', '13800000001', '110101199001010001', '', 1, 'admin@dispute.com', 4, 1, '系统管理员', '系统管理,权限管理', 116.4074000, 39.9042000, 0, 0.00, NULL, 1, NOW(), NOW()),
(2, 'director', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '张主任', '13800000002', '110101198501010002', '', 1, 'director@dispute.com', 1, 1, '综治中心主任', '全面管理,重大纠纷', 116.4074000, 39.9042000, 0, 0.00, NULL, 1, NOW(), NOW()),
(3, 'leader', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '李组长', '13800000003', '110101198801010003', '', 1, 'leader@dispute.com', 2, 2, '调解组组长', '审批管理,疑难案件', 116.4100000, 39.9080000, 15, 85.00, NULL, 1, NOW(), NOW()),
(4, 'mediator1', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '王调解员', '13800000004', '110101199001010004', '', 1, 'mediator1@dispute.com', 3, 2, '专职调解员', '邻里纠纷,婚姻家庭', 116.4100000, 39.9080000, 50, 90.00, NULL, 1, NOW(), NOW()),
(5, 'mediator2', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '赵调解员', '13800000005', '110101199101010005', '', 2, 'mediator2@dispute.com', 3, 2, '专职调解员', '物业纠纷,消费维权', 116.4100000, 39.9080000, 45, 88.00, NULL, 1, NOW(), NOW()),
(6, 'mediator3', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '刘调解员', '13800000006', '110101199201010006', '', 1, 'mediator3@dispute.com', 3, 4, '社区调解员', '劳动争议,合同纠纷', 116.4120000, 39.9100000, 40, 85.00, NULL, 1, NOW(), NOW());

-- =====================================================
-- 三、纠纷类型数据 (54条: 6大类 × 3中类 × 3小类)
-- =====================================================

-- 一级分类: 6大类
INSERT INTO dispute_type (id, type_code, type_name, parent_id, level, level_path, sort_order, description, status, created_at, updated_at) VALUES
(1, 'LINLI', '邻里纠纷', 0, 1, '/1', 1, '邻里之间因生活琐事产生的纠纷', 1, NOW(), NOW()),
(2, 'HUNYIN', '婚姻家庭', 0, 1, '/2', 2, '婚姻、家庭、抚养、赡养等纠纷', 1, NOW(), NOW()),
(3, 'WUYE', '物业纠纷', 0, 1, '/3', 3, '业主与物业公司之间的纠纷', 1, NOW(), NOW()),
(4, 'XIAOFEI', '消费维权', 0, 1, '/4', 4, '消费者与经营者之间的纠纷', 1, NOW(), NOW()),
(5, 'LAODONG', '劳动争议', 0, 1, '/5', 5, '劳动者与用人单位之间的纠纷', 1, NOW(), NOW()),
(6, 'HETONG', '合同纠纷', 0, 1, '/6', 6, '合同履行、违约等纠纷', 1, NOW(), NOW());

-- 二级分类: 每大类3个中类 (共18个)
INSERT INTO dispute_type (id, type_code, type_name, parent_id, level, level_path, sort_order, description, status, created_at, updated_at) VALUES
-- 邻里纠纷下的中类
(7, 'LINLI_ZAOYIN', '噪音扰民', 1, 2, '/1/7', 1, '装修噪音、生活噪音等', 1, NOW(), NOW()),
(8, 'LINLI_SHEBEI', '设施使用', 1, 2, '/1/8', 2, '公共设施、停车位等纠纷', 1, NOW(), NOW()),
(9, 'LINLI_QITA', '其他邻里', 1, 2, '/1/9', 3, '其他邻里间纠纷', 1, NOW(), NOW()),
-- 婚姻家庭下的中类
(10, 'HUNYIN_LIHUN', '离婚纠纷', 2, 2, '/2/10', 1, '离婚、财产分割、子女抚养', 1, NOW(), NOW()),
(11, 'HUNYIN_FUYANG', '抚养赡养', 2, 2, '/2/11', 2, '子女抚养、老人赡养', 1, NOW(), NOW()),
(12, 'HUNYIN_JIATING', '家庭矛盾', 2, 2, '/2/12', 3, '其他家庭内部矛盾', 1, NOW(), NOW()),
-- 物业纠纷下的中类
(13, 'WUYE_WUYEFEI', '物业费争议', 3, 2, '/3/13', 1, '物业费收取、标准争议', 1, NOW(), NOW()),
(14, 'WUYE_FUWU', '服务质量', 3, 2, '/3/14', 2, '物业服务不达标', 1, NOW(), NOW()),
(15, 'WUYE_GONGGONG', '公共区域', 3, 2, '/3/15', 3, '公共区域使用、维护', 1, NOW(), NOW()),
-- 消费维权下的中类
(16, 'XIAOFEI_ZHILIANG', '质量问题', 4, 2, '/4/16', 1, '商品质量不合格', 1, NOW(), NOW()),
(17, 'XIAOFEI_FUWU', '服务消费', 4, 2, '/4/17', 2, '服务质量、虚假宣传', 1, NOW(), NOW()),
(18, 'XIAOFEI_TUIK', '退款退货', 4, 2, '/4/18', 3, '退款、退货、换货', 1, NOW(), NOW()),
-- 劳动争议下的中类
(19, 'LAODONG_GONGZI', '工资报酬', 5, 2, '/5/19', 1, '拖欠工资、加班费', 1, NOW(), NOW()),
(20, 'LAODONG_LAOBAO', '劳保福利', 5, 2, '/5/20', 2, '社保、公积金、工伤', 1, NOW(), NOW()),
(21, 'LAODONG_JIECHU', '解除合同', 5, 2, '/5/21', 3, '解除劳动合同、经济补偿', 1, NOW(), NOW()),
-- 合同纠纷下的中类
(22, 'HETONG_MAIMAI', '买卖合同', 6, 2, '/6/22', 1, '商品买卖合同纠纷', 1, NOW(), NOW()),
(23, 'HETONG_ZULIN', '租赁合同', 6, 2, '/6/23', 2, '房屋、设备租赁', 1, NOW(), NOW()),
(24, 'HETONG_FUWU', '服务合同', 6, 2, '/6/24', 3, '服务类合同纠纷', 1, NOW(), NOW());

-- 三级分类: 每中类3个小类 (共54个三级分类)
INSERT INTO dispute_type (id, type_code, type_name, parent_id, level, level_path, sort_order, description, mediation_days, warning_days, status, created_at, updated_at) VALUES
-- 噪音扰民下的小类
(25, 'LINLI_ZAOYIN_ZHUANGXIU', '装修噪音', 7, 3, '/1/7/25', 1, '装修施工噪音扰民', 15, 3, 1, NOW(), NOW()),
(26, 'LINLI_ZAOYIN_SHENGHUO', '生活噪音', 7, 3, '/1/7/26', 2, '家庭生活噪音扰民', 10, 2, 1, NOW(), NOW()),
(27, 'LINLI_ZAOYIN_YINYUE', '娱乐噪音', 7, 3, '/1/7/27', 3, '音响、乐器等噪音', 10, 2, 1, NOW(), NOW()),
-- 设施使用下的小类
(28, 'LINLI_SHEBEI_TINGCHE', '停车纠纷', 8, 3, '/1/8/28', 1, '停车位占用纠纷', 15, 3, 1, NOW(), NOW()),
(29, 'LINLI_SHEBEI_GONGYONG', '公用设施', 8, 3, '/1/8/29', 2, '楼道、走廊等使用纠纷', 10, 2, 1, NOW(), NOW()),
(30, 'LINLI_SHEBEI_WATER', '水电使用', 8, 3, '/1/8/30', 3, '水、电、气使用纠纷', 15, 3, 1, NOW(), NOW()),
-- 其他邻里下的小类
(31, 'LINLI_QITA_ZHAIJI', '宅基地纠纷', 9, 3, '/1/9/31', 1, '宅基地边界纠纷', 30, 7, 1, NOW(), NOW()),
(32, 'LINLI_QITA_LIUJIN', '漏水纠纷', 9, 3, '/1/9/32', 2, '房屋漏水引发纠纷', 20, 5, 1, NOW(), NOW()),
(33, 'LINLI_QITA_QITA', '其他纠纷', 9, 3, '/1/9/33', 3, '其他邻里纠纷', 15, 3, 1, NOW(), NOW()),
-- 离婚纠纷下的小类
(34, 'HUNYIN_LIHUN_XIEYI', '协议离婚', 10, 3, '/2/10/34', 1, '协议离婚财产分配', 30, 7, 1, NOW(), NOW()),
(35, 'HUNYIN_LIHUN_CAICHAN', '财产分割', 10, 3, '/2/10/35', 2, '离婚财产分割纠纷', 45, 10, 1, NOW(), NOW()),
(36, 'HUNYIN_LIHUN_ZINV', '子女抚养', 10, 3, '/2/10/36', 3, '子女抚养权争议', 30, 7, 1, NOW(), NOW()),
-- 抚养赡养下的小类
(37, 'HUNYIN_FUYANG_FEI', '抚养费纠纷', 11, 3, '/2/11/37', 1, '抚养费支付争议', 30, 7, 1, NOW(), NOW()),
(38, 'HUNYIN_FUYANG_SHANYANG', '赡养纠纷', 11, 3, '/2/11/38', 2, '老人赡养纠纷', 30, 7, 1, NOW(), NOW()),
(39, 'HUNYIN_FUYANG_FUYANG', '扶养纠纷', 11, 3, '/2/11/39', 3, '夫妻间扶养纠纷', 30, 7, 1, NOW(), NOW()),
-- 家庭矛盾下的小类
(40, 'HUNYIN_JIATING_FUQI', '夫妻矛盾', 12, 3, '/2/12/40', 1, '夫妻日常生活矛盾', 15, 3, 1, NOW(), NOW()),
(41, 'HUNYIN_JIATING_POXI', '婆媳矛盾', 12, 3, '/2/12/41', 2, '婆媳关系纠纷', 20, 5, 1, NOW(), NOW()),
(42, 'HUNYIN_JIATING_YICHAN', '遗产继承', 12, 3, '/2/12/42', 3, '遗产继承纠纷', 45, 10, 1, NOW(), NOW()),
-- 物业费争议下的小类
(43, 'WUYE_WUYEFEI_BIAOZHUN', '收费标准', 13, 3, '/3/13/43', 1, '物业费标准争议', 30, 7, 1, NOW(), NOW()),
(44, 'WUYE_WUYEFEI_QIANFEI', '欠费纠纷', 13, 3, '/3/13/44', 2, '业主拖欠物业费', 30, 7, 1, NOW(), NOW()),
(45, 'WUYE_WUYEFEI_ZENGZHI', '增值服务', 13, 3, '/3/13/45', 3, '增值服务收费争议', 15, 3, 1, NOW(), NOW()),
-- 服务质量下的小类
(46, 'WUYE_FUWU_BAOJIE', '保洁服务', 14, 3, '/3/14/46', 1, '小区保洁不达标', 15, 3, 1, NOW(), NOW()),
(47, 'WUYE_FUWU_BAOAN', '安保服务', 14, 3, '/3/14/47', 2, '小区安保问题', 20, 5, 1, NOW(), NOW()),
(48, 'WUYE_FUWU_LVHU', '绿化养护', 14, 3, '/3/14/48', 3, '小区绿化问题', 15, 3, 1, NOW(), NOW()),
-- 公共区域下的小类
(49, 'WUYE_GONGGONG_DIAN', '公共用电', 15, 3, '/3/15/49', 1, '公摊电费争议', 20, 5, 1, NOW(), NOW()),
(50, 'WUYE_GONGGONG_WEI', '公共维修', 15, 3, '/3/15/50', 2, '公共设施维修', 20, 5, 1, NOW(), NOW()),
(51, 'WUYE_GONGGONG_YONGDI', '公共用地', 15, 3, '/3/15/51', 3, '公共空间占用纠纷', 20, 5, 1, NOW(), NOW()),
-- 质量问题下的小类
(52, 'XIAOFEI_ZHILIANG_SHANGPIN', '商品质量', 16, 3, '/4/16/52', 1, '商品质量不合格', 15, 3, 1, NOW(), NOW()),
(53, 'XIAOFEI_ZHILIANG_JIAHUO', '假冒伪劣', 16, 3, '/4/16/53', 2, '假冒伪劣商品', 20, 5, 1, NOW(), NOW()),
(54, 'XIAOFEI_ZHILIANG_SANBAO', '三包服务', 16, 3, '/4/16/54', 3, '三包服务纠纷', 15, 3, 1, NOW(), NOW()),
-- 服务消费下的小类
(55, 'XIAOFEI_FUWU_CANYIN', '餐饮服务', 17, 3, '/4/17/55', 1, '餐饮服务纠纷', 10, 2, 1, NOW(), NOW()),
(56, 'XIAOFEI_FUWU_MEIRONG', '美容美发', 17, 3, '/4/17/56', 2, '美容美发服务纠纷', 15, 3, 1, NOW(), NOW()),
(57, 'XIAOFEI_FUWU_YULE', '休闲娱乐', 17, 3, '/4/17/57', 3, '娱乐场所服务纠纷', 15, 3, 1, NOW(), NOW()),
-- 退款退货下的小类
(58, 'XIAOFEI_TUIK_TUIHUO', '退货纠纷', 18, 3, '/4/18/58', 1, '商品退货纠纷', 10, 2, 1, NOW(), NOW()),
(59, 'XIAOFEI_TUIK_TUIKUAN', '退款纠纷', 18, 3, '/4/18/59', 2, '服务退款纠纷', 10, 2, 1, NOW(), NOW()),
(60, 'XIAOFEI_TUIK_HUANHUO', '换货纠纷', 18, 3, '/4/18/60', 3, '商品换货纠纷', 10, 2, 1, NOW(), NOW()),
-- 工资报酬下的小类
(61, 'LAODONG_GONGZI_QIANGFA', '拖欠工资', 19, 3, '/5/19/61', 1, '用人单位拖欠工资', 30, 7, 1, NOW(), NOW()),
(62, 'LAODONG_GONGZI_JIABAN', '加班工资', 19, 3, '/5/19/62', 2, '加班费支付争议', 30, 7, 1, NOW(), NOW()),
(63, 'LAODONG_GONGZI_NIANZHONG', '年终奖金', 19, 3, '/5/19/63', 3, '年终奖发放争议', 30, 7, 1, NOW(), NOW()),
-- 劳保福利下的小类
(64, 'LAODONG_LAOBAO_SHEBAO', '社保缴纳', 20, 3, '/5/20/64', 1, '社保缴纳争议', 45, 10, 1, NOW(), NOW()),
(65, 'LAODONG_LAOBAO_GONGSHANG', '工伤赔偿', 20, 3, '/5/20/65', 2, '工伤认定及赔偿', 60, 15, 1, NOW(), NOW()),
(66, 'LAODONG_LAOBAO_GONGJIJIN', '公积金', 20, 3, '/5/20/66', 3, '公积金缴纳争议', 45, 10, 1, NOW(), NOW()),
-- 解除合同下的小类
(67, 'LAODONG_JIECHU_ZIYUAN', '自愿离职', 21, 3, '/5/21/67', 1, '员工主动离职纠纷', 15, 3, 1, NOW(), NOW()),
(68, 'LAODONG_JIECHU_CAITUI', '公司辞退', 21, 3, '/5/21/68', 2, '公司辞退员工争议', 30, 7, 1, NOW(), NOW()),
(69, 'LAODONG_JIECHU_BUCHANG', '经济补偿', 21, 3, '/5/21/69', 3, '经济补偿金争议', 30, 7, 1, NOW(), NOW()),
-- 买卖合同下的小类
(70, 'HETONG_MAIMAI_FANGWU', '房屋买卖', 22, 3, '/6/22/70', 1, '房屋买卖合同纠纷', 60, 15, 1, NOW(), NOW()),
(71, 'HETONG_MAIMAI_ERSH', '二手车交易', 22, 3, '/6/22/71', 2, '二手车买卖纠纷', 30, 7, 1, NOW(), NOW()),
(72, 'HETONG_MAIMAI_PIFA', '批发零售', 22, 3, '/6/22/72', 3, '批发零售合同纠纷', 30, 7, 1, NOW(), NOW()),
-- 租赁合同下的小类
(73, 'HETONG_ZULIN_ZUFANG', '房屋租赁', 23, 3, '/6/23/73', 1, '房屋租赁合同纠纷', 30, 7, 1, NOW(), NOW()),
(74, 'HETONG_ZULIN_SHANGPU', '商铺租赁', 23, 3, '/6/23/74', 2, '商铺租赁合同纠纷', 45, 10, 1, NOW(), NOW()),
(75, 'HETONG_ZULIN_SHEBEI', '设备租赁', 23, 3, '/6/23/75', 3, '设备租赁合同纠纷', 30, 7, 1, NOW(), NOW()),
-- 服务合同下的小类
(76, 'HETONG_FUWU_WULIU', '物流服务', 24, 3, '/6/24/76', 1, '物流快递服务纠纷', 15, 3, 1, NOW(), NOW()),
(77, 'HETONG_FUWU_ZHIXUN', '咨询服务', 24, 3, '/6/24/77', 2, '咨询服务合同纠纷', 30, 7, 1, NOW(), NOW()),
(78, 'HETONG_FUWU_JISHU', '技术服务', 24, 3, '/6/24/78', 3, '技术服务合同纠纷', 45, 10, 1, NOW(), NOW());

-- =====================================================
-- 四、工作流定义数据 (2条: 特急/一般)
-- =====================================================

INSERT INTO workflow_approval_definition (id, def_code, def_name, dispute_type_ids, flowable_process_key, approval_nodes, timeout_config, version, status, description, created_at, updated_at) VALUES
(1, 'WF_EXTRA_URGENT', '特急案件审批流程', '', 'dispute_approval_extra_urgent',
 '[{"nodeCode":"NODE_MEDIATOR","nodeName":"调解员提交","nodeType":1,"approverRole":3,"approverRoleCode":"ROLE_MEDIATOR","timeoutHours":4,"escalateTo":2},{"nodeCode":"NODE_LEADER","nodeName":"组长审批","nodeType":1,"approverRole":2,"approverRoleCode":"ROLE_LEADER","timeoutHours":8,"escalateTo":1},{"nodeCode":"NODE_DIRECTOR","nodeName":"主任审批","nodeType":1,"approverRole":1,"approverRoleCode":"ROLE_DIRECTOR","timeoutHours":12,"escalateTo":0}]',
 '{"timeoutLevels":[{"level":1,"hours":4,"action":"urge","notifyRoles":[2]},{"level":2,"hours":8,"action":"escalate","notifyRoles":[1]},{"level":3,"hours":12,"action":"alert","notifyRoles":[1]}]}',
 1, 1, '特急案件三级审批: 调解员→组长→主任,总时限24小时', NOW(), NOW()),
(2, 'WF_NORMAL', '一般案件审批流程', '', 'dispute_approval_normal',
 '[{"nodeCode":"NODE_MEDIATOR","nodeName":"调解员提交","nodeType":1,"approverRole":3,"approverRoleCode":"ROLE_MEDIATOR","timeoutHours":24,"escalateTo":2},{"nodeCode":"NODE_LEADER","nodeName":"组长审批","nodeType":1,"approverRole":2,"approverRoleCode":"ROLE_LEADER","timeoutHours":48,"escalateTo":1},{"nodeCode":"NODE_DIRECTOR","nodeName":"主任审批","nodeType":1,"approverRole":1,"approverRoleCode":"ROLE_DIRECTOR","timeoutHours":72,"escalateTo":0}]',
 '{"timeoutLevels":[{"level":1,"hours":24,"action":"urge","notifyRoles":[2]},{"level":2,"hours":48,"action":"escalate","notifyRoles":[1]},{"level":3,"hours":72,"action":"alert","notifyRoles":[1]}]}',
 1, 1, '一般案件三级审批: 调解员→组长→主任,总时限144小时', NOW(), NOW());

-- =====================================================
-- 五、法条初始数据 (10条常见民法典法条)
-- =====================================================

INSERT INTO ai_law_article (id, law_code, law_name, article_no, article_title, article_content, category, tags, status, created_at, updated_at) VALUES
(1, 'MINFADIAN', '中华人民共和国民法典', '第二百八十八条', '处理相邻关系的原则', '不动产的相邻权利人应当按照有利生产、方便生活、团结互助、公平合理的原则，正确处理相邻关系。', '物权编', '相邻关系,邻里纠纷,处理原则', 1, NOW(), NOW()),
(2, 'MINFADIAN', '中华人民共和国民法典', '第二百九十四条', '相邻不动产之间弃置废物等行为的禁止', '不动产权利人不得违反国家规定弃置固体废物，排放大气污染物、水污染物、土壤污染物、噪声、光辐射、电磁辐射等有害物质。', '物权编', '噪音,污染,相邻关系,扰民', 1, NOW(), NOW()),
(3, 'MINFADIAN', '中华人民共和国民法典', '第一千零七十六条', '协议离婚', '夫妻双方自愿离婚的，应当签订书面离婚协议，并亲自到婚姻登记机关申请离婚登记。离婚协议应当载明双方自愿离婚的意思表示和对子女抚养、财产以及债务处理等事项协商一致的意见。', '婚姻家庭编', '离婚,协议离婚,子女抚养,财产分割', 1, NOW(), NOW()),
(4, 'MINFADIAN', '中华人民共和国民法典', '第一千零七十九条', '诉讼离婚', '夫妻一方要求离婚的，可以由有关组织进行调解或者直接向人民法院提起离婚诉讼。人民法院审理离婚案件，应当进行调解；如果感情确已破裂，调解无效的，应当准予离婚。', '婚姻家庭编', '离婚,诉讼离婚,调解,感情破裂', 1, NOW(), NOW()),
(5, 'MINFADIAN', '中华人民共和国民法典', '第一千零八十五条', '离婚后子女抚养费的负担', '离婚后，子女由一方直接抚养的，另一方应当负担部分或者全部抚养费。负担费用的多少和期限的长短，由双方协议；协议不成的，由人民法院判决。前款规定的协议或者判决，不妨碍子女在必要时向父母任何一方提出超过协议或者判决原定数额的合理要求。', '婚姻家庭编', '抚养费,子女抚养,离婚', 1, NOW(), NOW()),
(6, 'MINFADIAN', '中华人民共和国民法典', '第一千零六十七条', '父母的抚养义务和子女的赡养义务', '父母不履行抚养义务的，未成年子女或者不能独立生活的成年子女，有要求父母给付抚养费的权利。成年子女不履行赡养义务的，缺乏劳动能力或者生活困难的父母，有要求成年子女给付赡养费的权利。', '婚姻家庭编', '抚养,赡养,父母子女,法定义务', 1, NOW(), NOW()),
(7, 'MINFADIAN', '中华人民共和国民法典', '第五百七十七条', '违约责任的基本形态', '当事人一方不履行合同义务或者履行合同义务不符合约定的，应当承担继续履行、采取补救措施或者赔偿损失等违约责任。', '合同编', '合同,违约,违约责任,赔偿损失', 1, NOW(), NOW()),
(8, 'MINFADIAN', '中华人民共和国民法典', '第九百四十四条', '业主支付物业费的义务', '业主应当按照约定向物业服务人支付物业费。物业服务人已经按照约定和有关规定提供服务的，业主不得以未接受或者无需接受相关物业服务为由拒绝支付物业费。业主违反约定逾期不支付物业费的，物业服务人可以催告其在合理期限内支付；合理期限届满仍不支付的，物业服务人可以提起诉讼或者申请仲裁。', '合同编', '物业费,物业服务,业主,支付义务', 1, NOW(), NOW()),
(9, 'MINFADIAN', '中华人民共和国民法典', '第六十二条', '劳务派遣单位、用工单位及劳动者的权利义务', '用工单位应当履行下列义务：（一）执行国家劳动标准，提供相应的劳动条件和劳动保护；（二）告知被派遣劳动者的工作要求和劳动报酬；（三）支付加班费、绩效奖金，提供与工作岗位相关的福利待遇；（四）对在岗被派遣劳动者进行工作岗位所必需的培训；（五）连续用工的，实行正常的工资调整机制。', '劳动法', '劳动报酬,加班费,劳动条件,用工单位', 1, NOW(), NOW()),
(10, 'MINFADIAN', '中华人民共和国民法典', '第一千一百八十四条', '财产损失计算方式', '侵害他人财产的，财产损失按照损失发生时的市场价格或者其他合理方式计算。', '侵权责任编', '财产损失,损害赔偿,市场价格,侵权', 1, NOW(), NOW());

-- =====================================================
-- 六、通知模板数据 (5条)
-- =====================================================

INSERT INTO notification_template (id, template_code, template_name, channel_type, title_template, content_template, params, status, created_at, updated_at) VALUES
(1, 'TPL_CASE_ASSIGN', '案件分派通知', 'app', '【案件分派】您有新的纠纷案件待处理',
 '尊敬的{{mediatorName}}您好：\n您被分派了新的纠纷案件，请及时处理。\n案件编号：{{caseNo}}\n案件标题：{{caseTitle}}\n纠纷类型：{{caseType}}\n紧急程度：{{caseLevel}}\n报案人：{{applicantName}}\n报案电话：{{applicantPhone}}\n分派时间：{{assignTime}}\n请登录系统查看案件详情并开展调解工作。',
 '["mediatorName","caseNo","caseTitle","caseType","caseLevel","applicantName","applicantPhone","assignTime"]',
 1, NOW(), NOW()),
(2, 'TPL_APPROVAL_TODO', '审批待办通知', 'app', '【审批待办】您有新的审批任务待处理',
 '尊敬的{{approverName}}您好：\n您有新的审批任务待处理。\n审批编号：{{approvalNo}}\n关联案件：{{caseNo}} - {{caseTitle}}\n审批节点：{{nodeName}}\n提交人：{{submitterName}}\n提交时间：{{submitTime}}\n请在{{deadline}}前完成审批，超时将自动升级。',
 '["approverName","approvalNo","caseNo","caseTitle","nodeName","submitterName","submitTime","deadline"]',
 1, NOW(), NOW()),
(3, 'TPL_TIMEOUT_REMIND', '超时提醒通知', 'sms,app', '【超时提醒】案件处理超时，请尽快处理',
 '【超时提醒】案件编号{{caseNo}}已超时{{overdueHours}}小时。\n案件标题：{{caseTitle}}\n当前节点：{{currentNode}}\n当前处理人：{{handlerName}}\n超时级别：{{timeoutLevel}}\n请尽快处理，避免案件再次升级。\n系统将在{{nextEscalateTime}}自动升级到{{nextEscalateRole}}。',
 '["caseNo","caseTitle","currentNode","handlerName","overdueHours","timeoutLevel","nextEscalateTime","nextEscalateRole"]',
 1, NOW(), NOW()),
(4, 'TPL_SATISFACTION_EVAL', '满意度评价邀请', 'sms,app,wechat', '【服务评价】诚邀您对本次调解服务进行评价',
 '尊敬的{{userName}}您好：\n您的纠纷案件（编号：{{caseNo}}）已结案。\n案件标题：{{caseTitle}}\n调解员：{{mediatorName}}\n结案时间：{{closeTime}}\n调解结果：{{mediationResult}}\n请您对本次调解服务进行满意度评价，您的反馈是我们改进服务的动力。\n点击链接参与评价：{{evalUrl}}\n本邀请72小时内有效。',
 '["userName","caseNo","caseTitle","mediatorName","closeTime","mediationResult","evalUrl"]',
 1, NOW(), NOW()),
(5, 'TPL_MEDIATION_SUCCESS', '调解成功通知', 'app,sms', '【调解成功】您的案件已调解成功',
 '尊敬的{{userName}}您好：\n您的纠纷案件调解成功。\n案件编号：{{caseNo}}\n案件标题：{{caseTitle}}\n调解员：{{mediatorName}}\n调解协议：{{agreementSummary}}\n协议签署时间：{{signTime}}\n请按照调解协议履行各自义务，如需帮助可联系社区调解委员会。\n联系电话：{{contactPhone}}',
 '["userName","caseNo","caseTitle","mediatorName","agreementSummary","signTime","contactPhone"]',
 1, NOW(), NOW());

-- =====================================================
-- 七、自助终端设备 (1条默认设备)
-- =====================================================

INSERT INTO kiosk_device (id, device_code, device_name, device_model, location, organization_id, longitude, latitude, ip_address, mac_address, last_heartbeat, status, status_detail, total_register_count, today_register_count, id_card_reader_status, printer_status, camera_status, remark, created_at, updated_at) VALUES
(1, 'KIOSK001', '阳光社区自助调解终端', 'ZK-MT-2024-Pro', '阳光社区服务大厅一楼', 4, 116.4120000, 39.9100000, '192.168.1.101', '00:1A:2B:3C:4D:5E', NOW(), 1, '设备运行正常', 0, 0, 1, 1, 1, '默认自助调解终端，支持身份证读取、材料扫描、视频调解', NOW(), NOW());

-- =====================================================
-- 八、角色权限数据
-- =====================================================

INSERT INTO sys_role_permission (id, role_code, role_name, permissions, description, created_at, updated_at) VALUES
(1, 'ROLE_ADMIN', '管理员', '["*"]', '系统超级管理员，拥有所有权限', NOW(), NOW()),
(2, 'ROLE_DIRECTOR', '主任', '["case:*","approval:*","mediation:*","stats:*","performance:*","user:read","org:read","notification:*"]', '综治中心主任，负责全面管理和审批', NOW(), NOW()),
(3, 'ROLE_LEADER', '组长', '["case:read","case:assign","case:update","approval:*","mediation:read","mediation:update","stats:read","performance:read","notification:read"]', '调解组组长，负责案件分派和一级审批', NOW(), NOW()),
(4, 'ROLE_MEDIATOR', '调解员', '["case:read","case:update","mediation:*","approval:submit","notification:read"]', '专职调解员，负责案件调解和审批提交', NOW(), NOW());
