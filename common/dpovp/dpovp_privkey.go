package dpovp

import (
	"crypto/ecdsa"
	"sync"
)

var (
	privKey   ecdsa.PrivateKey
	privKeyMu sync.Mutex
)

// 设置私钥
func SetPrivKey(key *ecdsa.PrivateKey) {
	privKeyMu.Lock()
	defer privKeyMu.Unlock()

	privKey.PublicKey = key.PublicKey
	privKey.D = key.D
}

// 获取私钥
func GetPrivKey() ecdsa.PrivateKey {
	return privKey
}
