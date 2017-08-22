package main

import (
	"github.com/henrylee2cn/faygo"
	"github.com/zhangweilun/tradeweb/router"
)

func main() {
	//53476085
	//GOOS=linux GOARCH=amd64 go build
	//scp ./tradeweb gls@gls_api:/data/tradego
	//screen -r es
	//ps -ef|grep tradeweb
	//kill -9 19781
	//chmod +x tradeweb
	//ps -ef|grep tradeweb|grep -v grep
	router.Route(faygo.New("tradeweb"))
	faygo.Run()
}
