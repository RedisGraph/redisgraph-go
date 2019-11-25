package redisgraph

import (
	"crypto/rand"
	"fmt"
	"strings"
	"strconv"
)

// go array to string is [1 2 3] for [1, 2, 3] array
// cypher expects comma separated array
func arrayToString(arr []interface{}) string {
	var arrayLength = len(arr)
	strArray := []string{}
	for i := 0; i < arrayLength; i++ {
		strArray = append(strArray, ToString(arr[i]))
	}
	return "[" + strings.Join(strArray, ",") + "]"
}

func ToString(i interface{}) string {
	if(i == nil) {
		return "null"
	}

	switch i.(type) {
	case string:
		s := i.(string)
		return strconv.Quote(s)
	case int:
		return strconv.Itoa(i.(int))
	case float64:
		return strconv.FormatFloat(i.(float64), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(i.(bool))
	case []interface {}:
		arr := i.([]interface{})
		return arrayToString(arr)
	default:
		panic("Unrecognized type to convert to string")
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
