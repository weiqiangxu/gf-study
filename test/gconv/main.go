package main

import (
	"fmt"

	"github.com/gogf/gf/v2/util/gconv"
)

type Data struct {
	Age int `json:"age"`
}
type Response struct {
	Name string `json:"passport"`
	Data Data   `json:"data"`
}

func main() {
	user := new(Response)
	c := map[string]interface{}{
		"data": map[string]interface{}{
			"age": 123,
		},
	}
	gconv.Struct(c, user)
	fmt.Println(user.Data.Age)

}
