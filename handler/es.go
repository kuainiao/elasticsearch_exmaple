package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/henrylee2cn/faygo"
	jsoniter "github.com/json-iterator/go"
	"github.com/zhangweilun/tradeweb/constants"
	"github.com/zhangweilun/tradeweb/model"
	util "github.com/zhangweilun/tradeweb/util"
	elastic "gopkg.in/olivere/elastic.v5"
	"log"
	"strconv"
	"strings"
	"time"
)

/**
*
* @author willian
* @created 2017-07-27 16:45
* @email 18702515157@163.com
**/

type detailQuery struct {
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

//首页
//产品描述或者公司名称搜索
var Search = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		SearchCtx context.Context
		cancel    context.CancelFunc
	)
	var param detailQuery
	err := ctx.BindJSON(&param)
	if err != nil {
		fmt.Println(err)
	}
	if param.TimeOut != 0 {
		SearchCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		SearchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	search := client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	query.QueryName("matchQuery")
	if param.ProKey != "" {
		query = query.Must(elastic.NewMatchQuery("ProDesc", param.ProKey))
		//同义词
		//words := util.Words()
		//if len(words) != 0 {
		//	for i := 0; i < len(words); i++ {
		//		query = query.Must(elastic.NewMatchQuery("ProDesc", words[i]))
		//	}
		//}
	}
	countSearch := client.Search().Index("trade").Type("frank")
	cardinality := elastic.NewCardinalityAggregation().Field("PurchaserId")
	count, _ := countSearch.Query(query).Aggregation("count", cardinality).Size(0).Do(SearchCtx)
	resCardinality, _ := count.Aggregations.Cardinality("count")
	total := resCardinality.Value
	fmt.Println(*total)
	agg := elastic.NewTermsAggregation()
	if param.CompanyType == 0 {
		agg.Field("PurchaserId")
		if param.CompanyName != "" {
			query = query.Must(elastic.NewTermQuery("Purchaser", param.CompanyName))
		}
	} else {
		agg.Field("SupplierId")
		if param.CompanyName != "" {
			query = query.Must(elastic.NewTermQuery("Supplier", param.CompanyName))
		}
	}
	search = search.Query(query)
	//https://www.elastic.co/guide/en/elasticsearch/reference/current/search-aggregations-bucket-terms-aggregation.html#_filtering_values_with_partitions
	//分页等分区处理
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
	search = search.Aggregation("search", agg).RequestCache(true)
	res, _ := search.Size(0).Do(SearchCtx)
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("search")
	//增加一个数组 容量等于前端请求的pageSize，循环purchaseId获取详细信息
	var franks []model.Frank
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

	//获取详细信息
	for i := 0; i < len(franks); i++ {
		search := client.Search().Index("trade").Type("frank")
		query := elastic.NewBoolQuery()
		query.QueryName("frankDetail")
		if param.CompanyType == 0 {
			query = query.Must(elastic.NewTermQuery("PurchaserId", franks[i].CompanyId))
		} else {
			query = query.Must(elastic.NewTermQuery("SupplierId", franks[i].CompanyId))
		}

		highlight := elastic.NewHighlight()
		highlight = highlight.PreTags(`<font color="#FF0000">`).PostTags("</font>")
		if param.ProKey != "" {
			query = query.Must(elastic.NewMatchQuery("ProDesc", param.ProKey))
			highlight.Field("ProDesc")
		}
		search.Query(query).Highlight(highlight).Sort("FrankTime", false).From(0).Size(1)
		search.RequestCache(true)
		res, _ := search.Do(SearchCtx)
		var frank model.Frank
		detail := res.Hits.Hits[0].Source
		jsonObject, _ := detail.MarshalJSON()
		json.Unmarshal(jsonObject, &frank)
		if param.CompanyType == 0 {
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
		//franks[i].PurchaserAddress = frank.PurchaserAddress
		franks[i].OrderNo = frank.OrderNo
		//franks[i].SupplierAddress = frank.SupplierAddress
		franks[i].ProKey = frank.ProKey
		//设置高亮
		hight := res.Hits.Hits[0].Highlight
		if param.ProKey != "" {
			fmt.Println(hight["ProDesc"][0])
			franks[i].ProDesc = hight["ProDesc"][0]
		}
	}
	response := model.Response{
		List:  franks,
		Code:  0,
		Total: int64(*total),
	}

	return ctx.JSON(200, response)

})

//首页
//搜索提单
var FrankDetail = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		FrankDetailCtx context.Context
		cancel         context.CancelFunc
	)
	var param queryParam
	err := ctx.BindJSON(&param)
	if err != nil {
		fmt.Println(err)
	}
	if param.TimeOut != 0 {
		FrankDetailCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		FrankDetailCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	search := client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	highlight := elastic.NewHighlight()
	if param.ProDesc != "" {
		query = query.Must(elastic.NewTermQuery("ProDesc", param.ProDesc))
		highlight.Field("ProDesc")
	}
	if param.Supplier != "" {
		query = query.Must(elastic.NewTermQuery("Supplier", param.Supplier))
		highlight.Field("Supplier")
	}
	if param.OriginalCountry != "" {
		query = query.Must(elastic.NewTermQuery("OriginalCountry", param.OriginalCountry))
		highlight.Field("OriginalCountry")
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
	response := model.Response{
		List:  franks,
		Total: res.Hits.TotalHits,
		Code:  0,
	}
	return ctx.JSON(200, response)
})

// TopTenProduct... 详情
//top10饼图和中间的商品名称 all product由前端展示
//传入参数公司id 公司类型
// status: ok
var TopTenProduct = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		TopTenProductCtx context.Context
		cancel           context.CancelFunc
	)
	var param detailQuery
	err := ctx.BindJSON(&param)
	if err != nil {
		fmt.Println(err)
	}
	if param.TimeOut != 0 {
		TopTenProductCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		TopTenProductCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	TopTenSearch := client.Search().Index("trade").Type("frank")
	var query *elastic.TermQuery
	if param.CompanyType == 0 {
		query = elastic.NewTermQuery("PurchaserId", param.CompanyId).QueryName("purchaserId")
	} else {
		query = elastic.NewTermQuery("SupplierId", param.CompanyId).QueryName("SupplierId")
	}
	agg := elastic.NewTermsAggregation().Field("ProductName.keyword").OrderByCount(false).Size(10)
	res, _ := TopTenSearch.Query(query).Aggregation("TopTen", agg).Size(0).Do(TopTenProductCtx)
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("TopTen")
	var topTenProducts []model.TopTenProduct
	for i := 0; i < len(terms.Buckets); i++ {
		ProductName := terms.Buckets[i].Key.(string)
		count := terms.Buckets[i].DocCount
		top10 := model.TopTenProduct{
			ProductName: ProductName,
			Count:       count,
		}
		topTenProducts = append(topTenProducts, top10)
	}
	response := model.Response{
		List: topTenProducts,
		Code: 0,
	}
	return ctx.JSON(200, response)
})

// CompanyRelations ... 详情
//公司关系图
//传入参数proKey 公司id 公司类型
//去重还未完成
var CompanyRelations = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		CompanyRelationsCtx context.Context
		cancel              context.CancelFunc
	)
	var param detailQuery
	err := ctx.BindJSON(&param)
	if err != nil {
		fmt.Println(err)
	}
	if param.TimeOut != 0 {
		CompanyRelationsCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		CompanyRelationsCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	relationship := model.Relationship{
		CompanyId:   param.CompanyId,
		CompanyName: param.CompanyName,
	}
	client := constants.Instance()
	CompanyRelationsSearch := client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	highlight := elastic.NewHighlight()
	if param.ProKey != "" {
		query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
		highlight.Field("ProDesc")
	}
	if param.CompanyType == 0 {
		query = query.Must(elastic.NewTermQuery("PurchaserId", param.CompanyId))
	} else {
		query = query.Must(elastic.NewTermQuery("SupplierId", param.CompanyId))
	}
	query = query.Boost(10)
	query = query.DisableCoord(true)
	query = query.QueryName("filter")
	res, err := CompanyRelationsSearch.Query(query).Size(10).Sort("FrankTime", false).Do(CompanyRelationsCtx)
	if err != nil {
		log.Println(err)
	}
	//一级 采购
	if param.CompanyType == 0 {
		//去重
		levelOne := make(map[string]int64)
		//levelTwo := make(map[string]int64)
		//levelThree := make(map[string]int64)
		for i := 0; i < len(res.Hits.Hits); i++ {
			detail := res.Hits.Hits[i].Source
			var frank model.Frank
			jsonObject, _ := detail.MarshalJSON()
			jsoniter.Unmarshal(jsonObject, &frank)
			levelOne[frank.Supplier] = frank.SupplierId
		}
		for k, v := range levelOne {
			relationship.Partner = append(relationship.Partner, model.Relationship{v, k, nil})
		}

		//查采购 二级
		serviceTwo := client.Search().Index("trade").Type("frank")
		serviceTwo.Size(1).Sort("FrankTime", false)
		for j := 0; j < len(relationship.Partner); j++ {
			query := elastic.NewBoolQuery()
			query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
			query = query.Must(elastic.NewTermQuery("SupplierId", relationship.Partner[j].CompanyId))
			query = query.QueryName("filter")
			res, err := serviceTwo.Query(query).Do(CompanyRelationsCtx)
			if err != nil {
				fmt.Println(err)
			}
			detail := res.Hits.Hits[0].Source
			var frank model.Frank
			jsonObject, _ := detail.MarshalJSON()
			jsoniter.Unmarshal(jsonObject, &frank)
			relationship.Partner[j].Partner = append(relationship.Partner[j].Partner, model.Relationship{frank.PurchaserId, frank.Purchaser, nil})

		}
		//查供应商 三级
		serviceThree := client.Search().Index("trade").Type("frank")
		serviceThree = serviceThree.Size(1).Sort("FrankTime", false)
		for j := 0; j < len(relationship.Partner); j++ {
			for i := 0; i < len(relationship.Partner[j].Partner); i++ {
				query := elastic.NewBoolQuery()
				query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
				query = query.Must(elastic.NewTermQuery("PurchaserId", relationship.Partner[j].Partner[i].CompanyId))
				query = query.QueryName("filter")
				res, err := serviceThree.Query(query).Do(CompanyRelationsCtx)
				if err != nil {
					fmt.Println(err)
				}
				detail := res.Hits.Hits[0].Source
				var frank model.Frank
				jsonObject, _ := detail.MarshalJSON()
				jsoniter.Unmarshal(jsonObject, &frank)
				relationship.Partner[j].Partner[i].Partner = append(relationship.Partner[j].Partner[i].Partner, model.Relationship{frank.SupplierId, frank.Supplier, nil})
			}
		}

	} else {
		//去重
		levelOne := make(map[string]int64)
		levelTwo := make(map[string]int64)
		levelThree := make(map[string]int64)
		for i := 0; i < len(res.Hits.Hits); i++ {
			detail := res.Hits.Hits[i].Source
			var frank model.Frank
			jsonObject, _ := detail.MarshalJSON()
			jsoniter.Unmarshal(jsonObject, &frank)
			levelOne[frank.Purchaser] = frank.PurchaserId
		}
		for k, v := range levelOne {
			relationship.Partner = append(relationship.Partner, model.Relationship{v, k, nil})
		}

		//查采购 二级
		serviceTwo := client.Search().Index("trade").Type("frank")
		for j := 0; j < len(relationship.Partner); j++ {
			query := elastic.NewBoolQuery()
			query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
			query = query.Must(elastic.NewTermQuery("SupplierId", relationship.Partner[j].CompanyId))
			query = query.QueryName("filter")
			res, err := serviceTwo.Query(query).Size(2).Sort("FrankTime", false).Do(CompanyRelationsCtx)
			if err != nil {
				fmt.Println(err)
			}
			if len(res.Hits.Hits) > 1 {
				for i := 0; i < len(res.Hits.Hits); i++ {
					detail := res.Hits.Hits[i].Source
					var frank model.Frank
					jsonObject, _ := detail.MarshalJSON()
					jsoniter.Unmarshal(jsonObject, &frank)
					levelTwo[frank.Supplier] = frank.SupplierId
				}
				for k, v := range levelTwo {
					relationship.Partner[j].Partner = append(relationship.Partner[j].Partner, model.Relationship{v, k, nil})
				}
			} else {
				detail := res.Hits.Hits[0].Source
				var frank model.Frank
				jsonObject, _ := detail.MarshalJSON()
				jsoniter.Unmarshal(jsonObject, &frank)
				relationship.Partner[j].Partner = append(relationship.Partner[j].Partner, model.Relationship{frank.PurchaserId, frank.Purchaser, nil})
			}

		}

		//查供应商 三级
		serviceThree := client.Search().Index("trade").Type("frank")
		for j := 0; j < len(relationship.Partner); j++ {
			for i := 0; i < len(relationship.Partner[j].Partner); i++ {
				query := elastic.NewBoolQuery()
				query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
				query = query.Must(elastic.NewTermQuery("PurchaserId", relationship.Partner[j].Partner[i].CompanyId))
				query = query.QueryName("filter")
				res, err := serviceThree.Query(query).Size(2).Sort("FrankTime", false).Do(CompanyRelationsCtx)
				if err != nil {
					fmt.Println(err)
				}
				if len(res.Hits.Hits) > 1 {
					for i := 0; i < len(res.Hits.Hits); i++ {
						detail := res.Hits.Hits[i].Source
						var frank model.Frank
						jsonObject, _ := detail.MarshalJSON()
						jsoniter.Unmarshal(jsonObject, &frank)
						levelThree[frank.Purchaser] = frank.PurchaserId
					}
					for k, v := range levelThree {
						relationship.Partner[j].Partner[i].Partner = append(relationship.Partner[j].Partner[i].Partner, model.Relationship{v, k, nil})
					}
				} else {
					detail := res.Hits.Hits[0].Source
					var frank model.Frank
					jsonObject, _ := detail.MarshalJSON()
					jsoniter.Unmarshal(jsonObject, &frank)
					relationship.Partner[j].Partner[i].Partner = append(relationship.Partner[j].Partner[i].Partner, model.Relationship{frank.SupplierId, frank.Supplier, nil})
				}
			}
		}

	}
	return ctx.JSON(200, model.Response{
		List: relationship,
		Code: 0,
	})
})

// GroupHistory... 详情
//Nearly a year of trading history
//通过proKey相关的公司的近一年的交易记录 如果是采购上进来 先查prokey 再通过group supplier分组来处理
//传入 参数 proKey 公司ID 公司类型
// not ok proKey:vacuum cleaner未决解
var GroupHistory = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		GroupHistoryCtx context.Context
		cancel          context.CancelFunc
	)
	var param detailQuery
	err := ctx.BindJSON(&param)
	if err != nil {
		fmt.Println(err)
	}
	if param.TimeOut != 0 {
		GroupHistoryCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		GroupHistoryCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	client := constants.Instance()
	GroupHistorySearch := client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	highlight := elastic.NewHighlight()
	var cardinality *elastic.CardinalityAggregation
	//过滤一年
	//query.Must(elastic.NewRangeQuery("FrankTime").From("now-1y").To("now"))
	query.Filter(elastic.NewRangeQuery("FrankTime").From("now-1y").To("now"))
	if param.ProKey != "" {
		query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
		highlight.Field("ProDesc")
		if param.CompanyType == 0 {
			if param.CompanyId != 0 {
				query = query.Must(elastic.NewTermQuery("PurchaserId", param.CompanyId))
				cardinality = elastic.NewCardinalityAggregation().Field("SupplierId")
			}
		} else {
			if param.CompanyId != 0 {
				query = query.Must(elastic.NewTermQuery("SupplierId", param.CompanyId))
				cardinality = elastic.NewCardinalityAggregation().Field("PurchaserId")
			}
		}
	}
	countSearch := client.Search().Index("trade").Type("frank")
	count, _ := countSearch.Query(query).Aggregation("count", cardinality).Size(0).Do(GroupHistoryCtx)
	resCardinality, _ := count.Aggregations.Cardinality("count")
	//采购进来的
	agg := elastic.NewTermsAggregation()
	if param.CompanyType == 0 {
		agg.Field("SupplierId")
	} else {
		agg.Field("PurchaserId")
	}
	//排序
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
	agg = agg.Size(param.PageSize * param.PageNo)
	weightAgg := elastic.NewSumAggregation().Field("OrderWeight")
	volumeAgg := elastic.NewSumAggregation().Field("OrderVolume")
	agg = agg.SubAggregation("weight", weightAgg)
	agg = agg.SubAggregation("volume", volumeAgg)
	agg = agg.OrderByCount(false)
	GroupHistorySearch = GroupHistorySearch.Query(query).Aggregation("search", agg).RequestCache(true)
	res, _ := GroupHistorySearch.Size(0).Do(GroupHistoryCtx)
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("search")
	//拿去聚合数据
	var franks []model.Frank
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
				frank.OrderVolume = volume
			}
			if k == "weight" {
				value := util.BytesString(data)
				weight, err := strconv.ParseFloat(value[strings.Index(value, ":")+1:len(value)-1], 10)
				if err != nil {
					log.Println(err)
				}
				frank.OrderWeight = weight
			}
		}
		franks = append(franks, frank)
	}

	//获取详细信息
	search := client.Search().Index("trade").Type("frank")
	search = search.Sort("FrankTime", false).From(0).Size(1)
	search = search.RequestCache(true)
	for i := 0; i < len(franks); i++ {
		queryDeatil := elastic.NewBoolQuery()
		queryDeatil.QueryName("frankDetail")
		if param.CompanyType == 0 {
			queryDeatil = queryDeatil.Must(elastic.NewTermQuery("SupplierId", franks[i].CompanyId))
		} else {
			queryDeatil = queryDeatil.Must(elastic.NewTermQuery("PurchaserId", franks[i].CompanyId))
		}
		if param.ProKey != "" {
			queryDeatil = queryDeatil.Must(elastic.NewMatchQuery("ProDesc", param.ProKey))
		}
		//search.Query(query).Sort("FrankTime", false).From(0).Size(1)
		search.Query(queryDeatil)
		res, _ := search.Do(GroupHistoryCtx)
		var frank model.Frank
		detail := res.Hits.Hits[0].Source
		jsonObject, _ := detail.MarshalJSON()
		jsoniter.Unmarshal(jsonObject, &frank)
		if param.CompanyType == 0 {
			franks[i].CompanyName = frank.Supplier
			franks[i].CompanyId = frank.SupplierId
			franks[i].CompanyAddress = frank.SupplierAddress
		} else {
			franks[i].CompanyName = frank.Purchaser
			franks[i].CompanyId = frank.PurchaserId
			franks[i].CompanyAddress = frank.PurchaserAddress
		}
		franks[i].FrankTime = frank.FrankTime
		franks[i].ProductName = frank.ProductName
		franks[i].QiyunPort = frank.QiyunPort
		franks[i].OrderId = frank.OrderId
		franks[i].ProDesc = frank.ProDesc
		franks[i].OriginalCountry = frank.OriginalCountry
		franks[i].MudiPort = frank.MudiPort
		franks[i].OrderNo = frank.OrderNo
		franks[i].SupplierId = frank.SupplierId
		franks[i].ProKey = frank.ProKey
	}
	response := model.Response{
		List:  franks,
		Total: int64(*resCardinality.Value),
		Code:  0,
	}
	return ctx.JSON(200, response)
})

//最新10条交易记录
//传入参数公司id 公司类型
//{ "company_type":0, "page_no":1, "page_size":10, "company_id":143382, "pro_key":"", "company_name":"", "time_out":5}
// status :ok
var NewTenFrank = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		NewTenFrankCtx context.Context
		cancel         context.CancelFunc
	)
	var param detailQuery
	err := ctx.BindJSON(&param)
	if err != nil {
		fmt.Println(err)
	}
	if param.TimeOut != 0 {
		NewTenFrankCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		NewTenFrankCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	NewTenFrankSearch := client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	if param.CompanyType == 0 {
		if param.CompanyId != 0 {
			query = query.Must(elastic.NewTermQuery("PurchaserId", param.CompanyId))
		}
	} else {
		if param.CompanyId != 0 {
			query = query.Must(elastic.NewTermQuery("SupplierId", param.CompanyId))
		}
	}
	from := (param.PageNo - 1) * param.PageSize
	NewTenFrankSearch.Query(query).Sort("FrankTime", false).From(from).Size(param.PageSize)
	NewTenFrankSearch.RequestCache(true)
	res, err := NewTenFrankSearch.Do(NewTenFrankCtx)
	if err != nil {
		fmt.Println(err)
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
		List: franks,
		Code: 0,
	})
})

//与供应商的最近一年交易情况
//参数 供应商id 采购商id 参数proKey
// status: ok
var InfoDetail = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		infoDetailCtx context.Context
		cancel        context.CancelFunc
	)
	var param detailQuery
	err := ctx.BindJSON(&param)
	if err != nil {
		fmt.Println(err)
	}
	if param.TimeOut != 0 {
		infoDetailCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		infoDetailCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	InfoDetailSearch := client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
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
		fmt.Println(err)
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
//	ProductListSearch := client.Search().Index("trade").Type("product")
//	query := elastic.NewBoolQuery()
//	query = query.Must(elastic.NewMatchQuery("ProductName", strings.ToLower(param.ProKey)))
//	from := (param.PageNo - 1) * param.PageSize
//	ProductListSearch.Query(query).From(from).Size(param.PageSize)
//	res, err := ProductListSearch.Do(IndexProductCtx)
//	return nil
//})

//首页
//得到产品列表
// 传入参数 prokey
//{ "company_type":0, "page_no":1, "page_size":10, "company_id":0, "pro_key":"q", "company_name":"", "time_out":5 }
var ProductList = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		ProductListCtx context.Context
		cancel         context.CancelFunc
	)
	var param detailQuery
	err := ctx.BindJSON(&param)
	if err != nil {
		fmt.Println(err)
	}
	if param.TimeOut != 0 {
		ProductListCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		ProductListCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	ProductListSearch := client.Search().Index("trade").Type("product")
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewWildcardQuery("ProductName", "*"+strings.ToLower(param.ProKey)+"*"))
	from := (param.PageNo - 1) * param.PageSize
	ProductListSearch.Query(query).From(from).Size(param.PageSize)
	res, err := ProductListSearch.Do(ProductListCtx)
	if err != nil {
		fmt.Println(err)
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
