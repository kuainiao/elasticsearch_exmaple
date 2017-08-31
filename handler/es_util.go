package handler

import elastic "gopkg.in/olivere/elastic.v5"

//date_type 0all year 1the last six month 2the last one year 3the last three year 时间过滤
func dataType(q *elastic.BoolQuery, dataType int) {
	if dataType == 1 {
		q = q.Filter(elastic.NewRangeQuery("FrankTime").From("now-6M").To("now"))
	} else if dataType == 2 {
		q = q.Filter(elastic.NewRangeQuery("FrankTime").From("now-1y").To("now"))
	} else if dataType == 3 {
		q = q.Filter(elastic.NewRangeQuery("FrankTime").From("now-3y").To("now"))
	}
}

func district(q *elastic.BoolQuery, districtId int, districtLevel int, ietype int) {
	var ids []string
	if ietype == 0 {
		ids = []string{"PurchaserDistrictId1", "PurchaserDistrictId2", "PurchaserDistrictId3"}
	} else {
		ids = []string{"SupplierDistrictId1", "SupplierDistrictId2", "SupplierDistrictId3"}
	}
	if districtLevel == 1 {
		q = q.Must(elastic.NewTermQuery(ids[0], districtId))
	} else if districtLevel == 2 {
		q = q.Must(elastic.NewTermQuery(ids[1], districtId))
	} else if districtLevel == 3 {
		q = q.Must(elastic.NewTermQuery(ids[2], districtId))
	}
}

func vwType(a *elastic.SumAggregation, vwtype int) {
	if vwtype == 0 {
		a = a.Field("OrderVolume")
	} else {
		a = a.Field("OrderWeight")
	}
}

func categoryFilter(q *elastic.BoolQuery, categoryId int) {
	q = q.Must(elastic.NewTermQuery("CategoryId", categoryId))
}

func productFilter(q *elastic.BoolQuery, categoryId int) {
	q = q.Must(elastic.NewTermQuery("ProductId", categoryId))

}
