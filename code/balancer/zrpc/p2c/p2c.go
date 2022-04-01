package p2c

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gozerosource/code/balancer/zrpc/internal/codes"

	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/timex"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

const (
	// Name is the name of p2c balancer.
	Name = "p2c_ewma"

	decayTime       = int64(time.Second * 10) // default value from finagle（衰退时间）
	forcePick       = int64(time.Second)      // 强制节点选取时间间隔
	initSuccess     = 1000                    // 初始连接健康值
	throttleSuccess = initSuccess / 2         // 连接非健康临界值
	penalty         = int64(math.MaxInt32)    // 负载状态最大值
	pickTimes       = 3                       // 随机选取节点次数
	logInterval     = time.Minute             // 输出节点状态间隔时间
)

var emptyPickResult balancer.PickResult

func init() {
	balancer.Register(newBuilder())
}

type p2cPickerBuilder struct{}

// gRPC 在节点有更新的时候会调用 Build 方法，传入所有节点信息，
// 我们在这里把每个节点信息用 subConn 结构保存起来。
// 并归并到一起用 p2cPicker 结构保存起来
func (b *p2cPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	readySCs := info.ReadySCs
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	var conns []*subConn
	for conn, connInfo := range readySCs {
		conns = append(conns, &subConn{
			addr:    connInfo.Address,
			conn:    conn,
			success: initSuccess,
		})
	}

	return &p2cPicker{
		conns: conns,
		r:     rand.New(rand.NewSource(time.Now().UnixNano())),
		stamp: syncx.NewAtomicDuration(),
	}
}

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, new(p2cPickerBuilder), base.Config{HealthCheck: true})
}

type p2cPicker struct {
	conns []*subConn // 保存所有节点的信息
	r     *rand.Rand
	stamp *syncx.AtomicDuration
	lock  sync.Mutex
}

// 选取节点算法（grpc 自定义负载均衡算法）
func (p *p2cPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	var chosen *subConn
	switch len(p.conns) {
	case 0: // 没有节点，返回错误
		return emptyPickResult, balancer.ErrNoSubConnAvailable
	case 1: // 有一个节点，直接返回这个节点
		chosen = p.choose(p.conns[0], nil)
	case 2: // 有两个节点，计算负载，返回负载低的节点
		chosen = p.choose(p.conns[0], p.conns[1])
	default: // 有多个节点，p2c 挑选两个节点，比较这两个节点的负载，返回负载低的节点
		var node1, node2 *subConn
		// 3次随机选择两个节点
		for i := 0; i < pickTimes; i++ {
			a := p.r.Intn(len(p.conns))
			b := p.r.Intn(len(p.conns) - 1)
			// 防止选择同一个
			if b >= a {
				b++
			}
			node1 = p.conns[a]
			node2 = p.conns[b]
			// 如果这次选择的节点达到了健康要求, 就中断选择
			if node1.healthy() && node2.healthy() {
				break
			}
		}

		chosen = p.choose(node1, node2)
	}

	atomic.AddInt64(&chosen.inflight, 1)
	atomic.AddInt64(&chosen.requests, 1)

	return balancer.PickResult{
		SubConn: chosen.conn,
		Done:    p.buildDoneFunc(chosen),
	}, nil
}

// grpc 请求结束时调用
// 存储本次请求耗时等信息，并计算出 EWMA值 保存起来，供下次请求时计算负载等情况的使用
func (p *p2cPicker) buildDoneFunc(c *subConn) func(info balancer.DoneInfo) {
	start := int64(timex.Now())
	return func(info balancer.DoneInfo) {
		// 正在处理的请求数减 1
		atomic.AddInt64(&c.inflight, -1)
		now := timex.Now()
		// 保存本次请求结束时的时间点，并取出上次请求时的时间点
		last := atomic.SwapInt64(&c.last, int64(now))
		td := int64(now) - last
		if td < 0 {
			td = 0
		}
		// 用牛顿冷却定律中的衰减函数模型计算EWMA算法中的β值
		// 牛顿冷却算法 https://www.ruanyifeng.com/blog/2012/03/ranking_algorithm_newton_s_law_of_cooling.html
		w := math.Exp(float64(-td) / float64(decayTime))
		// 保存本次请求的耗时
		lag := int64(now) - start
		if lag < 0 {
			lag = 0
		}
		olag := atomic.LoadUint64(&c.lag)
		if olag == 0 {
			w = 0
		}
		// 计算 EWMA 值
		// EWMA（指数加权移动平均算法） https://blog.csdn.net/mzpmzk/article/details/80085929
		atomic.StoreUint64(&c.lag, uint64(float64(olag)*w+float64(lag)*(1-w)))
		success := initSuccess
		if info.Err != nil && !codes.Acceptable(info.Err) {
			success = 0
		}
		// 健康状态
		osucc := atomic.LoadUint64(&c.success)
		atomic.StoreUint64(&c.success, uint64(float64(osucc)*w+float64(success)*(1-w)))

		stamp := p.stamp.Load()
		if now-stamp >= logInterval {
			if p.stamp.CompareAndSwap(stamp, now) {
				p.logStats()
			}
		}
	}
}

// 比较两个节点的负载情况，选择负载低的
func (p *p2cPicker) choose(c1, c2 *subConn) *subConn {
	start := int64(timex.Now())
	if c2 == nil {
		atomic.StoreInt64(&c1.pick, start)
		return c1
	}

	// c2 一直是负载较大的 node
	if c1.load() > c2.load() {
		c1, c2 = c2, c1 // 交换变量，方便判断
	}

	pick := atomic.LoadInt64(&c2.pick)
	// 如果(本次被选中的时间 - 上次被选中的时间 > forcePick && 本次与上次时间点不同)
	// return 负载较大 node
	if start-pick > forcePick && atomic.CompareAndSwapInt64(&c2.pick, pick, start) {
		return c2
	}

	// 返回负载较小 node
	atomic.StoreInt64(&c1.pick, start)
	return c1
}

// 输出所有节点状态信息
func (p *p2cPicker) logStats() {
	var stats []string

	p.lock.Lock()
	defer p.lock.Unlock()

	for _, conn := range p.conns {
		stats = append(stats, fmt.Sprintf("conn: %s, load: %d, reqs: %d",
			conn.addr.Addr, conn.load(), atomic.SwapInt64(&conn.requests, 0)))
	}

	fmt.Printf("p2c - %s", strings.Join(stats, "; "))
}

type subConn struct {
	lag      uint64 // 用来保存 ewma 值
	inflight int64  // 用在保存当前节点正在处理的请求总数
	success  uint64 // 用来标识一段时间内此连接的健康状态
	requests int64  // 用来保存请求总数
	last     int64  // 用来保存上一次请求耗时, 用于计算 ewma 值
	pick     int64  // 保存上一次被选中的时间点
	addr     resolver.Address
	conn     balancer.SubConn
}

// 节点健康情况
func (c *subConn) healthy() bool {
	return atomic.LoadUint64(&c.success) > throttleSuccess
}

// 节点负载情况
func (c *subConn) load() int64 {
	// plus one to avoid multiply zero
	// 通过 EWMA 计算节点的负载情况； 加 1 是为了避免为 0 的情况
	lag := int64(math.Sqrt(float64(atomic.LoadUint64(&c.lag) + 1)))
	load := lag * (atomic.LoadInt64(&c.inflight) + 1)
	if load == 0 {
		return penalty
	}

	return load
}
