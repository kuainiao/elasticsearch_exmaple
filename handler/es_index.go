package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
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
	search = client.Search().Index(constants.IndexName).Type("frank")
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
	search = client.Search().Index(constants.IndexName).Type("frank")
	res, _ := search.Query(query).Aggregation("vwCount", vwCount).RequestCache(true).Size(0).Do(AggCountCtx)
	terms, _ := res.Aggregations.Sum("vwCount")
	resultCount := terms.Value
	result["value"] = int64(*resultCount)
	result["code"] = 0
	jsonString, err := jsoniter.Marshal(result)
	if err != nil {
		ctx.Log().Error(err)
	}
	if ctx.HasData("redisKey") {
		redisKey = ctx.Data("redisKey").(string)
		err := redis.Set(redisKey, util.BytesString(jsonString), 1*time.Hour).Err()
		if err != nil {
			ctx.Log().Error(err)
		}
	}
	return ctx.String(200, util.BytesString(jsonString))
}

type CategoryTopTen struct {
	DistrictID    int `param:"<in:query> <name:did> <desc:地区id 0为全球>"`
	DistrictLevel int `param:"<in:query> <desc:地区等级> <name:dlevel>"`
	//CategoryID    int `param:"<in:query> <name:category_id> <desc:行业id 总共22个行业>"`
	VwType   int `param:"<in:query> <name:vwtype> <desc:volume weight类型 0volume>"`
	IeType   int `param:"<in:query> <name:ietype> <desc:import export 类型 0import>"`
	DateType int `param:"<in:query> <name:date_type> <desc:date_type 时间过滤>"`
}

//Serve 行业排名
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
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
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
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
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
		productNew.Id = int64(ProductId)
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
//	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
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

//Search 首页搜索的结构
type Search struct {
	CompanyType int           `param:"<in:formData> <name:company_type> <required:required> <err:company_type不能为空！>  <desc:0采购商 1供应商> "`
	CompanyName string        `param:"<in:formData> <name:company_name> <required:required>  <err:company_name不能为空！>  <desc:0采购商 1供应商> "`
	ProKey      string        `param:"<in:formData> <name:pro_key> <required:required>   <err:pro_key不能为空！>  <desc:产品描述>"`
	TimeOut     time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	PageNo      int           `param:"<in:formData> <name:page_no> <required:required>  <nonzero:nonzero> <range: 1:1000>  <err:page_no必须在1到1000之间>   <desc:分页页码>"`
	PageSize    int           `param:"<in:formData> <name:page_size> <required:required>  <nonzero:nonzero> <err:page_size不能为空！>  <desc:分页的页数>"`
	Sort        int           `param:"<in:formData> <name:sort> <required:required>  <err:sort不能为空！>  <desc:排序的参数 1 2 3>"`
}

func (s *Search) Serve(ctx *faygo.Context) error {
	var (
		SearchCtx   context.Context
		cancel      context.CancelFunc
		cardinality *elastic.CardinalityAggregation
		agg         *elastic.TermsAggregation
		search      *elastic.SearchService
		query       *elastic.BoolQuery
		redisKey    string
		total       float64
	)
	if s.PageNo > 1000 {
		s.PageNo = 1000
	}
	if s.TimeOut != 0 {
		SearchCtx, cancel = context.WithTimeout(context.Background(), s.TimeOut*time.Second)
	} else {
		SearchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query = elastic.NewBoolQuery()
	agg = elastic.NewTermsAggregation()
	if s.Sort == 2 {
		agg = agg.OrderByAggregation("volume", false)

	} else if s.Sort == -2 {
		agg = agg.OrderByAggregation("volume", true)

	} else if s.Sort == 3 {
		agg = agg.OrderByAggregation("weight", false)

	} else if s.Sort == -3 {
		agg = agg.OrderByAggregation("weight", true)

	} else if s.Sort == 1 {
		agg = agg.OrderByCount(false)

	} else if s.Sort == -1 {
		agg = agg.OrderByCount(true)

	} else {
		agg = agg.OrderByCount(false)
	}
	agg.Size(s.PageSize * s.PageNo)
	weightAgg := elastic.NewSumAggregation().Field("OrderWeight")
	volumeAgg := elastic.NewSumAggregation().Field("OrderVolume")
	agg = agg.SubAggregation("weight", weightAgg)
	agg = agg.SubAggregation("volume", volumeAgg)
	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))
	proKey, _ := url.PathUnescape(s.ProKey)
	proKey = util.TrimFrontBack(proKey)
	//判断是否全称存在
	if s.CompanyName != "" {
		//不能转小写，要全转大写
		if s.CompanyType == 0 {
			query = query.Filter(elastic.NewTermQuery("Purchaser.keyword", strings.ToUpper(s.CompanyName)))
		} else {
			query = query.Filter(elastic.NewTermQuery("Supplier.keyword", strings.ToUpper(s.CompanyName)))
		}
		if proKey != "" {
			query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
		}
		res, err := search.Query(query).Size(1).Do(SearchCtx)
		if err != nil {
			ctx.Log().Error(err)
		}
		if res.Hits.TotalHits > 0 {
			//存在全称匹配
			if s.CompanyType == 0 {
				cardinality = elastic.NewCardinalityAggregation().Field("PurchaserId")
				agg.Field("PurchaserId")
			} else {
				cardinality = elastic.NewCardinalityAggregation().Field("SupplierId")
				agg.Field("SupplierId")
			}
			count, _ := search.Query(query).Aggregation("count", cardinality).Size(0).Do(SearchCtx)
			resCardinality, _ := count.Aggregations.Cardinality("count")
			total = *resCardinality.Value
			search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
			search = search.Query(query)
			search = search.Aggregation("search", agg).RequestCache(true)
		} else {
			query = elastic.NewBoolQuery()
			query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))
			query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
			if s.CompanyType == 0 {
				cardinality = elastic.NewCardinalityAggregation().Field("PurchaserId")
				agg.Field("PurchaserId")
				if s.CompanyName != "" {
					query = query.Must(elastic.NewMatchQuery("Purchaser", strings.ToLower(s.CompanyName)))
				}
			} else {
				cardinality = elastic.NewCardinalityAggregation().Field("SupplierId")
				agg.Field("SupplierId")
				if s.CompanyName != "" {
					query = query.Must(elastic.NewMatchQuery("Supplier", strings.ToLower(s.CompanyName)))
				}
			}
			count, _ := search.Query(query).Aggregation("count", cardinality).Size(0).Do(SearchCtx)
			resCardinality, _ := count.Aggregations.Cardinality("count")
			total = *resCardinality.Value
			search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
			search = search.Query(query)
			search = search.Aggregation("search", agg).RequestCache(true)
		}
	} else {
		query = elastic.NewBoolQuery()
		query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))
		query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
		if s.CompanyType == 0 {
			cardinality = elastic.NewCardinalityAggregation().Field("PurchaserId")
			agg.Field("PurchaserId")
		} else {
			cardinality = elastic.NewCardinalityAggregation().Field("SupplierId")
			agg.Field("SupplierId")
		}
		count, _ := search.Query(query).Aggregation("count", cardinality).Size(0).Do(SearchCtx)
		resCardinality, _ := count.Aggregations.Cardinality("count")
		total = *resCardinality.Value
		search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
		search = search.Query(query)
		search = search.Aggregation("search", agg).RequestCache(true)
	}
	res, _ := search.Size(0).Do(SearchCtx)
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("search")
	//增加一个数组 容量等于前端请求的pageSize，循环purchaseId获取详细信息
	var franks []model.Frank
	for i := (s.PageNo - 1) * s.PageSize; i < len(terms.Buckets); i++ {
		companyId := terms.Buckets[i].Key.(float64)
		tradeNumber := terms.Buckets[i].DocCount
		frank := model.Frank{
			CompanyId:   int64(companyId),
			TradeNumber: tradeNumber,
		}
		for k, v := range terms.Buckets[i].Aggregations {
			data, _ := v.MarshalJSON()
			if k == "volume" {
				value := util.BytesString(data)
				volume, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					log.Println(err)
				}
				frank.OrderVolume = util.Round(volume, 2)
			}
			if k == "weight" {
				value := util.BytesString(data)
				weight, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					log.Println(err)
				}
				frank.OrderWeight = util.Round(weight, 2)
			}
		}
		franks = append(franks, frank)
	}
	for i := 0; i < len(franks); i++ {
		search := client.Search().Index(constants.IndexName).Type(constants.TypeName)
		query := elastic.NewBoolQuery()
		query.QueryName("frankDetail")
		highlight := elastic.NewHighlight()
		if s.CompanyType == 0 {
			query = query.Must(elastic.NewTermQuery("PurchaserId", franks[i].CompanyId))
			if s.CompanyName != "" {
				query = query.Must(elastic.NewMatchQuery("Purchaser", s.CompanyName))
				highlight.Field("Purchaser")
			}
		} else {
			query = query.Must(elastic.NewTermQuery("SupplierId", franks[i].CompanyId))
			if s.CompanyName != "" {
				query = query.Must(elastic.NewMatchQuery("Supplier", s.CompanyName))
				highlight.Field("Supplier")
			}
		}
		highlight = highlight.PreTags(`<font color="#FF0000">`).PostTags("</font>")
		proKey, _ := url.PathUnescape(s.ProKey)
		if s.ProKey != "" {
			query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
			highlight.Field("ProDesc")
		}
		search.Query(query).Highlight(highlight).Sort("FrankTime", false).From(0).Size(1)
		search.RequestCache(true)
		res, _ := search.Do(SearchCtx)
		var frank model.Frank
		detail := res.Hits.Hits[0].Source
		jsonObject, _ := detail.MarshalJSON()
		json.Unmarshal(jsonObject, &frank)
		if s.CompanyType == 0 {
			franks[i].CompanyName = frank.Purchaser
			franks[i].CompanyId = frank.PurchaserId
		} else {
			franks[i].CompanyName = frank.Supplier
			franks[i].CompanyId = frank.SupplierId
		}
		franks[i].FrankTime = frank.FrankTime
		franks[i].QiyunPort = frank.QiyunPort
		franks[i].OrderId = frank.OrderId
		franks[i].ProDesc = frank.ProDesc
		franks[i].OriginalCountry = frank.OriginalCountry
		franks[i].MudiPort = frank.MudiPort
		franks[i].OrderNo = frank.OrderNo
		franks[i].ProKey = frank.ProKey
		//设置高亮
		hight := res.Hits.Hits[0].Highlight
		if s.ProKey != "" {
			franks[i].ProDesc = hight["ProDesc"][0]
		}
		if s.CompanyName != "" {
			if s.CompanyType == 0 {
				franks[i].CompanyName = hight["Purchaser"][0]
			} else {
				franks[i].CompanyName = hight["Supplier"][0]
			}
		}
	}

	result, err := jsoniter.Marshal(model.Response{
		List:  franks,
		Code:  0,
		Total: int64(total),
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
