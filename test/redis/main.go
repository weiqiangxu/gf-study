package main

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

func main() {
	a,err := g.Redis().Do(context.Background(),"GET", "test")
	fmt.Println("a ==> ",gconv.String(a))
	if err != nil {
		panic(err)
	}
	fmt.Println("------------------------------213---------------------------------------------------------------------------------")
	config := gredis.Config{
		Address : "127.0.0.1:6379",
		Db   : 0,
	}
	group := "test"
	gredis.SetConfig(&config, group)

	redis := gredis.Instance(group)
	defer redis.Close(context.Background())

	// _, err := redis.Do("SET", "k", "v")
	// if err != nil {
	// 	panic(err)
	// }
	fmt.Println("34 000000000000000")
	r, err := redis.Do(context.Background(),"GET", "test")
	if err != nil {
		panic(err)
	}
	fmt.Println(gconv.String(r))


}