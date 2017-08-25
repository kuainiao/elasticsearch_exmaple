package middleware

import "github.com/henrylee2cn/faygo"

/**
*
* @author willian
* @created 2017-08-09 11:27
* @email 18702515157@163.com
**/
// CrossOrigin creates Cross-Domain middleware
var CrossOrigin = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	ctx.SetHeader(faygo.HeaderAccessControlAllowOrigin, "*")
	ctx.SetHeader(faygo.HeaderAccessControlAllowCredentials, "true")
	ctx.SetHeader(faygo.HeaderAccessControlAllowHeaders, "Origin, X-Requested-With, Content-Type, Accept,Authorization")
	return nil
})
