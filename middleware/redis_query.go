package middleware

import (
	"sort"
	"strings"

	"github.com/henrylee2cn/faygo"
	"github.com/zhangweilun/tradeweb/constants"

	"strconv"
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
	} else if ctx.Method() == "POST" {
		url := ctx.URI() + "?"
		var ids []int
		var idArray string
		params := ctx.FormParamAll()
		if v, ok := params["company_ids"]; ok {
			value := v[0]
			split := strings.Split(value, ",")
			for i := 0; i < len(split); i++ {
				atoi, err := strconv.Atoi(split[i])
				if err != nil {
					ctx.Log().Error(err)
				}
				ids = append(ids, atoi)
			}
		}
		var queryKey []string
		for k := range params {
			queryKey = append(queryKey, k)
		}
		sort.Strings(queryKey)
		for index := 0; index < len(queryKey); index++ {
			for k, v := range params {
				if queryKey[index] == k {
					if k != "token" && k != "userId" {
						if k == "company_ids" {
							for i := 0; i < len(ids); i++ {
								idArray = idArray + strconv.Itoa(ids[i])
							}
							url = url + queryKey[index] + "=" + idArray + "&"
						} else {
							url = url + queryKey[index] + "=" + v[0] + "&"
						}

					}
				}
			}
		}
		val, err := redis.Get("POST"+url[0:len(url)-1]).Result()
		if err == nil {
			ctx.Log().Info(err)
			ctx.Stop()
			return ctx.String(200, val)
		}
		ctx.SetData("redisKey", "POST"+url[0:len(url)-1])
	}
	return nil
})
