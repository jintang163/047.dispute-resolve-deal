#!/bin/bash
set -e

BACKEND_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICES_DIR="$BACKEND_DIR/services"
GATEWAY_DIR="$BACKEND_DIR/gateway"
LOGS_DIR="$BACKEND_DIR/logs"

mkdir -p "$LOGS_DIR"

SERVICES=(
    "user-service:9001:$SERVICES_DIR/user-service"
    "dispute-service:9002:$SERVICES_DIR/dispute-service"
    "workflow-service:9003:$SERVICES_DIR/workflow-service"
    "ai-service:9004:$SERVICES_DIR/ai-service"
    "notification-service:9005:$SERVICES_DIR/notification-service"
)

PIDS=()

echo "========================================"
echo -e "\033[36m  综治中心矛盾纠纷管理系统 - 后端启动\033[0m"
echo "========================================"
echo ""

cleanup() {
    echo ""
    echo -e "\033[33m正在停止所有服务...\033[0m"
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
            echo -e "\033[32m  ✓ 已停止 PID: $pid\033[0m"
        fi
    done
    echo -e "\033[32m所有服务已停止\033[0m"
    exit 0
}

trap cleanup SIGINT SIGTERM

check_port() {
    local port=$1
    if lsof -Pi ":$port" -sTCP:LISTEN -t >/dev/null 2>&1; then
        local pid=$(lsof -Pi ":$port" -sTCP:LISTEN -t)
        echo -e "\033[33m  端口 $port 被占用，正在终止进程 PID: $pid\033[0m"
        kill -9 "$pid" 2>/dev/null || true
        sleep 1
    fi
}

echo -e "\033[32m[1/6] 检查并清理端口...\033[0m"
for svc in "${SERVICES[@]}"; do
    IFS=':' read -r name port path <<< "$svc"
    check_port "$port"
done
check_port 8080
echo -e "\033[32m  端口清理完成\033[0m"
echo ""

echo -e "\033[32m[2/6] 启动微服务...\033[0m"
for svc in "${SERVICES[@]}"; do
    IFS=':' read -r name port path <<< "$svc"
    echo -e "\033[33m  正在启动 $name (端口: $port)...\033[0m"
    
    cd "$path"
    go run main.go > "$LOGS_DIR/$name.log" 2> "$LOGS_DIR/$name-err.log" &
    pid=$!
    PIDS+=("$pid")
    
    sleep 2
    
    if ! kill -0 "$pid" 2>/dev/null; then
        echo -e "\033[31m  ✗ $name 启动失败\033[0m"
        cat "$LOGS_DIR/$name-err.log"
        cleanup
        exit 1
    fi
    echo -e "\033[32m  ✓ $name 启动成功，PID: $pid\033[0m"
done
echo ""

echo -e "\033[32m[3/6] 启动 API Gateway (端口: 8080)...\033[0m"
cd "$GATEWAY_DIR"
go run main.go > "$LOGS_DIR/gateway.log" 2> "$LOGS_DIR/gateway-err.log" &
gateway_pid=$!
PIDS+=("$gateway_pid")

sleep 3

if ! kill -0 "$gateway_pid" 2>/dev/null; then
    echo -e "\033[31m  ✗ Gateway 启动失败\033[0m"
    cat "$LOGS_DIR/gateway-err.log"
    cleanup
    exit 1
fi
echo -e "\033[32m  ✓ Gateway 启动成功，PID: $gateway_pid\033[0m"
echo ""

echo "========================================"
echo -e "\033[32m  所有服务启动成功！\033[0m"
echo "========================================"
echo ""
echo "  服务列表："
for svc in "${SERVICES[@]}"; do
    IFS=':' read -r name port path <<< "$svc"
    echo "    • $name - 端口: $port"
done
echo "    • gateway - 端口: 8080"
echo ""
echo -e "\033[36m  API 地址: http://localhost:8080\033[0m"
echo ""
echo "  日志目录: $LOGS_DIR"
echo ""
echo -e "\033[33m  按 Ctrl+C 停止所有服务\033[0m"
echo ""

while true; do
    for pid in "${PIDS[@]}"; do
        if ! kill -0 "$pid" 2>/dev/null; then
            echo -e "\033[31m  ⚠ 进程 $pid 已意外退出\033[0m"
        fi
    done
    sleep 5
done
