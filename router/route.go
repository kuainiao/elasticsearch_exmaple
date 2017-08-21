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
			frame.NewPUT("/DetailList.go"),
			frame.NewPOST("/DetailList.go", &handler.FrankDetail),

			frame.NewOPTIONS("/search.go"),
			frame.NewPOST("/search.go", &handler.Search{}),

			frame.NewOPTIONS("/topTen.go"),
			frame.NewPOST("/topTen.go", &handler.TopTenProduct),

			frame.NewPOST("/CompanyRelations.go", &handler.CompanyRelations{}),
			frame.NewOPTIONS("/CompanyRelations.go"),

			frame.NewOPTIONS("/GroupHistory.go"),
			frame.NewPOST("/GroupHistory.go", &handler.GroupHistory),

			frame.NewOPTIONS("/NewTenFrank.go"),
			frame.NewPOST("/NewTenFrank.go", &handler.NewTenFrank),

			frame.NewOPTIONS("/ProductList.go"),
			frame.NewPOST("/ProductList.go", &handler.ProductList),

			frame.NewOPTIONS("/InfoDetail.go"),
			frame.NewPOST("/InfoDetail.go", &handler.InfoDetail),

			frame.NewOPTIONS("/DetailOne.go"),
			frame.NewGET("/DetailOne.go", &handler.DetailOne),

			frame.NewOPTIONS("/DetailTrend.go"),
			frame.NewPOST("/DetailTrend.go", &handler.DetailTrend{}),
		),

		frame.NewGroup("/index",
			frame.NewOPTIONS("/AggCount.go"),
			frame.NewGET("/AggCount.go", &handler.AggCount{}),

			frame.NewOPTIONS("/CategoryTopTen.go"),
			frame.NewGET("/CategoryTopTen.go", &handler.CategoryTopTen{}).Use(middleware.RedisCache),

			frame.NewOPTIONS("/CategoryProductTopTen.go"),
			frame.NewGET("/CategoryProductTopTen.go", &handler.CategoryProductTopTen{}).Use(middleware.RedisCache),
		),

		frame.NewNamedAPI("Index", "GET", "/", handler.Index).Use(middleware.DbQuery),
		frame.NewNamedAPI("test struct handler", "POST", "/test", &handler.Test{}).
			Use(middleware.Token),
	).Use(middleware.CrossOrigin).Use(middleware.RedisCache)
}
