package portforward

import (
	"math/rand"
)

// RandomPort will return a random port number.
func RandomPort() int {
	min := 32012
	max := 64319
	return (rand.Intn(max-min) + min)
}
