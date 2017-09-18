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
			frame.NewPOST("/CompanyRelations.go", &handler.CompanyRelations{}),

			frame.NewOPTIONS("/GroupHistory.go"),
			frame.NewPOST("/GroupHistory.go", &handler.GroupHistory{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/NewTenFrank.go"),
			frame.NewPOST("/NewTenFrank.go", &handler.NewTenFrank{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/ProductList.go"),
			frame.NewPOST("/ProductList.go", &handler.ProductList).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/InfoDetail.go"),
			frame.NewPOST("/InfoDetail.go", &handler.InfoDetail{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/DetailOne.go"),
			frame.NewGET("/DetailOne.go", &handler.DetailOne).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/DetailTrend.go"),
			frame.NewPOST("/DetailTrend.go", &handler.DetailTrend{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/CompanyInfo.html"),
			frame.NewPOST("/CompanyInfo.html", &handler.CompanyInfo{}).Use(middleware.Auth),

			frame.NewOPTIONS("/CompanyList.html"),
			frame.NewPOST("/CompanyList.html", &handler.CompanyList{}).Use(middleware.Auth),

			frame.NewOPTIONS("/CompanyDistrict.html"),
			frame.NewPOST("/CompanyDistrict.html", &handler.CompanyDistrict{}).Use(middleware.Auth),

			frame.NewOPTIONS("/CompanyContacts.html"),
			frame.NewPOST("/CompanyContacts.html", &handler.CompanyContacts{}).Use(middleware.Auth),
		),

		frame.NewGroup("/index",
			frame.NewOPTIONS("/AggCount.go"),
			frame.NewGET("/AggCount.go", &handler.AggCount{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/CategoryTopTen.go"),
			frame.NewGET("/CategoryTopTen.go", &handler.CategoryTopTen{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/CategoryProductTopTen.go"),
			frame.NewGET("/CategoryProductTopTen.go", &handler.CategoryProductTopTen{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/FindMaoInfo.html"),
			frame.NewPOST("/FindMaoInfo.html", &handler.FindMaoInfo{}).Use(middleware.Auth),

			frame.NewOPTIONS("/CategoryTopTenArea.html"),
			frame.NewPOST("/CategoryTopTenArea.html", &handler.CategoryTopTenArea{}).Use(middleware.RedisCache).Use(middleware.Auth),
			//frame.NewPOST("/CategoryTopTenArea.html", &handler.CategoryTopTenArea{}),

			frame.NewOPTIONS("/CategoryProductTopTenArea.html"),
			frame.NewPOST("/CategoryProductTopTenArea.html", &handler.CategoryProductTopTenArea{}).Use(middleware.RedisCache).Use(middleware.Auth),
			//frame.NewPOST("/CategoryProductTopTenArea.html", &handler.CategoryProductTopTenArea{}),

			frame.NewOPTIONS("/CategoryCompanyTopTen.html"),
			frame.NewPOST("/CategoryCompanyTopTen.html", &handler.CategoryCompanyTopTen{}).Use(middleware.RedisCache).Use(middleware.Auth),
			//frame.NewPOST("/CategoryCompanyTopTen.html", &handler.CategoryCompanyTopTen{}),

			//右侧第二个图
			frame.NewOPTIONS("/CategoryVwTimeFilter.html"),
			frame.NewPOST("/CategoryVwTimeFilter.html", &handler.CategoryVwTimeFilter{}).Use(middleware.RedisCache).Use(middleware.Auth),
			//frame.NewPOST("/CategoryVwTimeFilter.html", &handler.CategoryVwTimeFilter{}),

			//右侧第一个图
			frame.NewOPTIONS("/GlobalImport.html"),
			frame.NewPOST("/GlobalImport.html", &handler.GlobalImport{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/DistributionRegion.html"),
			frame.NewPOST("/DistributionRegion.html", &handler.DistributionRegion{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/DistrictCompanyList.html"),
			frame.NewPOST("/DistrictCompanyList.html", &handler.DistrictCompanyList{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/FindMapRelation.html"),
			frame.NewPOST("/FindMapRelation.html", &handler.FindMapRelation{}).Use(middleware.RedisCache).Use(middleware.Auth),
		),

		frame.NewGroup("/product",
			frame.NewOPTIONS("/CountryRank.html"),
			frame.NewPOST("/CountryRank.html", &handler.CountryRank{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/ProductCompanyTop.html"),
			frame.NewPOST("/ProductCompanyTop.html", &handler.ProductCompanyTop{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/ProductCompanyTrend.html"),
			frame.NewPOST("/ProductCompanyTrend.html", &handler.ProductCompanyTrend{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/ProductTrend.html"),
			frame.NewPOST("/ProductTrend.html", &handler.ProductTrend{}).Use(middleware.RedisCache).Use(middleware.Auth),

			frame.NewOPTIONS("/RegionTrend.html"),
			frame.NewPOST("/RegionTrend.html", &handler.RegionTrend{}),

			frame.NewOPTIONS("/RegionTop.html"),
			frame.NewPOST("/RegionTop.html", &handler.RegionTop{}).Use(middleware.RedisCache).Use(middleware.Auth),
		),

		frame.NewPOST("/login.html", &handler.Login{}),
		//倒的执行
		frame.NewNamedAPI("Index", "GET", "/", handler.Index).Use(middleware.RedisCache).Use(middleware.Auth),
		frame.NewNamedAPI("test struct handler", "POST", "/test", &handler.Test{}),
	).Use(middleware.CrossOrigin)

}
