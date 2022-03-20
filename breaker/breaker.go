package breaker

type (
	Breaker interface {
		// 熔断器名称
		Name() string

		// 熔断方法，执行请求时必须手动上报执行结果
		// 适用于简单无需自定义快速失败，无需自定义判定请求结果的场景
		// 相当于手动挡。。。
		Allow() (Promise, error)

		// 熔断方法，自动上报执行结果
		// 自动挡。。。
		Do(req func() error) error

		// 熔断方法
		// acceptable - 支持自定义判定执行结果
		DoWithAcceptable(req func() error, acceptable Acceptable) error

		// 熔断方法
		// fallback - 支持自定义快速失败
		DoWithFallback(req func() error, fallback func(err error) error) error

		// 熔断方法
		// fallback - 支持自定义快速失败
		// acceptable - 支持自定义判定执行结果
		DoWithFallbackAcceptable(req func() error, fallback func(err error) error, acceptable Acceptable) error
	}
	// Promise interface defines the callbacks that returned by Breaker.Allow.
	Promise internalPromise
)

func defaultAcceptable(err error) bool {
	return err == nil
}

type breaker struct {
	gb *googleBreaker
}

func newBreaker() *breaker {
	return &breaker{
		gb: newGoogleBreaker(),
	}
}

func (b *breaker) Name() string {
	return ""
}

func (b *breaker) Allow() (Promise, error) {
	return b.gb.allow()
}

func (b *breaker) Do(req func() error) error {
	return b.gb.doReq(req, nil, defaultAcceptable)
}

func (b *breaker) DoWithAcceptable(req func() error, acceptable Acceptable) error {
	return b.gb.doReq(req, nil, acceptable)
}

func (b *breaker) DoWithFallback(req func() error, fallback func(err error) error) error {
	return b.gb.doReq(req, fallback, defaultAcceptable)
}

func (b *breaker) DoWithFallbackAcceptable(req func() error, fallback func(err error) error, acceptable Acceptable) error {
	return b.gb.doReq(req, fallback, acceptable)
}
