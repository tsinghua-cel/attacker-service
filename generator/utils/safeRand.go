package utils

import "math/rand"

func SafeRand(n int) int {
	if n == 0 {
		return 0
	}
	r := rand.Intn(n)
	if r == 0 {
		r++
	}
	return r
}
