package handler

import (
	"github.com/henrylee2cn/faygo"
	"github.com/zhangweilun/tradeweb/service"
)

/*
Index
*/
var Index = faygo.HandlerFunc(func(ctx *faygo.Context) error {
	service.MoveFrank("2015-01-01", "2015-12-30")
	//return ctx.Render(200, faygo.JoinStatic("index.html"), faygo.Map{
	//	"TITLE":   "faygo",
	//	"VERSION": faygo.VERSION,
	//	"CONTENT": "Welcome To Faygo",
	//	"AUTHOR":  "HenryLee",
	//})
	return ctx.String(200, "hello,william")
})
