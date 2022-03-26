package breaker

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"

	"gozerosource/code/breaker/core/collection"
)

const (
	// 250ms for bucket duration
	windowSec  = time.Second * 10 // 窗口时间
	buckets    = 40               // bucket 数量
	k          = 1.5              // 倍值（越小越敏感）
	protection = 5
)

// A Proba is used to test if true on given probability.
type Proba struct {
	// rand.New(...) returns a non thread safe object
	r    *rand.Rand
	lock sync.Mutex
}

// NewProba returns a Proba.
func NewProba() *Proba {
	return &Proba{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// 检查给定概率是否为真
// TrueOnProba checks if true on given probability.
func (p *Proba) TrueOnProba(proba float64) (truth bool) {
	p.lock.Lock()
	truth = p.r.Float64() < proba
	p.lock.Unlock()
	return
}

// ErrServiceUnavailable is returned when the Breaker state is open.
var ErrServiceUnavailable = errors.New("circuit breaker is open")

// googleBreaker is a netflixBreaker pattern from google.
// see Client-Side Throttling section in https://landing.google.com/sre/sre-book/chapters/handling-overload/
type googleBreaker struct {
	k     float64
	stat  *collection.RollingWindow
	proba *Proba
}

func NewGoogleBreaker() *googleBreaker {
	bucketDuration := time.Duration(int64(windowSec) / int64(buckets))
	st := collection.NewRollingWindow(buckets, bucketDuration)
	return &googleBreaker{
		stat:  st,
		k:     k,
		proba: NewProba(),
	}
}

// 判断是否触发熔断
func (b *googleBreaker) accept() error {
	// 获取最近一段时间的统计数据
	accepts, total := b.History()
	// 计算动态熔断概率
	weightedAccepts := b.k * float64(accepts)
	// Google Sre过载保护算法 https://landing.google.com/sre/sre-book/chapters/handling-overload/#eq2101
	dropRatio := math.Max(0, (float64(total-protection)-weightedAccepts)/float64(total+1))
	if dropRatio <= 0 {
		return nil
	}
	// 随机产生0.0-1.0之间的随机数与上面计算出来的熔断概率相比较
	// 如果随机数比熔断概率小则进行熔断
	if b.proba.TrueOnProba(dropRatio) {
		return ErrServiceUnavailable
	}

	return nil
}

// 熔断方法，执行请求时必须手动上报执行结果
// 适用于简单无需自定义快速失败，无需自定义判定请求结果的场景
// 相当于手动挡。。。
// 返回一个promise异步回调对象，可由开发者自行决定是否上报结果到熔断器
func (b *googleBreaker) Allow() (internalPromise, error) {
	if err := b.accept(); err != nil {
		return nil, err
	}

	return googlePromise{
		b: b,
	}, nil
}

// 熔断方法，自动上报执行结果
// 自动挡。。。
// req 熔断对象方法
// fallback 自定义快速失败函数，可对熔断产生的err进行包装后返回
// acceptable 对本次未熔断时执行请求的结果进行自定义的判定，比如可以针对http.code,rpc.code,body.code
func (b *googleBreaker) DoReq(req func() error, fallback func(err error) error, acceptable Acceptable) error {
	// 判定是否熔断
	if err := b.accept(); err != nil {
		// 熔断中，如果有自定义的fallback则执行
		if fallback != nil {
			return fallback(err)
		}

		return err
	}
	// 如果执行req()过程发生了panic，依然判定本次执行失败上报至熔断器
	defer func() {
		if e := recover(); e != nil {
			b.markFailure()
			panic(e)
		}
	}()
	// 执行请求
	err := req()
	// 判定请求成功
	if acceptable(err) {
		b.markSuccess()
	} else {
		b.markFailure()
	}

	return err
}

// 上报成功
func (b *googleBreaker) markSuccess() {
	b.stat.Add(1)
}

// 上报失败
func (b *googleBreaker) markFailure() {
	b.stat.Add(0)
}

// 统计数据
// accepts 成功次数
// total 总次数
func (b *googleBreaker) History() (accepts, total int64) {
	b.stat.Reduce(func(b *collection.Bucket) {
		accepts += int64(b.Sum)
		total += b.Count
	})

	return
}

type googlePromise struct {
	b *googleBreaker
}

// 正常请求计数
func (p googlePromise) Accept() {
	p.b.markSuccess()
}

// 异常请求计数
func (p googlePromise) Reject() {
	p.b.markFailure()
}
