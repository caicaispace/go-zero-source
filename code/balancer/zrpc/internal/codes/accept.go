package codes

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 检测是否是可接受错误
// Acceptable checks if given error is acceptable.
func Acceptable(err error) bool {
	switch status.Code(err) {
	case codes.DeadlineExceeded, codes.Internal, codes.Unavailable, codes.DataLoss:
		return false
	default:
		return true
	}
}
