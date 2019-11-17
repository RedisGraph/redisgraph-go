package redisgraph

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"strings"
)

// go array to string is [1 2 3] for [1, 2, 3] array
// cypher expects comma separated array
func arrayToString(arr interface{}) interface{} {
	v := reflect.ValueOf(arr)
	var arrayLength = v.Len()
	strArray := []string{}
	for i := 0; i < arrayLength; i++ {
		strArray = append(strArray, fmt.Sprintf("%v", ToString(v.Index(i))))
	}
	return "[" + strings.Join(strArray[:], ",") + "]"
}

func ToString(i interface{}) interface{} {
	if(i == nil) {
		return "null"
	}
	v := reflect.ValueOf(i)
	switch reflect.TypeOf(i).Kind() {
	case reflect.String:
		s := v.String()
		if len(s) == 0 {
			return "\"\""
		}
		if s[0] != '"' {
			s = "\"" + s
		}
		if s[len(s)-1] != '"' {
			s += "\""
		}
		return s
	case reflect.Slice:
		return arrayToString(i)
	case reflect.Array:
		return arrayToString(i)
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

func BuildParamsHeader(params map[string]interface{}) (string) {
	header := "CYPHER "
	for key, value := range params {
		header += fmt.Sprintf("%s=%v ", key, ToString(value))
	}
	return header
}
