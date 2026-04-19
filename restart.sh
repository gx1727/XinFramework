#!/bin/bash

set -e

BINARY_NAME="xin-server"
BINARY_PATH="./out/${BINARY_NAME}"
PID_FILE="./out/${BINARY_NAME}.pid"
CONFIG_FILE="config/config.yaml"
TIMEOUT=30

usage() {
    echo "Usage: $0 {start|stop|restart|reload|status}"
    echo ""
    echo "  start   - Start the server"
    echo "  stop    - Graceful stop (SIGTERM, 30s timeout)"
    echo "  restart - Graceful stop then start"
    echo "  reload  - Hot reload without dropping connections (SIGUSR1)"
    echo "  status  - Show server status"
    exit 1
}

pid() {
    if [ -f "$PID_FILE" ]; then
        cat "$PID_FILE"
    fi
}

is_running() {
    local p=$(pid)
    if [ -n "$p" ] && kill -0 "$p" 2>/dev/null; then
        return 0
    fi
    return 1
}

wait_for_stop() {
    local p=$(pid)
    local waited=0

    while is_running && [ $waited -lt $TIMEOUT ]; do
        echo "Waiting for server to stop... ($waited/$TIMEOUT)"
        sleep 1
        waited=$((waited + 1))
    done

    if is_running; then
        echo "Force killing server..."
        kill -9 "$p" 2>/dev/null || true
        sleep 1
    fi

    rm -f "$PID_FILE"
}

start() {
    if is_running; then
        echo "Server is already running (PID: $(pid))"
        return 1
    fi

    if [ ! -f "$BINARY_PATH" ]; then
        echo "Binary not found: $BINARY_PATH"
        echo "Run 'build.sh' first"
        exit 1
    fi

    if [ ! -f "$CONFIG_FILE" ]; then
        echo "Config file not found: $CONFIG_FILE"
        exit 1
    fi

    echo "Starting $BINARY_NAME..."
    cd "$(dirname "$0")"

    nohup "$BINARY_PATH" > ./out/server.log 2>&1 &
    local new_pid=$!

    echo "$new_pid" > "$PID_FILE"

    sleep 1
    if kill -0 "$new_pid" 2>/dev/null; then
        echo "Server started (PID: $new_pid)"
        return 0
    else
        echo "Server failed to start"
        rm -f "$PID_FILE"
        cat ./out/server.log
        return 1
    fi
}

stop() {
    if ! is_running; then
        echo "Server is not running"
        rm -f "$PID_FILE"
        return 0
    fi

    local p=$(pid)
    echo "Stopping server (PID: $p)..."

    kill -TERM "$p" 2>/dev/null || true

    wait_for_stop

    echo "Server stopped"
}

reload() {
    if ! is_running; then
        echo "Server is not running"
        return 1
    fi

    local p=$(pid)
    echo "Sending SIGUSR1 to reload (PID: $p)..."
    kill -USR1 "$p"

    echo "Reload signal sent"
}

status() {
    if is_running; then
        local p=$(pid)
        echo "Server is running (PID: $p)"

        if [ -f "./out/server.log" ]; then
            echo ""
            echo "Last 5 lines of log:"
            tail -5 ./out/server.log
        fi
    else
        echo "Server is not running"
    fi
}

hot_restart() {
    if ! is_running; then
        echo "Server is not running, starting fresh..."
        start
        return
    fi

    local old_pid=$(pid)
    echo "Hot restart: starting new process..."

    cd "$(dirname "$0")"
    nohup "$BINARY_PATH" > ./out/server.log 2>&1 &
    local new_pid=$!

    echo "$new_pid" > "$PID_FILE"

    sleep 2

    if kill -0 "$new_pid" 2>/dev/null; then
        echo "New server started (PID: $new_pid), stopping old (PID: $old_pid)..."

        if kill -0 "$old_pid" 2>/dev/null; then
            kill -TERM "$old_pid" 2>/dev/null || true
            sleep 1
        fi

        wait_for_stop
        rm -f "$PID_FILE"
        echo "$new_pid" > "$PID_FILE"

        echo "Hot restart complete"
    else
        echo "New server failed to start, keeping old one"
        echo "$old_pid" > "$PID_FILE"
        cat ./out/server.log
        return 1
    fi
}

case "${1:-}" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop
        sleep 2
        start
        ;;
    reload)
        reload
        ;;
    status)
        status
        ;;
    hot-restart)
        hot_restart
        ;;
    *)
        usage
        ;;
esac
