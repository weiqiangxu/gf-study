package main

import (
	"context"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
)

func main() {

	// 连不上 - 不会告警
	_,err := g.Redis().Do(context.Background(),"GET", "test")
	if err != nil {
		panic(err)
	}

	// 连不上 - 会告警
	config := gredis.Config{
		Address : "127.0.0.2:6379",
		Db   : 0,
	}
	group := "test"
	gredis.SetConfig(&config, group)
	redis := gredis.Instance(group)
	defer redis.Close(context.Background())
	_, err = redis.Do(context.Background(),"GET", "test")
	if err != nil {
		panic(err)
	}
}