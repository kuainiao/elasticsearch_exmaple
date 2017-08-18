package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/henrylee2cn/faygo"
	"github.com/henrylee2cn/faygo/ext/db/xorm"
	jsoniter "github.com/json-iterator/go"
	"github.com/zhangweilun/tradeweb/constants"
	"github.com/zhangweilun/tradeweb/model"
	util "github.com/zhangweilun/tradeweb/util"
	elastic "gopkg.in/olivere/elastic.v5"
)

var redis = constants.Redis()

type AggCount struct {
	DistrictID    int `param:"<in:query> <name:did> <desc:地区id 0为全球>"`
	DistrictLevel int `param:"<in:query> <desc:地区等级> <name:dlevel>"`
	CategoryID    int `param:"<in:query> <name:category_id> <desc:行业id 总共22个行业>"`
	VwType        int `param:"<in:query> <name:vwtype> <desc:volume weight类型 0volume>"`
	IeType        int `param:"<in:query> <name:ietype> <desc:import export 类型 0import>"`
	DateType      int `param:"<in:query> <name:date_type> <desc:date_type 时间过滤>"`
}

//AggCount 全球采购商（进口总柜量）或供应商（出口总柜量）的总数
func (q *AggCount) Serve(ctx *faygo.Context) error {
	result := make(map[string]int64)
	var (
		AggCountCtx context.Context
		cancel      context.CancelFunc
		vwCount     *elastic.SumAggregation
		search      *elastic.SearchService
		redisKey    string
	)
	AggCountCtx, cancel = context.WithCancel(context.Background())
	defer cancel()
	client := constants.Instance()
	search = client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	dataType(query, q.DateType)
	district(query, q.DistrictID, q.DistrictLevel, q.IeType)
	cardinality := elastic.NewCardinalityAggregation()
	if q.IeType == 0 {
		cardinality.Field("PurchaserId")
	} else {
		cardinality.Field("SupplierId")
	}
	count, _ := search.Aggregation("count", cardinality).Query(query).Size(0).Do(AggCountCtx)
	resCardinality, _ := count.Aggregations.Cardinality("count")
	total := resCardinality.Value
	result["count"] = int64(*total)
	//柜量
	if q.VwType == 0 {
		vwCount = elastic.NewSumAggregation().Field("OrderVolume")
	} else {
		vwCount = elastic.NewSumAggregation().Field("OrderWeight")
	}
	search = client.Search().Index("trade").Type("frank")
	res, _ := search.Query(query).Aggregation("vwCount", vwCount).RequestCache(true).Size(0).Do(AggCountCtx)
	terms, _ := res.Aggregations.Sum("vwCount")
	resultCount := terms.Value
	result["value"] = int64(*resultCount)
	result["code"] = 0
	json, err := jsoniter.Marshal(result)
	if err != nil {
		ctx.Log().Error(err)
	}
	if ctx.HasData("redisKey") {
		redisKey = ctx.Data("redisKey").(string)
		err := redis.Set(redisKey, util.BytesString(json), 1*time.Hour).Err()
		if err != nil {
			ctx.Log().Error(err)
		}
	}
	return ctx.String(200, util.BytesString(json))
}

type CategoryTopTen struct {
	DistrictID    int `param:"<in:query> <name:did> <desc:地区id 0为全球>"`
	DistrictLevel int `param:"<in:query> <desc:地区等级> <name:dlevel>"`
	//CategoryID    int `param:"<in:query> <name:category_id> <desc:行业id 总共22个行业>"`
	VwType   int `param:"<in:query> <name:vwtype> <desc:volume weight类型 0volume>"`
	IeType   int `param:"<in:query> <name:ietype> <desc:import export 类型 0import>"`
	DateType int `param:"<in:query> <name:date_type> <desc:date_type 时间过滤>"`
}

//行业排名
//getIndustryTop10 get
func (c *CategoryTopTen) Serve(ctx *faygo.Context) error {
	var (
		AggCountCtx context.Context
		cancel      context.CancelFunc
		vwCount     *elastic.SumAggregation
		search      *elastic.SearchService
		categorys   []model.Category
		redisKey    string
	)
	AggCountCtx, cancel = context.WithCancel(context.Background())
	defer cancel()
	client := constants.Instance()
	search = client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	dataType(query, c.DateType)
	district(query, c.DistrictID, c.DistrictLevel, c.IeType)
	agg := elastic.NewTermsAggregation().Field("CategoryName.keyword")
	vwCount = elastic.NewSumAggregation()
	vwType(vwCount, c.VwType)
	agg = agg.SubAggregation("vwCount", vwCount).Size(10).OrderByAggregation("vwCount", false)
	search = search.Query(query).Aggregation("search", agg).RequestCache(true)
	res, err := search.Size(0).Do(AggCountCtx)
	if err != nil {
		ctx.Log().Error(err)
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
					ctx.Log().Error(err)
				}
				category.Value = util.Round(volume, 2)
			}
		}
		categorys = append(categorys, category)
	}
	result, err := jsoniter.Marshal(model.Response{
		Data: categorys,
	})
	if err != nil {
		ctx.Log().Error(err)
	}
	if ctx.HasData("redisKey") {
		redisKey = ctx.Data("redisKey").(string)
		err := redis.Set(redisKey, util.BytesString(result), 1*time.Hour).Err()
		if err != nil {
			ctx.Log().Error(err)
		}
	}
	return ctx.String(200, util.BytesString(result))
}

//行业下产品排名 首页 左二
type CategoryProductTopTen struct {
	DistrictID    int `param:"<in:query> <name:did> <desc:地区id 0为全球>"`
	DistrictLevel int `param:"<in:query> <desc:地区等级> <name:dlevel>"`
	CategoryID    int `param:"<in:query> <name:cid> <desc:行业id 总共22个行业>"`
	VwType        int `param:"<in:query> <name:vwtype> <desc:volume weight类型 0volume>"`
	IeType        int `param:"<in:query> <name:ietype> <desc:import export 类型 0import>"`
	DateType      int `param:"<in:query> <name:date_type> <desc:date_type 时间过滤>"`
}

//getIndustryProductTop10
func (p *CategoryProductTopTen) Serve(ctx *faygo.Context) error {
	var (
		AggCountCtx context.Context
		cancel      context.CancelFunc
		vwCount     *elastic.SumAggregation
		search      *elastic.SearchService
		products    []model.Product
		redisKey    string
	)
	AggCountCtx, cancel = context.WithCancel(context.Background())
	defer cancel()
	client := constants.Instance()
	search = client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	dataType(query, p.DateType)
	district(query, p.DistrictID, p.DistrictLevel, p.IeType)
	categoryFilter(query, p.CategoryID)
	agg := elastic.NewTermsAggregation().Field("ProductId")
	vwCount = elastic.NewSumAggregation()
	vwType(vwCount, p.VwType)
	agg = agg.SubAggregation("vwCount", vwCount).Size(10).OrderByAggregation("vwCount", false)
	search = search.Query(query).Aggregation("search", agg).RequestCache(true)
	res, err := search.Size(0).Do(AggCountCtx)
	if err != nil {
		ctx.Log().Errorf("CategoryProductTopTen%v", err)
	}
	terms, _ := res.Aggregations.Terms("search")
	db := xorm.MustDB("default")
	for i := 0; i < len(terms.Buckets); i++ {
		ProductId := terms.Buckets[i].Key.(float64)
		var productNew model.ProductNew
		productNew.Id = int(ProductId)
		ok, err := db.Get(&productNew)
		if !ok {
			ctx.Log().Error(err)
		}
		product := model.Product{
			ProId:       int64(ProductId),
			ProductName: productNew.Name,
		}
		for k, v := range terms.Buckets[i].Aggregations {
			data, _ := v.MarshalJSON()
			if k == "vwCount" {
				value := util.BytesString(data)
				volume, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					ctx.Log().Error(err)
				}
				product.Count = util.Round(volume, 2)
			}
		}
		products = append(products, product)
	}
	result, err := jsoniter.Marshal(model.Response{
		List: products,
	})
	if err != nil {
		ctx.Log().Error(err)
	}
	if ctx.HasData("redisKey") {
		redisKey = ctx.Data("redisKey").(string)
		err := redis.Set(redisKey, util.BytesString(result), 1*time.Hour).Err()
		if err != nil {
			ctx.Log().Error(err)
		}
	}
	return ctx.String(200, util.BytesString(result))
}

type vwDistributed struct {
	DistrictID    int `param:"<in:query> <name:did> <desc:地区id 0为全球>"`
	DistrictLevel int `param:"<in:query> <desc:地区等级> <name:dlevel>"`
	//CategoryID    int `param:"<in:query> <name:cid> <desc:行业id 总共22个行业>"`
	VwType   int `param:"<in:query> <name:vwtype> <desc:volume weight类型 0volume>"`
	IeType   int `param:"<in:query> <name:ietype> <desc:import export 类型 0import>"`
	DateType int `param:"<in:query> <name:date_type> <desc:date_type 时间过滤>"`
}

///getTurnoverDistributed
//func (vw *vwDistributed) Serve(ctx *faygo.Context) error {
//	var (
//		vwDistributedCtx context.Context
//		cancel           context.CancelFunc
//		vwCount          *elastic.SumAggregation
//		search           *elastic.SearchService
//		products         []model.Product
//		redisKey         string
//	)
//	vwDistributedCtx, cancel = context.WithCancel(context.Background())
//	defer cancel()
//	client := constants.Instance()
//	search = client.Search().Index("trade").Type("frank")
//	agg := elastic.NewTermsAggregation()
//	query := elastic.NewBoolQuery()
//	dataType(query, vw.DateType)
//	district(query, vw.DistrictID, vw.DistrictLevel, vw.IeType)
//	//查全球，过滤找不到国家的
//	if vw.DistrictID == 0 {
//		//进口
//		if vw.IeType == 0 {
//			query = query.Filter(elastic.NewTermQuery("PurchaserDistrictId1", 0))
//			agg.Field("PurchaserDistrictId1")
//		} else {
//			query = query.Filter(elastic.NewTermQuery("SupplierDistrictId1", 0))
//			agg.Field("SupplierDistrictId1")
//		}
//	}
//	//查国家
//	if vw.DistrictLevel == 1 {
//		if vw.IeType == 0 {
//			query = query.Filter(elastic.NewTermQuery("PurchaserDistrictId2", 0))
//			agg.Field("PurchaserDistrictId2")
//		} else {
//			query = query.Filter(elastic.NewTermQuery("SupplierDistrictId2", 0))
//			agg.Field("SupplierDistrictId2")
//		}
//	} else if vw.DistrictLevel == 2 {
//		if vw.IeType == 0 {
//			query = query.Filter(elastic.NewTermQuery("PurchaserDistrictId3", 0))
//			agg.Field("PurchaserDistrictId3")
//		} else {
//			query = query.Filter(elastic.NewTermQuery("SupplierDistrictId3", 0))
//			agg.Field("SupplierDistrictId3")
//		}
//	}
//	vwType(vwCount, vw.VwType)
//
//	agg = agg.SubAggregation("vwCount", vwCount).Size(10).OrderByAggregation("vwCount", false)
//	search = search.Query(query).Aggregation("search", agg).RequestCache(true)
//	return nil
//}
