#!/bin/bash

# 开发用的 LLM Mock 启动脚本
# 支持编译、启动、重启、重新构建并重启、查看日志等功能

set -e

ulimit -n 65535

# 配置
BINARY_PATH="/tmp/llm-mock"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_FILE="/tmp/llm-mock.pid"
LOG_FILE="/tmp/llm-mock.log"
MAIN_FILE="main.go"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查进程是否运行
is_running() {
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if ps -p "$pid" > /dev/null 2>&1; then
            return 0
        else
            rm -f "$PID_FILE"
            return 1
        fi
    fi
    return 1
}

# 获取进程PID
get_pid() {
    if [ -f "$PID_FILE" ]; then
        cat "$PID_FILE"
    fi
}

# 编译项目
compile() {
    print_info "开始编译项目..."

    cd "$PROJECT_DIR"

    # 如果二进制文件存在，先删除
    if [ -f "$BINARY_PATH" ]; then
        print_info "删除已存在的二进制文件: $BINARY_PATH"
        rm -f "$BINARY_PATH"
    fi

    # 编译项目
    print_info "编译中..."
    if go build -o "$BINARY_PATH" "$MAIN_FILE"; then
        print_success "编译成功! 二进制文件: $BINARY_PATH"
        return 0
    else
        print_error "编译失败!"
        return 1
    fi
}

# 启动服务
start() {
    if is_running; then
        local pid=$(get_pid)
        print_warning "服务已经在运行中 (PID: $pid)"
        return 1
    fi

    # 检查二进制文件是否存在
    if [ ! -f "$BINARY_PATH" ]; then
        print_warning "二进制文件不存在，尝试编译..."
        if ! compile; then
            return 1
        fi
    fi

    print_info "启动服务..."
    cd "$PROJECT_DIR"

    # 启动服务并记录PID
    nohup "$BINARY_PATH" > "$LOG_FILE" 2>&1 &
    local pid=$!
    echo $pid > "$PID_FILE"

    # 等待一下检查是否启动成功
    sleep 2
    if ps -p "$pid" > /dev/null 2>&1; then
        print_success "服务启动成功! (PID: $pid)"
        print_info "日志文件: $LOG_FILE"
    else
        print_error "服务启动失败!"
        rm -f "$PID_FILE"
        return 1
    fi
}

# 停止服务
stop() {
    if ! is_running; then
        print_warning "服务没有运行"
        return 1
    fi

    local pid=$(get_pid)
    print_info "停止服务 (PID: $pid)..."

    # 发送SIGTERM信号
    kill "$pid"

    # 等待进程结束
    local count=0
    while ps -p "$pid" > /dev/null 2>&1 && [ $count -lt 10 ]; do
        sleep 1
        count=$((count + 1))
    done

    # 如果进程还在运行，强制杀死
    if ps -p "$pid" > /dev/null 2>&1; then
        print_warning "进程没有正常退出，强制杀死..."
        kill -9 "$pid"
        sleep 1
    fi

    rm -f "$PID_FILE"
    print_success "服务已停止"
}

# 重启服务
restart() {
    print_info "重启服务..."
    stop
    sleep 1
    start
}

# 重新构建并重启
rebuild_restart() {
    print_info "重新构建并重启服务..."

    # 停止现有服务
    if is_running; then
        stop
    fi

    # 编译新版本
    if compile; then
        sleep 1
        start
    else
        print_error "重新构建失败!"
        return 1
    fi
}

# 查看日志
logs() {
    if [ ! -f "$LOG_FILE" ]; then
        print_warning "日志文件不存在: $LOG_FILE"
        return 1
    fi

    print_info "查看日志文件: $LOG_FILE"
    print_info "按 Ctrl+C 退出日志查看"
    echo "----------------------------------------"
    tail -n 100 -f "$LOG_FILE"
}

# 显示状态
status() {
    print_info "服务状态:"
    echo "二进制文件: $BINARY_PATH"
    echo "PID文件: $PID_FILE"
    echo "日志文件: $LOG_FILE"
    echo ""

    if [ -f "$BINARY_PATH" ]; then
        print_success "二进制文件存在"
    else
        print_warning "二进制文件不存在"
    fi

    if is_running; then
        local pid=$(get_pid)
        print_success "服务正在运行 (PID: $pid)"
    else
        print_warning "服务没有运行"
    fi

    if [ -f "$LOG_FILE" ]; then
        local log_size=$(du -h "$LOG_FILE" | cut -f1)
        print_info "日志文件大小: $log_size"
    fi
}

# 显示帮助
show_help() {
    echo "Silinex Router 开发启动脚本"
    echo ""
    echo "用法: $0 <命令>"
    echo ""
    echo "可用命令:"
    echo "  compile     编译项目到 $BINARY_PATH"
    echo "  start       启动服务"
    echo "  stop        停止服务"
    echo "  restart     重启服务"
    echo "  rebuild     重新构建并重启服务"
    echo "  logs        查看日志 (实时)"
    echo "  status      显示服务状态"
    echo "  help        显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 compile      # 仅编译"
    echo "  $0 start        # 启动服务"
    echo "  $0 restart      # 重启服务"
    echo "  $0 rebuild      # 重新构建并重启"
    echo "  $0 logs         # 查看日志"
}

# 主逻辑
case "$1" in
    compile)
        compile
        ;;
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    rebuild)
        rebuild_restart
        ;;
    logs)
        logs
        ;;
    status)
        status
        ;;
    help|--help|-h)
        show_help
        ;;
    "")
        print_error "请指定一个命令"
        echo ""
        show_help
        exit 1
        ;;
    *)
        print_error "未知命令: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
