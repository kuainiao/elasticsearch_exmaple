package handler

import (
	"github.com/henrylee2cn/faygo"
	"github.com/dgrijalva/jwt-go"
	"time"
	"crypto/md5"
	"encoding/hex"
	"github.com/json-iterator/go"
	"github.com/zhangweilun/tradeweb/constants"
	"github.com/zhangweilun/tradeweb/model"
	"github.com/zhangweilun/tradeweb/service"
)

/**
*
* @author willian
* @created 2017-08-25 14:31
* @email 18702515157@163.com
**/

type Login struct {
	Username string `json:"username" param:"<in:formData> <name:userName> <required:required> <nonzero:nonzero>  <err:userName不能为空!!>  <desc:公司类型>"`
	Password string `json:"password" param:"<in:formData> <name:passwd> <required:required>  <nonzero:nonzero> <err:passwd不能为空!!>  <desc:公司类型>"`
}

func (param *Login) Serve(ctx *faygo.Context) error {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(param.Password))
	cipherStr := md5Ctx.Sum(nil)
	users := model.Users{
		UserName: param.Username,
		Passwd:   hex.EncodeToString(cipherStr),
	}
	count, err := service.GetUserByCondition(&users)
	if err != nil {
		ctx.Log().Error(err)
	}
	if count == 0 {
		return ctx.JSON(200, model.Response{Error: "用户名或密码错误！"})
	} else {
		token := jwt.New(jwt.SigningMethodHS256)
		claims := make(jwt.MapClaims)
		claims["exp"] = time.Now().Add(time.Minute * time.Duration(1)).Unix()
		claims["iat"] = time.Now().Unix()
		token.Claims = claims
		tokenString, err := token.SignedString([]byte(constants.Secret))
		if err != nil {
			ctx.String(500, "Error while sign")
		}
		result := make(map[string]string)
		result["code"] = "0"
		result["auth"] = tokenString
		result["username"] = param.Username
		marshal, err := jsoniter.Marshal(result)
		if err != nil {
			ctx.Log().Error(err)
		}
		return ctx.Bytes(200, faygo.MIMEApplicationJSONCharsetUTF8, marshal)
	}
}
