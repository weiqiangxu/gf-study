适配器模式

解析gredis封装

基础类：

redis\redisConn

redis:其中就一个adapter，有对应的Conn（返回conn对象）、close关闭连接两个抽象func - 自身有实现 - 均指向自身的conn或adapter执行;

redisConn:继承了Redis的方法（需要实现redis抽象func）、增加待实现func有conn.Do、conn.Receive、conn.Close

适配器:

localAdapterGoRedisConn\AdapterGoRedis

具体实现的适配器对抽象类redis的适配器有具体实现 - 传入go-redis的对象 - 已经实现对应的方法;


localAdapterGoRedisConn实现了Receive、Do、Close三个func

AdapterGoRedis实现了Close、Conn两个func



----- 

理解：当New的时候将 *AdapterGoRedis 作为适配器用于redis基础类的实例化, redis已经实现的func都是依赖于该适配器执行

后续如果redis更改或者直接弃用redis换成其他的缓存中间件比如memcache之类,将起对应的适配器实例化的对象作为适配器传入Redis即可

