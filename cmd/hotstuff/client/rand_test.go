package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func Test_Rand(t *testing.T) {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	test := GenerateExpRand(0.1, r)
	fmt.Println(test)
}

func GenerateExpRand(lambda float64, r *rand.Rand) float64 {
	return r.ExpFloat64() / lambda
}
