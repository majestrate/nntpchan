package util

import (
	"crypto/sha1"
	"fmt"
)

// message id hash
func HashMessageID(msgid string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(msgid)))
}
