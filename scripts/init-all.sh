#!/bin/bash
# =====================================================
# 综治中心矛盾纠纷管理系统 - 一键初始化脚本 (Linux/Mac)
# 功能：创建目录、安装前端依赖、检查配置、启动提示
# =====================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo ""
echo "================================================"
echo -e "\033[36m  综治中心矛盾纠纷管理系统 - 初始化脚本\033[0m"
echo "================================================"
echo ""

echo -e "\033[33m[1/6] 创建后端日志目录...\033[0m"
LOG_DIR="$PROJECT_ROOT/backend/logs"
if [ ! -d "$LOG_DIR" ]; then
    mkdir -p "$LOG_DIR"
    echo -e "\033[32m    创建目录: $LOG_DIR\033[0m"
else
    echo -e "\033[90m    目录已存在: $LOG_DIR\033[0m"
fi
echo ""

echo -e "\033[33m[2/6] 安装管理端前端依赖...\033[0m"
ADMIN_DIR="$PROJECT_ROOT/frontend/admin"
if [ -d "$ADMIN_DIR" ]; then
    echo -e "\033[90m    进入目录: $ADMIN_DIR\033[0m"
    cd "$ADMIN_DIR"
    if npm install; then
        echo -e "\033[32m    管理端依赖安装成功\033[0m"
    else
        echo -e "\033[33m    [警告] 管理端依赖安装失败，请手动执行 npm install\033[0m"
    fi
    cd "$SCRIPT_DIR"
else
    echo -e "\033[33m    [警告] 管理端目录不存在: $ADMIN_DIR\033[0m"
fi
echo ""

echo -e "\033[33m[3/6] 安装自助终端前端依赖...\033[0m"
KIOSK_DIR="$PROJECT_ROOT/frontend/kiosk"
if [ -d "$KIOSK_DIR" ]; then
    echo -e "\033[90m    进入目录: $KIOSK_DIR\033[0m"
    cd "$KIOSK_DIR"
    if npm install; then
        echo -e "\033[32m    自助终端依赖安装成功\033[0m"
    else
        echo -e "\033[33m    [警告] 自助终端依赖安装失败，请手动执行 npm install\033[0m"
    fi
    cd "$SCRIPT_DIR"
else
    echo -e "\033[33m    [警告] 自助终端目录不存在: $KIOSK_DIR\033[0m"
fi
echo ""

echo -e "\033[33m[4/6] 检查数据库配置...\033[0m"
CONFIG_FILES=(
    "$PROJECT_ROOT/backend/gateway/conf/config.yaml"
    "$PROJECT_ROOT/backend/config/config.yaml"
)
for CONFIG_FILE in "${CONFIG_FILES[@]}"; do
    if [ -f "$CONFIG_FILE" ]; then
        echo ""
        echo -e "\033[90m    配置文件: $CONFIG_FILE\033[0m"
        echo ""
        echo -e "\033[33m    ================================================"
        echo "    ⚠  请修改以下数据库配置（如未修改）："
        echo -e "    ================================================\033[0m"
        echo "      database:"
        echo "        host:     127.0.0.1   ← 修改为你的MySQL/TiDB地址"
        echo "        port:     4000        ← 修改为端口(MySQL=3306, TiDB=4000)"
        echo "        user:     root        ← 修改为数据库用户名"
        echo "        password: 123456      ← 修改为数据库密码"
        echo "        dbname:   dispute_resolve"
        echo ""
        echo "      redis:"
        echo "        host:     127.0.0.1   ← 修改为Redis地址"
        echo "        port:     6379        ← 修改为Redis端口"
        echo ""
        echo "      rocketmq:"
        echo "        nameserver:"
        echo "          - 127.0.0.1:9876    ← 修改为RocketMQ地址"
        echo ""
        break
    fi
done
echo ""

echo -e "\033[33m[5/6] 数据库初始化步骤...\033[0m"
echo ""
echo -e "    执行顺序（\033[31m重要！\033[0m）："
echo ""
echo -e "    \033[37m步骤1: 先执行建表脚本\033[0m"
echo -e "      \033[90mmysql -h主机 -u用户 -p密码 < sql/01_init_schema.sql\033[0m"
echo -e "      \033[90m或:  mysql -h主机 -u用户 -p密码 -e 'source $PROJECT_ROOT/sql/01_init_schema.sql'\033[0m"
echo ""
echo -e "    \033[37m步骤2: 再执行初始化数据脚本\033[0m"
echo -e "      \033[90mmysql -h主机 -u用户 -p密码 < sql/init-data.sql\033[0m"
echo -e "      \033[90m或:  mysql -h主机 -u用户 -p密码 -e 'source $PROJECT_ROOT/sql/init-data.sql'\033[0m"
echo ""
echo -e "    \033[90m数据库: dispute_resolve\033[0m"
echo -e "    \033[90m字符集: utf8mb4\033[0m"
echo ""

echo -e "\033[33m[6/6] 启动顺序说明...\033[0m"
echo ""
echo -e "\033[36m    ================================================"
echo "    启动后端服务"
echo -e "    ================================================\033[0m"
echo "      cd $PROJECT_ROOT/backend"
echo "      chmod +x start-all.sh"
echo "      ./start-all.sh"
echo ""
echo "    服务列表："
echo "      - gateway           :8080  (HTTP网关)"
echo "      - user-service      :8081  (用户服务)"
echo "      - dispute-service   :8082  (纠纷服务)"
echo "      - workflow-service  :8083  (工作流服务)"
echo "      - ai-service        :8084  (AI服务)"
echo "      - notification-service :8085 (通知服务)"
echo ""
echo -e "\033[36m    ================================================"
echo "    启动前端服务"
echo -e "    ================================================\033[0m"
echo "      cd $PROJECT_ROOT/frontend"
echo "      chmod +x start-all.sh"
echo "      ./start-all.sh"
echo ""
echo "    前端列表："
echo "      - admin   管理后台   :5173"
echo "      - kiosk   自助终端   :5174"
echo ""

echo -e "\033[32m================================================"
echo "  默认账号密码（密码均为123456或admin123）"
echo "================================================"
echo ""
echo "  ┌────────────┬────────────┬─────────────────┐"
echo "  │ 账号       │ 密码       │ 角色             │"
echo "  ├────────────┼────────────┼─────────────────┤"
echo "  │ admin      │ admin123   │ 系统管理员       │"
echo "  ├────────────┼────────────┼─────────────────┤"
echo "  │ director   │ 123456     │ 综治中心主任     │"
echo "  ├────────────┼────────────┼─────────────────┤"
echo "  │ leader     │ 123456     │ 调解组组长       │"
echo "  ├────────────┼────────────┼─────────────────┤"
echo "  │ mediator1  │ 123456     │ 调解员(王)      │"
echo "  ├────────────┼────────────┼─────────────────┤"
echo "  │ mediator2  │ 123456     │ 调解员(赵)      │"
echo "  ├────────────┼────────────┼─────────────────┤"
echo "  │ mediator3  │ 123456     │ 调解员(刘)      │"
echo "  └────────────┴────────────┴─────────────────┘"
echo ""

echo -e "\033[32m================================================"
echo "  初始化脚本执行完成！"
echo "================================================"
echo -e "\033[0m"
echo ""
echo "  访问地址："
echo "    管理后台: http://localhost:5173"
echo "    自助终端: http://localhost:5174"
echo "    API文档 : http://localhost:8080/swagger/index.html"
echo ""
echo "  如有问题，请检查："
echo "    1. MySQL/TiDB 是否已启动"
echo "    2. Redis 是否已启动"
echo "    3. RocketMQ 是否已启动"
echo "    4. 配置文件 config.yaml 中的连接信息是否正确"
echo ""
