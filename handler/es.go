package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/henrylee2cn/faygo"
	"github.com/zhangweilun/tradeweb/constants"
	elastic "gopkg.in/olivere/elastic.v5"
	"github.com/zhangweilun/tradeweb/model"
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
	Purchase        string `json:"purchase"`
	Supplier        string `json:"supplier"`
	OriginalCountry string `json:"original_country"`
	DestinationPort string `json:"destination_port"`
}

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
	if param.ProDesc != "" {
		query = query.Must(elastic.NewTermQuery("ProDesc", param.ProDesc))
	}
	//query = query.MustNot(elastic.NewRangeQuery("OrderVolume").From(0.01).To(0.05))
	//query = query.Filter(NewTermqueryuery("account", "1"))
	//query = query.Should(elastic.NewTermQuery("OriginalCountry", "CN"), elastic.NewTermQuery("CompanyName", "a"))
	query = query.Boost(10)
	query = query.DisableCoord(true)
	query = query.QueryName("frankDetail")
	search = search.Query(query).From(0).Size(10)
	//search.Sort() //排序
	res, _ := search.Do(context)
	var franks []model.Frank
	for i := 0; i < len(res.Hits.Hits); i++ {
		detail := res.Hits.Hits[i].Source
		var frank model.Frank
		jsonObject, _ := detail.MarshalJSON()
		json.Unmarshal(jsonObject,&frank)
		//franks[i] = frank
		franks= append(franks, frank)
	}
	response := model.Response{
		List:     franks,
	}
	return ctx.JSON(0, response)
})
