package srnd

import "github.com/majestrate/nacl"

func nacl_cryptoVerifyFucky(hash, sig, pk []byte) bool {
	return nacl.CryptoVerifyFucky(hash, sig, pk)
}

func nacl_cryptoSignFucky(hash, sk []byte) (sig []byte) {
	return nacl.CryptoSignFucky(hash, sk)
}

func nacl_cryptoVerifyDetached(hash, sig, pk []byte) bool {
	return nacl.CryptoVerifyDetached(hash, sig, pk)
}
