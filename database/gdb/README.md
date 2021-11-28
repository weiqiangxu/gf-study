# GDB层级

gdb_core

gdb_driver

gdb_model

gdb_type_result


# 分组简介

DB struct{}

Core struct{ DB }

Model struct{ DB }



1. gdb.go

定义抽象类-DB - 定义ORM操作的接口（Query、insert、GetOne）
定义抽象类-Driver - 是将sql驱动程序集成到包gdb中的接口（只有一个待实现的new返回一个DB实现对象）
New\Interface\getCfg\ - 获取对象获取全局配置

2. gdb_core.go


3. core对象是一个通用的(group\debug\config...)

driver-mysql是最终的实现；Tables表格、TableFields表格字段获取


和redis差不多 - 使用传入的db对象执行对一等的func - 比如 gdb_model_delete

4. gdb_model_delete -- *Model.Delete

*Core.Model(table).Ctx(ctx).Where(condition, args...).Delete()

database/sql - 对象   - 最终指向github.com/go-sql-driver/mysql  - 其他官方包实现的func

gdb.sql.getSqlDb()获取database/sql的*DB;

如何将gihub.com/go-sql-driver注册到database/sql的呢?

在import第三方包github.com/go-sql的时候init func

通过调用database/sql.Register func将第三方包的mysqlDriver传递进入database/sql

5. 获取mysql连接对象的在gdb.go/getSqlDb func，其中上层是 Slave/Master;

而 All func之中指向gdb_core_underlying.go/DoQuery 才会通过gdb_core_utility.go/SlaveLink的func执行Open连接


