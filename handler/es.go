package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"net/url"

	"github.com/henrylee2cn/faygo"
	"github.com/henrylee2cn/faygo/ext/db/xorm"
	jsoniter "github.com/json-iterator/go"
	"github.com/zhangweilun/tradeweb/constants"
	"github.com/zhangweilun/tradeweb/model"
	util "github.com/zhangweilun/tradeweb/util"
	elastic "gopkg.in/olivere/elastic.v5"
)

/**
*
* @author willian
* @created 2017-07-27 16:45
* @email 18702515157@163.com
**/

type DetailQuery struct {
	CompanyType     int           `json:"company_type,omitempty"` // 0：采购商 1:供应商
	PageNo          int           `json:"page_no"`
	PageSize        int           `json:"page_size"`
	CompanyId       int64         `json:"company_id"`
	ProKey          string        `json:"pro_key"`
	TimeOut         time.Duration `json:"time_out"`
	CompanyName     string        `json:"company_name"`
	Sort            int           `json:"sort"` //1:tradeNumber 倒序 2volume 3weight
	LinkedCompanyId int           `json:"linked_company_id,omitempty"`
}

type queryParam struct {
	SupplierId      int           `json:"supplier_id"`
	ProKey          string        `json:"pro_key"`
	Country         string        `json:"country"` //属性一定要大写
	ProDesc         string        `json:"pro_desc"`
	StartDate       string        `json:"start_date"`
	EndDate         string        `json:"end_date"`
	Purchaser       string        `json:"purchaser"`
	Supplier        string        `json:"supplier"`
	OriginalCountry string        `json:"original_country"`
	DestinationPort string        `json:"destination_port"`
	PageNo          int           `json:"page_no"`
	PageSize        int           `json:"page_size"`
	Sort            int           `json:"sort"`
	TimeOut         time.Duration `json:"time_out"`
	PurchaserId     int           `json:"purchaser_id"`
}

//FrankDetail ...
//首页
//搜索提单
var FrankDetail = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		FrankDetailCtx context.Context
		cancel         context.CancelFunc
		redisKey       string
	)
	var param queryParam
	err := ctx.BindJSON(&param)
	if err != nil {
		ctx.Log().Error(err)
	}
	if param.PageNo > 1000 {
		param.PageNo = 1000
	}
	if param.TimeOut != 0 {
		FrankDetailCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		FrankDetailCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	search := client.Search().Index("trade").Type(constants.TypeName)
	query := elastic.NewBoolQuery()
	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))

	highlight := elastic.NewHighlight()
	if param.ProDesc != "" {
		query = query.Must(elastic.NewMatchQuery("ProDesc", param.ProDesc))
		highlight.Field("ProDesc")
	}
	if param.Supplier != "" {
		query = query.Must(elastic.NewMatchQuery("Supplier", param.Supplier))
		highlight.Field("Supplier")
	}
	if param.OriginalCountry != "" {
		query = query.Must(elastic.NewMatchQuery("OriginalCountry", param.OriginalCountry))
		highlight.Field("OriginalCountry")
	}
	if param.StartDate != "" && param.EndDate != "" {
		query = query.Filter(elastic.NewRangeQuery("FrankTime").From(param.StartDate).To(param.EndDate))
	} else if param.StartDate != "" && param.EndDate == "" {
		query = query.Filter(elastic.NewRangeQuery("FrankTime").From(param.StartDate).To(nil))
	} else if param.StartDate == "" && param.EndDate != "" {
		query = query.Filter(elastic.NewRangeQuery("FrankTime").From(nil).To(param.EndDate))
	}
	query = query.Boost(10)
	query = query.DisableCoord(true)
	query = query.QueryName("frankDetail")
	from := (param.PageNo - 1) * param.PageSize
	search = search.Query(query).Highlight(highlight).From(from).Size(param.PageSize)
	//search.Sort() //排序
	res, _ := search.Do(FrankDetailCtx)
	var franks []model.Frank
	for i := 0; i < len(res.Hits.Hits); i++ {
		detail := res.Hits.Hits[i].Source
		var frank model.Frank
		jsonObject, _ := detail.MarshalJSON()
		jsoniter.Unmarshal(jsonObject, &frank)
		//高亮
		hight := res.Hits.Hits[i].Highlight
		if param.ProDesc != "" {
			frank.ProDesc = hight["ProDesc"][0]
		}
		if param.Supplier != "" {
			frank.Supplier = hight["Supplier"][0]
		}
		if param.OriginalCountry != "" {
			frank.OriginalCountry = hight["OriginalCountry"][0]
		}
		franks = append(franks, frank)
	}

	result, err := jsoniter.Marshal(model.Response{
		List:  franks,
		Total: res.Hits.TotalHits,
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
})

// TopTenProduct ... 详情
//top10饼图和中间的商品名称 all product由前端展示
//传入参数公司id 公司类型
type TopTenProduct struct {
	CompanyType int           `param:"<in:formData> <name:company_type> <required:required>  <range: 0:2>  <err:company_type必须在0到2之间>  <desc:公司类型>"`
	CompanyID   int           `param:"<in:formData> <name:company_id> <required:required> <nonzero:nonzero>  <err:company_id不能为0>  <desc:公司类型>"`
	TimeOut     time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
}

func (param *TopTenProduct) Serve(ctx *faygo.Context) error {
	var (
		TopTenProductCtx context.Context
		cancel           context.CancelFunc
		redisKey         string
	)
	if param.TimeOut != 0 {
		TopTenProductCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		TopTenProductCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	TopTenSearch := client.Search().Index(constants.IndexName).Type(constants.TypeName)
	var query *elastic.TermQuery
	if param.CompanyType == 0 {
		query = elastic.NewTermQuery("PurchaserId", param.CompanyID).QueryName("purchaserId")
	} else {
		query = elastic.NewTermQuery("SupplierId", param.CompanyID).QueryName("SupplierId")
	}
	agg := elastic.NewTermsAggregation().Field("ProductId").OrderByCount(false).Size(11)
	res, _ := TopTenSearch.Query(query).Aggregation("TopTen", agg).Size(0).Do(TopTenProductCtx)
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("TopTen")
	var topTenProducts []model.TopTenProduct
	db := xorm.MustDB("default")
	var otherTotal int64
	for i := 0; i < len(terms.Buckets); i++ {
		ProductId := terms.Buckets[i].Key.(float64)
		count := terms.Buckets[i].DocCount
		if i > 9 {
			otherTotal = otherTotal + count
		} else {
			var productNew model.ProductNew
			productNew.Id = int64(ProductId)
			ok, err := db.Get(&productNew)
			if !ok {
				ctx.Log().Error(err)
			}
			top10 := model.TopTenProduct{
				ProductName: productNew.Name,
				Count:       count,
				ProId:       int64(ProductId),
			}
			topTenProducts = append(topTenProducts, top10)
		}
	}
	topTenProducts = append(topTenProducts, model.TopTenProduct{
		ProductName: "Other product",
		Count:       otherTotal,
		ProId:       0,
	})
	result, err := jsoniter.Marshal(model.Response{
		List: topTenProducts,
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

//NewTenFrank 最新10条交易记录
//传入参数公司id 公司类型
type NewTenFrank struct {
	CompanyType int           `param:"<in:formData> <name:company_type> <required:required>  <range: 0:2>  <err:company_type必须在0到2之间>  <desc:公司类型>"`
	CompanyID   int           `param:"<in:formData> <name:company_id> <required:required> <nonzero:nonzero>  <err:company_id不能为0>  <desc:公司类型>"`
	TimeOut     time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	PageNo      int           `param:"<in:formData> <name:page_no> <required:required>  <nonzero:nonzero> <range: 1:1000>  <err:page_no必须在1到1000之间>  <desc:分页页码>"`
	PageSize    int           `param:"<in:formData> <name:page_size> <required:required>  <nonzero:nonzero> <err:page_size不能为空！>  <desc:分页的页数>"`
}

func (param *NewTenFrank) Serve(ctx *faygo.Context) error {
	var (
		NewTenFrankCtx context.Context
		cancel         context.CancelFunc
		redisKey       string
	)
	if param.TimeOut != 0 {
		NewTenFrankCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		NewTenFrankCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	NewTenFrankSearch := client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query := elastic.NewBoolQuery()
	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))

	if param.CompanyType == 0 {
		if param.CompanyID != 0 {
			query = query.Must(elastic.NewTermQuery("PurchaserId", param.CompanyID))
		}
	} else {
		if param.CompanyID != 0 {
			query = query.Must(elastic.NewTermQuery("SupplierId", param.CompanyID))
		}
	}
	from := (param.PageNo - 1) * param.PageSize
	NewTenFrankSearch.Query(query).Sort("FrankTime", false).From(from).Size(param.PageSize)
	NewTenFrankSearch.RequestCache(true)
	res, err := NewTenFrankSearch.Do(NewTenFrankCtx)
	if err != nil {
		ctx.Log().Error(err)
	}
	var franks []model.Frank
	for i := 0; i < len(res.Hits.Hits); i++ {
		detail := res.Hits.Hits[i].Source
		var frank model.Frank
		jsonObject, _ := detail.MarshalJSON()
		jsoniter.Unmarshal(jsonObject, &frank)
		if param.CompanyType == 0 {
			frank.CompanyName = frank.Purchaser
			frank.CompanyId = frank.PurchaserId
			frank.CompanyAddress = frank.PurchaserAddress
		} else {
			frank.CompanyName = frank.Supplier
			frank.CompanyId = frank.SupplierId
			frank.CompanyAddress = frank.SupplierAddress
		}
		franks = append(franks, frank)
	}
	result, err := jsoniter.Marshal(model.Response{
		List: franks,
		Code: 0,
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

//InfoDetail ...
//参数 供应商id 采购商id 参数proKey
// status: ok
var InfoDetail = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		infoDetailCtx context.Context
		cancel        context.CancelFunc
	)
	var param DetailQuery
	err := ctx.BindJSON(&param)
	if err != nil {
		ctx.Log().Error(err)
	}
	if param.TimeOut != 0 {
		infoDetailCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		infoDetailCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	InfoDetailSearch := client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query := elastic.NewBoolQuery()
	proKey, err := url.PathUnescape(param.ProKey)
	if err != nil {
		ctx.Log().Error(err)
	}
	query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))

	query.Filter(elastic.NewRangeQuery("FrankTime").From("now-1y").To("now"))
	if param.CompanyType == 0 {
		query = query.Must(elastic.NewTermQuery("PurchaserId", param.CompanyId))
		query = query.Must(elastic.NewTermQuery("SupplierId", param.LinkedCompanyId))
	} else {
		query = query.Must(elastic.NewTermQuery("PurchaserId", param.LinkedCompanyId))
		query = query.Must(elastic.NewTermQuery("SupplierId", param.CompanyId))
	}
	from := (param.PageNo - 1) * param.PageSize
	res, err := InfoDetailSearch.Query(query).From(from).Size(param.PageSize).Do(infoDetailCtx)
	if err != nil {
		ctx.Log().Error(err)
	}
	var franks []model.Frank
	for i := 0; i < len(res.Hits.Hits); i++ {
		detail := res.Hits.Hits[i].Source
		var frank model.Frank
		jsonObject, _ := detail.MarshalJSON()
		jsoniter.Unmarshal(jsonObject, &frank)
		if param.CompanyType == 0 {
			frank.CompanyName = frank.Purchaser
			frank.CompanyId = frank.PurchaserId
			frank.CompanyAddress = frank.PurchaserAddress
		} else {
			frank.CompanyName = frank.Supplier
			frank.CompanyId = frank.SupplierId
			frank.CompanyAddress = frank.SupplierAddress
		}
		franks = append(franks, frank)
	}
	return ctx.JSON(200, model.Response{
		List:  franks,
		Code:  0,
		Total: res.Hits.TotalHits,
	})
})

//findSupplierTop10 ...
//{"did":0,"pid":"15","dlevel":0,"ietype":0,"vwtype":0,"token":"user:3183b6956804427f91d7b624db09e547","userId":"1","date_type":2}
var findSupplierTop10 = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	return nil
})

//findBusinessTrendInfo.php 找写的逻辑
//detailReq
//findbusinessDistribution.php
//regionMap
//首页产品搜索
//包括全球产品出口柜量占比（国家）
//包括全球产品进口柜量占比（国家）
//全球该商品的采购商总数
//全球该商品进口柜量总数
//全球进口该商品的top10采购商（柜量）
//全球进口该商品的采购商的占比图（柜量）
//var IndexProductSearch = faygo.HandlerFunc(func(ctx *faygo.Context) error {
//	var (
//		IndexProductCtx context.Context
//		cancel         context.CancelFunc
//	)
//	var param detailQuery
//	err := ctx.BindJSON(&param)
//	if err != nil {
//		fmt.Println(err)
//	}
//	if param.TimeOut != 0 {
//		IndexProductCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
//	} else {
//		IndexProductCtx, cancel = context.WithCancel(context.Background())
//	}
//	defer cancel()
//	client := constants.Instance()
//	ProductListSearch := client.Search().Index(constants.IndexName).Type("product")
//	query := elastic.NewBoolQuery()
//	query = query.Must(elastic.NewMatchQuery("ProductName", strings.ToLower(param.ProKey)))
//	from := (param.PageNo - 1) * param.PageSize
//	ProductListSearch.Query(query).From(from).Size(param.PageSize)
//	res, err := ProductListSearch.Do(IndexProductCtx)
//	return nil
//})

//DetailOne 通过orderid得出详情
var DetailOne = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	orderID, err := strconv.Atoi(ctx.QueryParam("orderId"))
	if err != nil {
		ctx.Log().Error(err)
	}
	var (
		DetailOneCtx context.Context
		cancel       context.CancelFunc
	)
	DetailOneCtx, cancel = context.WithCancel(context.Background())
	defer cancel()
	client := constants.Instance()
	search := client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("OrderId", orderID))
	res, err := search.Size(1).Query(query).Do(DetailOneCtx)
	if err != nil {
		ctx.Log().Error(err)
	}
	detail := res.Hits.Hits[0].Source
	var frank model.Frank
	jsonObject, _ := detail.MarshalJSON()
	jsoniter.Unmarshal(jsonObject, &frank)
	return ctx.JSON(200, model.Response{
		Code: 0,
		Data: frank,
	})
})

//ProductList ...
//首页
//得到产品列表
// 传入参数 prokey
//{ "company_type":0, "page_no":1, "page_size":10, "company_id":0, "pro_key":"q", "company_name":"", "time_out":5 }
var ProductList = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		ProductListCtx context.Context
		cancel         context.CancelFunc
	)
	var param DetailQuery
	err := ctx.BindJSON(&param)
	if err != nil {
		ctx.Log().Error(err)
	}
	if param.TimeOut != 0 {
		ProductListCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		ProductListCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	ProductListSearch := client.Search().Index(constants.IndexName).Type("product")
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewWildcardQuery("ProductName", "*"+strings.ToLower(param.ProKey)+"*"))
	from := (param.PageNo - 1) * param.PageSize
	ProductListSearch.Query(query).From(from).Size(param.PageSize)
	res, err := ProductListSearch.Do(ProductListCtx)
	if err != nil {
		ctx.Log().Error(err)
	}
	var franks []model.Product
	for i := 0; i < len(res.Hits.Hits); i++ {
		detail := res.Hits.Hits[i].Source
		var product model.Product
		jsonObject, _ := detail.MarshalJSON()
		jsoniter.Unmarshal(jsonObject, &product)
		franks = append(franks, product)
	}
	return ctx.JSON(200, model.Response{
		List: franks,
		Code: 0,
	})
})
