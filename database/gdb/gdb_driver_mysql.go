// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gdb

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/internal/intlog"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
)

// DriverMysql is the driver for mysql database.
// 是mysql数据库的驱动程序。
type DriverMysql struct {
	*Core
}

// New creates and returns a database object for mysql.
// 为mysql创建并返回一个数据库对象。
// It implements the interface of gdb.Driver for extra database driver installation.
// 它实现了gdb.Driver接口，用于安装额外的数据库驱动程序。
func (d *DriverMysql) New(core *Core, node *ConfigNode) (DB, error) {
	// 实例化mysql驱动实例
	return &DriverMysql{
		Core: core,
	}, nil
}

// Open creates and returns an underlying sql.DB object for mysql.
// 创建并返回mysql的底层sql.DB对象。
// Note that it converts time.Time argument to local timezone in default.
// 请注意，默认情况下，它将time.time参数转换为本地时区。
func (d *DriverMysql) Open(config *ConfigNode) (*sql.DB, error) {
	var source string
	if config.Link != "" {
		source = config.Link
		// Custom changing the schema in runtime.
		// 在运行时更改架构。
		if config.Name != "" {
			source, _ = gregex.ReplaceString(`/([\w\.\-]+)+`, "/"+config.Name, source)
		}
	} else {
		source = fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=%s",
			config.User, config.Pass, config.Host, config.Port, config.Name, config.Charset,
		)
		if config.Timezone != "" {
			source = fmt.Sprintf("%s&loc=%s", source, url.QueryEscape(config.Timezone))
		}
	}
	intlog.Printf(d.GetCtx(), "Open: %s", source)
	// 调用的sql是
	if db, err := sql.Open("mysql", source); err == nil {
		return db, nil
	} else {
		return nil, err
	}
}

// FilteredLink retrieves and returns filtered `linkInfo` that can be using for
// logging or tracing purpose.
// 检索并返回可用于日志记录或跟踪目的的筛选“linkInfo”。
func (d *DriverMysql) FilteredLink() string {
	linkInfo := d.GetConfig().Link
	if linkInfo == "" {
		return ""
	}
	s, _ := gregex.ReplaceString(
		`(.+?):(.+)@tcp(.+)`,
		`$1:xxx@tcp$3`,
		linkInfo,
	)
	return s
}

// GetChars returns the security char for this type of database.
// 返回此类型数据库的安全字符。
func (d *DriverMysql) GetChars() (charLeft string, charRight string) {
	return "`", "`"
}

// DoCommit handles the sql before posts it to database.
// 在将sql发布到数据库之前处理sql。
func (d *DriverMysql) DoCommit(ctx context.Context, link Link, sql string, args []interface{}) (newSql string, newArgs []interface{}, err error) {
	return d.Core.DoCommit(ctx, link, sql, args)
}

// Tables retrieves and returns the tables of current schema.
// 检索并返回当前架构的表。
// It's mainly used in cli tool chain for automatically generating the models.
func (d *DriverMysql) Tables(ctx context.Context, schema ...string) (tables []string, err error) {
	var result Result
	link, err := d.SlaveLink(schema...)
	if err != nil {
		return nil, err
	}
	result, err = d.DoGetAll(ctx, link, `SHOW TABLES`)
	if err != nil {
		return
	}
	for _, m := range result {
		for _, v := range m {
			tables = append(tables, v.String())
		}
	}
	return
}

// TableFields retrieves and returns the fields information of specified table of current
// 检索并返回当前表的指定表的字段信息
// schema.
//
// The parameter `link` is optional, if given nil it automatically retrieves a raw sql connection
// as its link to proceed necessary sql query.
// 参数'link'是可选的，如果给定nil，它将自动检索原始sql连接作为其链接，以进行必要的sql查询。
//
// Note that it returns a map containing the field name and its corresponding fields.
// 请注意，它返回一个包含字段名及其对应字段的映射。
// As a map is unsorted, the TableField struct has a "Index" field marks its sequence in
// the fields.
// 当映射未排序时，TableField结构有一个“索引”字段标记其在字段中的顺序。
//
// It's using cache feature to enhance the performance, which is never expired util the
// process restarts.
// 它使用缓存特性来增强性能，在进程重新启动之前，性能永远不会过期。
func (d *DriverMysql) TableFields(ctx context.Context, table string, schema ...string) (fields map[string]*TableField, err error) {
	charL, charR := d.GetChars()
	table = gstr.Trim(table, charL+charR)
	if gstr.Contains(table, " ") {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "function TableFields supports only single table operations")
	}
	useSchema := d.schema.Val()
	if len(schema) > 0 && schema[0] != "" {
		useSchema = schema[0]
	}
	tableFieldsCacheKey := fmt.Sprintf(
		`mysql_table_fields_%s_%s@group:%s`,
		table, useSchema, d.GetGroup(),
	)
	v := tableFieldsMap.GetOrSetFuncLock(tableFieldsCacheKey, func() interface{} {
		var (
			result    Result
			link, err = d.SlaveLink(useSchema)
		)
		if err != nil {
			return nil
		}
		result, err = d.DoGetAll(ctx, link, fmt.Sprintf(`SHOW FULL COLUMNS FROM %s`, d.QuoteWord(table)))
		if err != nil {
			return nil
		}
		fields = make(map[string]*TableField)
		for i, m := range result {
			fields[m["Field"].String()] = &TableField{
				Index:   i,
				Name:    m["Field"].String(),
				Type:    m["Type"].String(),
				Null:    m["Null"].Bool(),
				Key:     m["Key"].String(),
				Default: m["Default"].Val(),
				Extra:   m["Extra"].String(),
				Comment: m["Comment"].String(),
			}
		}
		return fields
	})
	if v != nil {
		fields = v.(map[string]*TableField)
	}
	return
}
