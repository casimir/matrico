package data

import "time"

func NowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond/time.Nanosecond)
}
