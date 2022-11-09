package common

const BLSPubkeyLength = 48

// ZeroSecretKey represents a zero secret key.
var ZeroSecretKey = [32]byte{}

// InfinitePublicKey represents an infinite public key (G1 Point at Infinity).
var InfinitePublicKey = [BLSPubkeyLength]byte{0xC0}

// InfiniteSignature represents an infinite signature (G2 Point at Infinity).
var InfiniteSignature = [96]byte{0xC0}
