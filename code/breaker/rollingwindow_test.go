package breaker

import (
	"fmt"
	"testing"
	"time"

	"gozerosource/code/breaker/core/collection"
)

// https://juejin.cn/post/6968606293191983111

const (
	// 250ms for bucket duration
	windowSec = time.Second * 10 // 窗口时间
	buckets   = 40               // bucket 数量
	k         = 1.5              // 倍值（越小越敏感）
)

func Test_RollingWindow(t *testing.T) {
	bucketDuration := time.Duration(int64(windowSec) / int64(buckets))
	fmt.Println(bucketDuration)
	// st := collection.NewRollingWindow(buckets, bucketDuration, collection.IgnoreCurrentBucket())
	st := collection.NewRollingWindow(buckets, bucketDuration)
	for i := 0; i < 50; i++ {
		time.Sleep(20 * time.Millisecond)
		st.Add(float64(i))
	}
	var accepts int64
	var total int64
	st.Reduce(func(b *collection.Bucket) {
		accepts += int64(b.Sum)
		total += b.Count
	})
	fmt.Println("accepts", accepts)
	fmt.Println("total", total)
}
