package utils

import (
	"math/rand"
)

func RandomNum(min, max int) int {
	return rand.Intn(max-min+1) + min
}
