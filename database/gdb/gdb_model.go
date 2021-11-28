// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gdb

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

// Model is core struct implementing the DAO for ORM.
// 是为 ORM 实现 DAO 的核心结构。
type Model struct {
	// 底层数据库接口
	db            DB                 // Underlying DB interface.
	// 底层TX事务相关接口
	tx            *TX                // Underlying TX interface.
	// rawSql 是原始 SQL 字符串，它标记了一个基于原始 SQL 的模型而不是基于表的模型。
	rawSql        string             // rawSql is the raw SQL string which marks a raw SQL based Model not a table based Model.
	schema        string             // Custom database schema.
	// 标记在主站或从站上操作。
	linkType      int                // Mark for operation on master or slave.
	// 模型初始化时的表名。
	tablesInit    string             // Table names when model initialization.
	// 操作表名，可以是多个表名和别名，如：“user”、“user u”、“user u、user_detail ud”。
	tables        string             // Operation table names, which can be more than one table names and aliases, like: "user", "user u", "user u, user_detail ud".
	// 操作字段，多个字段使用字符','连接。
	fields        string             // Operation fields, multiple fields joined using char ','.
	// 排除操作字段，多个字段使用字符','连接。
	fieldsEx      string             // Excluded operation fields, multiple fields joined using char ','.
	// With 功能的参数。
	withArray     []interface{}      // Arguments for With feature.
	// 对结构中具有“with”标记的所有对象启用模型关联操作。
	withAll       bool               // Enable model association operations on all objects that have "with" tag in the struct.
	// sql 的额外自定义参数，这些参数在 sql 提交给底层驱动程序之前添加到参数之前。
	extraArgs     []interface{}      // Extra custom arguments for sql, which are prepended to the arguments before sql committed to underlying driver.
	// where 操作的条件字符串。
	whereHolder   []ModelWhereHolder // Condition strings for where operation.
	// 用于“group by”语句。
	groupBy       string             // Used for "group by" statement.
	orderBy       string             // Used for "order by" statement.
	having        []interface{}      // Used for "having..." statement.
	start         int                // Used for "select ... start, limit ..." statement.
	limit         int                // Used for "select ... start, limit ..." statement.
	// 额外操作功能的选项。
	option        int                // Option for extra operation features.
	// 某些数据库语法的偏移量语句。
	offset        int                // Offset statement for some databases grammar.
	// 操作数据，可以是map/[]map/struct/*struct/string等类型。
	data          interface{}        // Data for operation, which can be type of map/[]map/struct/*struct/string, etc.
	// 批插入/替换/保存操作的批号。
	batch         int                // Batch number for batch Insert/Replace/Save operations.
	// 根据表的字段过滤数据和 where 键值对。
	filter        bool               // Filter data and where key-value pairs according to the fields of the table.
	// 强制查询仅返回不同的结果。
	distinct      string             // Force the query to only return distinct results.
	// 锁定更新或共享锁定。
	lockInfo      string             // Lock for update or in shared lock.
	// 启用 sql 结果缓存功能。
	cacheEnabled  bool               // Enable sql result cache feature.
	// 查询语句的缓存选项。
	cacheOption   CacheOption        // Cache option for query statement.
	// 在选择/删除操作时禁用软删除功能。
	unscoped      bool               // Disables soft deleting features when select/delete operations.
	// 如果为真，它会在操作完成时克隆并返回一个新的模型对象；否则它会改变当前模型的属性。
	safe          bool               // If true, it clones and returns a new model object whenever operation done; or else it changes the attribute of current model.
	// onDuplicate 用于 ON "DUPLICATE KEY UPDATE" 语句。
	onDuplicate   interface{}        // onDuplicate is used for ON "DUPLICATE KEY UPDATE" statement.
	// onDuplicateEx 用于在“DUPLICATE KEY UPDATE”语句中排除某些列。
	onDuplicateEx interface{}        // onDuplicateEx is used for excluding some columns ON "DUPLICATE KEY UPDATE" statement.
}

// ModelHandler is a function that handles given Model and returns a new Model that is custom modified.
// 是一个处理给定模型并返回自定义修改的新模型的函数。
type ModelHandler func(m *Model) *Model

// ChunkHandler is a function that is used in function Chunk, which handles given Result and error.
// 是在函数 Chunk 中使用的函数，它处理给定的结果和错误
// It returns true if it wants to continue chunking, or else it returns false to stop chunking.
// 如果要继续分块，则返回 true，否则返回 false 以停止分块。
type ChunkHandler func(result Result, err error) bool

// ModelWhereHolder is the holder for where condition preparing.
// 是 where 条件准备的持有者。
type ModelWhereHolder struct {
	Operator int           // Operator for this holder.
	Where    interface{}   // Where parameter, which can commonly be type of string/map/struct.
	Args     []interface{} // Arguments for where parameter.
	Prefix   string        // Field prefix, eg: "user.", "order.".
}

const (
	linkTypeMaster           = 1
	linkTypeSlave            = 2
	whereHolderOperatorWhere = 1
	whereHolderOperatorAnd   = 2
	whereHolderOperatorOr    = 3
	defaultFields            = "*"
)

// Model creates and returns a new ORM model from given schema.
// 从给定的模式创建并返回新的ORM模型。
// The parameter `tableNameQueryOrStruct` can be more than one table names, and also alias name, like:
// 可以是多个表名，也可以是别名
// 1. Model names:
//    db.Model("user")
//    db.Model("user u")
//    db.Model("user, user_detail")
//    db.Model("user u, user_detail ud")
// 2. Model name with alias:
//    db.Model("user", "u")
// 3. Model name with sub-query:
//    db.Model("? AS a, ? AS b", subQuery1, subQuery2)
// DB.Model具体实现
func (c *Core) Model(tableNameQueryOrStruct ...interface{}) *Model {
	fmt.Println("gdb_model.go 123 Model = ",tableNameQueryOrStruct[0])
	var (
		tableStr  string
		tableName string
		extraArgs []interface{}
	)
	// Model creation with sub-query.
	if len(tableNameQueryOrStruct) > 1 {
		conditionStr := gconv.String(tableNameQueryOrStruct[0])
		if gstr.Contains(conditionStr, "?") {
			tableStr, extraArgs = formatWhereHolder(c.db, formatWhereHolderInput{
				Where:     conditionStr,
				Args:      tableNameQueryOrStruct[1:],
				OmitNil:   false,
				OmitEmpty: false,
				Schema:    "",
				Table:     "",
			})
		}
	}
	// Normal model creation.
	if tableStr == "" {
		tableNames := make([]string, len(tableNameQueryOrStruct))
		for k, v := range tableNameQueryOrStruct {
			if s, ok := v.(string); ok {
				tableNames[k] = s
			} else if tableName = getTableNameFromOrmTag(v); tableName != "" {
				tableNames[k] = tableName
			}
		}
		if len(tableNames) > 1 {
			tableStr = fmt.Sprintf(
				`%s AS %s`, c.QuotePrefixTableName(tableNames[0]), c.QuoteWord(tableNames[1]),
			)
		} else if len(tableNames) == 1 {
			tableStr = c.QuotePrefixTableName(tableNames[0])
		}
	}
	m := &Model{
		db:         c.db,
		tablesInit: tableStr,
		tables:     tableStr,
		fields:     defaultFields,
		start:      -1,
		offset:     -1,
		filter:     true,
		extraArgs:  extraArgs,
	}
	if defaultModelSafe {
		m.safe = true
	}
	fmt.Println("gdb_model.go 174")
	return m
}

// Raw creates and returns a model based on a raw sql not a table.
// Example:
//     db.Raw("SELECT * FROM `user` WHERE `name` = ?", "john").Scan(&result)
// Raw基于原始sql而不是表创建并返回模型
func (c *Core) Raw(rawSql string, args ...interface{}) *Model {
	model := c.Model()
	model.rawSql = rawSql
	model.extraArgs = args
	return model
}

// Raw sets current model as a raw sql model.
// 将当前模型设置为原始 sql 模型。
// Example:
//     db.Raw("SELECT * FROM `user` WHERE `name` = ?", "john").Scan(&result)
// See Core.Raw.
func (m *Model) Raw(rawSql string, args ...interface{}) *Model {
	model := m.db.Raw(rawSql, args...)
	model.db = m.db
	model.tx = m.tx
	return model
}

func (tx *TX) Raw(rawSql string, args ...interface{}) *Model {
	return tx.Model().Raw(rawSql, args...)
}

// With creates and returns an ORM model based on metadata of given object.
// 基于给定对象的元数据创建并返回一个 ORM 模型。
func (c *Core) With(objects ...interface{}) *Model {
	return c.db.Model().With(objects...)
}

// Model acts like Core.Model except it operates on transaction.
// 模型的行为类似于 Core.Model，只是它对事务进行操作。
// See Core.Model.
func (tx *TX) Model(tableNameQueryOrStruct ...interface{}) *Model {
	model := tx.db.Model(tableNameQueryOrStruct...)
	model.db = tx.db
	model.tx = tx
	return model
}

// With acts like Core.With except it operates on transaction.
// With 就像 Core.With 一样，除了它对事务进行操作。
// See Core.With.
func (tx *TX) With(object interface{}) *Model {
	return tx.Model().With(object)
}

// Ctx sets the context for current operation.
// 设置当前操作的上下文
func (m *Model) Ctx(ctx context.Context) *Model {
	if ctx == nil {
		return m
	}
	model := m.getModel()
	model.db = model.db.Ctx(ctx)
	if m.tx != nil {
		model.tx = model.tx.Ctx(ctx)
	}
	return model
}

// GetCtx returns the context for current Model.
// GetCtx 返回当前模型的上下文。
// It returns `context.Background()` is there's no context previously set.
// 如果之前没有设置上下文，它返回`context.Background()`。
func (m *Model) GetCtx() context.Context {
	if m.tx != nil && m.tx.ctx != nil {
		return m.tx.ctx
	}
	return m.db.GetCtx()
}

// As sets an alias name for current table.
// 为当前表设置别名。
func (m *Model) As(as string) *Model {
	if m.tables != "" {
		model := m.getModel()
		split := " JOIN "
		if gstr.ContainsI(model.tables, split) {
			// For join table.
			array := gstr.Split(model.tables, split)
			array[len(array)-1], _ = gregex.ReplaceString(`(.+) ON`, fmt.Sprintf(`$1 AS %s ON`, as), array[len(array)-1])
			model.tables = gstr.Join(array, split)
		} else {
			// For base table.
			model.tables = gstr.TrimRight(model.tables) + " AS " + as
		}
		return model
	}
	return m
}

// DB sets/changes the db object for current operation.
// 设置/更改当前操作的 db 对象。
func (m *Model) DB(db DB) *Model {
	model := m.getModel()
	model.db = db
	return model
}

// TX sets/changes the transaction for current operation.
// 设置/更改当前操作的事务。
func (m *Model) TX(tx *TX) *Model {
	model := m.getModel()
	model.db = tx.db
	model.tx = tx
	return model
}

// Schema sets the schema for current operation.
// 设置当前操作的模式。
func (m *Model) Schema(schema string) *Model {
	model := m.getModel()
	model.schema = schema
	return model
}

// Clone creates and returns a new model which is a clone of current model.
// 创建并返回一个新模型，它是当前模型的克隆。
// Note that it uses deep-copy for the clone.
// 请注意，它对克隆使用深拷贝。
func (m *Model) Clone() *Model {
	newModel := (*Model)(nil)
	if m.tx != nil {
		newModel = m.tx.Model(m.tablesInit)
	} else {
		newModel = m.db.Model(m.tablesInit)
	}
	*newModel = *m
	// Shallow copy slice attributes.
	if n := len(m.extraArgs); n > 0 {
		newModel.extraArgs = make([]interface{}, n)
		copy(newModel.extraArgs, m.extraArgs)
	}
	if n := len(m.whereHolder); n > 0 {
		newModel.whereHolder = make([]ModelWhereHolder, n)
		copy(newModel.whereHolder, m.whereHolder)
	}
	if n := len(m.withArray); n > 0 {
		newModel.withArray = make([]interface{}, n)
		copy(newModel.withArray, m.withArray)
	}
	return newModel
}

// Master marks the following operation on master node.
// 标记主节点上的以下操作。
func (m *Model) Master() *Model {
	model := m.getModel()
	model.linkType = linkTypeMaster
	return model
}

// Slave marks the following operation on slave node.
// Note that it makes sense only if there's any slave node configured.
func (m *Model) Slave() *Model {
	model := m.getModel()
	model.linkType = linkTypeSlave
	return model
}

// Safe marks this model safe or unsafe. If safe is true, it clones and returns a new model object
// 标记此模型安全或不安全。如果 safe 为真，它会克隆并返回一个新的模型对象
// whenever the operation done, or else it changes the attribute of current model.
// 每当操作完成，否则它会更改当前模型的属性。
func (m *Model) Safe(safe ...bool) *Model {
	if len(safe) > 0 {
		m.safe = safe[0]
	} else {
		m.safe = true
	}
	return m
}

// Args sets custom arguments for model operation.
// 为模型操作设置自定义参数。
func (m *Model) Args(args ...interface{}) *Model {
	model := m.getModel()
	model.extraArgs = append(model.extraArgs, args)
	return model
}

// Handler calls each of `handlers` on current Model and returns a new Model.
// 调用当前模型上的每个“处理程序”并返回一个新模型。
// ModelHandler is a function that handles given Model and returns a new Model that is custom modified.
// ModelHandler 是一个处理给定模型并返回自定义修改的新模型的函数。
func (m *Model) Handler(handlers ...ModelHandler) *Model {
	model := m.getModel()
	for _, handler := range handlers {
		model = handler(model)
	}
	return model
}
