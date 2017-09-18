package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zhangweilun/tradeweb/service"

	"github.com/henrylee2cn/faygo"
	"github.com/henrylee2cn/faygo/errors"
	"github.com/henrylee2cn/faygo/ext/db/xorm"
	"github.com/json-iterator/go"
	"github.com/zhangweilun/tradeweb/constants"
	"github.com/zhangweilun/tradeweb/model"
	"github.com/zhangweilun/tradeweb/util"

	"gopkg.in/olivere/elastic.v5"
)

var redis = constants.Redis()

type AggCount struct {
	DistrictID    int `param:"<in:query> <name:did> <desc:地区id 0为全球>"`
	DistrictLevel int `param:"<in:query> <desc:地区等级> <name:dlevel>"`
	CategoryID    int `param:"<in:query> <name:category_id> <desc:行业id 总共22个行业>"`
	VwType        int `param:"<in:query> <name:vwtype> <desc:volume weight类型 0volume>"`
	IeType        int `param:"<in:query> <name:ietype> <desc:import export 类型 0import>"`
	DateType      int `param:"<in:query> <name:date_type> <desc:date_type 时间过滤>"`
	Pid           int `param:"<in:query> <name:pid>   <err:pid不能为空！>  <desc:排序的参数 1 2 3>"`
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
	if q.Pid != 0 {
		productFilter(query, q.Pid)
	}
	if q.CategoryID != 0 {
		categoryFilter(query, q.CategoryID)
	}
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
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, jsonString)
}

//CategoryTopTen 左边第一个表
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
func (param *CategoryTopTen) Serve(ctx *faygo.Context) error {
	var (
		searchCtx context.Context
		cancel    context.CancelFunc
		redisKey  string
	)
	searchCtx, cancel = context.WithCancel(context.Background())
	defer cancel()
	categorys := GetIndustryTop10(param.DistrictID, param.DistrictLevel, param.VwType, param.IeType,param.DateType, searchCtx)
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
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, result)
}

//行业下产品排名 首页 左边第二个图
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
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, result)
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
		franks      []model.Frank
	)
	if s.CompanyName == "" && s.ProKey == "" {
		return ctx.String(400, "prokey和公司名不能同时为空!!!")
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

	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))
	proKey, _ := url.PathUnescape(s.ProKey)
	if proKey != "" {
		proKey = util.TrimFrontBack(proKey)
	}
	//判断是否全称存在
	if s.CompanyName != "" {
		if s.PageNo == 1 {
			company, err := getSpecificCompany(client, s.CompanyType, s.Sort, s.CompanyName, proKey, SearchCtx)
			if err != nil {
				ctx.Log().Error(err.Error())
			} else {
				total = total + 1
				franks = append(franks, *company)
			}
		}
	}
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
	if proKey != "" {
		query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
	}
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
	total = total + *resCardinality.Value
	if len(franks) > 0 {
		//已经有完整匹配
		agg.Size(s.PageSize*s.PageNo - 1)
	} else {
		agg.Size(s.PageSize * s.PageNo)
	}
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	search = search.Query(query)
	search = search.Aggregation("search", agg).RequestCache(true)
	res, _ := search.Size(0).Do(SearchCtx)
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("search")
	//增加一个数组 容量等于前端请求的pageSize，循环purchaseId获取详细信息
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
				query = query.Must(elastic.NewMatchQuery("Purchaser", strings.ToLower(s.CompanyName)))
				highlight.Field("Purchaser")
			}
		} else {
			query = query.Must(elastic.NewTermQuery("SupplierId", franks[i].CompanyId))
			if s.CompanyName != "" {
				query = query.Must(elastic.NewMatchQuery("Supplier", strings.ToLower(s.CompanyName)))
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
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, result)
}

//getSpecificCompany 完整匹配
func getSpecificCompany(client *elastic.Client, CompanyType, Sort int, CompanyName string, proKey string, SearchCtx context.Context) (*model.Frank, error) {
	search := client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query := elastic.NewBoolQuery()
	agg := elastic.NewTermsAggregation()
	agg.Size(1)
	if Sort == 2 {
		agg = agg.OrderByAggregation("volume", false)

	} else if Sort == -2 {
		agg = agg.OrderByAggregation("volume", true)

	} else if Sort == 3 {
		agg = agg.OrderByAggregation("weight", false)

	} else if Sort == -3 {
		agg = agg.OrderByAggregation("weight", true)

	} else if Sort == 1 {
		agg = agg.OrderByCount(false)

	} else if Sort == -1 {
		agg = agg.OrderByCount(true)

	} else {
		agg = agg.OrderByCount(false)
	}
	if CompanyType == 0 {
		query = query.Filter(elastic.NewTermQuery("Purchaser.keyword", strings.ToUpper(CompanyName)))
	} else {
		query = query.Filter(elastic.NewTermQuery("Supplier.keyword", strings.ToUpper(CompanyName)))
	}
	if proKey != "" {
		query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
	}
	//存在全称匹配
	if CompanyType == 0 {
		agg.Field("PurchaserId")
	} else {
		agg.Field("SupplierId")
	}
	search = search.Aggregation("search", agg).RequestCache(true)
	res, _ := search.Query(query).Size(0).Do(SearchCtx)
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("search")
	if len(terms.Buckets) == 0 {
		return nil, errors.New("没有完整匹配")
	}
	//增加一个数组 容量等于前端请求的pageSize，循环purchaseId获取详细信息
	companyId := terms.Buckets[0].Key.(float64)
	tradeNumber := terms.Buckets[0].DocCount
	frank := model.Frank{
		CompanyId:   int64(companyId),
		TradeNumber: tradeNumber,
	}
	for k, v := range terms.Buckets[0].Aggregations {
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
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query = elastic.NewBoolQuery()
	query.QueryName("frankDetail")
	highlight := elastic.NewHighlight()
	if CompanyType == 0 {
		query = query.Must(elastic.NewTermQuery("PurchaserId", frank.CompanyId))
		if CompanyName != "" {
			for i := 0; i < len(constants.Stopwords); i++ {
				if strings.Contains(strings.ToLower(CompanyName), constants.Stopwords[i]) {
					query = query.Must(elastic.NewMatchQuery("Purchaser", CompanyName))
					highlight.Field("Purchaser")
				}
			}
		}

	} else {
		query = query.Must(elastic.NewTermQuery("SupplierId", frank.CompanyId))
		if CompanyName != "" {
			for i := 0; i < len(constants.Stopwords); i++ {
				query = query.Must(elastic.NewMatchQuery("Supplier", CompanyName))
				highlight.Field("Supplier")
			}
		}
	}
	highlight = highlight.PreTags(`<font color="#FF0000">`).PostTags("</font>")

	if proKey != "" {
		query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
		highlight.Field("ProDesc")
	}
	search.Query(query).Highlight(highlight).Sort("FrankTime", false).From(0).Size(1)
	search.RequestCache(true)
	res, _ = search.Do(SearchCtx)
	var frankDetail model.Frank
	detail := res.Hits.Hits[0].Source
	jsonObject, _ := detail.MarshalJSON()
	json.Unmarshal(jsonObject, &frankDetail)
	if CompanyType == 0 {
		frank.CompanyName = frankDetail.Purchaser
		frank.CompanyId = frankDetail.PurchaserId
	} else {
		frank.CompanyName = frankDetail.Supplier
		frank.CompanyId = frankDetail.SupplierId
	}
	frank.FrankTime = frankDetail.FrankTime
	frank.QiyunPort = frankDetail.QiyunPort
	frank.OrderId = frankDetail.OrderId
	frank.ProDesc = frankDetail.ProDesc
	frank.OriginalCountry = frankDetail.OriginalCountry
	frank.MudiPort = frankDetail.MudiPort
	frank.OrderNo = frankDetail.OrderNo
	frank.ProKey = frankDetail.ProKey
	//设置高亮
	hight := res.Hits.Hits[0].Highlight
	if proKey != "" {
		frank.ProDesc = hight["ProDesc"][0]
	}
	if CompanyName != "" {
		if CompanyType == 0 {
			frank.CompanyName = hight["Purchaser"][0]
		} else {
			frank.CompanyName = hight["Supplier"][0]
		}
	}

	return &frank, nil
}

//FindMaoInfo findMapInfo.php
type FindMaoInfo struct {
	DistrictID    int `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:0采购商 1供应商> "`
	CategoryID    int `param:"<in:formData> <name:cid> <required:required> <err:cid不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:0采购商 1供应商> "`
	IeType        int `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	VwType        int `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0采购商 1供应商> "`
	//CategoryLevel int           `param:"<in:formData> <name:clevel> <required:required>  <err:clevel不能为空！>  <desc:0采购商 1供应商> "`
	TimeOut  time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	Pid      int           `param:"<in:formData> <name:pid> <required:required>  <err:pid不能为空！>  <desc:排序的参数 1 2 3>"`
	DateType int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:排序的参数 1 2 3>"`
}

//Serve s
func (param *FindMaoInfo) Serve(ctx *faygo.Context) error {
	var result []byte
	var err error
	mapInfo := service.GetMapInfo(param.IeType, param.DateType, param.Pid, param.DistrictLevel, param.CategoryID, param.DistrictID)
	result, err = jsoniter.Marshal(model.Response{
		List: mapInfo,
	})
	if err != nil {
		ctx.Log().Error(err)
	}
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, result)
}

//CategoryTopTenArea 国家分布 首页和产品页的下面的第一个表
type CategoryTopTenArea struct {
	CategoryID    int           `param:"<in:formData> <name:cid> <err:cid不能为空！>  <desc:0采购商 1供应商> "`
	VwType        int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0采购商 1供应商> "`
	Ietype        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	DateType      int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:排序的参数 1 2 3>"`
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:0采购商 1供应商> "`
	Pid           int           `param:"<in:formData> <name:pid> <err:pid不能为空！>  <desc:0采购商 1供应商> "`
}

func (param *CategoryTopTenArea) Serve(ctx *faygo.Context) error {
	var (
		SearchCtx context.Context
		cancel    context.CancelFunc
		agg       *elastic.TermsAggregation
		search    *elastic.SearchService
		query     *elastic.BoolQuery
		redisKey  string
	)
	if param.TimeOut != 0 {
		SearchCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		SearchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query = elastic.NewBoolQuery()
	agg = elastic.NewTermsAggregation()
	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))
	if param.CategoryID != 0 {
		categoryFilter(query, param.CategoryID)
	}
	if param.Pid != 0 {
		productFilter(query, param.Pid)
	}
	dataType(query, param.DateType)
	district(query, param.DistrictID, param.DistrictLevel, param.Ietype)
	if param.Ietype == 0 {
		if param.DistrictLevel == 0 {
			agg.Field("PurchaserDistrictId1")
		} else if param.DistrictLevel == 1 {
			agg.Field("PurchaserDistrictId2")
		} else if param.DistrictLevel == 2 {
			agg.Field("PurchaserDistrictId3")
		}
	} else {
		if param.DistrictLevel == 0 {
			agg.Field("SupplierDistrictId1")
		} else if param.DistrictLevel == 1 {
			agg.Field("SupplierDistrictId2")
		} else if param.DistrictLevel == 2 {
			agg.Field("SupplierDistrictId3")
		}
	}
	agg.Size(10)
	if param.VwType == 0 {
		volumeAgg := elastic.NewSumAggregation().Field("OrderVolume")
		agg = agg.SubAggregation("volume", volumeAgg)
		agg.OrderByAggregation("volume", false)
	} else {
		weightAgg := elastic.NewSumAggregation().Field("OrderWeight")
		agg = agg.SubAggregation("weight", weightAgg)
		agg.OrderByAggregation("weight", false)
	}
	res, err := search.Query(query).Aggregation("CategoryTopTenArea", agg).Size(0).RequestCache(true).Do(SearchCtx)
	if err != nil {
		ctx.Log().Error(err)
	}
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("CategoryTopTenArea")
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
			if k == "volume" {
				value := util.BytesString(data)
				volume, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					log.Println(err)
				}
				category.Value = util.Round(volume, 2)
			}
			if k == "weight" {
				value := util.BytesString(data)
				weight, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					log.Println(err)
				}
				category.Value = util.Round(weight, 2)
			}
		}
		districts = append(districts, category)
	}
	result, err := jsoniter.Marshal(model.Response{
		List: districts,
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
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, result)
}

//CategoryProductTopTenArea top10产品的国家分布 下面第二个表
type CategoryProductTopTenArea struct {
	ProductID     int           `param:"<in:formData> <name:pid> <required:required>  <err:pid不能为空！>  <desc:0采购商 1供应商> "`
	DateType      int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:排序的参数 1 2 3>"`
	VwType        int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0采购商 1供应商> "`
	Ietype        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:0采购商 1供应商> "`
}

//Serve 处理方法
func (param *CategoryProductTopTenArea) Serve(ctx *faygo.Context) error {
	var (
		SearchCtx context.Context
		cancel    context.CancelFunc
		agg       *elastic.TermsAggregation
		search    *elastic.SearchService
		query     *elastic.BoolQuery
		redisKey  string
	)
	if param.TimeOut != 0 {
		SearchCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		SearchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query = elastic.NewBoolQuery()
	agg = elastic.NewTermsAggregation()
	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))
	dataType(query, param.DateType)
	district(query, param.DistrictID, param.DistrictLevel, param.Ietype)
	productFilter(query, param.ProductID)
	if param.Ietype == 0 {
		if param.DistrictLevel == 0 {
			agg.Field("PurchaserDistrictId1")
		} else if param.DistrictLevel == 1 {
			agg.Field("PurchaserDistrictId2")
		} else if param.DistrictLevel == 2 {
			agg.Field("PurchaserDistrictId3")
		}
	} else {
		if param.DistrictLevel == 0 {
			agg.Field("SupplierDistrictId1")
		} else if param.DistrictLevel == 1 {
			agg.Field("SupplierDistrictId2")
		} else if param.DistrictLevel == 2 {
			agg.Field("SupplierDistrictId3")
		}
	}
	agg.Size(10)
	if param.VwType == 0 {
		volumeAgg := elastic.NewSumAggregation().Field("OrderVolume")
		agg = agg.SubAggregation("volume", volumeAgg)
		agg.OrderByAggregation("volume", false)
	} else {
		weightAgg := elastic.NewSumAggregation().Field("OrderWeight")
		agg = agg.SubAggregation("weight", weightAgg)
		agg.OrderByAggregation("weight", false)
	}
	res, err := search.Query(query).Aggregation("CategoryTopTenArea", agg).Size(0).RequestCache(true).Do(SearchCtx)
	if err != nil {
		ctx.Log().Error(err)
	}
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("CategoryTopTenArea")
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
			if k == "volume" {
				value := util.BytesString(data)
				volume, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					log.Println(err)
				}
				category.Value = util.Round(volume, 2)
			}
			if k == "weight" {
				value := util.BytesString(data)
				weight, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					log.Println(err)
				}
				category.Value = util.Round(weight, 2)
			}
		}
		districts = append(districts, category)
	}
	result, err := jsoniter.Marshal(model.Response{
		List: districts,
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
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, result)
}

//CategorySupplierTopTen 每一个行业的top10供应商 下面第三个表
type CategoryCompanyTopTen struct {
	CategoryID    int `param:"<in:formData> <name:cid> <required:required> <err:cid不能为空！>  <desc:0采购商 1供应商> "`
	DistrictID    int `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:0采购商 1供应商> "`
	//VwType     int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0采购商 1供应商> "`
	Ietype   int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	DateType int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:排序的参数 1 2 3>"`
	TimeOut  time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
}

func (param *CategoryCompanyTopTen) Serve(ctx *faygo.Context) error {
	var (
		SearchCtx context.Context
		cancel    context.CancelFunc
		agg       *elastic.TermsAggregation
		search    *elastic.SearchService
		query     *elastic.BoolQuery
		redisKey  string
		aggFiled  string
	)
	if param.TimeOut != 0 {
		SearchCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		SearchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query = elastic.NewBoolQuery()
	agg = elastic.NewTermsAggregation()
	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))
	district(query, param.DistrictID, param.DistrictLevel, param.Ietype)
	dataType(query, param.DateType)
	categoryFilter(query, param.CategoryID)
	if param.Ietype == 0 {
		aggFiled = "PurchaserId"
		if param.DistrictLevel == 0 {
			agg.Field("PurchaserDistrictId1")
		} else if param.DistrictLevel == 1 {
			agg.Field("PurchaserDistrictId2")
		} else if param.DistrictLevel == 2 {
			agg.Field("PurchaserDistrictId3")
		}
		purchaser := elastic.NewTermsAggregation().Field(aggFiled)
		purchaser.Size(1)
		agg = agg.SubAggregation(aggFiled, purchaser)
		agg.OrderByCountDesc()
	} else {
		aggFiled = "SupplierId"
		if param.DistrictLevel == 0 {
			agg.Field("SupplierDistrictId1")
		} else if param.DistrictLevel == 1 {
			agg.Field("SupplierDistrictId2")
		} else if param.DistrictLevel == 2 {
			agg.Field("SupplierDistrictId3")
		}
		supplier := elastic.NewTermsAggregation().Field(aggFiled)
		supplier.Size(1)
		agg = agg.SubAggregation(aggFiled, supplier)
		agg.OrderByCountDesc()
	}
	agg.Size(10)
	res, err := search.Query(query).Aggregation("CategoryCompanyTopTen", agg).Size(0).RequestCache(true).Do(SearchCtx)
	if err != nil {
		ctx.Log().Error(err)
	}
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("CategoryCompanyTopTen")
	var districts []model.Category
	//增加一个数组 容量等于前端请求的pageSize，循环purchaseId获取详细信息
	for i := 0; i < len(terms.Buckets); i++ {
		DistrictID := terms.Buckets[i].Key.(float64)
		category := model.Category{
			Did: int64(DistrictID),
		}
		category.Dname = service.GetDidNameByDid(int64(DistrictID))
		termsChilren, _ := terms.Buckets[i].Aggregations.Terms(aggFiled)
		count := termsChilren.SumOfOtherDocCount
		category.Value = float64(count)
		districts = append(districts, category)
	}
	result, err := jsoniter.Marshal(model.Response{
		List: districts,
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
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, result)
}

//CategoryVwTimeFilter 首页右侧第二个图
type CategoryVwTimeFilter struct {
	VwType        int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0volume 1weight> "`
	Ietype        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0 import 1 export 0代表查询采购商> "`
	DateType      int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:date_type 时间过滤>"`
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:地区id> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:地区等级> "`
	Pid           int           `param:"<in:formData> <name:pid>  <err:pid不能为空！>  <desc:产品id> "`
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
}

func (param *CategoryVwTimeFilter) Serve(ctx *faygo.Context) error {
	var (
		searchCtx context.Context
		cancel    context.CancelFunc
		search    *elastic.SearchService
		query     *elastic.BoolQuery
		agg       *elastic.SumAggregation
		dateAgg   *elastic.DateHistogramAggregation
		redisKey  string
	)
	if param.TimeOut != 0 {
		searchCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		searchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	agg = elastic.NewSumAggregation()
	dateAgg = elastic.NewDateHistogramAggregation()
	dateAgg.Field("FrankTime").Interval("month").Format("yyyy-MM-dd")
	var results []model.Category
	if param.VwType == 0 {
		agg.Field("OrderVolume")
	} else {
		agg.Field("OrderWeight")
	}
	dateAgg.SubAggregation("vwCount", agg)
	categorys := *GetIndustryTop10(param.DistrictID, param.DistrictLevel, param.VwType, param.Ietype,param.DateType, searchCtx)
	for i := 0; i < len(categorys); i++ {
		query = elastic.NewBoolQuery()
		district(query, param.DistrictID, param.DistrictLevel, param.Ietype)
		search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
		dataType(query, param.DateType)
		query.Filter(elastic.NewTermQuery("CategoryId", categorys[i].CategoryId))
		res, err := search.Query(query).Aggregation("indexTime", dateAgg).Do(searchCtx)
		if err != nil {
			ctx.Log().Error(err)
		}
		category := model.Category{
			CategoryId:   categorys[i].CategoryId,
			CategoryName: categorys[i].CategoryName,
		}
		aggregations := res.Aggregations
		terms, _ := aggregations.DateHistogram("indexTime")
		for i := 0; i < len(terms.Buckets); i++ {
			dateString := terms.Buckets[i].KeyAsString
			detail := model.DetailTrand{YearMonth: *dateString}
			for k, v := range terms.Buckets[i].Aggregations {
				data, _ := v.MarshalJSON()
				if k == "vwCount" {
					value := util.BytesString(data)
					volume, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
					if err != nil {
						ctx.Log().Error(err)
					}
					detail.Value = util.Round(volume, 2)
				}
			}
			category.StatisList = append(category.StatisList, detail)
		}
		results = append(results, category)

	}
	jsonResult, err := jsoniter.Marshal(model.Response{
		List: results,
	})
	if err != nil {
		ctx.Log().Error(err)
	}
	if ctx.HasData("redisKey") {
		redisKey = ctx.Data("redisKey").(string)
		err := redis.Set(redisKey, util.BytesString(jsonResult), 1*time.Hour).Err()
		if err != nil {
			ctx.Log().Error(err)
		}
	}
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, jsonResult)
}

//GlobalImport 右侧第一个图
type GlobalImport struct {
	VwType        int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0volume 1weight> "`
	Ietype        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:地区id> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:地区等级> "`
	DateType      int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:排序的参数 1 2 3>"`
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
}

func (param *GlobalImport) Serve(ctx *faygo.Context) error {
	var (
		searchCtx context.Context
		cancel    context.CancelFunc
		search    *elastic.SearchService
		query     *elastic.BoolQuery
		agg       *elastic.SumAggregation
		dateAgg   *elastic.DateHistogramAggregation
		redisKey  string
	)
	if param.TimeOut != 0 {
		searchCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		searchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	agg = elastic.NewSumAggregation()
	query = elastic.NewBoolQuery()
	dateAgg = elastic.NewDateHistogramAggregation()
	dateAgg.Field("FrankTime").Interval("month").Format("yyyy-MM-dd")
	var trends []model.DetailTrand
	dataType(query, param.DateType)
	district(query, param.DistrictID, param.DistrictLevel, param.Ietype)
	if param.VwType == 0 {
		agg.Field("OrderVolume")
	} else {
		agg.Field("OrderWeight")
	}
	dateAgg.SubAggregation("vwCount", agg)
	res, err := search.Query(query).Aggregation("GlobalImport", dateAgg).Do(searchCtx)
	if err != nil {
		ctx.Log().Error(err)
	}
	aggregations := res.Aggregations
	terms, _ := aggregations.DateHistogram("GlobalImport")
	for i := 0; i < len(terms.Buckets); i++ {
		dateString := terms.Buckets[i].KeyAsString
		detail := model.DetailTrand{YearMonth: *dateString}
		for k, v := range terms.Buckets[i].Aggregations {
			data, _ := v.MarshalJSON()
			if k == "vwCount" {
				value := util.BytesString(data)
				volume, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					ctx.Log().Error(err)
				}

				detail.Value = util.Round(volume, 2)
			}
		}
		trends = append(trends, detail)
	}
	result, err := jsoniter.Marshal(model.Response{
		List: trends,
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
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, result)
}

//DistributionRegion  下面第四个图
type DistributionRegion struct {
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:地区id> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <range 0:2> <err:dlevel不能为空！>  <desc:地区等级> "`
	VwType        int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0volume 1weight> "`
	Ietype        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
}

func (param *DistributionRegion) Serve(ctx *faygo.Context) error {
	var (
		searchCtx context.Context
		cancel    context.CancelFunc
		search    *elastic.SearchService
		query     *elastic.BoolQuery
		agg       *elastic.TermsAggregation
		redisKey  string
	)
	if param.TimeOut != 0 {
		searchCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		searchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	alldistrictName := service.GetAllDistrictName(param.DistrictLevel)
	size := len(*alldistrictName)
	query = elastic.NewBoolQuery()
	agg = elastic.NewTermsAggregation()
	if param.Ietype == 0 {
		if param.DistrictLevel == 0 {
			agg.Field("PurchaserDistrictId1")
		} else if param.DistrictLevel == 1 {
			agg.Field("PurchaserDistrictId2")
		} else if param.DistrictLevel == 2 {
			agg.Field("PurchaserDistrictId3")
		}
	} else {
		if param.DistrictLevel == 0 {
			agg.Field("SupplierDistrictId1")
		} else if param.DistrictLevel == 1 {
			agg.Field("SupplierDistrictId2")
		} else if param.DistrictLevel == 2 {
			agg.Field("SupplierDistrictId3")
		}
	}
	vwTypeAgg := elastic.NewSumAggregation()
	vwType(vwTypeAgg, param.VwType)
	agg.SubAggregation("vwCount", vwTypeAgg)
	agg.OrderByAggregation("vwCount", false).Size(size)
	district(query, param.DistrictID, param.DistrictLevel, param.Ietype)
	res, err := search.Query(query).Aggregation("DistributionRegion", agg).Size(0).RequestCache(true).Do(searchCtx)
	if err != nil {
		ctx.Log().Error(err)
	}
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("DistributionRegion")
	var districts []model.Category
	//增加一个数组 容量等于前端请求的pageSize，循环purchaseId获取详细信息
	for i := 0; i < len(terms.Buckets); i++ {
		DistrictID := terms.Buckets[i].Key.(float64)
		category := model.Category{
			Did: int64(DistrictID),
		}
		for i := 0; i < size; i++ {
			district := *alldistrictName
			if district[i].Did == category.Did {
				category.Dname = district[i].DnameEn
			}
		}
		//category.Dname = service.GetDidNameByDid(int64(DistrictID))
		for k, v := range terms.Buckets[i].Aggregations {
			data, _ := v.MarshalJSON()
			if k == "vwCount" {
				value := util.BytesString(data)
				volume, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					log.Println(err)
				}
				category.Value = util.Round(volume, 2)
			}
		}
		districts = append(districts, category)
	}
	result, err := jsoniter.Marshal(districts)
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
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, result)
}

func (param *DistributionRegion) Doc() faygo.Doc {
	return faygo.Doc{
		// API接口说明
		Note: "首页下面第四个图",
		// 响应说明或示例
		Return: "返回json数组，数组key为value，地区英文名，和did",
	}
}

//CompanyList 首页单击地图地区弹窗
//getBuyerListTest dialog.js
type DistrictCompanyList struct {
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	PageNo        int           `param:"<in:formData> <name:page_no> <required:required>  <nonzero:nonzero> <range: 1:1000>  <err:page_no必须在1到1000之间>   <desc:分页页码>"`
	PageSize      int           `param:"<in:formData> <name:page_size> <required:required>  <nonzero:nonzero> <err:page_size不能为空！>  <desc:分页的页数>"`
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:地区id> "`
	Ietype        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:地区等级> "`
	Sort          int           `param:"<in:formData> <name:sort> <required:required>  <err:sort不能为空！>  <desc:排序的参数 1 2 3>"`
}

func (param *DistrictCompanyList) Serve(ctx *faygo.Context) error {
	var (
		searchCtx   context.Context
		cancel      context.CancelFunc
		search      *elastic.SearchService
		query       *elastic.BoolQuery
		agg         *elastic.TermsAggregation
		redisKey    string
		cardinality *elastic.CardinalityAggregation
		total       float64
		franks      []model.Frank
	)
	if param.TimeOut != 0 {
		searchCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		searchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query = elastic.NewBoolQuery()
	district(query, param.DistrictID, param.DistrictLevel, param.Ietype)
	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))
	agg = elastic.NewTermsAggregation()
	if param.Sort == 2 {
		agg = agg.OrderByAggregation("volume", false)

	} else if param.Sort == -2 {
		agg = agg.OrderByAggregation("volume", true)

	} else if param.Sort == 3 {
		agg = agg.OrderByAggregation("weight", false)

	} else if param.Sort == -3 {
		agg = agg.OrderByAggregation("weight", true)

	} else if param.Sort == 1 {
		agg = agg.OrderByCount(false)

	} else if param.Sort == -1 {
		agg = agg.OrderByCount(true)

	} else {
		agg = agg.OrderByCount(false)
	}
	agg.Size(param.PageSize * param.PageNo)
	weightAgg := elastic.NewSumAggregation().Field("OrderWeight")
	volumeAgg := elastic.NewSumAggregation().Field("OrderVolume")
	agg = agg.SubAggregation("weight", weightAgg)
	agg = agg.SubAggregation("volume", volumeAgg)
	if param.Ietype == 0 {
		cardinality = elastic.NewCardinalityAggregation().Field("PurchaserId")
		agg.Field("PurchaserId")
	} else {
		cardinality = elastic.NewCardinalityAggregation().Field("SupplierId")
		agg.Field("SupplierId")
	}
	count, _ := search.Query(query).Aggregation("count", cardinality).Size(0).Do(searchCtx)
	resCardinality, _ := count.Aggregations.Cardinality("count")
	total = total + *resCardinality.Value
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	search = search.Query(query)
	search = search.Aggregation("search", agg).RequestCache(true)
	res, _ := search.Size(0).Do(searchCtx)
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("search")
	//增加一个数组 容量等于前端请求的pageSize，循环purchaseId获取详细信息
	for i := (param.PageNo - 1) * param.PageSize; i < len(terms.Buckets); i++ {
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
		if param.Ietype == 0 {
			query = query.Must(elastic.NewTermQuery("PurchaserId", franks[i].CompanyId))
		} else {
			query = query.Must(elastic.NewTermQuery("SupplierId", franks[i].CompanyId))
		}
		search.Query(query).Sort("FrankTime", false).From(0).Size(1)
		search.RequestCache(true)
		res, _ := search.Do(searchCtx)
		var frank model.Frank
		detail := res.Hits.Hits[0].Source
		jsonObject, _ := detail.MarshalJSON()
		json.Unmarshal(jsonObject, &frank)
		if param.Ietype == 0 {
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
	}
	result, err := jsoniter.Marshal(model.Response{
		List:  franks,
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
	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, result)
}

func (param *DistrictCompanyList) Doc() faygo.Doc {
	return faygo.Doc{
		// API接口说明
		Note: "单击地区弹出框",
		// 响应说明或示例
		Return: "返回json",
	}
}

type FindMapRelation struct {
	DateType      int `param:"<in:formData> <name:date_type> <desc:date_type 时间过滤>"`
	DistrictID    int `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:地区id> "`
	Ietype        int `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:地区等级> "`
}

func (param *FindMapRelation) Serve(ctx *faygo.Context) error {
	mapRelation := *service.GetMapRelation(param.Ietype, param.DateType, param.DistrictID)
	info := service.GetMapClickInfo(int64(param.DistrictID))
	value := 0
	for i := 0; i < len(mapRelation); i++ {
		value = value + mapRelation[i].Value
	}
	clickInfo := model.MapClickInfo{
		Did:       info.Did,
		Title:     info.DnameEn,
		Longitude: info.Longitude,
		Latitude:  info.Latitude,
		Value:     int64(value),
		Maps:      mapRelation,
	}
	result, err := jsoniter.Marshal(model.Response{
		Data: clickInfo,
	})
	if err != nil {
		ctx.Log().Error(err)
	}

	return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, result)

}
