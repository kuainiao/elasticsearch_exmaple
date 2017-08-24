package handler

import (
	"context"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/henrylee2cn/faygo"
	"github.com/json-iterator/go"
	"github.com/zhangweilun/tradeweb/constants"

	"github.com/zhangweilun/gor"
	"github.com/zhangweilun/tradeweb/model"
	"github.com/zhangweilun/tradeweb/service"
	util "github.com/zhangweilun/tradeweb/util"
	"gopkg.in/olivere/elastic.v5"
)

// CompanyRelations ... 详情
//公司关系图
//传入参数proKey 公司id 公司类型
//https://elasticsearch.cn/article/132
//contentType: 'application/x-www-form-urlencoded',
//data: info,
type CompanyRelations struct {
	ProKey      string        `param:"<in:formData> <name:pro_key> <required:required>  <nonzero:nonzero> <err:pro_key不能为空！>  <desc:产品描述>"`
	CompanyID   int64         `param:"<in:formData> <name:company_id> <required:required> <nonzero:nonzero>  <desc:采购商或者供应商公司id> "`
	CompanyType int           `param:"<in:formData> <name:company_type> <required:required>   <desc:0采购商 1供应商> "`
	TimeOut     time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
}

//Serve 处理方法
func (c *CompanyRelations) Serve(ctx *faygo.Context) error {
	var (
		CompanyRelationsCtx context.Context
		cancel              context.CancelFunc
		search              *elastic.SearchService
		query               *elastic.BoolQuery
		collapse            *elastic.CollapseBuilder
	)
	if c.TimeOut != 0 {
		CompanyRelationsCtx, cancel = context.WithTimeout(context.Background(), c.TimeOut*time.Second)
	} else {
		CompanyRelationsCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	//companyName待查
	relationship := model.Relationship{
		CompanyID:  c.CompanyID,
		ParentID:   0,
		ParentName: "",
	}
	client := constants.Instance()
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query = elastic.NewBoolQuery()
	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))
	proKey, _ := url.PathUnescape(c.ProKey)
	proKey = util.TrimFrontBack(proKey)
	if proKey != "All Product" {
		query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
	}
	if c.CompanyType == 0 {
		query = query.Must(elastic.NewTermQuery("PurchaserId", c.CompanyID))
		collapse = elastic.NewCollapseBuilder("SupplierId").
			InnerHit(elastic.NewInnerHit().Name("SupplierId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
	} else {
		query = query.Must(elastic.NewTermQuery("SupplierId", c.CompanyID))
		collapse = elastic.NewCollapseBuilder("PurchaserId").
			InnerHit(elastic.NewInnerHit().Name("PurchaserId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
	}
	query = query.Boost(10)
	query = query.DisableCoord(true)
	query = query.QueryName("filter")
	res, err := search.Query(query).Sort("FrankTime", false).Size(10).Collapse(collapse).Do(CompanyRelationsCtx)
	if err != nil {
		ctx.Log().Error(err)
	}
	//一级 采购
	if c.CompanyType == 0 {
		//查供应商 一级
		if len(res.Hits.Hits) > 0 {
			for i := 0; i < len(res.Hits.Hits); i++ {
				detail := res.Hits.Hits[i].Source
				var frank model.Frank
				jsonObject, _ := detail.MarshalJSON()
				jsoniter.Unmarshal(jsonObject, &frank)
				if relationship.CompanyName == "" {
					relationship.CompanyName = frank.Purchaser
				}
				relationship.Partner = append(relationship.Partner, model.Relationship{
					CompanyID:   frank.SupplierId,
					CompanyName: frank.Supplier,
					ParentID:    relationship.CompanyID,
					ParentName:  relationship.CompanyName,
					Partner:     nil,
				})
			}
		}
		//查采购商 二级 去掉反对应关系
		serviceTwo := client.Search().Index(constants.IndexName).Type(constants.TypeName).Sort("FrankTime", false).Size(5)
		collapse = elastic.NewCollapseBuilder("PurchaserId").
			InnerHit(elastic.NewInnerHit().Name("PurchaserId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
		for i := 0; i < len(relationship.Partner); i++ {
			query := elastic.NewBoolQuery()
			query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
			query = query.Must(elastic.NewTermQuery("SupplierId", relationship.Partner[i].CompanyID))
			query = query.QueryName("filter")
			res, err := serviceTwo.Query(query).Collapse(collapse).Do(CompanyRelationsCtx)
			if err != nil {
				ctx.Log().Error(err)
			}
			if len(res.Hits.Hits) > 0 {
				for q := 0; q < len(res.Hits.Hits); q++ {
					detail := res.Hits.Hits[q].Source
					jsonObject, _ := detail.MarshalJSON()
					var frank model.Frank
					jsoniter.Unmarshal(jsonObject, &frank)
					if frank.PurchaserId != relationship.Partner[i].ParentID {
						relationship.Partner[i].Partner = append(relationship.Partner[i].Partner,
							model.Relationship{
								CompanyID:   frank.PurchaserId,
								CompanyName: frank.Purchaser,
								ParentID:    relationship.Partner[i].CompanyID,
								ParentName:  relationship.Partner[i].CompanyName,
								Partner:     nil,
							})
					}
				}
			}
		}
		//查供应商  三级
		serviceThree := client.Search().Index(constants.IndexName).Type(constants.TypeName).Sort("FrankTime", false).Size(10)
		collapse = elastic.NewCollapseBuilder("SupplierId").
			InnerHit(elastic.NewInnerHit().Name("SupplierId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
		for i := 0; i < len(relationship.Partner); i++ {
			for j := 0; j < len(relationship.Partner[i].Partner); j++ {
				query := elastic.NewBoolQuery()
				query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
				query = query.Must(elastic.NewTermQuery("PurchaserId", relationship.Partner[i].Partner[j].CompanyID))
				query = query.QueryName("filter")
				res, err := serviceThree.Query(query).Collapse(collapse).Do(CompanyRelationsCtx)
				if err != nil {
					ctx.Log().Error(err)
				}
				if len(res.Hits.Hits) > 0 {
					for q := 0; q < len(res.Hits.Hits); q++ {
						detail := res.Hits.Hits[q].Source
						jsonObject, _ := detail.MarshalJSON()
						var frank model.Frank
						jsoniter.Unmarshal(jsonObject, &frank)
						if frank.SupplierId != relationship.Partner[i].Partner[j].ParentID {
							relationship.Partner[i].Partner[j].Partner = append(relationship.Partner[i].Partner[j].Partner,
								model.Relationship{
									CompanyID:   frank.SupplierId,
									CompanyName: frank.Supplier,
									ParentID:    relationship.Partner[i].Partner[j].CompanyID,
									ParentName:  relationship.Partner[i].Partner[j].CompanyName,
									Partner:     nil,
								})
						}
					}
				}
			}
		}

	} else {
		//查采购商 一级
		if len(res.Hits.Hits) > 0 {
			for i := 0; i < len(res.Hits.Hits); i++ {
				detail := res.Hits.Hits[i].Source
				var frank model.Frank
				jsonObject, _ := detail.MarshalJSON()
				jsoniter.Unmarshal(jsonObject, &frank)
				if relationship.CompanyName == "" {
					relationship.CompanyName = frank.Supplier
				}
				relationship.Partner = append(relationship.Partner, model.Relationship{
					CompanyID:   frank.PurchaserId,
					CompanyName: frank.Purchaser,
					ParentID:    relationship.CompanyID,
					ParentName:  relationship.CompanyName,
					Partner:     nil,
				})
			}
		}
		//查供应商 二级
		serviceTwo := client.Search().Index(constants.IndexName).Type(constants.TypeName).Sort("FrankTime", false).Size(5)
		collapse = elastic.NewCollapseBuilder("PurchaserId").
			InnerHit(elastic.NewInnerHit().Name("PurchaserId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
		for i := 0; i < len(relationship.Partner); i++ {
			query := elastic.NewBoolQuery()
			query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
			query = query.Must(elastic.NewTermQuery("SupplierId", relationship.Partner[i].CompanyID))
			query = query.QueryName("filter")
			res, err := serviceTwo.Query(query).Collapse(collapse).Do(CompanyRelationsCtx)
			if err != nil {
				ctx.Log().Error(err)
			}
			if len(res.Hits.Hits) > 0 {
				for q := 0; q < len(res.Hits.Hits); q++ {
					detail := res.Hits.Hits[q].Source
					jsonObject, _ := detail.MarshalJSON()
					var frank model.Frank
					jsoniter.Unmarshal(jsonObject, &frank)
					relationship.Partner[i].Partner = append(relationship.Partner[i].Partner,
						model.Relationship{
							CompanyID:   frank.SupplierId,
							CompanyName: frank.Supplier,
							ParentID:    relationship.Partner[i].CompanyID,
							ParentName:  relationship.Partner[i].CompanyName,
							Partner:     nil,
						})
				}
			}
		}
		//查供应商  三级
		serviceThree := client.Search().Index(constants.IndexName).Type(constants.TypeName).Sort("FrankTime", false).Size(10)
		collapse = elastic.NewCollapseBuilder("SupplierId").
			InnerHit(elastic.NewInnerHit().Name("SupplierId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
		for i := 0; i < len(relationship.Partner); i++ {
			for j := 0; j < len(relationship.Partner[i].Partner); j++ {
				query := elastic.NewBoolQuery()
				query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
				query = query.Must(elastic.NewTermQuery("PurchaserId", relationship.Partner[i].Partner[j].CompanyID))
				query = query.QueryName("filter")
				res, err := serviceThree.Query(query).Collapse(collapse).Do(CompanyRelationsCtx)
				if err != nil {
					ctx.Log().Error(err)
				}
				if len(res.Hits.Hits) > 0 {
					for q := 0; q < len(res.Hits.Hits); q++ {
						detail := res.Hits.Hits[q].Source
						jsonObject, _ := detail.MarshalJSON()
						var frank model.Frank
						jsoniter.Unmarshal(jsonObject, &frank)
						relationship.Partner[i].Partner[j].Partner = append(relationship.Partner[i].Partner[j].Partner,
							model.Relationship{
								CompanyID:   frank.PurchaserId,
								CompanyName: frank.Purchaser,
								ParentID:    relationship.Partner[i].Partner[j].CompanyID,
								ParentName:  relationship.Partner[i].Partner[j].CompanyName,
								Partner:     nil,
							})
					}
				}
			}
		}

	}
	result, err := jsoniter.Marshal(model.Response{
		List: relationship,
		Code: 0,
	})
	if err != nil {
		ctx.Log().Error(err)
	}
	return ctx.String(200, util.BytesString(result))
}

//DetailTrend ...
//findBusinessTrendInfo.php
type DetailTrend struct {
	ProKey         string        `param:"<in:formData> <name:pro_key> <required:required>  <nonzero:nonzero> <err:pro_key不能为空或者空字符串！>  <desc:产品描述>"`
	CompanyIDArray string        `param:"<in:formData> <name:company_ids> <required:required> <nonzero:nonzero>  <desc:采购商或者供应商公司id> "`
	CompanyType    int           `param:"<in:formData> <name:company_type> <required:required>   <desc:0采购商 1供应商> "`
	TimeOut        time.Duration `param:"<in:formData>  <name:time_out>  <desc:该接口的最大响应时间> "`
	DateType       int           `param:"<in:formData> <name:date_time> <required:required> <range: 0:1> <err:date_type必须在0到1之间> <desc:date_type 0月(最近一年) 1年(12到17年)> " `
	Vwtype         int           `param:"<in:formData> <name:vwtype>  <required:required> <range: 0:1> <err:vwtypee必须在0到1之间> <desc:vwtype 0volume 1weight>"`
}

//Serve 访问逻辑
func (detailTrend *DetailTrend) Serve(ctx *faygo.Context) error {
	var (
		detailTrendCtx context.Context
		cancel         context.CancelFunc
		search         *elastic.SearchService
		query          *elastic.BoolQuery
		agg            *elastic.SumAggregation
		dateAgg        *elastic.DateHistogramAggregation
		redisKey       string
	)
	if detailTrend.TimeOut != 0 {
		detailTrendCtx, cancel = context.WithTimeout(context.Background(), detailTrend.TimeOut*time.Second)
	} else {
		detailTrendCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query = elastic.NewBoolQuery()
	agg = elastic.NewSumAggregation()
	dateAgg = elastic.NewDateHistogramAggregation()
	ids := strings.Split(detailTrend.CompanyIDArray, ",")
	proKey, _ := url.PathUnescape(detailTrend.ProKey)
	proKey = util.TrimFrontBack(proKey)
	query = query.Filter(elastic.NewMatchQuery("ProDesc", proKey))
	var results []model.DetailInfo
	if detailTrend.Vwtype == 0 {
		agg.Field("OrderVolume")
	} else {
		agg.Field("OrderWeight")
	}
	if detailTrend.DateType == 0 {
		dateAgg.Field("FrankTime").Interval("month").Format("yyyy-MM-dd")
		var forCount int
		if len(ids) > 10 {
			forCount = 10
		}
		for Index := 0; Index < forCount; Index++ {
			query = elastic.NewBoolQuery()
			search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
			if proKey != "All Product" {
				query = query.Filter(elastic.NewMatchQuery("ProDesc", proKey))
			}
			if detailTrend.CompanyType == 0 {
				query = query.Filter(elastic.NewTermQuery("PurchaserId", ids[Index]))
			} else {
				query = query.Filter(elastic.NewTermQuery("SupplierId", ids[Index]))
			}
			companyId, _ := strconv.Atoi(ids[Index])
			result := model.DetailInfo{ID: companyId}
			// 最近一年
			if detailTrend.Vwtype == 0 {
				agg.Field("OrderVolume")
			} else {
				agg.Field("OrderWeight")
			}
			query = query.Filter(elastic.NewRangeQuery("FrankTime").From("now-1y").To("now"))
			dateAgg.SubAggregation("vwCount", agg)
			res, err := search.Query(query).Size(1).Aggregation("detailTrend", dateAgg).Do(detailTrendCtx)
			if err != nil {
				ctx.Log().Error(err)
			}
			var frank model.Frank
			if len(res.Hits.Hits) > 0 {
				detail := res.Hits.Hits[0].Source
				jsonObject, _ := detail.MarshalJSON()
				jsoniter.Unmarshal(jsonObject, &frank)
				if detailTrend.CompanyType == 0 {
					result.Name = frank.Purchaser
				} else {
					result.Name = frank.Supplier
				}
			} else {
				continue
			}
			aggregations := res.Aggregations
			terms, _ := aggregations.DateHistogram("detailTrend")
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
				result.Trends = append(result.Trends, detail)
			}
			results = append(results, result)

		}

	} else {
		dateAgg.Field("FrankTime").Interval("year").Format("yyyy-MM-dd")
		var forCount int
		if len(ids) > 10 {
			forCount = 10
		}
		for Index := 0; Index < forCount; Index++ {
			if detailTrend.CompanyType == 0 {
				query = query.Filter(elastic.NewTermQuery("PurchaserId", ids[Index]))
			} else {
				query = query.Filter(elastic.NewTermQuery("SupplierId", ids[Index]))
			}
			companyId, _ := strconv.Atoi(ids[Index])
			result := model.DetailInfo{ID: companyId}
			// 最近一年
			if detailTrend.Vwtype == 0 {
				agg.Field("OrderVolume")
			} else {
				agg.Field("OrderWeight")
			}
			dateAgg.SubAggregation("vwCount", agg)
			res, err := search.Query(query).Size(1).Aggregation("detailTrend", dateAgg).Do(detailTrendCtx)
			if err != nil {
				ctx.Log().Error(err)
			}
			var frank model.Frank
			if len(res.Hits.Hits) > 0 {
				detail := res.Hits.Hits[0].Source
				jsonObject, _ := detail.MarshalJSON()
				jsoniter.Unmarshal(jsonObject, &frank)
				if detailTrend.CompanyType == 0 {
					result.Name = frank.Purchaser
				} else {
					result.Name = frank.Supplier
				}
			} else {
				continue
			}
			aggregations := res.Aggregations
			terms, _ := aggregations.DateHistogram("detailTrend")
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
				result.Trends = append(result.Trends, detail)
			}
			results = append(results, result)
			query = elastic.NewBoolQuery()
			search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
			query = query.Filter(elastic.NewMatchQuery("ProDesc", proKey))
		}
	}
	json, err := jsoniter.Marshal(model.Response{
		Data: results,
	})
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

type GroupHistory struct {
	ProKey      string        `param:"<in:formData> <name:pro_key> <required:required>  <nonzero:nonzero>  <err:pro_key不能为空或者空字符串！>  <desc:产品描述>"`
	PageNo      int           `param:"<in:formData> <name:page_no> <required:required>  <nonzero:nonzero> <range: 1:1000>  <err:page_no必须在1到1000之间>  <desc:分页页码>"`
	PageSize    int           `param:"<in:formData> <name:page_size> <required:required>  <nonzero:nonzero> <err:page_size不能为空！>  <desc:分页的页数>"`
	CompanyType int           `param:"<in:formData> <name:company_type> <required:required>  <range: 0:2>  <err:company_type必须在0到2之间>  <desc:公司类型>"`
	CompanyID   int           `param:"<in:formData> <name:company_id> <required:required> <nonzero:nonzero>  <err:company_id不能为0>  <desc:公司类型>"`
	TimeOut     time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	Sort        int           `param:"<in:formData> <name:sort> <required:required>  <err:sort不能为空！>  <desc:排序的参数 1 2 3>"`
}

// GroupHistory ... 详情
//Nearly a year of trading history
//通过proKey相关的公司的近一年的交易记录 如果是采购上进来 先查prokey 再通过group supplier分组来处理
//传入 参数 proKey 公司ID 公司类型
func (param *GroupHistory) Serve(ctx *faygo.Context) error {
	var (
		GroupHistoryCtx context.Context
		cancel          context.CancelFunc
		redisKey        string
	)
	if param.TimeOut != 0 {
		GroupHistoryCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		GroupHistoryCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	GroupHistorySearch := client.Search().Index(constants.IndexName).Type(constants.TypeName)
	query := elastic.NewBoolQuery()
	highlight := elastic.NewHighlight()
	var cardinality *elastic.CardinalityAggregation
	query.Filter(elastic.NewRangeQuery("FrankTime").From("now-1y").To("now"))
	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))
	proKey, err := url.PathUnescape(param.ProKey)
	proKey = util.TrimFrontBack(proKey)
	if err != nil {
		ctx.Log().Error(err)
	}
	if proKey != "All Product" {
		query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
	}
	highlight.Field("ProDesc")
	if param.CompanyType == 0 {
		query = query.Must(elastic.NewTermQuery("PurchaserId", param.CompanyID))
		cardinality = elastic.NewCardinalityAggregation().Field("SupplierId")
	} else {
		query = query.Must(elastic.NewTermQuery("SupplierId", param.CompanyID))
		cardinality = elastic.NewCardinalityAggregation().Field("PurchaserId")
	}
	countSearch := client.Search().Index(constants.IndexName).Type(constants.TypeName)
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
					ctx.Log().Error(err)
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
	search := client.Search().Index(constants.IndexName).Type(constants.TypeName)
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
		if proKey != "" {
			queryDeatil = queryDeatil.Must(elastic.NewMatchQuery("ProDesc", proKey))
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
			franks[i].CompanyAddress = service.GetDidNameByDid(frank.PurchaserDistrictId1)
		} else {
			franks[i].CompanyName = frank.Purchaser
			franks[i].CompanyId = frank.PurchaserId
			franks[i].CompanyAddress = service.GetDidNameByDid(frank.SupplierDistrictId1)
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
	result, err := jsoniter.Marshal(model.Response{
		List:  franks,
		Total: int64(*resCardinality.Value),
		Code:  0,
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

//CompanyInfo 得到公司信息
type CompanyInfo struct {
	CompanyType int           `param:"<in:formData> <name:company_type> <required:required>  <range: 0:2>  <err:company_type必须在0到2之间>  <desc:公司类型>"`
	CompanyID   int           `param:"<in:formData> <name:company_id> <required:required> <nonzero:nonzero>  <err:company_id不能为0>  <desc:公司类型>"`
	TimeOut     time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
}

func (param *CompanyInfo) Serve(ctx *faygo.Context) error {
	if param.CompanyType == 0 {
		//service.GetBuyer(param.CompanyID)
		return ctx.JSON(200, model.Response{Data: service.GetBuyer(param.CompanyID)})
	} else {
		return ctx.JSON(200, model.Response{Data: service.GetSupplier(param.CompanyID)})
	}
}

//公司列表
type CompanyList struct {
	CompanyType int           `param:"<in:formData> <name:company_type> <required:required>  <range: 0:2>  <err:company_type必须在0到2之间>  <desc:公司类型>"`
	CompanyID   int           `param:"<in:formData> <name:company_id> <required:required> <nonzero:nonzero>  <err:company_id不能为0>  <desc:公司类型>"`
	TimeOut     time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
}

func (param *CompanyList) Serve(ctx *faygo.Context) error {
	if param.CompanyType == 0 {
		contacts := service.GetBuyerContacts(param.CompanyID)
		if len(*contacts) == 0 {
			//发送请求
			post, err := gor.Post(constants.Config.GlsUrl,
				&gor.Request_options{
					Json: map[string]string{
						"ietype":     "0",
						"shangJiaId": strconv.Itoa(param.CompanyID),
						//"shangJiaId": "896373",
						"date_type": "2",
					},
					Is_ajax: true,
				})
			if err != nil {
				ctx.Log().Error(err)
			}
			return ctx.String(200, post.String())
		}
		return ctx.JSON(200, model.Response{Data: service.GetBuyerContacts(param.CompanyID)})
	} else {
		contacts := service.GetSupplierContacts(param.CompanyID)
		if len(*contacts) == 0 {
			//发送请求
			post, err := gor.Post(constants.Config.GlsUrl,
				&gor.Request_options{
					Json: map[string]string{
						"ietype":     "1",
						"shangJiaId": strconv.Itoa(param.CompanyID),
						//"shangJiaId": "896373",
						"date_type": "2",
					},
					Is_ajax: true,
				})
			if err != nil {
				ctx.Log().Error(err)
			}
			return ctx.String(200, post.String())
		}
		return ctx.JSON(200, model.Response{Data: service.GetSupplierContacts(param.CompanyID)})

	}
}

//采购商或者供应商地图分布
type CompanyDistrict struct {
	CompanyIDArray string        `param:"<in:formData> <name:company_ids> <required:required> <nonzero:nonzero> <err:company_ids不能为空或空字符串> <desc:采购商或者供应商公司id> "`
	CompanyType    int           `param:"<in:formData> <name:company_type> <required:required>  <range: 0:2>  <err:company_type必须在0到2之间>  <desc:公司类型>"`
	TimeOut        time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
}

func (param *CompanyDistrict) Serve(ctx *faygo.Context) error {
	info := service.GetCompanyDistrictInfo(param.CompanyIDArray, param.CompanyType)
	result, err := jsoniter.Marshal(model.Response{
		List: info,
	})
	if err != nil {
		ctx.Log().Error(err)
	}
	return ctx.String(200, util.BytesString(result))
}

//得到公司联系人
type CompanyContacts struct {
	Guid        int `param:"<in:formData> <name:guid> <required:required> <nonzero:nonzero>  <err:guid不能为空!!>  <desc:公司类型>"`
	PageNo      int `param:"<in:formData> <name:page_no> <required:required>  <nonzero:nonzero> <range: 1:1000>  <err:page_no必须在1到1000之间>   <desc:分页页码>"`
	PageSize    int `param:"<in:formData> <name:page_size> <required:required>  <nonzero:nonzero> <err:page_size不能为空！>  <desc:分页的页数>"`
	CompanyID   int `param:"<in:formData> <name:company_id> <required:required> <nonzero:nonzero>  <err:company_id不能为0>  <desc:公司类型>"`
	CompanyType int `param:"<in:formData> <name:company_type> <required:required>  <range: 0:2>  <err:company_type必须在0到2之间>  <desc:公司类型>"`
}
