#!/usr/bin/env bash
echo '请输入要提交的注释信息:'

read msg

git pull origin master

git add .

git commit -a -m "$msg"

git push origin master

cd $GOPATH/src/github.com/zhangweilun/tradeweb
GOOS=linux GOARCH=amd64 go build
scp ./tradeweb gls@gls_api:/data/tradego