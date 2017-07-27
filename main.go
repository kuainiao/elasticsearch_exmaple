package main

import (
	"github.com/henrylee2cn/faygo"
	"github.com/zhangweilun/tradeweb/router"
)

func main() {
	router.Route(faygo.New("tradeweb"))
	faygo.Run()
}
