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

\databases\gdb -- 数据库ORM



# 为什么要解析框架源码，我问你几个问题

1. goframe的文件配置热更新的吗？如何实现的，带来的额外消耗如何

2. redis的文档好空白，想要一些集群相关的功能，我该如何扩展

3. db连接对象，在MySQL端短连接还是长连接,从http进去开协程，处理一个请求，期间连接redis、mysql等资源是直接释放吗还是不停复用呢

4. goframe的各个垃圾回收怎么样，是否存在运行久了就会变得内存爆满的问题


# 结语

我不会直接拼接http请求、https请求各个握手之间的、也没处理过udp、tcp包丢包的问题，但是我想保证我的go程序这一块，是ok的