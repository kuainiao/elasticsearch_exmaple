package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/henrylee2cn/faygo"
	"github.com/json-iterator/go"
	"github.com/zhangweilun/tradeweb/constants"
	"github.com/zhangweilun/tradeweb/model"
	util "github.com/zhangweilun/tradeweb/util"
	elastic "gopkg.in/olivere/elastic.v5"
	"log"
	"strconv"
	"strings"
)

/**
*
* @author willian
* @created 2017-07-27 16:45
* @email 18702515157@163.com
**/

type queryParam struct {
	Country         string `json:"country"` //属性一定要大写
	ProDesc         string `json:"pro_desc"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	Purchaser       string `json:"purchaser"`
	Supplier        string `json:"supplier"`
	OriginalCountry string `json:"original_country"`
	DestinationPort string `json:"destination_port"`
	PageNo          int    `json:"page_no"`
	PageSize        int    `json:"page_size"`
	Sort            int    `json:"sort"`
}

//产品描述或者公司名称搜索
var Search = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var param queryParam
	err := ctx.BindJSON(&param)
	if err != nil {
		fmt.Println(err)
	}
	client := constants.Instance()
	context := context.Background()
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
	if param.Purchaser != "" {
		query = query.Must(elastic.NewTermQuery("Purchaser", param.Supplier))
		highlight.Field("Purchaser")
	}
	query = query.Boost(10)
	query = query.DisableCoord(true)
	query = query.QueryName("filter")
	from := (param.PageNo - 1) * param.PageSize
	search = search.Query(query).Highlight(highlight).From(from).Size(param.PageSize)
	countSearch := client.Search().Index("trade").Type("frank")
	cardinality := elastic.NewCardinalityAggregation().Field("PurchaserId")
	count, _ := countSearch.Aggregation("count", cardinality).Do(context)
	resCardinality, _ := count.Aggregations.Cardinality("count")
	total := resCardinality.Value
	fmt.Println(*total)
	//https://www.elastic.co/guide/en/elasticsearch/reference/current/search-aggregations-bucket-terms-aggregation.html#_filtering_values_with_partitions
	agg := elastic.NewTermsAggregation().Field("PurchaserId")
	if param.Sort == 2 {
		agg = agg.OrderByAggregation("weight", false).
			Size(param.PageSize)
	} else if param.Sort == -2 {
		agg = agg.OrderByAggregation("weight", true).
			Size(param.PageSize)
	} else if param.Sort == 3 {
		agg = agg.OrderByAggregation("volume", false).
			Size(param.PageSize)
	} else if param.Sort == -3 {
		agg = agg.OrderByAggregation("volume", true).
			Size(param.PageSize)
	} else if param.Sort == 1 {
		agg = agg.OrderByCount(false).
			Size(param.PageSize)
	} else {
		agg = agg.OrderByCount(true).
			Size(param.PageSize)
	}
	//agg := elastic.NewTermsAggregation().Field("PurchaserId").OrderByAggregation("weight", false).
	//	Size(param.PageSize)
	weightAgg := elastic.NewSumAggregation().Field("OrderWeight")
	volumeAgg := elastic.NewSumAggregation().Field("OrderVolume")
	agg = agg.SubAggregation("weight", weightAgg)
	agg = agg.SubAggregation("volume", volumeAgg)
	search = search.Aggregation("search", agg).RequestCache(true)
	res, _ := search.Size(0).Do(context)
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("search")
	//增加一个数组 容量等于前端请求的pageSize，循环purchaseId获取详细信息
	var franks []model.Frank
	for i := 0; i < len(terms.Buckets); i++ {
		purchaseId := terms.Buckets[i].Key.(float64)
		tradeNumber := terms.Buckets[i].DocCount
		frank := model.Frank{
			PurchaserId: int64(purchaseId),
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
	for i := 0; i < len(franks); i++ {
		search := client.Search().Index("trade").Type("frank")
		query := elastic.NewBoolQuery()
		query.QueryName("frankDetail")
		query = query.Must(elastic.NewTermQuery("PurchaserId", franks[i].PurchaserId))
		highlight := elastic.NewHighlight()
		if param.ProDesc != "" {
			query = query.Must(elastic.NewTermQuery("ProDesc", param.ProDesc))
			highlight.Field("ProDesc")
		}
		if param.Supplier != "" {
			query = query.Must(elastic.NewTermQuery("Supplier", param.Supplier))
			highlight.Field("Supplier")
		}
		if param.Purchaser != "" {
			query = query.Must(elastic.NewTermQuery("Purchase", param.Supplier))
			highlight.Field("Purchase")
		}
		search.Query(query).Highlight(highlight).Sort("FrankTime", false).From(0).Size(1)
		search.RequestCache(true)
		res, _ := search.Do(context)
		var frank model.Frank
		detail := res.Hits.Hits[0].Source
		jsonObject, _ := detail.MarshalJSON()
		json.Unmarshal(jsonObject, &frank)

		franks[i].Purchaser = frank.Purchaser
		franks[i].FrankTime = frank.FrankTime
		franks[i].QiyunPort = frank.QiyunPort
		franks[i].OrderId = frank.OrderId
		franks[i].ProDesc = frank.ProDesc
		franks[i].Supplier = frank.Supplier
		franks[i].OriginalCountry = frank.OriginalCountry
		franks[i].MudiPort = frank.MudiPort
		franks[i].PurchaserAddress = frank.PurchaserAddress
		franks[i].OrderNo = frank.OrderNo
		franks[i].SupplierId = frank.SupplierId
		franks[i].SupplierAddress = frank.SupplierAddress
		//设置高亮
		hight := res.Hits.Hits[0].Highlight
		if param.ProDesc != "" {
			frank.ProDesc = hight["ProDesc"][0]
		}
		if param.Supplier != "" {
			frank.Supplier = hight["Supplier"][0]
		}
		if param.Purchaser != "" {
			frank.Purchaser = hight["Purchaser"][0]
		}

	}
	response := model.Response{
		List:  franks,
		Code:  0,
		Total: int64(*total),
	}

	return ctx.JSON(200, response)

})

var FrankDetail = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var param queryParam
	err := ctx.BindJSON(&param)
	if err != nil {
		fmt.Println(err)
	}
	client := constants.Instance()
	context := context.Background()
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
		query = query.Must(elastic.NewTermQuery("OriginalCountry", param.Supplier))
		highlight.Field("OriginalCountry")
	}
	query = query.Boost(10)
	query = query.DisableCoord(true)
	query = query.QueryName("frankDetail")
	from := (param.PageNo - 1) * param.PageSize
	search = search.Query(query).Highlight(highlight).From(from).Size(param.PageSize)
	//search.Sort() //排序
	res, _ := search.Do(context)
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
