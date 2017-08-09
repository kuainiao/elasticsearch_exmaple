package middleware

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/henrylee2cn/faygo"
	"github.com/json-iterator/go"
	"github.com/zhangweilun/tradeweb/util"
	"strings"

)

/**
*
* @author willian
* @created 2017-08-07 09:35
* @email 18702515157@163.com
**/

var Sign = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	path := ctx.Path()
	secretId := "zhangwewilun"
	var secretKey = []byte("Gu5t9xGARNpq86cd98joQYCN3Cozk1qA")
	var content = []byte(ctx.Method() + path + "?id=" + secretId)
	fmt.Println(string(content))
	hash := hmac.New(sha1.New, secretKey)
	hash.Write(content)
	encodeString := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	fmt.Println(encodeString)
	sign := ctx.QueryParam("sign")
	if sign == "" {
		result := make(map[string]string)
		result["msg"] = "接口鉴权失败!!"
		json, err := jsoniter.Marshal(result)
		if err != nil {
			fmt.Println(err)
		}
		ctx.Stop()
		return ctx.JSON(200, util.BytesString(json))
	}
	if strings.Contains(sign, "%2B") {
		sign = strings.Replace(sign, "%2B", "+", -1)
	}
	if strings.Contains(sign, "%26") {
		sign = strings.Replace(sign, "%26", "&", -1)
	}
	fmt.Println(sign)
	if sign != encodeString {
		result := make(map[string]string)
		result["msg"] = "接口鉴权失败!!"
		json, err := jsoniter.Marshal(result)
		if err != nil {
			fmt.Println(err)
		}
		ctx.Stop()
		return ctx.JSON(200, util.BytesString(json))
	}
	return nil
})
