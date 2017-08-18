package middleware

import (
	"github.com/henrylee2cn/faygo"
	"github.com/zhangweilun/tradeweb/constants"
	"strings"
)

/**
*
* @author willian
* @created 2017-08-17 14:08
* @email 18702515157@163.com
**/

//RedisCache redis缓存
var RedisCache = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	redis := constants.Redis()
	//判断请求方法和请求参数
	if ctx.Method() == "GET" {
		url := ctx.URI()
		if strings.Contains(url, "token") {
			index := strings.Index(url, "token")
			url = url[0 : index-1]
		}
		val, err := redis.Get("GET" + url).Result()
		if err == nil {
			ctx.Stop()
			return ctx.String(200, val)
		}
		ctx.SetData("redisKey", "GET"+url)
	}
	return nil
})
