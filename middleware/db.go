package middleware

import (
	"time"

	"github.com/henrylee2cn/faygo"
	"github.com/henrylee2cn/faygo/ext/db/xorm"
	"github.com/zhangweilun/tradeweb/model"
)

//DbQuery query the DataBase
var DbQuery = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	db := xorm.MustDB("default")
	_, err := db.Insert(&model.CompanyDetail{
		CompanyId:       0,
		CompanyType:     0,
		CompanyRelation: "{}",
		TopTen:          "{}",
		ComputerTime:    time.Now(),
	})
	if err != nil {
		ctx.Log().Error(err)
	}
	return nil
})
