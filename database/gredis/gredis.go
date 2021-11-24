// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package gredis provides convenient client for redis server.
//
// Redis Client.
//
// Redis Commands Official: https://redis.io/commands
//
// Redis Chinese Documentation: http://redisdoc.com/
package gredis

import (
	"context"
	"crypto/tls"
	"fmt"
	"reflect"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gogf/gf/v2"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/internal/intlog"
	"github.com/gogf/gf/v2/internal/json"
	"github.com/gogf/gf/v2/internal/utils"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// 实现关系是

// redis{adapter} > RedisConn{do receive close}

// AdapterGoRedis > localAdapterGoRedisConn{do receive close}

const (
	errorNilRedis = `the Redis object is nil`
)
const (
	defaultPoolMaxIdle     = 10
	defaultPoolMaxActive   = 100
	defaultPoolIdleTimeout = 10 * time.Second
	defaultPoolWaitTimeout = 10 * time.Second
	defaultPoolMaxLifeTime = 30 * time.Second
)
const (
	DefaultGroupName = "default" // Default configuration group name.
)
const (
	tracingInstrumentName               = "github.com/gogf/gf/v2/database/gredis"
	tracingAttrRedisAddress             = "redis.address"
	tracingAttrRedisDb                  = "redis.db"
	tracingEventRedisExecution          = "redis.execution"
	tracingEventRedisExecutionCommand   = "redis.execution.command"
	tracingEventRedisExecutionCost      = "redis.execution.cost"
	tracingEventRedisExecutionArguments = "redis.execution.arguments"
)

// redis群组配置 - 在一个go程序实例之中-当前包下只有一个群组配置 - 线程安全集合
// 全局对象g对redis cfg手动set也是对这个进行操作
var (
	// Configuration groups.
	localConfigMap = gmap.NewStrAnyMap(true)
)

// Redis client.
// 带有适配器的redis客户端  - 抽象接口
// 所有的g.Redis返回的对象都是这个的实现的实例
// 其自身实现了 - Redis == {SetAdapter设置适配器\GetAdapter获取适配器\Conn返回适配器指向的conn\Do指向适配器执行\Close指向适配器执行}
type Redis struct {
	adapter Adapter
}

// Adapter is an interface for universal redis operations.
// redis适配器 - 传入的适配器实现接口
type Adapter interface {
	// Conn retrieves and returns a connection object for continuous operations.
	// Note that you should call Close function manually if you do not use this connection any further.
	Conn(ctx context.Context) (conn Conn, err error)

	// Close closes current redis client, closes its connection pool and releases all its related resources.
	Close(ctx context.Context) (err error)
}


// RedisConn is a connection of redis client.
// redis客户端的连接实现
// 包含有一个待实现redis抽象类
// RedisConn == {Conn(do reveive close) + redis(adapter(Conn Close))}
//  ---- redis基类的实现
// 其自身有func - {Do指向自己conn执行\Receive\Close\addTracingItem}
type RedisConn struct {
	conn  Conn
	redis *Redis
}


// Conn is an interface of a connection from universal redis client -  待实现的抽象接口-连接对象
// 来自universal redis客户端的连接接口
type Conn interface {
	// Do sends a command to the server and returns the received reply.
	// It uses json.Marshal for struct/slice/map type values before committing them to redis.
	Do(ctx context.Context, command string, args ...interface{}) (result *gvar.Var, err error)

	// Receive receives a single reply as gvar.Var from the Redis server.
	Receive(ctx context.Context) (result *gvar.Var, err error)

	// Close puts the connection back to connection pool.
	Close(ctx context.Context) (err error)
}

// Subscription received after a successful subscription to channel.
// 成功订阅频道后收到 - 结果集
type Subscription struct {
	Kind    string // Can be "subscribe", "unsubscribe", "psubscribe" or "punsubscribe".
	Channel string // Channel name we have subscribed to.
	Count   int    // Number of channels we are currently subscribed to.
}

// 带有发布订阅实现和适配器redis的连接 - 已经具体实现抽象接口的类
type localAdapterGoRedisConn struct {
	ps    *redis.PubSub
	redis *AdapterGoRedis
}

// AdapterGoRedis is an implement of Adapter using go-redis - 已经具体实现的类
// go-redis实现和cofig
type AdapterGoRedis struct {
	client redis.UniversalClient
	config *Config
}


// tracingItem holds the information for redis tracing.
// 保存用于redis跟踪的信息。
type tracingItem struct {
	err       error
	command   string
	args      []interface{}
	costMilli int64
}

// Message received as result of a PUBLISH command issued by another client.
// 作为另一个客户端发出的发布命令的结果接收
type Message struct {
	Channel      string
	Pattern      string
	Payload      string
	PayloadSlice []string
}


// Config is redis configuration.
type Config struct {
	Address         string        `json:"address"`         // It supports single and cluster redis server. Multiple addresses joined with char ','.
	Db              int           `json:"db"`              // Redis db.
	Pass            string        `json:"pass"`            // Password for AUTH.
	MinIdle         int           `json:"minIdle"`         // Minimum number of connections allowed to be idle (default is 0)
	MaxIdle         int           `json:"maxIdle"`         // Maximum number of connections allowed to be idle (default is 10)
	MaxActive       int           `json:"maxActive"`       // Maximum number of connections limit (default is 0 means no limit).
	MaxConnLifetime time.Duration `json:"maxConnLifetime"` // Maximum lifetime of the connection (default is 30 seconds, not allowed to be set to 0)
	IdleTimeout     time.Duration `json:"idleTimeout"`     // Maximum idle time for connection (default is 10 seconds, not allowed to be set to 0)
	WaitTimeout     time.Duration `json:"waitTimeout"`     // Timed out duration waiting to get a connection from the connection pool.
	DialTimeout     time.Duration `json:"dialTimeout"`     // Dial connection timeout for TCP.
	ReadTimeout     time.Duration `json:"readTimeout"`     // Read timeout for TCP.
	WriteTimeout    time.Duration `json:"writeTimeout"`    // Write timeout for TCP.
	MasterName      string        `json:"masterName"`      // Used in Redis Sentinel mode.
	TLS             bool          `json:"tls"`             // Specifies whether TLS should be used when connecting to the server.
	TLSSkipVerify   bool          `json:"tlsSkipVerify"`   // Disables server name verification when connecting over TLS.
	TLSConfig       *tls.Config   `json:"-"`               // TLS Config to use. When set TLS will be negotiated.
}

// New creates and returns a redis client.
// It creates a default redis adapter of go-redis.
// 返回redis实现对象
func New(config ...*Config) (*Redis, error) {
	if len(config) > 0 {
		return &Redis{adapter: NewAdapterGoRedis(config[0])}, nil
	}
	configFromGlobal, ok := GetConfig()
	if !ok {
		return nil, gerror.NewCode(
			gcode.CodeMissingConfiguration,
			`configuration not found for creating Redis client`,
		)
	}
	// 适配器模式 - 直接以一个已经实现的具体对象 - 作为抽象类
	// 传入Adapter适配器具体实现 - go-redis扩展类对象
	return &Redis{adapter: NewAdapterGoRedis(configFromGlobal)}, nil
}

// NewWithAdapter creates and returns a redis client with given adapter.
// 创建并返回具有给定适配器的redis客户端。
func NewWithAdapter(adapter Adapter) *Redis {
	return &Redis{adapter: adapter}
}

// 当前func直接调用go-redis为适配器实现的redis-client执行redis命令 - 最终传入go-redis对象执行go-redis的func
// Do sends a command to the server and returns the received reply.
// 向服务器发送命令并返回收到的回复
// It uses json.Marshal for struct/slice/map type values before committing them to redis.
// 在将struct/slice/map类型的值提交到redis之前，它使用json.Marshal处理这些值。
func (c *localAdapterGoRedisConn) Do(ctx context.Context, command string, args ...interface{}) (reply *gvar.Var, err error) {
	switch gstr.ToLower(command) {
	case `subscribe`:
		c.ps = c.redis.client.Subscribe(ctx, gconv.Strings(args)...)

	case `psubscribe`:
		c.ps = c.redis.client.PSubscribe(ctx, gconv.Strings(args)...)

	case `unsubscribe`:
		if c.ps != nil {
			err = c.ps.Unsubscribe(ctx, gconv.Strings(args)...)
		}

	case `punsubscribe`:
		if c.ps != nil {
			err = c.ps.PUnsubscribe(ctx, gconv.Strings(args)...)
		}

	default:
		arguments := make([]interface{}, len(args)+1)
		copy(arguments, []interface{}{command})
		copy(arguments[1:], args)
		cc := c.redis.client.Do(ctx, arguments...)
		fmt.Println("cc == ",cc)
		a,b := cc.Result()
		reply, err = c.resultToVar(a,b)
	}

	return
}

// Receive receives a single reply as gvar.Var from the Redis server.
// 当前知识转换了一下结果集为gvar格式的结果
// 从Redis服务器接收一个作为gvar.Var的回复
func (c *localAdapterGoRedisConn) Receive(ctx context.Context) (*gvar.Var, error) {
	if c.ps != nil {
		return c.resultToVar(c.ps.Receive(ctx))
	}
	return nil, nil
}

// Close closes current PubSub or puts the connection back to connection pool.
// 关闭当前PubSub或将连接放回连接池
func (c *localAdapterGoRedisConn) Close(ctx context.Context) error {
	if c.ps != nil {
		return c.ps.Close()
	}
	return nil
}

// resultToVar converts redis operation result to gvar.Var.
// 将redis操作结果转换为gvar.Var
func (c *localAdapterGoRedisConn) resultToVar(result interface{}, err error) (*gvar.Var, error) {
	if err == redis.Nil {
		err = nil
	}
	if err == nil {
		// 仍然类型断言
		switch v := result.(type) {
		case []byte:
			return gvar.New(string(v)), err

		case []interface{}:
			return gvar.New(gconv.Strings(v)), err

		case *redis.Message:
			result = &Message{
				Channel:      v.Channel,
				Pattern:      v.Pattern,
				Payload:      v.Payload,
				PayloadSlice: v.PayloadSlice,
			}

		case *redis.Subscription:
			result = &Subscription{
				Kind:    v.Kind,
				Channel: v.Channel,
				Count:   v.Count,
			}
		}
	}
	return gvar.New(result), err
}





// NewAdapterGoRedis creates and returns a redis adapter using go-redis.
// 使用go-redis创建并返回redis适配器。
func NewAdapterGoRedis(config *Config) *AdapterGoRedis {
	// 填充默认参数
	fillWithDefaultConfiguration(config)
	// 调用go-redis实例化对象获取
	aaa := &redis.UniversalOptions{
		Addrs:        gstr.SplitAndTrim(config.Address, ","),
		Password:     config.Pass,
		DB:           config.Db,
		MinIdleConns: config.MinIdle,
		MaxConnAge:   config.MaxConnLifetime,
		IdleTimeout:  config.IdleTimeout,
		PoolTimeout:  config.WaitTimeout,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		MasterName:   config.MasterName,
		TLSConfig:    config.TLSConfig,
	}
	fmt.Println("aaa ===> ",aaa)
	client := redis.NewUniversalClient(aaa)
	// 返回go-redis为实现的redis适配器
	return &AdapterGoRedis{
		client: client,
		config: config,
	}
}

// Close closes the redis connection pool, which will release all connections reserved by this pool.
// 关闭redis连接池，这将释放此池保留的所有连接
// It is commonly not necessary to call Close manually.
// 通常不需要手动调用Close。
func (r *AdapterGoRedis) Close(ctx context.Context) error {
	return r.client.Close()
}

// Conn retrieves and returns a connection object for continuous operations.
// etrieve并返回连续操作的连接对象
// Note that you should call Close function manually if you do not use this connection any further.
// 请注意，如果不再使用此连接，则应手动调用Close函数。
func (r *AdapterGoRedis) Conn(ctx context.Context) (Conn, error) {
	return &localAdapterGoRedisConn{
		redis: r,
	}, nil
}

// 填充默认参数
func fillWithDefaultConfiguration(config *Config) {
	// The MaxIdle is the most important attribute of the connection pool.
	// Only if this attribute is set, the created connections from client
	// can not exceed the limit of the server.
	if config.MaxIdle == 0 {
		config.MaxIdle = defaultPoolMaxIdle
	}
	// This value SHOULD NOT exceed the connection limit of redis server.
	if config.MaxActive == 0 {
		config.MaxActive = defaultPoolMaxActive
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = defaultPoolIdleTimeout
	}
	if config.WaitTimeout == 0 {
		config.WaitTimeout = defaultPoolWaitTimeout
	}
	if config.MaxConnLifetime == 0 {
		config.MaxConnLifetime = defaultPoolMaxLifeTime
	}
}


// SetConfig sets the global configuration for specified group.
// 设置指定组的全局配置。
// If `name` is not passed, it sets configuration for the default group name.
// 如果未传递，它将为默认组名设置配置。
func SetConfig(config *Config, name ...string) {
	group := DefaultGroupName
	if len(name) > 0 {
		group = name[0]
	}
	localConfigMap.Set(group, config)

	intlog.Printf(context.TODO(), `SetConfig for group "%s": %+v`, group, config)
}

// SetConfigByMap sets the global configuration for specified group with map.
// 使用映射设置指定组的全局配置。
// If `name` is not passed, it sets configuration for the default group name.
// 如果未传递，它将为默认组名设置配置。
func SetConfigByMap(m map[string]interface{}, name ...string) error {
	group := DefaultGroupName
	if len(name) > 0 {
		group = name[0]
	}
	config, err := ConfigFromMap(m)
	if err != nil {
		return err
	}
	localConfigMap.Set(group, config)
	return nil
}

// ConfigFromMap parses and returns config from given map.
// 解析并返回给定映射中的配置。
func ConfigFromMap(m map[string]interface{}) (config *Config, err error) {
	config = &Config{}
	if err = gconv.Scan(m, config); err != nil {
		err = gerror.NewCodef(gcode.CodeInvalidConfiguration, `invalid redis configuration: "%+v"`, m)
	}
	if config.DialTimeout < 1000 {
		config.DialTimeout = config.DialTimeout * time.Second
	}
	if config.WaitTimeout < 1000 {
		config.WaitTimeout = config.WaitTimeout * time.Second
	}
	if config.WriteTimeout < 1000 {
		config.WriteTimeout = config.WriteTimeout * time.Second
	}
	if config.ReadTimeout < 1000 {
		config.ReadTimeout = config.ReadTimeout * time.Second
	}
	if config.IdleTimeout < 1000 {
		config.IdleTimeout = config.IdleTimeout * time.Second
	}
	if config.MaxConnLifetime < 1000 {
		config.MaxConnLifetime = config.MaxConnLifetime * time.Second
	}
	return
}

// GetConfig returns the global configuration with specified group name.
// 返回具有指定组名的全局配置。
// If `name` is not passed, it returns configuration of the default group name.
// 如果未传递'name'，则返回默认组名的配置
func GetConfig(name ...string) (config *Config, ok bool) {
	group := DefaultGroupName
	if len(name) > 0 {
		group = name[0]
	}
	if v := localConfigMap.Get(group); v != nil {
		return v.(*Config), true
	}
	return &Config{}, false
}

// RemoveConfig removes the global configuration with specified group.
// 删除具有指定组的全局配置。
// If `name` is not passed, it removes configuration of the default group name.
// 如果未传递'name'，它将删除默认组名的配置。
func RemoveConfig(name ...string) {
	group := DefaultGroupName
	if len(name) > 0 {
		group = name[0]
	}
	localConfigMap.Remove(group)

	intlog.Printf(context.TODO(), `RemoveConfig: %s`, group)
}

// ClearConfig removes all configurations of redis.
// 删除redis的所有配置。
func ClearConfig() {
	localConfigMap.Clear()
}


var (
	localInstances = gmap.NewStrAnyMap(true)
)

// Instance returns an instance of redis client with specified group.
// 返回具有指定组的redis客户端实例。
// The `name` param is unnecessary, if `name` is not passed,
// 如果未传递'name'，则不需要'name'参数，
// it returns a redis instance with default configuration group.
// 它返回一个带有默认配置组的redis实例。
func Instance(name ...string) *Redis {
	group := DefaultGroupName
	if len(name) > 0 && name[0] != "" {
		group = name[0]
	}
	v := localInstances.GetOrSetFuncLock(group, func() interface{} {
		if config, ok := GetConfig(group); ok {
			r, err := New(config)
			if err != nil {
				intlog.Error(context.TODO(), err)
				return nil
			}
			return r
		}
		return nil
	})
	if v != nil {
		return v.(*Redis)
	}
	return nil
}


// RedisConn == {Do指向自己conn执行\Receive\Close\addTracingItem}
// Do sends a command to the server and returns the received reply.
// 向服务器发送命令并返回收到的回复。
// It uses json.Marshal for struct/slice/map type values before committing them to redis.
// 在将struct/slice/map类型的值提交到redis之前，它使用json.Marshal处理这些值。
func (c *RedisConn) Do(ctx context.Context, command string, args ...interface{}) (reply *gvar.Var, err error) {
	for k, v := range args {
		var (
			reflectInfo = utils.OriginTypeAndKind(v)
		)
		switch reflectInfo.OriginKind {
		case
			reflect.Struct,
			reflect.Map,
			reflect.Slice,
			reflect.Array:
			// Ignore slice type of: []byte.
			if _, ok := v.([]byte); !ok {
				if args[k], err = json.Marshal(v); err != nil {
					return nil, err
				}
			}
		}
	}
	timestampMilli1 := gtime.TimestampMilli()
	if len(args)>2{
	}
	reply, err = c.conn.Do(ctx, command, args...)
	timestampMilli2 := gtime.TimestampMilli()

	// Tracing.
	c.addTracingItem(ctx, &tracingItem{
		err:       err,
		command:   command,
		args:      args,
		costMilli: timestampMilli2 - timestampMilli1,
	})
	return
}

// Receive receives a single reply as gvar.Var from the Redis server.
// Receive从Redis服务器接收一个作为gvar.Var的回复。
func (c *RedisConn) Receive(ctx context.Context) (*gvar.Var, error) {
	return c.conn.Receive(ctx)
}

// Close puts the connection back to connection pool.
// 关闭将连接放回连接池。
func (c *RedisConn) Close(ctx context.Context) error {
	return c.conn.Close(ctx)
}

// addTracingItem checks and adds redis tracing information to OpenTelemetry.
// 检查并向OpenTelemetry添加redis跟踪信息
func (c *RedisConn) addTracingItem(ctx context.Context, item *tracingItem) {
	if !gtrace.IsTracingInternal() || !gtrace.IsActivated(ctx) {
		return
	}
	tr := otel.GetTracerProvider().Tracer(
		tracingInstrumentName,
		trace.WithInstrumentationVersion(gf.VERSION),
	)
	if ctx == nil {
		ctx = context.Background()
	}
	_, span := tr.Start(ctx, "Redis."+item.command, trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	if item.err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf(`%+v`, item.err))
	}

	span.SetAttributes(gtrace.CommonLabels()...)

	if adapter, ok := c.redis.GetAdapter().(*AdapterGoRedis); ok {
		span.SetAttributes(
			attribute.String(tracingAttrRedisAddress, adapter.config.Address),
			attribute.Int(tracingAttrRedisDb, adapter.config.Db),
		)
	}

	jsonBytes, _ := json.Marshal(item.args)
	span.AddEvent(tracingEventRedisExecution, trace.WithAttributes(
		attribute.String(tracingEventRedisExecutionCommand, item.command),
		attribute.String(tracingEventRedisExecutionCost, fmt.Sprintf(`%d ms`, item.costMilli)),
		attribute.String(tracingEventRedisExecutionArguments, string(jsonBytes)),
	))
}

// Redis == {SetAdapter设置适配器\GetAdapter获取适配器\Conn返回适配器指向的conn\Do指向适配器执行\Close指向适配器执行}
// SetAdapter sets custom adapter for current redis client.
// 为当前redis客户端设置自定义适配器。
func (r *Redis) SetAdapter(adapter Adapter) {
	if r == nil {
		return
	}
	r.adapter = adapter
}

// GetAdapter returns the adapter that is set in current redis client.
// 返回在当前redis客户端中设置的适配器。
func (r *Redis) GetAdapter() Adapter {
	if r == nil {
		return nil
	}
	return r.adapter
}

// Conn retrieves and returns a connection object for continuous operations.
// 检索并返回连续操作的连接对象。
// Note that you should call Close function manually if you do not use this connection any further.
// 请注意，如果不再使用此连接，则应手动调用Close函数。
func (r *Redis) Conn(ctx context.Context) (*RedisConn, error) {
	if r == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, errorNilRedis)
	}
	conn, err := r.adapter.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return &RedisConn{
		conn:  conn,
		redis: r,
	}, nil
}

// Do sends a command to the server and returns the received reply.
// 向服务器发送命令并返回收到的回复。
// It uses json.Marshal for struct/slice/map type values before committing them to redis.
// 在将struct/slice/map类型的值提交到redis之前，它使用json.Marshal处理这些值。
func (r *Redis) Do(ctx context.Context, command string, args ...interface{}) (*gvar.Var, error) {
	if r == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, errorNilRedis)
	}
	conn, err := r.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := conn.Close(ctx); err != nil {
			intlog.Error(ctx, err)
		}
	}()
	return conn.Do(ctx, command, args...)
}

// Close closes current redis client, closes its connection pool and releases all its related resources.
// 关闭当前redis客户端，关闭其连接池并释放其所有相关资源。
func (r *Redis) Close(ctx context.Context) error {
	if r == nil {
		return gerror.NewCode(gcode.CodeInvalidParameter, errorNilRedis)
	}
	return r.adapter.Close(ctx)
}

// String converts current object to a readable string.
// 将当前对象转换为可读字符串。
func (m *Subscription) String() string {
	return fmt.Sprintf("%s: %s", m.Kind, m.Channel)
}
