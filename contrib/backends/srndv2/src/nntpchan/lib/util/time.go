package util

import "time"

// time for right now as int64
func TimeNow() int64 {
	return time.Now().UTC().Unix()
}
