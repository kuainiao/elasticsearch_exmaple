package handler

import (
	"context"
	"fmt"
	"github.com/zhangweilun/tradeweb/constants"
	"github.com/zhangweilun/tradeweb/model"
	"github.com/zhangweilun/tradeweb/service"
	"github.com/zhangweilun/tradeweb/util"
	"gopkg.in/olivere/elastic.v5"
	"strconv"
	"strings"
)

//date_type 0all year 1the last six month 2the last one year 3the last three year 时间过滤
func dataType(q *elastic.BoolQuery, dataType int) {
	if dataType == 1 {
		q = q.Filter(elastic.NewRangeQuery("FrankTime").From("now-6M").To("now"))
	} else if dataType == 2 {
		q = q.Filter(elastic.NewRangeQuery("FrankTime").From("now-1y").To("now"))
	} else if dataType == 3 {
		q = q.Filter(elastic.NewRangeQuery("FrankTime").From("now-3y").To("now"))
	}
}

func district(q *elastic.BoolQuery, districtId int, districtLevel int, ietype int) {
	var ids []string
	if ietype == 0 {
		ids = []string{"PurchaserDistrictId1", "PurchaserDistrictId2", "PurchaserDistrictId3"}
	} else {
		ids = []string{"SupplierDistrictId1", "SupplierDistrictId2", "SupplierDistrictId3"}
	}
	//0 全球 1 国家 2 省
	if districtLevel == 0 {
		//q = q.Must(elastic.NewTermQuery(ids[0], districtId))
		q = q.MustNot(elastic.NewTermQuery(ids[0], 0))
	} else if districtLevel == 1 {
		q = q.Must(elastic.NewTermQuery(ids[0], districtId))
		q = q.MustNot(elastic.NewTermQuery(ids[1], 0))
	} else if districtLevel == 2 {
		q = q.Must(elastic.NewTermQuery(ids[1], districtId))
		q = q.MustNot(elastic.NewTermQuery(ids[2], 0))
	}
}

func vwType(a *elastic.SumAggregation, vwtype int) {
	if vwtype == 0 {
		a = a.Field("OrderVolume")
	} else {
		a = a.Field("OrderWeight")
	}
}

func categoryFilter(q *elastic.BoolQuery, categoryId int) {
	q = q.Must(elastic.NewTermQuery("CategoryId", categoryId))
}

func productFilter(q *elastic.BoolQuery, productId int) {
	q = q.Must(elastic.NewTermQuery("ProductId", productId))
}

//得到特定产品在特定地区进出口前10的公司id
func GetTopTenCompanyId(pid, vwtype, dlevle, ietype, did, dateType int, searchCtx context.Context) *[]model.CompanyInfo {
	var (
		aggField string
		result   []model.CompanyInfo
	)
	client := constants.Instance()
	search := client.Search().Index(constants.IndexName).Type(constants.TypeName)
	agg := elastic.NewTermsAggregation()
	vw := elastic.NewSumAggregation()
	query := elastic.NewBoolQuery()
	productFilter(query, pid)
	district(query, did, dlevle, ietype)
	dataType(query, dateType)
	if vwtype == 0 {
		vw.Field("OrderVolume")
	} else {
		vw.Field("OrderWeight")
	}
	if ietype == 0 {
		aggField = "PurchaserId"
	} else {
		aggField = "SupplierId"
	}
	agg.Field(aggField).Size(10).SubAggregation("vwCount", vw).OrderByAggregation("vwCount", false)
	res, err := search.Query(query).Aggregation("GetTopTenCompanyId", agg).Size(0).RequestCache(true).Do(searchCtx)
	if err != nil {
		fmt.Println(err)
	}
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("GetTopTenCompanyId")
	for i := 0; i < len(terms.Buckets); i++ {
		companyId := terms.Buckets[i].Key.(float64)
		company := model.CompanyInfo{
			ID: int64(companyId),
		}
		if ietype == 0 {
			company.Name = service.GetBuyer(int(company.ID)).OriginName
		} else {
			company.Name = service.GetSupplier(int(company.ID)).OriginName
		}

		for k, v := range terms.Buckets[i].Aggregations {
			data, _ := v.MarshalJSON()
			if k == "vwCount" {
				value := util.BytesString(data)
				volume, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					fmt.Println(err)
				}
				company.Value = util.Round(volume, 2)
			}

		}
		result = append(result, company)
	}
	return &result
}

//GetRegionTop 得到产品地区页面的该产品的top5
func GetRegionTop(pid, vwtype, dlevle, ietype, did, dateType int, searchCtx context.Context) *[]model.Category {
	client := constants.Instance()
	search := client.Search().Index(constants.IndexName).Type(constants.TypeName)
	agg := elastic.NewTermsAggregation()
	sum := elastic.NewSumAggregation()
	query := elastic.NewBoolQuery()
	productFilter(query,pid)
	dataType(query,dateType)
	district(query,did,dlevle,ietype)
	if ietype == 0 {
		if dlevle == 0 {
			agg.Field("PurchaserDistrictId1")
		} else if dlevle == 1 {
			agg.Field("PurchaserDistrictId2")
		} else if dlevle == 2 {
			agg.Field("PurchaserDistrictId3")
		}
	} else {
		if dlevle == 0 {
			agg.Field("SupplierDistrictId1")
		} else if dlevle == 1 {
			agg.Field("SupplierDistrictId2")
		} else if dlevle == 2 {
			agg.Field("SupplierDistrictId3")
		}
	}
	if vwtype == 0 {
		sum.Field("OrderVolume")
	} else {
		sum.Field("OrderWeight")
	}
	agg.SubAggregation("vwCount", sum).OrderByAggregation("vwCount", false).Size(5)
	res, err := search.Query(query).Aggregation("GetRegionTop", agg).Size(0).RequestCache(true).Do(searchCtx)
	if err != nil {
		fmt.Println(err)
	}
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("GetRegionTop")
	var districts []model.Category
	//增加一个数组 容量等于前端请求的pageSize，循环purchaseId获取详细信息
	for i := 0; i < len(terms.Buckets); i++ {
		DistrictID := terms.Buckets[i].Key.(float64)
		category := model.Category{
			Did: int64(DistrictID),
		}
		category.Dname = service.GetDidNameByDid(int64(DistrictID))
		for k, v := range terms.Buckets[i].Aggregations {
			data, _ := v.MarshalJSON()
			if k == "vwCount" {
				value := util.BytesString(data)
				volume, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					fmt.Println(err)
				}
				category.Value = util.Round(volume, 2)
			}
		}
		districts = append(districts, category)
	}
	return &districts
}

func GetIndustryTop10(did, dlevel, vwtype, ietype ,dateType int ,searchCtx context.Context) *[]model.Category {
	var (
		vwCount     *elastic.SumAggregation
		search      *elastic.SearchService
		categorys   []model.Category
	)
	client := constants.Instance()
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query := elastic.NewBoolQuery()
	dataType(query, dateType)
	district(query, did, dlevel, ietype)
	agg := elastic.NewTermsAggregation().Field("CategoryName.keyword")
	vwCount = elastic.NewSumAggregation()
	vwType(vwCount, vwtype)
	agg = agg.SubAggregation("vwCount", vwCount).Size(10).OrderByAggregation("vwCount", false)
	search = search.Query(query).Aggregation("search", agg).RequestCache(true)
	res, err := search.Size(0).Do(searchCtx)
	if err != nil {
		fmt.Println(err)
	}
	terms, _ := res.Aggregations.Terms("search")
	for i := 0; i < len(terms.Buckets); i++ {
		CategoryName := terms.Buckets[i].Key.(string)
		category := model.Category{
			CategoryName: CategoryName,
			CategoryId:   constants.CategoryMap[CategoryName],
		}
		for k, v := range terms.Buckets[i].Aggregations {
			data, _ := v.MarshalJSON()
			if k == "vwCount" {
				value := util.BytesString(data)
				volume, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					fmt.Println(err)
				}
				category.Value = util.Round(volume, 2)
			}
		}
		categorys = append(categorys, category)
	}
	return &categorys
}