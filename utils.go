package redisgraph

import (
	"crypto/rand"
)

func QuoteString(i interface{}) interface{} {
	switch x := i.(type) {
	case string:
		if len(x) == 0 {
			return "\"\""
		}
		if x[0] != '"' {
			x = "\"" + x
		}
		if x[len(x)-1] != '"' {
			x += "\""
		}
		return x
	default:
		return i
	}
}

// https://medium.com/@kpbird/golang-generate-fixed-size-random-string-dd6dbd5e63c0
func RandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	output := make([]byte, n)
	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)
	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}
	l := len(letterBytes)
	// fill output
	for pos := range output {
		// get random item
		random := uint8(randomness[pos])
		// random % 64
		randomPos := random % uint8(l)
		// put into output
		output[pos] = letterBytes[randomPos]
	}
	return string(output)
}