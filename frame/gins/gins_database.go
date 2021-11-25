// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gins

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/internal/intlog"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gutil"
)

const (
	frameCoreComponentNameDatabase = "gf.core.component.database"
	configNodeNameDatabase         = "database"
)

// Database returns an instance of database ORM object with specified configuration group name.
// 返回具有指定配置组名称的数据库ORM对象的实例
// Note that it panics if any error occurs duration instance creating.
// 请注意，如果在创建实例期间发生任何错误，它将陷入panic
func Database(name ...string) gdb.DB {
	var (
		ctx   = context.Background()
		group = gdb.DefaultGroupName
	)
	if len(name) > 0 && name[0] != "" {
		group = name[0]
	}
	instanceKey := fmt.Sprintf("%s.%s", frameCoreComponentNameDatabase, group)
	// GetOrSetFuncLock按键返回值，或者使用回调函数“f”的返回值设置值（如果不存在），然后返回此值。
	// GetOrSetFuncLock函数与GetOrSetFunc函数的不同之处在于，它使用哈希映射的mutex.Lock执行函数'f'。
	db := localInstances.GetOrSetFuncLock(instanceKey, func() interface{} {
		// It ignores returned error to avoid file no found error while it's not necessary.
		// 它忽略返回的错误，以避免在不需要时出现文件未找到错误。
		var (
			configMap     map[string]interface{}
			configNodeKey = configNodeNameDatabase
		)
		fmt.Println("configNodeNameDatabase --- ",configNodeNameDatabase)
		// It firstly searches the configuration of the instance name.
		// 它首先搜索实例名称的配置。
		if configData, _ := Config().Data(ctx); len(configData) > 0 {
			// MapPossibleItemByKey尝试为给定的键查找可能的键值对，忽略大小写和符号。
			if v, _ := gutil.MapPossibleItemByKey(configData, configNodeNameDatabase); v != "" {
				configNodeKey = v
			}
		}
		// Get按指定的“模式”检索并返回值。如果'pattern'为空或字符串'，则返回当前Json对象的所有值。如果“pattern”未找到值，则返回nil。
		if v, _ := Config().Get(ctx, configNodeKey); !v.IsEmpty() {
			configMap = v.Map()
		}
		if len(configMap) == 0 && !gdb.IsConfigured() {
			// File configuration object checks.
			// 文件配置检查
			var (
				err            error
				configFilePath string
			)
			if fileConfig, ok := Config().GetAdapter().(*gcfg.AdapterFile); ok {
				if configFilePath, err = fileConfig.GetFilePath(); configFilePath == "" {
					exampleFileName := "config.example.toml"
					if exampleConfigFilePath, _ := fileConfig.GetFilePath(exampleFileName); exampleConfigFilePath != "" {
						// 配置文件未找到
						err = gerror.WrapCodef(
							gcode.CodeMissingConfiguration,
							err,
							`configuration file "%s" not found, but found "%s", did you miss renaming the example configuration file?`,
							fileConfig.GetFileName(),
							exampleFileName,
						)
					} else {
						// 配置文件里面的数据库配置 - 未找到
						err = gerror.WrapCodef(
							gcode.CodeMissingConfiguration,
							err,
							`configuration file "%s" not found, did you miss the configuration file or the misspell the configuration file name?`,
							fileConfig.GetFileName(),
						)
					}
					if err != nil {
						panic(err)
					}
				}
			}
			// Panic if nothing found in Config object or in gdb configuration.
			if len(configMap) == 0 && !gdb.IsConfigured() {
				err = gerror.WrapCodef(
					gcode.CodeMissingConfiguration,
					err,
					`database initialization failed: "%s" node not found, is configuration file or configuration node missing?`,
					configNodeNameDatabase,
				)
				panic(err)
			}
		}

		if len(configMap) == 0 {
			configMap = make(map[string]interface{})
		}
		// Parse `m` as map-slice and adds it to global configurations for package gdb.
		// 将'm'解析为映射片，并将其添加到包gdb的全局配置中。
		for g, groupConfig := range configMap {
			cg := gdb.ConfigGroup{}
			switch value := groupConfig.(type) {
			case []interface{}:
				for _, v := range value {
					if node := parseDBConfigNode(v); node != nil {
						cg = append(cg, *node)
					}
				}
			case map[string]interface{}:
				if node := parseDBConfigNode(value); node != nil {
					cg = append(cg, *node)
				}
			}
			if len(cg) > 0 {
				if gdb.GetConfig(group) == nil {
					intlog.Printf(ctx, "add configuration for group: %s, %#v", g, cg)
					gdb.SetConfigGroup(g, cg)
				} else {
					intlog.Printf(ctx, "ignore configuration as it already exists for group: %s, %#v", g, cg)
					intlog.Printf(ctx, "%s, %#v", g, cg)
				}
			}
		}
		// Parse `m` as a single node configuration,
		// 将'm'解析为单节点配置
		// which is the default group configuration.
		// 这是默认的组配置
		if node := parseDBConfigNode(configMap); node != nil {
			cg := gdb.ConfigGroup{}
			if node.Link != "" || node.Host != "" {
				cg = append(cg, *node)
			}

			if len(cg) > 0 {
				if gdb.GetConfig(group) == nil {
					intlog.Printf(ctx, "add configuration for group: %s, %#v", gdb.DefaultGroupName, cg)
					gdb.SetConfigGroup(gdb.DefaultGroupName, cg)
				} else {
					intlog.Printf(ctx, "ignore configuration as it already exists for group: %s, %#v", gdb.DefaultGroupName, cg)
					intlog.Printf(ctx, "%s, %#v", gdb.DefaultGroupName, cg)
				}
			}
		}

		// Create a new ORM object with given configurations.
		// 创建具有给定配置的新ORM对象。
		fmt.Println("name ==>  ",name)
		if db, err := gdb.New(name...); err == nil {
			// Initialize logger for ORM.
			// 初始化ORM的记录器。
			var (
				loggerConfigMap map[string]interface{}
				loggerNodeName  = fmt.Sprintf("%s.%s", configNodeKey, configNodeNameLogger)
			)
			if v, _ := Config().Get(ctx, loggerNodeName); !v.IsEmpty() {
				loggerConfigMap = v.Map()
			}
			if len(loggerConfigMap) == 0 {
				if v, _ := Config().Get(ctx, configNodeKey); !v.IsEmpty() {
					loggerConfigMap = v.Map()
				}
			}
			if len(loggerConfigMap) > 0 {
				if err = db.GetLogger().SetConfigWithMap(loggerConfigMap); err != nil {
					panic(err)
				}
			}
			return db
		} else {
			// If panics, often because it does not find its configuration for given group.
			panic(err)
		}
		return nil
	})
	if db != nil {
		return db.(gdb.DB)
	}
	return nil
}

func parseDBConfigNode(value interface{}) *gdb.ConfigNode {
	nodeMap, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}
	node := &gdb.ConfigNode{}
	err := gconv.Struct(nodeMap, node)
	if err != nil {
		panic(err)
	}
	// Be compatible with old version.
	if _, v := gutil.MapPossibleItemByKey(nodeMap, "LinkInfo"); v != nil {
		node.Link = gconv.String(v)
	}
	if _, v := gutil.MapPossibleItemByKey(nodeMap, "Link"); v != nil {
		node.Link = gconv.String(v)
	}
	// Parse link syntax.
	if node.Link != "" && node.Type == "" {
		match, _ := gregex.MatchString(`([a-z]+):(.+)`, node.Link)
		if len(match) == 3 {
			node.Type = gstr.Trim(match[1])
			node.Link = gstr.Trim(match[2])
		}
	}
	return node
}
