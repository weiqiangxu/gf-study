package main

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
)

func main() {

	// CREATE TABLE user ( id INT(11),name VARCHAR(25));

	// 获取默认配置的数据库对象(配置名称为"default")
	db := g.DB()
	fmt.Println(db)
	users,err := db.Model("user").Limit(0, 10).All()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(users)

	

	
	// 获取配置分组名称为"user-center"的数据库对象
	// db := g.DB("user-center")

	// 使用原生New方法创建数据库对象
	// gdb.SetConfig(gdb.Config {
	// 	"default" : gdb.ConfigGroup {
	// 		gdb.ConfigNode {
	// 			Host     : "127.0.0.1",
	// 			Port     : "3306",
	// 			User     : "root",
	// 			Pass     : "123456",
	// 			Name     : "test",
	// 			Type     : "mysql",
	// 			Role     : "master",
	// 			Weight   : 100,
	// 		},
	// 	},
	// })
	// db, err := gdb.New()
	// fmt.Println(err)
	// users,err := db.Model("user").Limit(0, 10).All()
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	// fmt.Println(users)
	
	// db, err := gdb.New("user-center")

	// 使用原生单例管理方法获取数据库对象单例
	// db, err := gdb.Instance()
	// db, err := gdb.Instance("user-center")

	// 注意不用的时候不需要使用Close方法关闭数据库连接(并且gdb也没有提供Close方法)，
	// 数据库引擎底层采用了链接池设计，当链接不再使用时会自动关闭

}
