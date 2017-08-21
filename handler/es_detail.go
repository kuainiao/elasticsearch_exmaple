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
	"github.com/zhangweilun/tradeweb/model"
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
	search = client.Search().Index("trade").Type("frank")
	query = elastic.NewBoolQuery()
	query = query.MustNot(elastic.NewMatchQuery("Supplier", "UNAVAILABLE"), elastic.NewMatchQuery("Purchaser", "UNAVAILABLE"))

	proKey, _ := url.PathUnescape(c.ProKey)
	query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(proKey)))
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
		serviceTwo := client.Search().Index("trade").Type("frank").Sort("FrankTime", false).Size(5)
		collapse = elastic.NewCollapseBuilder("PurchaserId").
			InnerHit(elastic.NewInnerHit().Name("PurchaserId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
		for i := 0; i < len(relationship.Partner); i++ {
			query := elastic.NewBoolQuery()
			query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(c.ProKey)))
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
		serviceThree := client.Search().Index("trade").Type("frank").Sort("FrankTime", false).Size(10)
		collapse = elastic.NewCollapseBuilder("SupplierId").
			InnerHit(elastic.NewInnerHit().Name("SupplierId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
		for i := 0; i < len(relationship.Partner); i++ {
			for j := 0; j < len(relationship.Partner[i].Partner); j++ {
				query := elastic.NewBoolQuery()
				query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(c.ProKey)))
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
		serviceTwo := client.Search().Index("trade").Type("frank").Sort("FrankTime", false).Size(5)
		collapse = elastic.NewCollapseBuilder("PurchaserId").
			InnerHit(elastic.NewInnerHit().Name("PurchaserId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
		for i := 0; i < len(relationship.Partner); i++ {
			query := elastic.NewBoolQuery()
			query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(c.ProKey)))
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
		serviceThree := client.Search().Index("trade").Type("frank").Sort("FrankTime", false).Size(10)
		collapse = elastic.NewCollapseBuilder("SupplierId").
			InnerHit(elastic.NewInnerHit().Name("SupplierId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
		for i := 0; i < len(relationship.Partner); i++ {
			for j := 0; j < len(relationship.Partner[i].Partner); j++ {
				query := elastic.NewBoolQuery()
				query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(c.ProKey)))
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
	ctx.Log().Error(err)
	return ctx.String(200, util.BytesString(result))
}

//DetailTrend ...
//findBusinessTrendInfo.php
type DetailTrend struct {
	ProKey         string        `param:"<in:formData> <name:pro_key> <required:required>  <nonzero:nonzero> <err:pro_key不能为空或者空字符串！>  <desc:产品描述>"`
	CompanyIDArray string        `param:"<in:formData> <name:company_ids> <required:required> <nonzero:nonzero>  <desc:采购商或者供应商公司id> "`
	CompanyType    int           `param:"<in:formData> <name:company_type> <required:required>   <desc:0采购商 1供应商> "`
	TimeOut        time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	DateType       int           `param:"<in:formData> <name:date_type> <required:required>  <desc:date_type 0月(最近一年) 1年(12到17年)>"`
	Vwtype         int           `param:"<in:formData> <name:vwtype>  <required:required>  <desc:vwtype 0volume 1weight>"`
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
	search = client.Search().Index("trade").Type("frank")
	search.Size(1)
	query = elastic.NewBoolQuery()
	agg = elastic.NewSumAggregation()
	dateAgg = elastic.NewDateHistogramAggregation()
	ids := strings.Split(detailTrend.CompanyIDArray, ",")
	query = query.Must(elastic.NewMatchQuery("ProDesc", detailTrend.ProKey))
	var results []model.DetailInfo
	if detailTrend.Vwtype == 0 {
		agg.Field("OrderVolume")
	} else {
		agg.Field("OrderWeight")
	}
	if detailTrend.DateType == 0 {
		dateAgg.Field("FrankTime").Interval("month").Format("yyyy-MM-dd")
		for Index := 0; Index < len(ids); Index++ {
			if detailTrend.CompanyType == 0 {
				query = query.Must(elastic.NewTermQuery("PurchaserId", ids[Index]))
			} else {
				query = query.Must(elastic.NewTermQuery("SupplierId", ids[Index]))
			}
			companyId, _ := strconv.Atoi(ids[Index])
			result := model.DetailInfo{ID: companyId}
			// 最近一年
			query = query.Filter(elastic.NewRangeQuery("FrankTime").From("now-1y").To("now"))
			dateAgg.SubAggregation("vwCount", agg)
			res, err := search.Query(query).Aggregation("detailTrend", dateAgg).Do(detailTrendCtx)
			if err != nil {
				ctx.Log().Error(err)
			}
			var frank model.Frank
			if len(res.Hits.Hits) > 0 {
				detail := res.Hits.Hits[0].Source
				jsonObject, _ := detail.MarshalJSON()
				jsoniter.Unmarshal(jsonObject, &frank)
				result.Name = frank.Purchaser
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
							log.Println(err)
						}

						detail.Value = util.Round(volume, 2)
					}
				}
				result.Trends = append(result.Trends, detail)
			}
			results = append(results, result)
			query = elastic.NewBoolQuery()
			search = client.Search().Index("trade").Type("frank")
			query = query.Must(elastic.NewMatchQuery("ProDesc", detailTrend.ProKey))
		}

	} else {
		dateAgg.Field("FrankTime").Interval("year").Format("yyyy-MM-dd")
		for Index := 0; Index < len(ids); Index++ {
			if detailTrend.CompanyType == 0 {
				query = query.Must(elastic.NewTermQuery("PurchaserId", ids[Index]))
			} else {
				query = query.Must(elastic.NewTermQuery("SupplierId", ids[Index]))
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
			res, err := search.Query(query).Aggregation("detailTrend", dateAgg).Do(detailTrendCtx)
			if err != nil {
				ctx.Log().Error(err)
			}
			var frank model.Frank
			if len(res.Hits.Hits) > 0 {
				detail := res.Hits.Hits[0].Source
				jsonObject, _ := detail.MarshalJSON()
				jsoniter.Unmarshal(jsonObject, &frank)
				result.Name = frank.Supplier
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
							log.Println(err)
						}

						detail.Value = util.Round(volume, 2)
					}
				}
				result.Trends = append(result.Trends, detail)
			}
			results = append(results, result)
			query = elastic.NewBoolQuery()
			search = client.Search().Index("trade").Type("frank")
			query = query.Must(elastic.NewMatchQuery("ProDesc", detailTrend.ProKey))
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
