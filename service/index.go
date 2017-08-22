package service

import (
	"fmt"
	"github.com/henrylee2cn/faygo/ext/db/xorm"
	"github.com/zhangweilun/tradeweb/model"
)

/**
*
* @author willian
* @created 2017-08-18 17:48
* @email 18702515157@163.com
**/
var db = xorm.MustDB("default")

func GetAllCountry() {

}

func MoveFrank(startTime string, endTime string) []model.FranklyFive {
	var franks []model.FranklyFive
	total, err := db.Table("frankly_oredr_new").Alias("f").Where("f.frankly_time >?", "2015-01-01").And("f.frankly_time <?", "2015-12-30").Count()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(total)
	allPage := 130000
	for i := 0; i < allPage; i++ {
		db.Table("frankly_oredr_new").Alias("f").Select("f.*").Limit(100, 100*(i)).
			Where("f.frankly_time >?", "2015-01-01").
			And("f.frankly_time <?", "2015-12-30").Find(&franks)
		fmt.Println(len(franks))
		if db.SupportInsertMany() {
			db.Insert(&franks)
		} else {
			db.Insert(&franks)
		}
	}
	return nil
}
