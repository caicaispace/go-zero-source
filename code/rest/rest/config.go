package rest

import (
	"time"

	"github.com/zeromicro/go-zero/core/service"
)

type (
	// A PrivateKeyConf is a private key config.
	PrivateKeyConf struct {
		Fingerprint string // 指纹
		KeyFile     string // key 文件
	}

	// A SignatureConf is a signature config.
	SignatureConf struct {
		Strict      bool             `json:",default=false"` // 是否严格
		Expiry      time.Duration    `json:",default=1h"`    // 过期时间
		PrivateKeys []PrivateKeyConf // 私有 keys
	}

	// A RestConf is a http service config.
	// Why not name it as Conf, because we need to consider usage like:
	//  type Config struct {
	//     zrpc.RpcConf
	//     rest.RestConf
	//  }
	// if with the name Conf, there will be two Conf inside Config.
	// rest 服务配置
	RestConf struct {
		service.ServiceConf        // 业务服务配置
		Host                string `json:",default=0.0.0.0"` // host
		Port                int    // port
		CertFile            string `json:",optional"` // cret 文件
		KeyFile             string `json:",optional"` // key 文件
		Verbose             bool   `json:",optional"`
		MaxConns            int    `json:",default=10000"`   // 单服务可承载最大并发数
		MaxBytes            int64  `json:",default=1048576"` // 单服务单次可承载最大数据
		// milliseconds
		Timeout      int64         `json:",default=3000"`               // 服务超时时间
		CpuThreshold int64         `json:",default=900,range=[0:1000]"` // 服务熔断阀值
		Signature    SignatureConf `json:",optional"`                   // 签名配置
	}
)
