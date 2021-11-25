1. gdb.go

定义抽象类-DB - 定义ORM操作的接口（Query、insert、GetOne）
定义抽象类-Driver - 是将sql驱动程序集成到包gdb中的接口（只有一个待实现的new返回一个DB实现对象）
New\Interface\getCfg\ - 获取对象获取全局配置

2. gdb_core.go

