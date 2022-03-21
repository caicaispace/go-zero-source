package internal

import (
	"strings"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"

	"google.golang.org/grpc/resolver"
)

type discovBuilder struct{}

func (b *discovBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (
	resolver.Resolver, error,
) {
	hosts := strings.FieldsFunc(target.Authority, func(r rune) bool {
		return r == EndpointSepChar
	})
	// 获取服务列表
	sub, err := discov.NewSubscriber(hosts, target.Endpoint)
	if err != nil {
		return nil, err
	}

	update := func() {
		var addrs []resolver.Address
		for _, val := range subset(sub.Values(), subsetSize) {
			addrs = append(addrs, resolver.Address{
				Addr: val,
			})
		}
		// 调用UpdateState方法更新
		if err := cc.UpdateState(resolver.State{
			Addresses: addrs,
		}); err != nil {
			logx.Errorf("%s", err)
		}
	}
	// 添加监听，当服务地址发生变化会触发更新
	sub.AddListener(update)
	// 更新服务列表
	update()

	return &nopResolver{cc: cc}, nil
}

func (b *discovBuilder) Scheme() string {
	return DiscovScheme
}
