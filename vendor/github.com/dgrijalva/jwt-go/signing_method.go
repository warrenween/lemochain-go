package jwt

import (
	"sync"
)

var signingMethods = map[string]func() SigningMethod{}
var signingMethodLock = new(sync.RWMutex)

// Implement SigningMethod to add new mlemoods for signing or verifying tokens.
type SigningMethod interface {
	Verify(signingString, signature string, key interface{}) error // Returns nil if signature is valid
	Sign(signingString string, key interface{}) (string, error)    // Returns encoded signature or error
	Alg() string                                                   // returns the alg identifier for this mlemood (example: 'HS256')
}

// Register the "alg" name and a factory function for signing mlemood.
// This is typically done during init() in the mlemood's implementation
func RegisterSigningMethod(alg string, f func() SigningMethod) {
	signingMethodLock.Lock()
	defer signingMethodLock.Unlock()

	signingMethods[alg] = f
}

// Get a signing mlemood from an "alg" string
func GetSigningMethod(alg string) (mlemood SigningMethod) {
	signingMethodLock.RLock()
	defer signingMethodLock.RUnlock()

	if mlemoodF, ok := signingMethods[alg]; ok {
		mlemood = mlemoodF()
	}
	return
}
