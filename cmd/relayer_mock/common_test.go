package main

import (
	"fmt"
	"math/big"
	"testing"
)

func Test_moneyChange(t *testing.T) {
	a := new(big.Int).SetInt64(100000)
	b := new(big.Int).SetInt64(100001)
	fmt.Println(a.Abs(a.Sub(a, b)))
	a = new(big.Int).SetInt64(100001)
	b = new(big.Int).SetInt64(100000)
	fmt.Println(a.Abs(a.Sub(a, b)))
	a = new(big.Int).SetInt64(100000)
	b = new(big.Int).SetInt64(100000)
	fmt.Println(a.Abs(a.Sub(a, b)))
}
func Test_floatPrint(t *testing.T) {
	fmt.Println(new(big.Float).SetInt64(111))
	fmt.Println(*new(big.Float).SetInt64(111))
}
