package handler

import (
	"context"
	"github.com/henrylee2cn/faygo"
	"github.com/json-iterator/go"
	"github.com/zhangweilun/tradeweb/constants"
	"github.com/zhangweilun/tradeweb/model"
	"gopkg.in/olivere/elastic.v5"
	"strings"
	"time"
)

// CompanyRelations ... 详情
//公司关系图
//传入参数proKey 公司id 公司类型
//https://elasticsearch.cn/article/132
//	b := NewCollapseBuilder("user").
//InnerHit(NewInnerHit().Name("last_tweets").Size(5).Sort("date", true)).
//MaxConcurrentGroupRequests(4) 去重查询
var CompanyRelations = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	var (
		CompanyRelationsCtx context.Context
		cancel              context.CancelFunc
	)
	var param DetailQuery
	err := ctx.BindJSON(&param)
	if err != nil {
		ctx.Log().Error(err)
	}
	if param.TimeOut != 0 {
		CompanyRelationsCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		CompanyRelationsCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	//companyName待查
	relationship := model.Relationship{
		CompanyID:   param.CompanyId,
		CompanyName: param.CompanyName,
		ParentID:    0,
		ParentName:  "",
	}
	client := constants.Instance()
	CompanyRelationsSearch := client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	query = query.MustNot(elastic.NewTermQuery("Supplier", "UNAVAILABLE"))
	query = query.MustNot(elastic.NewTermQuery("Purchaser", "UNAVAILABLE"))
	var collapse *elastic.CollapseBuilder
	if param.ProKey != "" {
		query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
	}
	if param.CompanyType == 0 {
		query = query.Must(elastic.NewTermQuery("PurchaserId", param.CompanyId))
		collapse = elastic.NewCollapseBuilder("SupplierId").
			InnerHit(elastic.NewInnerHit().Name("SupplierId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
	} else {
		query = query.Must(elastic.NewTermQuery("SupplierId", param.CompanyId))
		collapse = elastic.NewCollapseBuilder("PurchaserId").
			InnerHit(elastic.NewInnerHit().Name("PurchaserId").Size(0).Sort("FrankTime", false)).
			MaxConcurrentGroupRequests(4)
	}
	query = query.Boost(10)
	query = query.DisableCoord(true)
	query = query.QueryName("filter")
	res, err := CompanyRelationsSearch.Query(query).Sort("FrankTime", false).Size(10).Collapse(collapse).Do(CompanyRelationsCtx)
	if err != nil {
		ctx.Log().Error(err)

	}
	//一级 采购
	if param.CompanyType == 0 {
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
			query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
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
				query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
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
			query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
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
				query = query.Must(elastic.NewMatchQuery("ProDesc", strings.ToLower(param.ProKey)))
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
	return ctx.JSON(200, model.Response{
		List: relationship,
		Code: 0,
	})
})