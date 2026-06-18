# 三端闭环流程 API 联调测试文档

## 概述

本文档描述了**自助终端(Kiosk) → 管理端(Admin) → 调解员(Mediator) → 审批流程 → 小程序端(MiniApp)** 的完整闭环业务流程的API测试方法。

---

## 1. 启动后端服务

### Windows 环境

```powershell
cd backend
.\start-all.ps1
```

### Linux/Mac 环境

```bash
cd backend
chmod +x start-all.sh
./start-all.sh
```

服务启动后可访问健康检查接口验证：

```bash
curl -X GET http://localhost:8080/api/v1/public/health
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok",
    "time": 1718688000,
    "service": "dispute-gateway"
  }
}
```

---

## 2. 自助终端提交案件

### 2.1 获取纠纷类型列表

```bash
curl -X GET http://localhost:8080/api/v1/public/dispute/types
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "typeName": "邻里纠纷",
      "parentId": 0,
      "level": 1,
      "status": 1
    },
    {
      "id": 2,
      "typeName": "物业纠纷",
      "parentId": 0,
      "level": 1,
      "status": 1
    }
  ]
}
```

### 2.2 自助终端提交纠纷案件

```bash
curl -X POST http://localhost:8080/api/v1/public/dispute/kiosk/create \
  -H "Content-Type: application/json" \
  -d '{
    "title": "楼上漏水导致楼下财产损失纠纷",
    "description": "楼上住户卫生间防水破损，漏水导致楼下天花板及家具受损，双方就赔偿金额无法达成一致。",
    "typeId": 1,
    "caseLevel": 2,
    "caseSource": 1,
    "reporterName": "张三",
    "reporterPhone": "13800138001",
    "reporterIdCard": "110101199001011234",
    "reporterAddress": "北京市朝阳区某小区1号楼101室",
    "respondentName": "李四",
    "respondentPhone": "13800138002",
    "respondentAddress": "北京市朝阳区某小区1号楼201室",
    "occurAddress": "北京市朝阳区某小区1号楼",
    "occurTime": "2024-06-15 14:30:00",
    "expectation": "要求楼上住户修复防水并赔偿损失5000元",
    "longitude": 116.4074,
    "latitude": 39.9042,
    "evidenceIds": []
  }'
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1001,
    "caseNo": "JJ202406180001",
    "title": "楼上漏水导致楼下财产损失纠纷",
    "status": 1,
    "statusName": "待分派",
    "createdAt": "2024-06-18 10:30:00"
  }
}
```

---

## 3. 管理端登录与案件分派

### 3.1 管理端登录（组长/主任账号）

```bash
curl -X POST http://localhost:8080/api/v1/public/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "leader01",
    "password": "Admin@123",
    "captcha": ""
  }'
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 7200,
    "userInfo": {
      "userId": 2,
      "username": "leader01",
      "realName": "王组长",
      "phone": "13900139001",
      "avatar": "",
      "role": 3,
      "roleName": "组长",
      "organizationId": 1,
      "orgName": "朝阳区调解中心",
      "position": "调解组长"
    }
  }
}
```

> **注意**：请保存返回的 `token`，后续所有需要鉴权的接口都要在请求头中携带：`Authorization: Bearer <token>`

### 3.2 获取待分派案件列表

```bash
curl -X GET "http://localhost:8080/api/v1/dispute?page=1&pageSize=10&status=1" \
  -H "Authorization: Bearer <token>"
```

**查询参数说明：**
- `page`: 页码，默认1
- `pageSize`: 每页条数，默认10
- `status`: 案件状态（1=待分派，2=调解中，3=待审批，4=已完成，5=已终止）
- `keyword`: 关键词搜索
- `typeId`: 纠纷类型ID
- `startTime`: 开始时间
- `endTime`: 结束时间

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1001,
        "caseNo": "JJ202406180001",
        "title": "楼上漏水导致楼下财产损失纠纷",
        "typeId": 1,
        "typeName": "邻里纠纷",
        "status": 1,
        "statusName": "待分派",
        "caseLevel": 2,
        "caseLevelName": "一般",
        "reporterName": "张三",
        "reporterPhone": "13800138001",
        "mediatorId": 0,
        "mediatorName": null,
        "organizationId": 1,
        "orgName": "朝阳区调解中心",
        "createdAt": "2024-06-18 10:30:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

### 3.3 获取调解员列表（用于分派）

```bash
curl -X GET "http://localhost:8080/api/v1/system/mediator?organizationId=1&specialtyId=0" \
  -H "Authorization: Bearer <token>"
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 101,
      "realName": "赵调解员",
      "username": "mediator01",
      "phone": "13700137001",
      "role": 2,
      "avatar": "",
      "organizationId": 1,
      "specialtyIds": [1, 2],
      "currentLoad": 3,
      "status": 1
    }
  ]
}
```

### 3.4 分派案件给调解员

```bash
curl -X POST http://localhost:8080/api/v1/dispute/1001/assign \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "mediatorId": 101,
    "reason": "该调解员擅长邻里纠纷调解，当前案件负载适中"
  }'
```

**URL参数说明：**
- `1001`: 案件ID

**响应示例：**
```json
{
  "code": 0,
  "message": "案件分派成功",
  "data": {
    "id": 1001,
    "status": 2,
    "statusName": "调解中",
    "mediatorId": 101,
    "mediatorName": "赵调解员",
    "assignedAt": "2024-06-18 11:00:00"
  }
}
```

---

## 4. 调解员记录调解记录

### 4.1 调解员账号登录

```bash
curl -X POST http://localhost:8080/api/v1/public/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "mediator01",
    "password": "Mediator@123",
    "captcha": ""
  }'
```

### 4.2 记录调解过程

```bash
curl -X POST http://localhost:8080/api/v1/dispute/1001/mediation \
  -H "Authorization: Bearer <mediator_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "recordType": 1,
    "mediationTime": "2024-06-18 14:00:00",
    "mediationPlace": "朝阳区调解中心第一调解室",
    "mediationDuration": 120,
    "processContent": "双方当事人到场，先分别了解各方诉求。申请人张三诉求为赔偿5000元及修复防水；被申请人李四认可漏水事实，但认为赔偿金额过高，只愿承担2000元。经过反复沟通协商，释明相关法律规定...",
    "disputeFocus": "1. 漏水原因及责任认定；2. 财产损失金额认定；3. 修复方案及期限",
    "mediationOpinion": "本案事实清楚，责任明确，建议双方各让一步，被申请人在30日内完成防水修复，并一次性赔偿申请人3500元。",
    "agreementContent": "1. 李四于2024年7月18日前聘请有资质的施工队完成卫生间防水重做，并承担全部费用；2. 李四于本协议签订之日起3日内一次性支付张三赔偿金人民币3500元整；3. 本协议履行完毕后，双方就此事不再有任何争议。",
    "result": 1,
    "nextStep": "等待双方签署调解协议后提交审批",
    "participants": ["张三", "李四", "赵调解员"],
    "assistMediators": []
  }'
```

**请求参数说明：**
- `recordType`: 记录类型（1=初次调解，2=再次调解，3=现场调解，4=线上调解）
- `mediationDuration`: 调解时长（分钟）
- `result`: 调解结果（1=达成协议，2=部分达成，3=未达成，4=继续调解）

**响应示例：**
```json
{
  "code": 0,
  "message": "调解记录保存成功",
  "data": {
    "id": 2001,
    "caseId": 1001,
    "result": 1,
    "createdAt": "2024-06-18 16:00:00"
  }
}
```

---

## 5. 三级审批流程

### 5.1 调解员提交审批

```bash
curl -X POST http://localhost:8080/api/v1/dispute/1001/approval/submit \
  -H "Authorization: Bearer <mediator_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "caseId": 1001,
    "defCode": "DEFAULT_APPROVAL",
    "remark": "本案调解过程规范，双方自愿达成协议，事实清楚，适用法律正确，提请审批。"
  }'
```

**响应示例：**
```json
{
  "code": 0,
  "message": "审批已提交，等待组长复核",
  "data": {
    "approvalId": 3001,
    "instanceNo": "AP202406180001",
    "status": 2,
    "statusName": "组长复核中",
    "currentNode": "组长复核",
    "currentApprover": "王组长",
    "submittedAt": "2024-06-18 17:00:00"
  }
}
```

### 5.2 组长复核（通过）

使用组长账号登录后执行：

```bash
curl -X POST http://localhost:8080/api/v1/dispute/1001/approval/approve \
  -H "Authorization: Bearer <leader_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "action": 1,
    "opinion": "经复核，本案调解程序合法，协议内容公平合理，符合相关法律规定，同意调解协议，提请主任审批。",
    "nextApproverId": 0
  }'
```

**请求参数说明：**
- `action`: 审批动作（1=通过，2=驳回，3=退回修改，4=加签，5=转办）

**响应示例：**
```json
{
  "code": 0,
  "message": "组长复核通过，等待主任审批",
  "data": {
    "approvalId": 3001,
    "status": 3,
    "statusName": "主任审批中",
    "currentNode": "主任审批",
    "currentApprover": "刘主任"
  }
}
```

### 5.3 主任审批（通过）

使用主任账号登录后执行：

```bash
curl -X POST http://localhost:8080/api/v1/dispute/1001/approval/approve \
  -H "Authorization: Bearer <director_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "action": 1,
    "opinion": "同意调解协议。调解中心公章已加盖，本调解协议具有法律约束力，双方当事人应当按照约定履行。",
    "nextApproverId": 0
  }'
```

**响应示例：**
```json
{
  "code": 0,
  "message": "审批完成，案件已结案",
  "data": {
    "approvalId": 3001,
    "status": 5,
    "statusName": "审批通过",
    "caseStatus": 4,
    "caseStatusName": "已完成",
    "approvedAt": "2024-06-19 09:30:00"
  }
}
```

### 5.4 查看审批进度

```bash
curl -X GET http://localhost:8080/api/v1/dispute/1001/approval \
  -H "Authorization: Bearer <token>"
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "approvalId": 3001,
    "instanceNo": "AP202406180001",
    "status": 5,
    "statusName": "审批通过",
    "currentNode": "已完成",
    "nodes": [
      {
        "nodeName": "提交申请",
        "approverName": "赵调解员",
        "action": "提交",
        "opinion": "本案调解过程规范...",
        "operatedAt": "2024-06-18 17:00:00"
      },
      {
        "nodeName": "组长复核",
        "approverName": "王组长",
        "action": "通过",
        "opinion": "经复核，本案调解程序合法...",
        "operatedAt": "2024-06-18 18:00:00"
      },
      {
        "nodeName": "主任审批",
        "approverName": "刘主任",
        "action": "通过",
        "opinion": "同意调解协议...",
        "operatedAt": "2024-06-19 09:30:00"
      }
    ]
  }
}
```

---

## 6. 小程序端进度查询与评价

### 6.1 查询案件调解进度（公开接口，无需登录）

```bash
curl -X GET "http://localhost:8080/api/v1/public/dispute/progress?caseNo=JJ202406180001&phone=13800138001"
```

**查询参数说明：**
- `caseNo`: 案件编号
- `phone`: 申请人手机号（用于验证身份）

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "caseId": 1001,
    "caseNo": "JJ202406180001",
    "title": "楼上漏水导致楼下财产损失纠纷",
    "status": 4,
    "statusName": "已完成",
    "mediatorName": "赵调解员",
    "mediatorPhone": "13700137001",
    "timeline": [
      {
        "time": "2024-06-18 10:30:00",
        "title": "案件提交",
        "description": "您已成功提交纠纷申请，案件编号JJ202406180001"
      },
      {
        "time": "2024-06-18 11:00:00",
        "title": "案件分派",
        "description": "案件已分派给赵调解员处理"
      },
      {
        "time": "2024-06-18 16:00:00",
        "title": "调解完成",
        "description": "双方已达成调解协议"
      },
      {
        "time": "2024-06-19 09:30:00",
        "title": "案件结案",
        "description": "审批通过，案件已完成，请按协议履行"
      }
    ],
    "agreementSummary": "李四于2024年7月18日前完成卫生间防水修复，并支付张三赔偿金人民币3500元整。",
    "canRate": true,
    "rating": null
  }
}
```

### 6.2 满意度评价

（此接口需要小程序登录获取token，此处演示评价提交格式）

```bash
curl -X POST http://localhost:8080/api/v1/dispute/1001/rate \
  -H "Authorization: Bearer <miniapp_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "rating": 5,
    "mediatorRating": 5,
    "processRating": 5,
    "resultRating": 5,
    "comment": "调解员非常专业耐心，全程跟进，最终帮我们圆满解决了问题。调解中心的服务态度和效率都值得称赞！",
    "suggestions": "希望可以增加线上视频调解功能，不用每次都跑现场。"
  }'
```

**请求参数说明：**
- `rating`: 总体评分（1-5星）
- `mediatorRating`: 调解员评分
- `processRating`: 流程便捷性评分
- `resultRating`: 结果满意度评分

**响应示例：**
```json
{
  "code": 0,
  "message": "感谢您的评价！",
  "data": {
    "ratingId": 4001,
    "ratedAt": "2024-06-20 10:00:00"
  }
}
```

---

## 7. 管理端数据大屏统计

### 7.1 获取Dashboard统计数据

```bash
curl -X GET "http://localhost:8080/api/v1/stats/dashboard?period=month&organizationId=1" \
  -H "Authorization: Bearer <admin_token>"
```

**查询参数说明：**
- `period`: 统计周期（week=本周，month=本月，quarter=本季度，year=本年，all=全部）
- `organizationId`: 组织机构ID（0表示当前用户所属机构）

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "summary": {
      "totalCases": 1286,
      "pendingCases": 86,
      "mediatingCases": 152,
      "completedCases": 986,
      "successRate": 76.7,
      "avgDurationDays": 12.5,
      "newThisMonth": 128,
      "completedThisMonth": 96
    },
    "trend": {
      "dates": ["06-12", "06-13", "06-14", "06-15", "06-16", "06-17", "06-18"],
      "newCases": [18, 22, 15, 25, 20, 16, 12],
      "completedCases": [12, 18, 14, 20, 22, 15, 10]
    },
    "typeDistribution": [
      { "name": "邻里纠纷", "value": 356, "percentage": 27.7 },
      { "name": "物业纠纷", "value": 298, "percentage": 23.2 },
      { "name": "婚姻家庭", "value": 186, "percentage": 14.5 },
      { "name": "劳务纠纷", "value": 168, "percentage": 13.1 },
      { "name": "合同纠纷", "value": 142, "percentage": 11.0 },
      { "name": "其他", "value": 136, "percentage": 10.6 }
    ],
    "organizationRanking": [
      { "orgName": "朝阳区调解中心", "completed": 312, "successRate": 78.5 },
      { "orgName": "海淀区调解中心", "completed": 268, "successRate": 75.2 },
      { "orgName": "东城区调解中心", "completed": 206, "successRate": 80.1 }
    ]
  }
}
```

### 7.2 获取调解员绩效排行

```bash
curl -X GET "http://localhost:8080/api/v1/stats/mediator-ranking?period=month&organizationId=1&limit=10" \
  -H "Authorization: Bearer <admin_token>"
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "rank": 1,
      "userId": 101,
      "realName": "赵调解员",
      "avatar": "",
      "orgName": "朝阳区调解中心",
      "casesAccepted": 28,
      "casesCompleted": 24,
      "successRate": 85.7,
      "avgDuration": 8.2,
      "avgRating": 4.9,
      "performanceScore": 96.5
    },
    {
      "rank": 2,
      "userId": 102,
      "realName": "钱调解员",
      "avatar": "",
      "orgName": "朝阳区调解中心",
      "casesAccepted": 25,
      "casesCompleted": 20,
      "successRate": 80.0,
      "avgDuration": 9.5,
      "avgRating": 4.7,
      "performanceScore": 91.2
    }
  ]
}
```

### 7.3 获取年度对比数据

```bash
curl -X GET "http://localhost:8080/api/v1/stats/yearly-comparison?organizationId=1" \
  -H "Authorization: Bearer <admin_token>"
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "currentYear": 2024,
    "previousYear": 2023,
    "comparison": [
      { "month": "1月", "currentYear": 98, "previousYear": 86 },
      { "month": "2月", "currentYear": 76, "previousYear": 72 },
      { "month": "3月", "currentYear": 112, "previousYear": 95 },
      { "month": "4月", "currentYear": 108, "previousYear": 101 },
      { "month": "5月", "currentYear": 125, "previousYear": 110 },
      { "month": "6月", "currentYear": 128, "previousYear": 0 }
    ],
    "growth": {
      "casesGrowth": 15.8,
      "successRateGrowth": 3.2,
      "avgDurationImprove": -1.8
    }
  }
}
```

---

## 附录：常用测试账号

| 角色 | 用户名 | 密码 | 说明 |
|------|--------|------|------|
| 主任 | director01 | Director@123 | 最高权限，最终审批 |
| 组长 | leader01 | Leader@123 | 案件分派、组长复核 |
| 调解员 | mediator01 | Mediator@123 | 案件调解、记录、提交审批 |
| 调解员 | mediator02 | Mediator@123 | 备用调解员账号 |
| 管理员 | admin01 | Admin@123 | 系统管理、数据统计 |

---

## 附录：案件状态流转

```
待分派(1) → 分派 → 调解中(2) → 达成协议 → 提交审批 → 组长复核中(审批节点1)
                                                    ↓ 通过
                                                主任审批中(审批节点2)
                                                    ↓ 通过
                                                  已完成(4)

调解中(2) → 未达成 → 终止(5)
调解中(2) → 部分达成 → 继续调解(仍为2)
审批任意节点 → 驳回 → 退回调解中(2)
```

---

## 附录：HTTP状态码说明

| HTTP状态码 | 业务Code | 说明 |
|-----------|----------|------|
| 200 | 0 | 成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未登录或Token过期 |
| 403 | 403 | 无权限访问 |
| 404 | 404 | 资源不存在 |
| 500 | 500 | 服务器内部错误 |
