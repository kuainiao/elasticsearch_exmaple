package router

import (
	"github.com/henrylee2cn/faygo"
	"github.com/zhangweilun/tradeweb/handler"
	"github.com/zhangweilun/tradeweb/middleware"
)

// Route register router in a tree style.
func Route(frame *faygo.Framework) {
	frame.Route(
		frame.NewNamedAPI("Index", "GET", "/", handler.Index),
		frame.NewNamedAPI("test struct handler", "POST", "/test", &handler.Test{}).Use(middleware.Token),
		frame.NewPOST("/frank/DetailList.go", &handler.FrankDetail),
		frame.NewPOST("/frank/search.go", &handler.Search),
	)
}
