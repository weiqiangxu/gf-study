package main

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
    s := g.Server()
	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(MiddlewareCORS)
		// 后台模块
		group.ALL("/test", List)
	})
    s.Run()
}

// MiddlewareCORS 允许跨域
func MiddlewareCORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

func List(r *ghttp.Request) {
	fmt.Println(r.Get("name"))
	r.Response.Write("hello world")
}
