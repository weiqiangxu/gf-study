// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.
//

package gdb

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/internal/intlog"
	"github.com/gogf/gf/v2/internal/utils"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

// GetCore returns the underlying *Core object.
// 返回基础*Core核心对象。
func (c *Core) GetCore() *Core {
	return c
}

// Ctx is a chaining function, which creates and returns a new DB that is a shallow copy
// 是一个链接函数，它创建并返回一个浅层副本的新数据库
// of current DB object and with given context in it.
// 当前数据库对象，并且其中包含给定的上下文
// Note that this returned DB object can be used only once, so do not assign it to
// 请注意，此返回的DB对象只能使用一次，因此不要将其分配给
// a global or package variable for long using.
// 长期使用的全局或包变量。
func (c *Core) Ctx(ctx context.Context) DB {
	if ctx == nil {
		return c.db
	}
	// It makes a shallow copy of current db and changes its context for next chaining operation.
	// 它创建当前数据库的浅层副本，并为下一个链接操作更改其上下文。
	var (
		err        error
		newCore    = &Core{}
		configNode = c.db.GetConfig()
	)
	*newCore = *c
	newCore.ctx = ctx
	// It creates a new DB object, which is commonly a wrapper for object `Core`.
	// 它创建了一个新的DB对象，它通常是对象“Core”的包装器。
	newCore.db, err = driverMap[configNode.Type].New(newCore, configNode)
	if err != nil {
		// It is really a serious error here.
		// Do not let it continue.
		panic(err)
	}
	return newCore.db
}

// GetCtx returns the context for current DB.
// 返回当前数据库的上下文。
// It returns `context.Background()` is there's no context previously set.
// 它返回`context.Background（）`就是之前没有设置上下文。
func (c *Core) GetCtx() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.TODO()
}

// GetCtxTimeout returns the context and cancel function for specified timeout type.
// GetCtxTimeout返回指定超时类型的上下文和取消函数。
func (c *Core) GetCtxTimeout(timeoutType int, ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = c.GetCtx()
	} else {
		ctx = context.WithValue(ctx, "WrappedByGetCtxTimeout", nil)
	}
	switch timeoutType {
	case ctxTimeoutTypeExec:
		if c.db.GetConfig().ExecTimeout > 0 {
			return context.WithTimeout(ctx, c.db.GetConfig().ExecTimeout)
		}
	case ctxTimeoutTypeQuery:
		if c.db.GetConfig().QueryTimeout > 0 {
			return context.WithTimeout(ctx, c.db.GetConfig().QueryTimeout)
		}
	case ctxTimeoutTypePrepare:
		if c.db.GetConfig().PrepareTimeout > 0 {
			return context.WithTimeout(ctx, c.db.GetConfig().PrepareTimeout)
		}
	default:
		panic(gerror.NewCodef(gcode.CodeInvalidParameter, "invalid context timeout type: %d", timeoutType))
	}
	return ctx, func() {}
}

// Close closes the database and prevents new queries from starting.
// 关闭数据库并阻止启动新查询。
// Close then waits for all queries that have started processing on the server
// to finish.
// 然后等待服务器上已开始处理的所有查询
//
// It is rare to Close a DB, as the DB handle is meant to be
// long-lived and shared between many goroutines.
// 关闭一个DB是很少见的，因为DB句柄是长寿命的，并且在许多Goroutine之间共享。
func (c *Core) Close(ctx context.Context) (err error) {
	c.links.LockFunc(func(m map[string]interface{}) {
		for k, v := range m {
			if db, ok := v.(*sql.DB); ok {
				err = db.Close()
				intlog.Printf(ctx, `close link: %s, err: %v`, k, err)
				if err != nil {
					return
				}
				delete(m, k)
			}
		}
	})
	return
}

// Master creates and returns a connection from master node if master-slave configured.
// 如果配置了主从式，则从主节点创建并返回连接。
// It returns the default connection if master-slave not configured.
// 如果未配置主从，则返回默认连接。
func (c *Core) Master(schema ...string) (*sql.DB, error) {
	fmt.Println("gdb_core.go 133 ")
	useSchema := ""
	if len(schema) > 0 && schema[0] != "" {
		useSchema = schema[0]
	} else {
		useSchema = c.schema.Val()
	}
	return c.getSqlDb(true, useSchema)
}

// Slave creates and returns a connection from slave node if master-slave configured.
// 如果配置了主从节点，则创建并返回从节点的连接。
// It returns the default connection if master-slave not configured.
// 如果未配置主从，则返回默认连接。
func (c *Core) Slave(schema ...string) (*sql.DB, error) {
	fmt.Println("gdb_core.go Slave")
	useSchema := ""
	if len(schema) > 0 && schema[0] != "" {
		useSchema = schema[0]
	} else {
		useSchema = c.schema.Val()
	}
	return c.getSqlDb(false, useSchema)
}

// GetAll queries and returns data records from database.
// GetAll查询并从数据库返回数据记录。
func (c *Core) GetAll(ctx context.Context, sql string, args ...interface{}) (Result, error) {
	return c.db.DoGetAll(ctx, nil, sql, args...)
}

// DoGetAll queries and returns data records from database.
// 查询并返回数据库中的数据记录。
func (c *Core) DoGetAll(ctx context.Context, link Link, sql string, args ...interface{}) (result Result, err error) {
	fmt.Println("gdb_core.go 165 ",reflect.TypeOf(c.db))
	return c.db.DoQuery(ctx, link, sql, args...)
}

// GetOne queries and returns one record from database.
// 查询并从数据库返回一条记录。
func (c *Core) GetOne(ctx context.Context, sql string, args ...interface{}) (Record, error) {
	list, err := c.db.GetAll(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		return list[0], nil
	}
	return nil, nil
}

// GetArray queries and returns data values as slice from database.
// 从数据库中查询并返回数据值作为切片。
// Note that if there are multiple columns in the result, it returns just one column values randomly.
// 请注意，如果结果中有多个列，它将随机返回一列值。
func (c *Core) GetArray(ctx context.Context, sql string, args ...interface{}) ([]Value, error) {
	all, err := c.db.DoGetAll(ctx, nil, sql, args...)
	if err != nil {
		return nil, err
	}
	return all.Array(), nil
}

// GetStruct queries one record from database and converts it to given struct.
// 从数据库中查询一条记录并将其转换为给定的结构。
// The parameter `pointer` should be a pointer to struct.
// 参数“pointer”应该是指向struct的指针。
func (c *Core) GetStruct(ctx context.Context, pointer interface{}, sql string, args ...interface{}) error {
	one, err := c.db.GetOne(ctx, sql, args...)
	if err != nil {
		return err
	}
	return one.Struct(pointer)
}

// GetStructs queries records from database and converts them to given struct.
// 从数据库查询记录并将其转换为给定结构。
// The parameter `pointer` should be type of struct slice: []struct/[]*struct.
// 参数“pointer”应为结构片的类型：[]struct/[]*struct。
func (c *Core) GetStructs(ctx context.Context, pointer interface{}, sql string, args ...interface{}) error {
	all, err := c.db.GetAll(ctx, sql, args...)
	if err != nil {
		return err
	}
	return all.Structs(pointer)
}

// GetScan queries one or more records from database and converts them to given struct or
// GetScan从数据库中查询一个或多个记录，并将它们转换为给定的结构
// struct array.
//
// If parameter `pointer` is type of struct pointer, it calls GetStruct internally for
// the conversion. If parameter `pointer` is type of slice, it calls GetStructs internally
// for conversion.
// 如果参数`pointer`是结构指针的类型，它将在内部调用GetStruct进行转换。若参数`pointer`是切片类型，它将在内部调用GetStructs进行转换。
func (c *Core) GetScan(ctx context.Context, pointer interface{}, sql string, args ...interface{}) error {
	reflectInfo := utils.OriginTypeAndKind(pointer)
	if reflectInfo.InputKind != reflect.Ptr {
		return gerror.NewCodef(
			gcode.CodeInvalidParameter,
			"params should be type of pointer, but got: %v",
			reflectInfo.InputKind,
		)
	}
	switch reflectInfo.OriginKind {
	case reflect.Array, reflect.Slice:
		return c.db.GetCore().GetStructs(ctx, pointer, sql, args...)

	case reflect.Struct:
		return c.db.GetCore().GetStruct(ctx, pointer, sql, args...)
	}
	return gerror.NewCodef(
		gcode.CodeInvalidParameter,
		`in valid parameter type "%v", of which element type should be type of struct/slice`,
		reflectInfo.InputType,
	)
}

// GetValue queries and returns the field value from database.
// GetValue查询并返回数据库中的字段值。
// The sql should query only one field from database, or else it returns only one
// field of the result.
// sql应该只查询数据库中的一个字段，否则它只返回结果的一个字段。
func (c *Core) GetValue(ctx context.Context, sql string, args ...interface{}) (Value, error) {
	one, err := c.db.GetOne(ctx, sql, args...)
	if err != nil {
		return gvar.New(nil), err
	}
	for _, v := range one {
		return v, nil
	}
	return gvar.New(nil), nil
}

// GetCount queries and returns the count from database.
// 查询并从数据库返回计数。
func (c *Core) GetCount(ctx context.Context, sql string, args ...interface{}) (int, error) {
	// If the query fields do not contains function "COUNT",
	// 如果查询字段不包含函数“COUNT”，
	// it replaces the sql string and adds the "COUNT" function to the fields.
	// 它替换sql字符串并将“COUNT”函数添加到字段中。
	if !gregex.IsMatchString(`(?i)SELECT\s+COUNT\(.+\)\s+FROM`, sql) {
		sql, _ = gregex.ReplaceString(`(?i)(SELECT)\s+(.+)\s+(FROM)`, `$1 COUNT($2) $3`, sql)
	}
	value, err := c.db.GetValue(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return value.Int(), nil
}

// Union does "(SELECT xxx FROM xxx) UNION (SELECT xxx FROM xxx) ..." statement.
// Union does“（从xxx中选择xxx）Union（从xxx中选择xxx）…”语句。
func (c *Core) Union(unions ...*Model) *Model {
	return c.doUnion(unionTypeNormal, unions...)
}

// UnionAll does "(SELECT xxx FROM xxx) UNION ALL (SELECT xxx FROM xxx) ..." statement.
func (c *Core) UnionAll(unions ...*Model) *Model {
	return c.doUnion(unionTypeAll, unions...)
}

func (c *Core) doUnion(unionType int, unions ...*Model) *Model {
	var (
		unionTypeStr   string
		composedSqlStr string
		composedArgs   = make([]interface{}, 0)
	)
	if unionType == unionTypeAll {
		unionTypeStr = "UNION ALL"
	} else {
		unionTypeStr = "UNION"
	}
	for _, v := range unions {
		sqlWithHolder, holderArgs := v.getFormattedSqlAndArgs(queryTypeNormal, false)
		if composedSqlStr == "" {
			composedSqlStr += fmt.Sprintf(`(%s)`, sqlWithHolder)
		} else {
			composedSqlStr += fmt.Sprintf(` %s (%s)`, unionTypeStr, sqlWithHolder)
		}
		composedArgs = append(composedArgs, holderArgs...)
	}
	return c.db.Raw(composedSqlStr, composedArgs...)
}

// PingMaster pings the master node to check authentication or keeps the connection alive.
// ping主节点以检查身份验证或保持连接处于活动状态。
func (c *Core) PingMaster() error {
	if master, err := c.db.Master(); err != nil {
		return err
	} else {
		return master.Ping()
	}
}

// PingSlave pings the slave node to check authentication or keeps the connection alive.
// ping从属节点以检查身份验证或保持连接处于活动状态。
func (c *Core) PingSlave() error {
	fmt.Println("ping slave ")
	if slave, err := c.db.Slave(); err != nil {
		return err
	} else {
		return slave.Ping()
	}
}

// Insert does "INSERT INTO ..." statement for the table.
// Insert对表执行“Insert INTO…”语句。
// If there's already one unique record of the data in the table, it returns error.
// 如果表中已经有一条唯一的数据记录，它将返回错误。
//
// The parameter `data` can be type of map/gmap/struct/*struct/[]map/[]struct, etc.
// 参数'data'可以是map/gmap/struct/*struct/[]map/[]struct等类型。
// Eg:
// Data(g.Map{"uid": 10000, "name":"john"})
// Data(g.Slice{g.Map{"uid": 10000, "name":"john"}, g.Map{"uid": 20000, "name":"smith"})
//
// The parameter `batch` specifies the batch operation count when given data is slice.
// 参数“batch”指定给定数据为slice时的批操作计数。
func (c *Core) Insert(ctx context.Context, table string, data interface{}, batch ...int) (sql.Result, error) {
	if len(batch) > 0 {
		return c.Model(table).Ctx(ctx).Data(data).Batch(batch[0]).Insert()
	}
	return c.Model(table).Ctx(ctx).Data(data).Insert()
}

// InsertIgnore does "INSERT IGNORE INTO ..." statement for the table.
// If there's already one unique record of the data in the table, it ignores the inserting.
// 如果表中已经有一条唯一的数据记录，它将忽略插入。
//
// The parameter `data` can be type of map/gmap/struct/*struct/[]map/[]struct, etc.
// Eg:
// Data(g.Map{"uid": 10000, "name":"john"})
// Data(g.Slice{g.Map{"uid": 10000, "name":"john"}, g.Map{"uid": 20000, "name":"smith"})
//
// The parameter `batch` specifies the batch operation count when given data is slice.
func (c *Core) InsertIgnore(ctx context.Context, table string, data interface{}, batch ...int) (sql.Result, error) {
	if len(batch) > 0 {
		return c.Model(table).Ctx(ctx).Data(data).Batch(batch[0]).InsertIgnore()
	}
	return c.Model(table).Ctx(ctx).Data(data).InsertIgnore()
}

// InsertAndGetId performs action Insert and returns the last insert id that automatically generated.
// 执行操作插入并返回自动生成的最后一个插入id。
func (c *Core) InsertAndGetId(ctx context.Context, table string, data interface{}, batch ...int) (int64, error) {
	if len(batch) > 0 {
		return c.Model(table).Ctx(ctx).Data(data).Batch(batch[0]).InsertAndGetId()
	}
	return c.Model(table).Ctx(ctx).Data(data).InsertAndGetId()
}

// Replace does "REPLACE INTO ..." statement for the table.
// If there's already one unique record of the data in the table, it deletes the record
// and inserts a new one.
// 如果表中已经有一条唯一的数据记录，它将删除该记录并插入一条新记录。
//
// The parameter `data` can be type of map/gmap/struct/*struct/[]map/[]struct, etc.
// Eg:
// Data(g.Map{"uid": 10000, "name":"john"})
// Data(g.Slice{g.Map{"uid": 10000, "name":"john"}, g.Map{"uid": 20000, "name":"smith"})
//
// The parameter `data` can be type of map/gmap/struct/*struct/[]map/[]struct, etc.
// 参数'data'可以是map/gmap/struct/*struct/[]map/[]struct等类型。
// If given data is type of slice, it then does batch replacing, and the optional parameter
// 如果给定的数据是切片的类型，那么它将进行批量替换，并使用可选参数
// `batch` specifies the batch operation count.
// `batch`指定批处理操作计数。
func (c *Core) Replace(ctx context.Context, table string, data interface{}, batch ...int) (sql.Result, error) {
	if len(batch) > 0 {
		return c.Model(table).Ctx(ctx).Data(data).Batch(batch[0]).Replace()
	}
	return c.Model(table).Ctx(ctx).Data(data).Replace()
}

// Save does "INSERT INTO ... ON DUPLICATE KEY UPDATE..." statement for the table.
// Save对表执行“在重复密钥更新时插入…”语句。
// It updates the record if there's primary or unique index in the saving data,
// 如果保存的数据中有主索引或唯一索引，则会更新记录，
// or else it inserts a new record into the table.
// 否则，它会在表中插入一条新记录。
//
// The parameter `data` can be type of map/gmap/struct/*struct/[]map/[]struct, etc.
// Eg:
// Data(g.Map{"uid": 10000, "name":"john"})
// Data(g.Slice{g.Map{"uid": 10000, "name":"john"}, g.Map{"uid": 20000, "name":"smith"})
//
// If given data is type of slice, it then does batch saving, and the optional parameter
// `batch` specifies the batch operation count.
// 如果给定的数据是切片类型，则会执行批保存，可选参数“batch”指定批操作计数。
func (c *Core) Save(ctx context.Context, table string, data interface{}, batch ...int) (sql.Result, error) {
	if len(batch) > 0 {
		return c.Model(table).Ctx(ctx).Data(data).Batch(batch[0]).Save()
	}
	return c.Model(table).Ctx(ctx).Data(data).Save()
}

// DoInsert inserts or updates data forF given table.
// 插入或更新给定表的数据。
// This function is usually used for custom interface definition, you do not need call it manually.
// 此函数通常用于自定义接口定义，不需要手动调用。
// The parameter `data` can be type of map/gmap/struct/*struct/[]map/[]struct, etc.
// 参数'data'可以是map/gmap/struct/*struct/[]map/[]struct等类型。
// Eg:
// Data(g.Map{"uid": 10000, "name":"john"})
// Data(g.Slice{g.Map{"uid": 10000, "name":"john"}, g.Map{"uid": 20000, "name":"smith"})
//
// The parameter `option` values are as follows:
// 参数'option'值如下所示：
// 0: insert:  just insert, if there's unique/primary key in the data, it returns error;
// 1: replace: if there's unique/primary key in the data, it deletes it from table and inserts a new one;
// 2: save:    if there's unique/primary key in the data, it updates it or else inserts a new one;
// 3: ignore:  if there's unique/primary key in the data, it ignores the inserting;
func (c *Core) DoInsert(ctx context.Context, link Link, table string, list List, option DoInsertOption) (result sql.Result, err error) {
	var (
		keys           []string      // Field names.
		values         []string      // Value holder string array, like: (?,?,?)
		params         []interface{} // Values that will be committed to underlying database driver.
		onDuplicateStr string        // onDuplicateStr is used in "ON DUPLICATE KEY UPDATE" statement.
	)
	// Handle the field names and placeholders.
	for k := range list[0] {
		keys = append(keys, k)
	}
	// Prepare the batch result pointer.
	var (
		charL, charR = c.db.GetChars()
		batchResult  = new(SqlResult)
		keysStr      = charL + strings.Join(keys, charR+","+charL) + charR
		operation    = GetInsertOperationByOption(option.InsertOption)
	)
	if option.InsertOption == insertOptionSave {
		onDuplicateStr = c.formatOnDuplicate(keys, option)
	}
	var (
		listLength  = len(list)
		valueHolder = make([]string, 0)
	)
	for i := 0; i < listLength; i++ {
		values = values[:0]
		// Note that the map type is unordered,
		// so it should use slice+key to retrieve the value.
		for _, k := range keys {
			if s, ok := list[i][k].(Raw); ok {
				values = append(values, gconv.String(s))
			} else {
				values = append(values, "?")
				params = append(params, list[i][k])
			}
		}
		valueHolder = append(valueHolder, "("+gstr.Join(values, ",")+")")
		// Batch package checks: It meets the batch number, or it is the last element.
		if len(valueHolder) == option.BatchCount || (i == listLength-1 && len(valueHolder) > 0) {
			r, err := c.db.DoExec(ctx, link, fmt.Sprintf(
				"%s INTO %s(%s) VALUES%s %s",
				operation, c.QuotePrefixTableName(table), keysStr,
				gstr.Join(valueHolder, ","),
				onDuplicateStr,
			), params...)
			if err != nil {
				return r, err
			}
			if n, err := r.RowsAffected(); err != nil {
				return r, err
			} else {
				batchResult.result = r
				batchResult.affected += n
			}
			params = params[:0]
			valueHolder = valueHolder[:0]
		}
	}
	return batchResult, nil
}

func (c *Core) formatOnDuplicate(columns []string, option DoInsertOption) string {
	var onDuplicateStr string
	if option.OnDuplicateStr != "" {
		onDuplicateStr = option.OnDuplicateStr
	} else if len(option.OnDuplicateMap) > 0 {
		for k, v := range option.OnDuplicateMap {
			if len(onDuplicateStr) > 0 {
				onDuplicateStr += ","
			}
			switch v.(type) {
			case Raw, *Raw:
				onDuplicateStr += fmt.Sprintf(
					"%s=%s",
					c.QuoteWord(k),
					v,
				)
			default:
				onDuplicateStr += fmt.Sprintf(
					"%s=VALUES(%s)",
					c.QuoteWord(k),
					c.QuoteWord(gconv.String(v)),
				)
			}
		}
	} else {
		for _, column := range columns {
			// If it's SAVE operation, do not automatically update the creating time.
			if c.isSoftCreatedFieldName(column) {
				continue
			}
			if len(onDuplicateStr) > 0 {
				onDuplicateStr += ","
			}
			onDuplicateStr += fmt.Sprintf(
				"%s=VALUES(%s)",
				c.QuoteWord(column),
				c.QuoteWord(column),
			)
		}
	}
	return fmt.Sprintf("ON DUPLICATE KEY UPDATE %s", onDuplicateStr)
}

// Update does "UPDATE ... " statement for the table.
//
// The parameter `data` can be type of string/map/gmap/struct/*struct, etc.
// Eg: "uid=10000", "uid", 10000, g.Map{"uid": 10000, "name":"john"}
//
// The parameter `condition` can be type of string/map/gmap/slice/struct/*struct, etc.
// It is commonly used with parameter `args`.
// Eg:
// "uid=10000",
// "uid", 10000
// "money>? AND name like ?", 99999, "vip_%"
// "status IN (?)", g.Slice{1,2,3}
// "age IN(?,?)", 18, 50
// User{ Id : 1, UserName : "john"}.
func (c *Core) Update(ctx context.Context, table string, data interface{}, condition interface{}, args ...interface{}) (sql.Result, error) {
	return c.Model(table).Ctx(ctx).Data(data).Where(condition, args...).Update()
}

// DoUpdate does "UPDATE ... " statement for the table.
// This function is usually used for custom interface definition, you do not need to call it manually.
func (c *Core) DoUpdate(ctx context.Context, link Link, table string, data interface{}, condition string, args ...interface{}) (result sql.Result, err error) {
	table = c.QuotePrefixTableName(table)
	var (
		rv   = reflect.ValueOf(data)
		kind = rv.Kind()
	)
	if kind == reflect.Ptr {
		rv = rv.Elem()
		kind = rv.Kind()
	}
	var (
		params  []interface{}
		updates = ""
	)
	switch kind {
	case reflect.Map, reflect.Struct:
		var (
			fields         []string
			dataMap        = ConvertDataForTableRecord(data)
			counterHandler = func(column string, counter Counter) {
				if counter.Value != 0 {
					var (
						column    = c.QuoteWord(column)
						columnRef = c.QuoteWord(counter.Field)
						columnVal = counter.Value
						operator  = "+"
					)
					if columnVal < 0 {
						operator = "-"
						columnVal = -columnVal
					}
					fields = append(fields, fmt.Sprintf("%s=%s%s?", column, columnRef, operator))
					params = append(params, columnVal)
				}
			}
		)

		for k, v := range dataMap {
			switch value := v.(type) {
			case *Counter:
				counterHandler(k, *value)

			case Counter:
				counterHandler(k, value)

			default:
				if s, ok := v.(Raw); ok {
					fields = append(fields, c.QuoteWord(k)+"="+gconv.String(s))
				} else {
					fields = append(fields, c.QuoteWord(k)+"=?")
					params = append(params, v)
				}
			}
		}
		updates = strings.Join(fields, ",")

	default:
		updates = gconv.String(data)
	}
	if len(updates) == 0 {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "data cannot be empty")
	}
	if len(params) > 0 {
		args = append(params, args...)
	}
	// If no link passed, it then uses the master link.
	if link == nil {
		if link, err = c.MasterLink(); err != nil {
			return nil, err
		}
	}
	return c.db.DoExec(ctx, link, fmt.Sprintf("UPDATE %s SET %s%s", table, updates, condition), args...)
}

// Delete does "DELETE FROM ... " statement for the table.
//
// The parameter `condition` can be type of string/map/gmap/slice/struct/*struct, etc.
// It is commonly used with parameter `args`.
// Eg:
// "uid=10000",
// "uid", 10000
// "money>? AND name like ?", 99999, "vip_%"
// "status IN (?)", g.Slice{1,2,3}
// "age IN(?,?)", 18, 50
// User{ Id : 1, UserName : "john"}.
func (c *Core) Delete(ctx context.Context, table string, condition interface{}, args ...interface{}) (result sql.Result, err error) {
	return c.Model(table).Ctx(ctx).Where(condition, args...).Delete()
}

// DoDelete does "DELETE FROM ... " statement for the table.
// This function is usually used for custom interface definition, you do not need call it manually.
func (c *Core) DoDelete(ctx context.Context, link Link, table string, condition string, args ...interface{}) (result sql.Result, err error) {
	if link == nil {
		if link, err = c.MasterLink(); err != nil {
			return nil, err
		}
	}
	table = c.QuotePrefixTableName(table)
	return c.db.DoExec(ctx, link, fmt.Sprintf("DELETE FROM %s%s", table, condition), args...)
}

// MarshalJSON implements the interface MarshalJSON for json.Marshal.
// MarshalJSON实现json.Marshal的接口MarshalJSON
// It just returns the pointer address.
// 它只返回指针地址。
//
// Note that this interface implements mainly for workaround for a json infinite loop bug
// of Golang version < v1.14.
// 请注意，此接口主要用于解决Golang版本<v1.14的json无限循环错误。
func (c *Core) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`%+v`, c)), nil
}

// writeSqlToLogger outputs the Sql object to logger.
// writeSqlToLogger将Sql对象输出到记录器。
// It is enabled only if configuration "debug" is true.
// 仅当配置“调试”为真时才启用该选项。
func (c *Core) writeSqlToLogger(ctx context.Context, sql *Sql) {
	var (
		sqlTypeKey       string
		transactionIdStr string
	)
	switch sql.Type {
	case sqlTypeQueryContext:
		sqlTypeKey = `rows`
	default:
		sqlTypeKey = `rows`
	}
	if sql.IsTransaction {
		if v := ctx.Value(transactionIdForLoggerCtx); v != nil {
			transactionIdStr = fmt.Sprintf(`[txid:%d] `, v.(uint64))
		}
	}
	s := fmt.Sprintf(
		"[%3d ms] [%s] [%s:%-3d] %s%s",
		sql.End-sql.Start, sql.Group, sqlTypeKey, sql.RowsAffected, transactionIdStr, sql.Format,
	)
	if sql.Error != nil {
		s += "\nError: " + sql.Error.Error()
		c.logger.Error(ctx, s)
	} else {
		c.logger.Debug(ctx, s)
	}
}

// HasTable determine whether the table name exists in the database.
// 确定数据库中是否存在表名。
func (c *Core) HasTable(name string) (bool, error) {
	tableList, err := c.db.Tables(c.GetCtx())
	if err != nil {
		return false, err
	}
	for _, table := range tableList {
		if table == name {
			return true, nil
		}
	}
	return false, nil
}

// isSoftCreatedFieldName checks and returns whether given filed name is an automatic-filled created time.
// 检查并返回给定的文件名是否为自动填充创建时间。
func (c *Core) isSoftCreatedFieldName(fieldName string) bool {
	if fieldName == "" {
		return false
	}
	if config := c.db.GetConfig(); config.CreatedAt != "" {
		if utils.EqualFoldWithoutChars(fieldName, config.CreatedAt) {
			return true
		}
		return gstr.InArray(append([]string{config.CreatedAt}, createdFiledNames...), fieldName)
	}
	for _, v := range createdFiledNames {
		if utils.EqualFoldWithoutChars(fieldName, v) {
			return true
		}
	}
	return false
}
