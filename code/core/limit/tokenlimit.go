package limit

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	xrate "golang.org/x/time/rate"
)

/*

-- 返回是否可以活获得预期的token
每秒生成token数量即token生成速度
local rate = tonumber(ARGV[1])
-- 桶容量
local capacity = tonumber(ARGV[2])
-- 当前时间戳
local now = tonumber(ARGV[3])
-- 当前请求token数量
local requested = tonumber(ARGV[4])

-- fill_time：填满 token_bucket 需要多久
local fill_time = capacity/rate
-- 向下取整,ttl为填满时间的2倍
local ttl = math.floor(fill_time*2)
-- 当前时间桶容量
-- 获取目前 token_bucket 中剩余 token 数
-- 如果是第一次进入，则设置 token_bucket 数量为 令牌桶最大值
local last_tokens = tonumber(redis.call("get", KEYS[1]))
-- 如果当前桶容量为0,说明是第一次进入,则默认容量为桶的最大容量
if last_tokens == nil then
    last_tokens = capacity
end

-- 上一次更新 token_bucket 的时间
local last_refreshed = tonumber(redis.call("get", KEYS[2]))
-- 第一次进入则设置刷新时间为0
if last_refreshed == nil then
    last_refreshed = 0
end

-- 距离上次请求的时间跨度
local delta = math.max(0, now-last_refreshed)
-- 通过当前时间与上一次更新时间的跨度，以及生产token的速率，计算出新的token数
-- 如果超过 max_burst，多余生产的token会被丢弃
local filled_tokens = math.min(capacity, last_tokens+(delta*rate))
-- 本次请求token数量是否足够
local allowed = filled_tokens >= requested
-- 桶剩余数量
local new_tokens = filled_tokens
-- 允许本次token申请,计算剩余数量
if allowed then
    new_tokens = filled_tokens - requested
end

-- 更新新的token数，以及更新时间
-- 设置剩余token数量
redis.call("setex", KEYS[1], ttl, new_tokens)
--设置刷新时间
redis.call("setex", KEYS[2], ttl, now)

return allowed

*/

const (
	// to be compatible with aliyun redis, we cannot use `local key = KEYS[1]` to reuse the key
	// KEYS[1] as tokens_key
	// KEYS[2] as timestamp_key
	script = `local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local fill_time = capacity/rate
local ttl = math.floor(fill_time*2)
local last_tokens = tonumber(redis.call("get", KEYS[1]))
if last_tokens == nil then
    last_tokens = capacity
end

local last_refreshed = tonumber(redis.call("get", KEYS[2]))
if last_refreshed == nil then
    last_refreshed = 0
end

local delta = math.max(0, now-last_refreshed)
local filled_tokens = math.min(capacity, last_tokens+(delta*rate))
local allowed = filled_tokens >= requested
local new_tokens = filled_tokens
if allowed then
    new_tokens = filled_tokens - requested
end

redis.call("setex", KEYS[1], ttl, new_tokens)
redis.call("setex", KEYS[2], ttl, now)

return allowed`
	tokenFormat     = "{%s}.tokens"
	timestampFormat = "{%s}.ts"
	pingInterval    = time.Millisecond * 100
)

// A TokenLimiter controls how frequently events are allowed to happen with in one second.
type TokenLimiter struct {
	rate           int            // 每秒生产速率
	burst          int            // 桶容量
	store          *redis.Redis   // 存储容器
	tokenKey       string         // redis key
	timestampKey   string         // 桶刷新时间key
	rescueLock     sync.Mutex     // lock
	redisAlive     uint32         // redis健康标识
	rescueLimiter  *xrate.Limiter // redis故障时采用进程内 令牌桶限流器
	monitorStarted bool           // redis监控探测任务标识
}

// NewTokenLimiter returns a new TokenLimiter that allows events up to rate and permits
// bursts of at most burst tokens.
func NewTokenLimiter(rate, burst int, store *redis.Redis, key string) *TokenLimiter {
	tokenKey := fmt.Sprintf(tokenFormat, key)
	timestampKey := fmt.Sprintf(timestampFormat, key)

	return &TokenLimiter{
		rate:          rate,
		burst:         burst,
		store:         store,
		tokenKey:      tokenKey,
		timestampKey:  timestampKey,
		redisAlive:    1,
		rescueLimiter: xrate.NewLimiter(xrate.Every(time.Second/time.Duration(rate)), burst),
	}
}

// Allow is shorthand for AllowN(time.Now(), 1).
func (lim *TokenLimiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

// AllowN reports whether n events may happen at time now.
// Use this method if you intend to drop / skip events that exceed the rate.
// Otherwise, use Reserve or Wait.
func (lim *TokenLimiter) AllowN(now time.Time, n int) bool {
	return lim.reserveN(now, n)
}

func (lim *TokenLimiter) reserveN(now time.Time, n int) bool {
	// 判断redis是否健康
	// redis故障时采用进程内限流器
	// 兜底保障
	if atomic.LoadUint32(&lim.redisAlive) == 0 {
		return lim.rescueLimiter.AllowN(now, n)
	}
	// 执行脚本获取令牌
	resp, err := lim.store.Eval(
		script,
		[]string{
			lim.tokenKey,
			lim.timestampKey,
		},
		[]string{
			strconv.Itoa(lim.rate),
			strconv.Itoa(lim.burst),
			strconv.FormatInt(now.Unix(), 10),
			strconv.Itoa(n),
		})
	// redis allowed == false
	// Lua boolean false -> r Nil bulk reply
	// 特殊处理key不存在的情况
	if err == redis.Nil {
		return false
	}
	if err != nil {
		logx.Errorf("fail to use rate limiter: %s, use in-process limiter for rescue", err)
		// 执行异常，开启redis健康探测任务
		// 同时采用进程内限流器作为兜底
		lim.startMonitor()
		return lim.rescueLimiter.AllowN(now, n)
	}

	code, ok := resp.(int64)
	if !ok {
		logx.Errorf("fail to eval redis script: %v, use in-process limiter for rescue", resp)
		lim.startMonitor()
		return lim.rescueLimiter.AllowN(now, n)
	}

	// redis allowed == true
	// Lua boolean true -> r integer reply with value of 1
	return code == 1
}

// 开启redis健康探测
func (lim *TokenLimiter) startMonitor() {
	lim.rescueLock.Lock()
	defer lim.rescueLock.Unlock()
	// 防止重复开启
	if lim.monitorStarted {
		return
	}
	// 设置任务和健康标识
	lim.monitorStarted = true
	atomic.StoreUint32(&lim.redisAlive, 0)
	// 健康探测
	go lim.waitForRedis()
}

// redis健康探测定时任务
func (lim *TokenLimiter) waitForRedis() {
	// 健康探测成功时回调此函数
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		lim.rescueLock.Lock()
		lim.monitorStarted = false
		lim.rescueLock.Unlock()
	}()

	for range ticker.C {
		// ping属于redis内置健康探测命令
		if lim.store.Ping() {
			// 健康探测成功，设置健康标识
			atomic.StoreUint32(&lim.redisAlive, 1)
			return
		}
	}
}
