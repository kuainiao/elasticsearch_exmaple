#!/usr/bin/env bash

echo 'pack'
cd $GOPATH/src/github.com/zhangweilun/tradeweb
GOOS=linux GOARCH=amd64 go build
# tar -czf tradeweb.tar.gz ./tradeweb
# tar -xzvf file.tar.gz
upx  ./tradeweb
# scp ./tradeweb.tar.gz gls@123.207.252.117:/data/tradego
echo 'upload'
scp ./tradeweb gls@123.207.252.117:/data/tradego
# ssh gls@123.207.252.117 'cd /data/tradego && sh restart.sh /data/tradego/tradeweb'
# ssh gls@123.207.252.117 'cd /data/tradego && tar -xzvf tradeweb.tar.gz && chmod +x tradeweb && /data/tradego/tradeweb &'