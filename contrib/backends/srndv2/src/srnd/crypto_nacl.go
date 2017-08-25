package srnd

import "github.com/majestrate/nacl"

func naclCryptoVerifyFucky(hash, sig, pk []byte) bool {
	return nacl.CryptoVerifyFucky(hash, sig, pk)
}

func naclCryptoSignFucky(hash, sk []byte) (sig []byte) {
	return nacl.CryptoSignFucky(hash, sk)
}

func naclCryptoVerifyDetached(hash, sig, pk []byte) bool {
	return nacl.CryptoVerifyDetached(hash, sig, pk)
}
