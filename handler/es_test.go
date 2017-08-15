package handler

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/json-iterator/go"
	"github.com/zhangweilun/tradeweb/model"
	"gopkg.in/olivere/elastic.v5"
)

/**
*
* @author willian
* @created 2017-08-14 09:54
* @email 18702515157@163.com
**/

func TestCollaps(t *testing.T) {

	client, _ := elastic.NewClient(
		elastic.SetURL("http://"+"es.g2l-service.com"),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
		elastic.SetInfoLog(log.New(os.Stdout, "", log.LstdFlags)),
		elastic.SetTraceLog(log.New(os.Stderr, "[[ELASTIC]]", 0)),
		elastic.SetBasicAuth("admin", "4Dm1n.3s"),
		elastic.SetSniff(false),
	)

	CompanyRelationsSearch := client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("ProDesc", "bike"))
	query = query.Must(elastic.NewTermQuery("PurchaserId", 305971))
	b := elastic.NewCollapseBuilder("SupplierId").
		InnerHit(elastic.NewInnerHit().Name("SupplierId").Size(0).Sort("FrankTime", false)).
		MaxConcurrentGroupRequests(4)
	res, _ := CompanyRelationsSearch.Query(query).Sort("FrankTime", false).Size(10).Collapse(b).Do(context.Background())
	hits := res.Hits.Hits
	for i := 0; i < len(hits); i++ {
		innerHits := hits[i].InnerHits["SupplierId"]
		for j := 0; j < len(innerHits.Hits.Hits); j++ {
			detail := innerHits.Hits.Hits[j].Source
			jsonObject, _ := detail.MarshalJSON()
			var frank model.Frank
			jsoniter.Unmarshal(jsonObject, &frank)
			fmt.Println(frank)
		}

		//for k ,v := range hits[i].InnerHits{
		//	fmt.Println("k=",k)
		//	fmt.Println("v=",v)
		//}
	}
	fmt.Println(res.Hits.Hits)
}
