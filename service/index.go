package service

import (
	"github.com/henrylee2cn/faygo/ext/db/xorm"
	"github.com/zhangweilun/tradeweb/model"
	"fmt"
)

/**
*
* @author willian
* @created 2017-08-18 17:48
* @email 18702515157@163.com
**/
var db = xorm.MustDB("default")

var db_usa = xorm.MustDB("usa")

func GetAllCountry() {

}

func MoveFrank(startTime string, endTime string) []model.Frankly2015 {
	var franks []model.Frankly2015
	total, err := db_usa.Table("frankly_oredr_new").Alias("f").Where("f.frankly_time >?", "2015-01-01").And("f.frankly_time <?", "2015-12-30").Count()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(total)
	allPage := total / 1000
	var i int64
	for ; i < allPage; i++ {
		db_usa.Table("frankly_oredr_new").Alias("f").Select("f.*").Limit(int(i), 100).
			Where("f.frankly_time >?", startTime).
			And("f.frankly_time <?", endTime).Find(franks)
		if db.SupportInsertMany() {
			db.Insert(franks)
		} else {
			db.Insert(franks)
		}
	}
	return nil
}
