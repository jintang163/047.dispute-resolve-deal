#!/bin/bash

set +e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
GRAY='\033[0;90m'
MAGENTA='\033[0;35m'
NC='\033[0m'

ADMIN_PID=""
KIOSK_PID=""

cleanup() {
    echo ""
    echo -e "${YELLOW}[SHUTDOWN] Stopping all frontend services...${NC}"

    if [ -n "$ADMIN_PID" ] && kill -0 "$ADMIN_PID" 2>/dev/null; then
        kill "$ADMIN_PID" 2>/dev/null || true
        wait "$ADMIN_PID" 2>/dev/null || true
        echo -e "${GRAY}[SHUTDOWN] Admin UI stopped${NC}"
    fi

    if [ -n "$KIOSK_PID" ] && kill -0 "$KIOSK_PID" 2>/dev/null; then
        kill "$KIOSK_PID" 2>/dev/null || true
        wait "$KIOSK_PID" 2>/dev/null || true
        echo -e "${GRAY}[SHUTDOWN] Kiosk UI stopped${NC}"
    fi

    pkill -f "vite.*admin" 2>/dev/null || true
    pkill -f "vite.*kiosk" 2>/dev/null || true

    echo ""
    echo -e "${GRAY}[SHUTDOWN] All frontend services stopped.${NC}"
    exit 0
}

trap cleanup SIGINT SIGTERM EXIT

echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  Dispute Resolve - Frontend Launcher    ${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

start_frontend() {
    local name="$1"
    local dir="$2"
    local port="$3"
    local color="$4"
    local full_path="$SCRIPT_DIR/$dir"
    local log_file="$SCRIPT_DIR/$name.log"

    if [ ! -d "$full_path" ]; then
        echo -e "${RED}[ERROR] $name directory not found: $full_path${NC}"
        echo ""
        return 1
    fi

    echo -e "${color}[$name] Starting on port $port...${NC}"
    touch "$log_file"

    (
        cd "$full_path"
        echo -e "${color}[$name] Installing dependencies...${NC}"
        npm install
        echo -e "${color}[$name] Starting dev server on port $port...${NC}"
        if [ "$name" = "kiosk" ]; then
            npm run dev -- --port "$port"
        else
            npm run dev
        fi
    ) >> "$log_file" 2>&1 &

    local pid=$!
    sleep 3

    if ! kill -0 "$pid" 2>/dev/null; then
        echo -e "${RED}[$name] Failed to start. Check $log_file for details.${NC}"
        return 1
    fi

    echo -e "${GREEN}[$name] Started (PID: $pid, Log: $log_file)${NC}"
    echo "$pid"
    return 0
}

echo -e "${YELLOW}[1/3] Starting Admin Frontend...${NC}"
ADMIN_PID=$(start_frontend "admin" "admin" "5173" "$CYAN")

echo ""
echo -e "${YELLOW}[2/3] Starting Kiosk Frontend...${NC}"
KIOSK_PID=$(start_frontend "kiosk" "kiosk" "5174" "$MAGENTA")

echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  Frontend Services Status               ${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

if [ -n "$ADMIN_PID" ]; then
    echo -e "${GREEN}[OK] Admin UI    : http://localhost:5173${NC}"
else
    echo -e "${RED}[FAIL] Admin UI failed to start${NC}"
fi

if [ -n "$KIOSK_PID" ]; then
    echo -e "${GREEN}[OK] Kiosk UI    : http://localhost:5174${NC}"
else
    echo -e "${RED}[FAIL] Kiosk UI failed to start${NC}"
fi

echo ""
echo -e "${YELLOW}[INFO] MiniApp: Please use Kuikly CLI to compile and run${NC}"
echo -e "${GRAY}       cd miniapp${NC}"
echo -e "${GRAY}       kuikly dev --platform wechat     (WeChat MiniApp)${NC}"
echo -e "${GRAY}       kuikly dev --platform alipay     (Alipay MiniApp)${NC}"
echo -e "${GRAY}       kuikly build --platform harmonyos (HarmonyOS)${NC}"
echo ""
echo -e "${GRAY}Press Ctrl+C to stop all services...${NC}"
echo ""

while true; do
    sleep 2

    if [ -n "$ADMIN_PID" ] && ! kill -0 "$ADMIN_PID" 2>/dev/null; then
        echo -e "\n${YELLOW}[WARN] Admin UI service has stopped.${NC}"
        ADMIN_PID=""
    fi

    if [ -n "$KIOSK_PID" ] && ! kill -0 "$KIOSK_PID" 2>/dev/null; then
        echo -e "\n${YELLOW}[WARN] Kiosk UI service has stopped.${NC}"
        KIOSK_PID=""
    fi

    if [ -z "$ADMIN_PID" ] && [ -z "$KIOSK_PID" ]; then
        echo -e "\n${YELLOW}[WARN] All services have stopped.${NC}"
        break
    fi
done
