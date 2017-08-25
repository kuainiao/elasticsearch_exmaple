package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/henrylee2cn/faygo"
	"github.com/zhangweilun/tradeweb/constants"
	"net/http"

)

/**
*
* @author willian
* @created 2017-08-25 15:07
* @email 18702515157@163.com
**/

var Auth = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	token, err := request.ParseFromRequest(ctx.R, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(constants.Secret), nil
		})
	if err != nil {
		ctx.Stop()
		ctx.String(http.StatusUnauthorized, "Unauthorized access to this resource,请重新登录")

	} else {
		if token.Valid {
		} else {
			ctx.Stop()
			ctx.String(http.StatusBadRequest, "auth bad request")
		}
	}
	return nil
})
