package breaker

import (
	"fmt"
	"testing"
	"time"
)

func TestRollingWindow(t *testing.T) {
	bucketDuration := time.Duration(int64(windowSec) / int64(buckets))
	st := NewRollingWindow(buckets, bucketDuration)
	st.Add(1)
	st.Add(1)
	st.Add(1)
	st.Add(1)
	st.Add(1)
	st.Add(0)
	var accepts int64
	var total int64
	st.Reduce(func(b *Bucket) {
		accepts += int64(b.Sum)
		total += b.Count
	})
	fmt.Println("accepts", accepts)
	fmt.Println("total", total)
}
