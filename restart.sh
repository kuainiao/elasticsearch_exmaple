#!/bin/bash
#Program
# The program is used to restart a process
#Auth:
# Jimmy
#Date:
# 2014-11-18
#检查是否有输入参数
[ -z "$1" ] && echo "you should input the process name !" && exit 0
#检查输入参数是否指向可执行文件
if test ! -e $1 || test ! -x $1;then
        echo "please input the absolute path of your process"
        exit 0
fi
#杀进程
kill -9 `ps -u | grep $1 | grep -v grep | grep -v sh| awk '{print $2}'`
#重启进程
# $1 &
#检查进程是否启动
# [ -z "`ps -u | grep $1 | grep -v grep | grep -v sh | awk '{print $2}'`" ]  && echo "process $1 is not running" || echo "process $1 is running"