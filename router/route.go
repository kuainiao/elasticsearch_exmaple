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
			frame.NewPOST("/DetailList.go", &handler.FrankDetail),
			frame.NewOPTIONS("/search.go").Use(middleware.CrossOrigin),
			frame.NewPOST("/search.go", &handler.Search).Use(middleware.CrossOrigin),
			frame.NewPOST("/topTen.go", &handler.TopTenProduct),
			frame.NewPOST("/CompanyRelations.go", &handler.CompanyRelations).Use(middleware.CrossOrigin),
			frame.NewOPTIONS("/CompanyRelations.go").Use(middleware.CrossOrigin),
			frame.NewPOST("/GroupHistory.go", &handler.GroupHistory),
			frame.NewPOST("/NewTenFrank.go", &handler.NewTenFrank),
			frame.NewPOST("/ProductList.go", &handler.ProductList),),
		frame.NewNamedAPI("Index", "GET", "/", handler.Index),
		frame.NewNamedAPI("test struct handler", "POST", "/test", &handler.Test{}).Use(middleware.Token),
		//frame.NewPOST("/frank/DetailList.go", &handler.FrankDetail),
		//frame.NewOPTIONS("/frank/search.go").Use(middleware.CrossOrigin),
		//frame.NewPOST("/frank/search.go", &handler.Search).Use(middleware.CrossOrigin),
		//frame.NewPOST("/frank/topTen.go", &handler.TopTenProduct),
		//frame.NewPOST("/frank/CompanyRelations.go", &handler.CompanyRelations),
		//frame.NewPOST("/frank/GroupHistory.go", &handler.GroupHistory),
		//frame.NewPOST("/frank/NewTenFrank.go", &handler.NewTenFrank),
		//frame.NewPOST("/frank/ProductList.go", &handler.ProductList),
	)
}
