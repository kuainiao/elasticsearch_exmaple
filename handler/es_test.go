package handler

import (
	"testing"
	"github.com/zhangweilun/tradeweb/constants"
	"context"
	"gopkg.in/olivere/elastic.v5"
	"fmt"
)

/**
* 
* @author willian
* @created 2017-08-14 09:54
* @email 18702515157@163.com  
**/

func TestCollaps(t *testing.T) {
	client := constants.Instance()
	CompanyRelationsSearch := client.Search().Index("trade").Type("frank")
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("ProDesc", "bike"))
	b := elastic.NewCollapseBuilder("SupplierId").
		InnerHit(elastic.NewInnerHit().Name("SupplierId").Size(10).Sort("FrankTime", false)).
		MaxConcurrentGroupRequests(4)
	res, _ := CompanyRelationsSearch.Query(query).Collapse(b).Do(context.Background())
	fmt.Println(res)
}