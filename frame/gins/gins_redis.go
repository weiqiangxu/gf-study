// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gins

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gutil"
)

const (
	frameCoreComponentNameRedis = "gf.core.component.redis"
	configNodeNameRedis         = "redis"
)

// Redis returns an instance of redis client with specified configuration group name.
// Note that it panics if any error occurs duration instance creating.
func Redis(name ...string) *gredis.Redis {
	var (
		ctx   = context.Background()
		group = gredis.DefaultGroupName
	)
	if len(name) > 0 && name[0] != "" {
		group = name[0]
	}
	instanceKey := fmt.Sprintf("%s.%s", frameCoreComponentNameRedis, group)
	result := localInstances.GetOrSetFuncLock(instanceKey, func() interface{} {
		// If already configured, it returns the redis instance.
		if _, ok := gredis.GetConfig(group); ok {
			// 获取到配置- 依据配置实例化对象 - 注意这里的配置对象是一个结构体
			return gredis.Instance(group)
		}
		// Or else, it parses the default configuration file and returns a new redis instance.
		var (
			configMap map[string]interface{}
		)

		if configData, err := Config().Data(ctx); err != nil {
			panic(gerror.WrapCode(gcode.CodeOperationFailed, err, `retrieving redis configuration failed - 配置文件找不到`))
		} else {
			fmt.Println("configData == ",configData)
			if _, v := gutil.MapPossibleItemByKey(configData, configNodeNameRedis); v != nil {
				configMap = gconv.Map(v)
			}
		}

		fmt.Println("configMap = ",configMap)
		if len(configMap) > 0 {
			if v, ok := configMap[group]; ok {
				fmt.Println(v)
				redisConfig, err := gredis.ConfigFromMap(gconv.Map(v))
				if err != nil {
					panic(err)
				}
				redisClient, err := gredis.New(redisConfig)
				if err != nil {
					panic(err)
				}
				return redisClient
			} else {
				panic(fmt.Sprintf(`missing configuration for redis group "%s"`, group))
			}
		} else {
			panic(`missing configuration for redis: "redis" node not found`)
		}
		return nil
	})
	if result != nil {
		return result.(*gredis.Redis)
	}
	return nil
}
