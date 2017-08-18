#!/usr/bin/env bash

cd $GOPATH/src/github.com/zhangweilun/tradeweb
GOOS=linux GOARCH=amd64 go build
scp ./tradeweb gls@gls_api:/data/tradego