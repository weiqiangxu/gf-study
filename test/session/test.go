package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gsession"
)

func main() {
	s := g.Server()

	config := gredis.Config{
		Address: "127.0.0.1:6379",
		Db:      0,
	}
	group := "test"
	gredis.SetConfig(&config, group)
	redis := gredis.Instance(group)
	defer redis.Close(context.Background())
	s.SetConfigWithMap(g.Map{
		"SessionMaxAge":  time.Minute * 60,
		"SessionStorage": gsession.NewStorageRedis(redis),
		"dumpRouterMap":  true,
	})
	s.SetSessionIdName("si-pro_sid")
	// 分组路由注册方式
	s.Group("/web", func(group *ghttp.RouterGroup) {
		group.ALL("/login", login)
	})
	s.Run()
}

func login(r *ghttp.Request) {
	fmt.Println("41")
	go func() {
		defer func() {
			if e := recover(); e != nil {
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				fmt.Println(string(buf[:n]))
			}
		}()
		panic(123)
	}()
	fmt.Println(r.Server.GetSessionIdName())
	r.Cookie.SetHttpCookie(
		&http.Cookie{
			Name:     r.Server.GetSessionIdName(),
			Value:    r.GetSessionId(),
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			HttpOnly: true,
		},
	)
	data := map[string]interface{}{
		"name": "jack",
	}
	r.Session.SetMap(data)
	d := map[string]interface{}{"code": 0}
	var bb = []byte{}
	// 这里又忘了 bufio \ bytes怎么用了?
	cc := bytes.NewBuffer(bb)
	r.Response.Header().Write(cc)
	fmt.Println("header ", cc.String())
	r.Response.WriteJsonExit(d)
}
