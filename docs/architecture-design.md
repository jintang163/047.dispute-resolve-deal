# 系统架构详细设计

## 1. 整体架构设计

### 1.1 架构风格
采用**领域驱动设计(DDD)** + **微服务架构**，基于**CloudWeGo**生态构建高性能分布式系统。

### 1.2 架构分层

| 层级 | 职责 | 技术选型 |
|------|------|----------|
| 接入层 | 多端接入、协议转换 | Hertz HTTP Gateway |
| 应用层 | 业务编排、API聚合 | Gateway聚合层 + BFF |
| 领域层 | 核心业务逻辑 | Kitex微服务 + Flowable |
| 基础设施层 | 数据持久化、外部依赖 | TiDB + Redis + ES + RocketMQ |

### 1.3 服务划分

#### 核心服务
1. **用户服务 (user-service)** - 用户管理、权限、组织架构
2. **纠纷服务 (dispute-service)** - 纠纷CRUD、分派、调解记录
3. **工作流服务 (workflow-service)** - 审批流、Flowable集成
4. **AI服务 (ai-service)** - DeepSeek LLM、Milvus向量检索
5. **通知服务 (notification-service)** - RocketMQ消息、短信、微信通知

#### 公共组件
- **common/database** - TiDB ORM封装
- **common/cache** - Redis操作封装
- **common/mq** - RocketMQ生产者/消费者
- **common/auth** - JWT认证、鉴权中间件

---

## 2. 关键技术方案

### 2.1 API网关设计 (Hertz)

**核心职责**:
- 统一接入入口，路由分发到各微服务
- 统一认证鉴权 (JWT + RBAC)
- 限流熔断、超时控制
- 请求日志、链路追踪
- 协议转换 (HTTP → gRPC)

**路由设计**:
```
/api/v1/user/*           → user-service
/api/v1/dispute/*        → dispute-service
/api/v1/workflow/*       → workflow-service
/api/v1/ai/*             → ai-service
/api/v1/notification/*   → notification-service
/ws/*                    → WebSocket长连接
```

### 2.2 微服务通信 (Kitex)

**通信协议**: gRPC + Protobuf
**服务发现**: Consul / Etcd
**负载均衡**: 客户端负载均衡
**熔断降级**: Hystrix / Sentinel

**IDL示例 (user.thrift)**:
```thrift
namespace go user

struct User {
    1: i64 id
    2: string username
    3: string real_name
    4: string phone
    5: i32 role
    6: i64 organization_id
}

service UserService {
    User GetUser(1: i64 id)
    User Login(1: string username, 2: string password)
    list<User> ListUser(1: i32 page, 2: i32 page_size)
}
```

### 2.3 工作流引擎集成 (Flowable)

**审批流程定义**:
- 调解完成 → 组长复核 → 主任审批 → 结案
- 支持: 通过、驳回、退回修改、加签、转审
- 可配置不同纠纷类型的审批链

**超时升级机制**:
- 调解员24h未处理 → 自动升级至组长
- 组长48h未处理 → 自动升级至主任
- 主任72h未处理 → 系统告警

### 2.4 智能分派算法

基于以下因素自动分派案件：
1. 调解员当前负载（案件数量）
2. 调解员专业领域（纠纷类型匹配度）
3. 调解员地理位置（就近原则）
4. 历史调解成功率
5. 案件紧急程度

---

## 3. 数据架构设计

### 3.1 TiDB 分布式数据库

**核心表设计**:
- `sys_user` - 用户表
- `sys_organization` - 组织架构表
- `dispute_case` - 纠纷案件主表
- `dispute_type` - 纠纷类型表（三级分类）
- `dispute_evidence` - 证据材料表
- `dispute_mediation_record` - 调解记录表
- `workflow_approval` - 审批记录表
- `workflow_urge` - 催办记录表
- `performance_score` - 绩效考核表

### 3.2 Redis Cluster

**使用场景**:
- 用户会话缓存 (Session)
- 热点数据缓存（纠纷类型、字典表）
- 分布式锁（案件分派、审批防重）
- 接口限流计数器
- 排行榜、热度统计

### 3.3 Elasticsearch

**使用场景**:
- 纠纷案件全文检索
- 调解记录模糊搜索
- 统计分析（聚合查询）
- 日志存储分析

### 3.4 Milvus 向量数据库

**使用场景**:
- 法律咨询语义匹配
- 历史案例相似度检索
- 法条智能推荐

---

## 4. 安全设计

### 4.1 认证授权
- 基于JWT的无状态认证
- RBAC权限模型（角色-权限-用户）
- 数据权限隔离（组织架构维度）
- 接口签名校验

### 4.2 数据安全
- 敏感数据加密存储（身份证、手机号）
- HTTPS全链路加密
- 接口脱敏返回
- 操作审计日志

### 4.3 防攻击
- 接口限流（IP+用户维度）
- SQL注入防护
- XSS攻击防护
- CSRF防护

---

## 5. 非功能设计

### 5.1 性能指标
- API响应时间: P95 < 200ms
- 并发支持: ≥ 1000 QPS
- 页面加载时间: < 2s
- 视频调解延迟: < 500ms

### 5.2 高可用
- 服务无状态，支持水平扩展
- 数据库多副本 + 自动故障转移
- 消息队列ACK机制保证可靠性
- 多级缓存降级策略

### 5.3 可观测性
- 分布式链路追踪 (OpenTelemetry)
- 服务监控 (Prometheus + Grafana)
- 日志聚合 (ELK)
- 告警机制

---

## 6. 部署架构

### 6.1 容器化部署
- Docker镜像 + Kubernetes编排
- 服务网格 (Istio) 可选
- 蓝绿发布 / 灰度发布

### 6.2 环境划分
- DEV → SIT → UAT → PROD
- 各环境网络隔离
- 配置中心 (Nacos / Apollo)
