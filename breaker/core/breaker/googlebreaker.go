package breaker

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"

	"gozerosource/breaker/core/collection"
)

const (
	// 250ms for bucket duration
	windowSec  = time.Second * 10
	buckets    = 40
	k          = 1.5
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

type internalPromise interface {
	Accept()
	Reject()
}

type Acceptable func(err error) bool

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
	accepts, total := b.History()
	weightedAccepts := b.k * float64(accepts)
	// Google Sre过载保护算法 https://landing.google.com/sre/sre-book/chapters/handling-overload/#eq2101
	dropRatio := math.Max(0, (float64(total-protection)-weightedAccepts)/float64(total+1))
	if dropRatio <= 0 {
		return nil
	}

	if b.proba.TrueOnProba(dropRatio) {
		return ErrServiceUnavailable
	}

	return nil
}

// 熔断方法，执行请求时必须手动上报执行结果
// 适用于简单无需自定义快速失败，无需自定义判定请求结果的场景
// 相当于手动挡。。。
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
func (b *googleBreaker) DoReq(req func() error, fallback func(err error) error, acceptable Acceptable) error {
	if err := b.accept(); err != nil {
		if fallback != nil {
			return fallback(err)
		}

		return err
	}

	defer func() {
		if e := recover(); e != nil {
			b.markFailure()
			panic(e)
		}
	}()

	err := req()
	if acceptable(err) {
		b.markSuccess()
	} else {
		b.markFailure()
	}

	return err
}

// 正常请求计数
func (b *googleBreaker) markSuccess() {
	b.stat.Add(1)
}

// 异常请求计数
func (b *googleBreaker) markFailure() {
	b.stat.Add(0)
}

// 历史数据
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
