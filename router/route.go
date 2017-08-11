package router

import (
	"github.com/henrylee2cn/faygo"
	"github.com/zhangweilun/tradeweb/handler"
	"github.com/zhangweilun/tradeweb/middleware"
)

// Route register router in a tree style.
func Route(frame *faygo.Framework) {
	frame.Route(
		frame.NewGroup("/frank",
		    frame.NewPUT("/DetailList.go").Use(middleware.CrossOrigin),
			frame.NewPOST("/DetailList.go", &handler.FrankDetail).Use(middleware.CrossOrigin),

			frame.NewOPTIONS("/search.go").Use(middleware.CrossOrigin),
			frame.NewPOST("/search.go", &handler.Search).Use(middleware.CrossOrigin),

			frame.NewOPTIONS("/topTen.go").Use(middleware.CrossOrigin),
			frame.NewPOST("/topTen.go", &handler.TopTenProduct),

			frame.NewPOST("/CompanyRelations.go", &handler.CompanyRelations).Use(middleware.CrossOrigin),
			frame.NewOPTIONS("/CompanyRelations.go").Use(middleware.CrossOrigin),

			frame.NewOPTIONS("/GroupHistory.go").Use(middleware.CrossOrigin),
			frame.NewPOST("/GroupHistory.go", &handler.GroupHistory).Use(middleware.CrossOrigin),

			frame.NewOPTIONS("/NewTenFrank.go").Use(middleware.CrossOrigin),
			frame.NewPOST("/NewTenFrank.go", &handler.NewTenFrank).Use(middleware.CrossOrigin),

			frame.NewOPTIONS("/ProductList.go").Use(middleware.CrossOrigin),
			frame.NewPOST("/ProductList.go", &handler.ProductList).Use(middleware.CrossOrigin),
			frame.NewOPTIONS("/InfoDetail.go").Use(middleware.CrossOrigin),
			frame.NewPOST("/InfoDetail.go",&handler.InfoDetail).Use(middleware.CrossOrigin),
		),

		frame.NewNamedAPI("Index", "GET", "/", handler.Index),
		frame.NewNamedAPI("test struct handler", "POST", "/test", &handler.Test{}).Use(middleware.Token),
	)
}
