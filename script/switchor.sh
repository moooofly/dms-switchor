#!/usr/bin/env bash

SWITCHOR_ROOT_DIR=$(pwd)
SWITCHOR_WORKING_DIR=${SWITCHOR_ROOT_DIR}'/src'
ELECTOR_CLIENT_DIR=${SWITCHOR_ROOT_DIR}/../elector

SWITCHOR_SCRIPT_NAME='switchor.py'

SWITCHOR_SCRIPT=${SWITCHOR_WORKING_DIR}'/'${SWITCHOR_SCRIPT_NAME}
SWITCHOR_PID_FILE='/run/dms-switchor.pid'

SWITCHOR_CFG=$2


case $1 in
start)
    echo "starting switchor..."
    echo "using configure file: ${SWITCHOR_CFG}"

    if [ -f ${SWITCHOR_PID_FILE} ]; then
        if kill -0 $(cat ${SWITCHOR_PID_FILE}) > /dev/null 2>&1; then
            pid=$(cat ${SWITCHOR_PID_FILE})
            echo "switchor already running as process ${pid}"
            exit -1
        else
            echo "pid file exists but process does not exist, cleaning pid file"
            rm -f ${SWITCHOR_PID_FILE}
        fi
    fi

    PYTHONPATH=${PYTHONPATH}:${ELECTOR_CLIENT_DIR} python3 ${SWITCHOR_SCRIPT} ${SWITCHOR_CFG} &

    if [ $? -eq 0 ]; then
        pid=$!
        if echo ${pid} > ${SWITCHOR_PID_FILE}; then
            echo "switchor started, as pid ${pid}"
            exit 0
        else
            echo "failed to write pid file"
            exit -1
        fi
    else
        echo "switchor failed to start"
        exit -1
    fi
    ;;

stop)
    echo "stopping switchor..."

    if [ ! -f ${SWITCHOR_PID_FILE} ]; then
        echo "no switchor to stop (could not find file ${SWITCHOR_PID_FILE})"
        exit -1
    else
        kill -9 $(cat ${SWITCHOR_PID_FILE})
        rm -f ${SWITCHOR_PID_FILE}
        echo "switchor stopped"
        exit 0
    fi
    ;;

status)
    if [ ! -f ${SWITCHOR_PID_FILE} ]; then
        echo "no switchor running"
        exit -1
    else
        if kill -0 $(cat ${SWITCHOR_PID_FILE}) > /dev/null 2>&1; then
            pid=$(cat ${SWITCHOR_PID_FILE})
            echo "switchor running, as pid ${pid}"
            exit 0
        else
            echo "pid file exists but process does not exist, cleaning pid file"
            rm -f ${SWITCHOR_PID_FILE}
            exit -1
        fi
    fi
    ;;

version)
    PYTHONPATH=${PYTHONPATH}:${ELECTOR_CLIENT_DIR} python3 ${SWITCHOR_SCRIPT} --version
    ;;

*)
    echo "usage: $0 {start|stop|status|version}" >&2

esac
