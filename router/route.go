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
			frame.NewOPTIONS("/DetailList.go"),
			frame.NewPOST("/DetailList.go", &handler.FrankDetail).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/search.go"),
			frame.NewPOST("/search.go", &handler.Search{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/topTen.go"),
			frame.NewPOST("/topTen.go", &handler.TopTenProduct{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/CompanyRelations.go"),
			frame.NewPOST("/CompanyRelations.go", &handler.CompanyRelations{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/GroupHistory.go"),
			frame.NewPOST("/GroupHistory.go", &handler.GroupHistory{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/NewTenFrank.go"),
			frame.NewPOST("/NewTenFrank.go", &handler.NewTenFrank{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/ProductList.go"),
			frame.NewPOST("/ProductList.go", &handler.ProductList).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/InfoDetail.go"),
			frame.NewPOST("/InfoDetail.go", &handler.InfoDetail).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/DetailOne.go"),
			frame.NewGET("/DetailOne.go", &handler.DetailOne).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/DetailTrend.go"),
			frame.NewPOST("/DetailTrend.go", &handler.DetailTrend{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewPOST("/CompanyInfo.html", &handler.CompanyInfo{}).Use(middleware.Auth),

			frame.NewPOST("/CompanyList.html", &handler.CompanyList{}).Use(middleware.Auth),

			frame.NewPOST("/CompanyDistrict.html", &handler.CompanyDistrict{}).Use(middleware.Auth),

			frame.NewPOST("/CompanyContacts.html", &handler.CompanyContacts{}).Use(middleware.Auth),
		),

		frame.NewGroup("/index",
			frame.NewOPTIONS("/AggCount.go"),
			frame.NewGET("/AggCount.go", &handler.AggCount{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/CategoryTopTen.go"),
			frame.NewGET("/CategoryTopTen.go", &handler.CategoryTopTen{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/CategoryProductTopTen.go"),
			frame.NewGET("/CategoryProductTopTen.go", &handler.CategoryProductTopTen{}).Use(middleware.RedisCache).Use(middleware.Auth),
		),

		frame.NewPOST("/login.html", &handler.Login{}),
		//倒的执行
		frame.NewNamedAPI("Index", "GET", "/", handler.Index).Use(middleware.RedisCache).Use(middleware.Auth),
		frame.NewNamedAPI("test struct handler", "POST", "/test", &handler.Test{}).
			Use(middleware.Token),
	).Use(middleware.CrossOrigin)

}
