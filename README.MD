# GoFrame

GOFrame源码解析

简化代码

找出设计巧妙的地方

以设计模式的角度读懂代码


# 终极奥义

搬砖过来封装一套属于自己的框架 

宝宝再也不用担心框架更新后一堆func undefined～～～ 


# 各个路径解析

\frame\g  -  这个包提供了常用的公用函数（IsEmpty...）以及全局对象获取（以key为idx从各个包package下的var(线程安全集合)之中获取对象）

\frame\gins - 提供redis、config、database等全局对象的获取(会加载配置然后初始化)

\container\ - 特殊diy数据结构（集合、链表、gpool、gmap等）


\util - 提供工具包(gconv格式转换工具、grand随机数生成工具、、、)

\os - 系统相关工具包-gcron定时器、gcmd命令行包、