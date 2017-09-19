package handler

import (
	"context"
	"github.com/henrylee2cn/faygo"
	"github.com/json-iterator/go"
	"github.com/zhangweilun/tradeweb/constants"
	"github.com/zhangweilun/tradeweb/model"
	"github.com/zhangweilun/tradeweb/service"
	"github.com/zhangweilun/tradeweb/util"
	"gopkg.in/olivere/elastic.v5"
	"strconv"
	"strings"
	"time"
)

/**
*
* @author willian
* @created 2017-09-05 15:29
* @email 18702515157@163.com
**/

type CountryRank struct {
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	Pid           int           `param:"<in:formData> <name:pid> <required:required>  <err:pid不能为空！>  <desc:排序的参数 1 2 3>"`
	DateType      int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:排序的参数 1 2 3>"`
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:0采购商 1供应商> "`
	IeType        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	VwType        int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0采购商 1供应商> "`
}

func (param *CountryRank) Serve(ctx *faygo.Context) error {
	var (
		searchCtx context.Context
		cancel    context.CancelFunc
		search    *elastic.SearchService
		query     *elastic.BoolQuery
		agg       *elastic.TermsAggregation
		sum       *elastic.SumAggregation
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
	agg = elastic.NewTermsAggregation()
	if param.IeType == 0 {
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
	sum = elastic.NewSumAggregation()
	query = elastic.NewBoolQuery()
	dataType(query, param.DateType)
	district(query, param.DistrictID, param.DistrictLevel, param.IeType)
	if param.Pid != 0 {
		productFilter(query, param.Pid)
	}
	if param.VwType == 0 {
		sum.Field("OrderVolume")
	} else {
		sum.Field("OrderWeight")
	}
	agg.SubAggregation("vwCount", sum).OrderByAggregation("vwCount", false)
	res, err := search.Query(query).Aggregation("CountryRank", agg).Size(0).RequestCache(true).Do(searchCtx)
	if err != nil {
		ctx.Log().Error(err)
	}
	aggregations := res.Aggregations
	terms, _ := aggregations.Terms("CountryRank")
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
					ctx.Log().Error(err)
				}
				category.Value = util.Round(volume, 2)
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

func (param *CountryRank) Doc() faygo.Doc {
	return faygo.Doc{
		// API接口说明
		Note: "国家进出口某产品国家排行，产品页左一图",
		// 响应说明或示例
		Return: "返回json",
	}
}

//ProductCompanyTop 产品页下面第三个表
type ProductCompanyTop struct {
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	Pid           int           `param:"<in:formData> <name:pid> <required:required>  <err:pid不能为空！>  <desc:排序的参数 1 2 3>"`
	DateType      int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:排序的参数 1 2 3>"`
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:0采购商 1供应商> "`
	IeType        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	VwType        int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0采购商 1供应商> "`
}

func (param *ProductCompanyTop) Serve(ctx *faygo.Context) error {
	var (
		searchCtx context.Context
		cancel    context.CancelFunc
		redisKey  string
	)
	if param.TimeOut != 0 {
		searchCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		searchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	tenCompanyId := *GetTopTenCompanyId(param.Pid, param.VwType, param.DistrictLevel, param.IeType, param.DistrictID, param.DateType, searchCtx)
	result, err := jsoniter.Marshal(model.Response{
		List: tenCompanyId,
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

func (param *ProductCompanyTop) Doc() faygo.Doc {
	return faygo.Doc{
		// API接口说明
		Note: "某产品采购商或者供应商的卖出或者买入总和，产品页面下面第三个图",
		// 响应说明或示例
		Return: "返回json",
	}
}

//ProductCompanyTrend 右边第二个图
type ProductCompanyTrend struct {
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	Pid           int           `param:"<in:formData> <name:pid> <required:required>  <err:pid不能为空！>  <desc:排序的参数 1 2 3>"`
	DateType      int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:排序的参数 1 2 3>"`
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:0采购商 1供应商> "`
	IeType        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	VwType        int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0采购商 1供应商> "`
}

func (param *ProductCompanyTrend) Serve(ctx *faygo.Context) error {
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
	var results []model.DetailInfo
	if param.VwType == 0 {
		agg.Field("OrderVolume")
	} else {
		agg.Field("OrderWeight")
	}
	dateAgg.SubAggregation("vwCount", agg)
	tenCompanyId := *GetTopTenCompanyId(param.Pid, param.VwType,
		param.DistrictLevel, param.IeType, param.DistrictID, param.DateType, searchCtx)
	for i := 0; i < len(tenCompanyId); i++ {
		query = elastic.NewBoolQuery()
		district(query, param.DistrictID, param.DistrictLevel, param.IeType)
		search = client.Search().Index(constants.IndexName).Type(constants.TypeName)
		dataType(query, param.DateType)
		if param.IeType == 0 {
			query.Filter(elastic.NewTermQuery("PurchaserId", tenCompanyId[i].ID))
		} else {
			query.Filter(elastic.NewTermQuery("SupplierId", tenCompanyId[i].ID))
		}
		res, err := search.Query(query).Aggregation("indexTime", dateAgg).Do(searchCtx)
		if err != nil {
			ctx.Log().Error(err)
		}
		category := model.DetailInfo{
			ID:   int(tenCompanyId[i].ID),
			Name: tenCompanyId[i].Name,
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
			category.Trends = append(category.Trends, detail)
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

func (param *ProductCompanyTrend) Doc() faygo.Doc {
	return faygo.Doc{
		// API接口说明
		Note: "某产品top10公司的按月份的进出口趋势",
		// 响应说明或示例
		Return: "返回json",
	}
}

type ProductTrend struct {
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	Pid           int           `param:"<in:formData> <name:pid> <required:required>  <err:pid不能为空！>  <desc:排序的参数 1 2 3>"`
	DateType      int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:排序的参数 1 2 3>"`
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:0采购商 1供应商> "`
	IeType        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	VwType        int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0采购商 1供应商> "`
}

func (param *ProductTrend) Serve(ctx *faygo.Context) error {
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
	dateAgg = elastic.NewDateHistogramAggregation()
	dateAgg.Field("FrankTime").Interval("month").Format("yyyy-MM-dd")
	var results []model.DetailTrand
	if param.VwType == 0 {
		agg.Field("OrderVolume")
	} else {
		agg.Field("OrderWeight")
	}
	dateAgg.SubAggregation("vwCount", agg)
	query = elastic.NewBoolQuery()
	district(query, param.DistrictID, param.DistrictLevel, param.IeType)
	dataType(query, param.DateType)
	res, err := search.Query(query).Aggregation("indexTime", dateAgg).Do(searchCtx)
	if err != nil {
		ctx.Log().Error(err)
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
		results = append(results, detail)
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

func (param *ProductTrend) Doc() faygo.Doc {
	return faygo.Doc{
		// API接口说明
		Note: "公司产品按月份进出口数据",
		// 响应说明或示例
		Return: "返回json",
	}
}

type RegionTop struct {
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	Pid           int           `param:"<in:formData> <name:pid> <required:required>  <err:pid不能为空！>  <desc:排序的参数 1 2 3>"`
	DateType      int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:排序的参数 1 2 3>"`
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:0采购商 1供应商> "`
	IeType        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	VwType        int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0采购商 1供应商> "`
}

func (param *RegionTop) Serve(ctx *faygo.Context) error {
	var (
		searchCtx context.Context
		cancel    context.CancelFunc
		redisKey  string
	)
	if param.TimeOut != 0 {
		searchCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		searchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	regionTop := GetRegionTop(param.Pid, param.VwType, param.DistrictLevel, param.IeType, param.DistrictID, param.DateType, searchCtx)
	result, err := jsoniter.Marshal(model.Response{
		List: regionTop,
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

func (param *RegionTop) Doc() faygo.Doc {
	return faygo.Doc{
		// API接口说明
		Note: "产品地区页面top5地区",
		// 响应说明或示例
		Return: "返回json",
	}
}

type RegionTrend struct {
	TimeOut       time.Duration `param:"<in:formData>  <name:time_out> <desc:该接口的最大响应时间> "`
	Pid           int           `param:"<in:formData> <name:pid> <required:required>  <err:pid不能为空！>  <desc:排序的参数 1 2 3>"`
	DateType      int           `param:"<in:formData> <name:date_type> <required:required>  <err:date_type不能为空！>  <desc:排序的参数 1 2 3>"`
	DistrictID    int           `param:"<in:formData> <name:did> <required:required> <err:did不能为空！>  <desc:0采购商 1供应商> "`
	DistrictLevel int           `param:"<in:formData> <name:dlevel> <required:required> <err:dlevel不能为空！>  <desc:0采购商 1供应商> "`
	IeType        int           `param:"<in:formData> <name:ietype> <required:required> <err:ietype不能为空！>  <desc:0采购商 1供应商> "`
	VwType        int           `param:"<in:formData> <name:vwtype> <required:required> <err:vwType不能为空！>  <desc:0采购商 1供应商> "`
}

func (param *RegionTrend) Serve(ctx *faygo.Context) error {
	var (
		searchCtx context.Context
		cancel    context.CancelFunc
		redisKey  string
		query     *elastic.BoolQuery
		agg       *elastic.SumAggregation
		dateAgg   *elastic.DateHistogramAggregation
	)
	if param.TimeOut != 0 {
		searchCtx, cancel = context.WithTimeout(context.Background(), param.TimeOut*time.Second)
	} else {
		searchCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	client := constants.Instance()
	regionTop := *GetRegionTop(param.Pid, param.VwType, param.DistrictLevel, param.IeType, param.DistrictID, param.DateType, searchCtx)
	agg = elastic.NewSumAggregation()
	dateAgg = elastic.NewDateHistogramAggregation()
	dateAgg.Field("FrankTime").Interval("month").Format("yyyy-MM-dd")
	var results []model.DetailInfo
	if param.VwType == 0 {
		agg.Field("OrderVolume")
	} else {
		agg.Field("OrderWeight")
	}
	dateAgg.SubAggregation("vwCount", agg)
	for i := 0; i < len(regionTop); i++ {
		query = elastic.NewBoolQuery()
		district(query, param.DistrictID, param.DistrictLevel, param.IeType)
		search := client.Search().Index(constants.IndexName).Type(constants.TypeName)
		dataType(query, param.DateType)
		if param.IeType == 0 {
			if param.DistrictLevel == 0 {
				query.Filter(elastic.NewTermQuery("PurchaserDistrictId1", regionTop[i].Did))
			} else if param.DistrictLevel == 1 {
				query.Filter(elastic.NewTermQuery("PurchaserDistrictId2", regionTop[i].Did))
			} else if param.DistrictLevel == 2 {
				query.Filter(elastic.NewTermQuery("PurchaserDistrictId3", regionTop[i].Did))
			}
		} else {
			if param.DistrictLevel == 0 {
				query.Filter(elastic.NewTermQuery("SupplierDistrictId1", regionTop[i].Did))
			} else if param.DistrictLevel == 1 {
				query.Filter(elastic.NewTermQuery("SupplierDistrictId2", regionTop[i].Did))
			} else if param.DistrictLevel == 2 {
				query.Filter(elastic.NewTermQuery("SupplierDistrictId3", regionTop[i].Did))
			}
		}
		res, err := search.Query(query).Aggregation("indexTime", dateAgg).Do(searchCtx)
		if err != nil {
			ctx.Log().Error(err)
		}
		category := model.DetailInfo{
			ID:   int(regionTop[i].Did),
			Name: regionTop[i].Dname,
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
			category.Trends = append(category.Trends, detail)
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

func (param *RegionTrend) Doc() faygo.Doc {
	return faygo.Doc{
		// API接口说明
		Note: "产品地区页面top5地区量总和",
		// 响应说明或示例
		Return: "返回json",
	}
}
