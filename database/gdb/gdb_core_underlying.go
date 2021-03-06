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

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/internal/intlog"
	"github.com/gogf/gf/v2/os/gtime"
)

// Query commits one query SQL to underlying driver and returns the execution result.
// 向底层驱动程序提交一个查询SQL并返回执行结果。
// It is most commonly used for data querying.
func (c *Core) Query(ctx context.Context, sql string, args ...interface{}) (result Result, err error) {
	return c.db.DoQuery(ctx, nil, sql, args...)
}

// DoQuery commits the sql string and its arguments to underlying driver
// 将sql字符串及其参数提交给基础驱动程序
// through given link object and returns the execution result.
// 将sql字符串及其参数提交给基础驱动程序
func (c *Core) DoQuery(ctx context.Context, link Link, sql string, args ...interface{}) (result Result, err error) {
	// Transaction checks.
	fmt.Println("gdb_core_underlying 32 ")
	if link == nil {
		if tx := TXFromCtx(ctx, c.db.GetGroup()); tx != nil {
			// 首先，从上下文中检查和检索交易链接。
			// Firstly, check and retrieve transaction link from context.
			link = &txLink{tx.tx}
		} else if link, err = c.SlaveLink(); err != nil {
			// 在这里执行 - open - dial
			// 否则它会从主节点创建一个。
			// Or else it creates one from master node.
			return nil, err
		}
	} else if !link.IsTransaction() {
		// If current link is not transaction link, it checks and retrieves transaction from context.
		// 如果当前链接不是事务链接，它会检查并从上下文中检索事务。
		if tx := TXFromCtx(ctx, c.db.GetGroup()); tx != nil {
			link = &txLink{tx.tx}
		}
	}
	fmt.Println("gdb_core_underlying 51 ",reflect.TypeOf(link))
	// db_core_underlying 51  *gdb.dbLink
	if c.GetConfig().QueryTimeout > 0 {
		ctx, _ = context.WithTimeout(ctx, c.GetConfig().QueryTimeout)
	}

	// Link execution.
	sql, args = formatSql(sql, args)
	// *gdb.DriverMysql
	fmt.Println(reflect.TypeOf(c.db))
	// 这是一个空func
	sql, args, err = c.db.DoCommit(ctx, link, sql, args)
	if err != nil {
		return nil, err
	}
	fmt.Println("underlying 64 = ",sql)
	// underlying 64 =  SELECT * FROM `user` LIMIT 0,10

	mTime1 := gtime.TimestampMilli()
	fmt.Println("gdb_cire_underlying 71 ")
	rows, err := link.QueryContext(ctx, sql, args...)
	// github.com/go-mysql.QueryContext
	mTime2 := gtime.TimestampMilli()
	if err == nil {
		result, err = c.convertRowsToResult(ctx, rows)
	}
	sqlObj := &Sql{
		Sql:           sql,
		Type:          sqlTypeQueryContext,
		Args:          args,
		Format:        FormatSqlWithArgs(sql, args),
		Error:         err,
		Start:         mTime1,
		End:           mTime2,
		Group:         c.db.GetGroup(),
		IsTransaction: link.IsTransaction(),
		RowsAffected:  int64(result.Len()),
	}
	// Tracing and logging.
	c.addSqlToTracing(ctx, sqlObj)
	if c.db.GetDebug() {
		c.writeSqlToLogger(ctx, sqlObj)
	}
	if err == nil {
		return result, nil
	} else {
		err = formatError(err, sql, args...)
	}
	return nil, err
}

// Exec commits one query SQL to underlying driver and returns the execution result.
// Exec 向底层驱动提交一条查询 SQL 并返回执行结果。
// It is most commonly used for data inserting and updating.
// 它最常用于数据插入和更新。
func (c *Core) Exec(ctx context.Context, sql string, args ...interface{}) (result sql.Result, err error) {
	return c.db.DoExec(ctx, nil, sql, args...)
}

// DoExec commits the sql string and its arguments to underlying driver
// DoExec 将 sql 字符串及其参数提交给底层驱动程序
// through given link object and returns the execution result.
// 通过给定的链接对象并返回执行结果。
func (c *Core) DoExec(ctx context.Context, link Link, sql string, args ...interface{}) (result sql.Result, err error) {
	// Transaction checks.
	if link == nil {
		if tx := TXFromCtx(ctx, c.db.GetGroup()); tx != nil {
			// Firstly, check and retrieve transaction link from context.
			link = &txLink{tx.tx}
		} else if link, err = c.MasterLink(); err != nil {
			// Or else it creates one from master node.
			return nil, err
		}
	} else if !link.IsTransaction() {
		// If current link is not transaction link, it checks and retrieves transaction from context.
		if tx := TXFromCtx(ctx, c.db.GetGroup()); tx != nil {
			link = &txLink{tx.tx}
		}
	}

	if c.GetConfig().ExecTimeout > 0 {
		var cancelFunc context.CancelFunc
		ctx, cancelFunc = context.WithTimeout(ctx, c.GetConfig().ExecTimeout)
		defer cancelFunc()
	}

	// Link execution.
	sql, args = formatSql(sql, args)
	sql, args, err = c.db.DoCommit(ctx, link, sql, args)
	if err != nil {
		return nil, err
	}

	mTime1 := gtime.TimestampMilli()
	if !c.db.GetDryRun() {
		result, err = link.ExecContext(ctx, sql, args...)
	} else {
		result = new(SqlResult)
	}
	mTime2 := gtime.TimestampMilli()
	var rowsAffected int64
	if err == nil {
		rowsAffected, err = result.RowsAffected()
	}
	sqlObj := &Sql{
		Sql:           sql,
		Type:          sqlTypeExecContext,
		Args:          args,
		Format:        FormatSqlWithArgs(sql, args),
		Error:         err,
		Start:         mTime1,
		End:           mTime2,
		Group:         c.db.GetGroup(),
		IsTransaction: link.IsTransaction(),
		RowsAffected:  rowsAffected,
	}
	// Tracing and logging.
	c.addSqlToTracing(ctx, sqlObj)
	if c.db.GetDebug() {
		c.writeSqlToLogger(ctx, sqlObj)
	}
	return result, formatError(err, sql, args...)
}

// DoCommit is a hook function, which deals with the sql string before it's committed to underlying driver.
// The parameter `link` specifies the current database connection operation object. You can modify the sql
// string `sql` and its arguments `args` as you wish before they're committed to driver.
func (c *Core) DoCommit(ctx context.Context, link Link, sql string, args []interface{}) (newSql string, newArgs []interface{}, err error) {
	fmt.Println("gdb_core_underlying.go 177 ")
	return sql, args, nil
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the
// returned statement.
// The caller must call the statement's Close method
// when the statement is no longer needed.
//
// The parameter `execOnMaster` specifies whether executing the sql on master node,
// or else it executes the sql on slave node if master-slave configured.
func (c *Core) Prepare(ctx context.Context, sql string, execOnMaster ...bool) (*Stmt, error) {
	var (
		err  error
		link Link
	)
	if len(execOnMaster) > 0 && execOnMaster[0] {
		if link, err = c.MasterLink(); err != nil {
			return nil, err
		}
	} else {
		if link, err = c.SlaveLink(); err != nil {
			return nil, err
		}
	}
	return c.db.DoPrepare(ctx, link, sql)
}

// DoPrepare calls prepare function on given link object and returns the statement object.
func (c *Core) DoPrepare(ctx context.Context, link Link, sql string) (*Stmt, error) {
	// Transaction checks.
	if link == nil {
		if tx := TXFromCtx(ctx, c.db.GetGroup()); tx != nil {
			// Firstly, check and retrieve transaction link from context.
			link = &txLink{tx.tx}
		} else {
			// Or else it creates one from master node.
			var err error
			if link, err = c.MasterLink(); err != nil {
				return nil, err
			}
		}
	} else if !link.IsTransaction() {
		// If current link is not transaction link, it checks and retrieves transaction from context.
		if tx := TXFromCtx(ctx, c.db.GetGroup()); tx != nil {
			link = &txLink{tx.tx}
		}
	}

	if c.GetConfig().PrepareTimeout > 0 {
		// DO NOT USE cancel function in prepare statement.
		ctx, _ = context.WithTimeout(ctx, c.GetConfig().PrepareTimeout)
	}

	var (
		mTime1    = gtime.TimestampMilli()
		stmt, err = link.PrepareContext(ctx, sql)
		mTime2    = gtime.TimestampMilli()
		sqlObj    = &Sql{
			Sql:           sql,
			Type:          sqlTypePrepareContext,
			Args:          nil,
			Format:        FormatSqlWithArgs(sql, nil),
			Error:         err,
			Start:         mTime1,
			End:           mTime2,
			Group:         c.db.GetGroup(),
			IsTransaction: link.IsTransaction(),
		}
	)
	// Tracing and logging.
	c.addSqlToTracing(ctx, sqlObj)
	if c.db.GetDebug() {
		c.writeSqlToLogger(ctx, sqlObj)
	}
	return &Stmt{
		Stmt: stmt,
		core: c,
		link: link,
		sql:  sql,
	}, err
}

// convertRowsToResult converts underlying data record type sql.Rows to Result type.
func (c *Core) convertRowsToResult(ctx context.Context, rows *sql.Rows) (Result, error) {
	if rows == nil {
		return nil, nil
	}
	defer func() {
		if err := rows.Close(); err != nil {
			intlog.Error(ctx, err)
		}
	}()
	if !rows.Next() {
		return nil, nil
	}
	// Column names and types.
	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	var (
		columnTypes = make([]string, len(columns))
		columnNames = make([]string, len(columns))
	)
	for k, v := range columns {
		columnTypes[k] = v.DatabaseTypeName()
		columnNames[k] = v.Name()
	}
	var (
		values   = make([]interface{}, len(columnNames))
		result   = make(Result, 0)
		scanArgs = make([]interface{}, len(values))
	)
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for {
		if err = rows.Scan(scanArgs...); err != nil {
			return result, err
		}
		record := Record{}
		for i, value := range values {
			if value == nil {
				record[columnNames[i]] = gvar.New(nil)
			} else {
				record[columnNames[i]] = gvar.New(c.convertFieldValueToLocalValue(value, columnTypes[i]))
			}
		}
		result = append(result, record)
		if !rows.Next() {
			break
		}
	}
	return result, nil
}
