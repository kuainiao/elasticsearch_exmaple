package middleware

import (
	"fmt"
	"github.com/henrylee2cn/faygo"
	"github.com/henrylee2cn/faygo/ext/db/xorm"
	"github.com/zhangweilun/tradeweb/model"
	"time"
)

//DbQuery query the DataBase
var DbQuery = faygo.HandlerFunc(func(ctx *faygo.Context) error {

	db := xorm.MustDB("default")

	//db.Alias("c")
	//var param handler.DetailQuery
	//err := ctx.BindJSON(&param)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//if param.CompanyId == 0 || param.ProKey == "" || param.ProKey == `""` {
	//	response.Code = -1
	//	response.Error = "请求信息不能为0或者空"
	//} else {
	//	db.Where("c.company_id = ? ", param.CompanyId)
	//}

	_, err := db.Insert(&model.CompanyDetail{
		CompanyId:       0,
		CompanyType:     0,
		CompanyRelation: "{}",
		TopTen:          "{}",
		ComputerTime:    time.Now(),
	})

	fmt.Println(err)
	return nil
})
