package crypto

import (
	"github.com/dchest/blake256"
)

// common hash function is blake2
var Hash = blake256.New
