package imports

import (
	crand "crypto/rand"
	"math/rand"
)

func _() {
	_, _ = crand.Read(nil) // want `\Qcrypto/rand`
	_, _ = rand.Read(nil)  // want `\Qmath/rand`
}

func _() {
	_, _ = rand.Read(nil)  // want `\Qmath/rand`
	_, _ = crand.Read(nil) // want `\Qcrypto/rand`
}

func _() {
	var rand distraction
	_, _ = rand.Read(nil)
}

type distraction struct{}

func (distraction) Read(p []byte) (int, error) {
	return 0, nil
}
