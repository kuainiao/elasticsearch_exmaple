package middleware

import "github.com/henrylee2cn/faygo"

//DbQuery query the DataBase
var DbQuery = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	return nil
})
