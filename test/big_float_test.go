package test

import (
	"math/big"
	"testing"

	dec "github.com/shopspring/decimal"

	"github.com/mapprotocol/atlas/helper/decimal"
	"github.com/mapprotocol/atlas/params"

)

func decimalFloat64(number *big.Int, digits int64) float64 {
	amount, _ := dec.NewFromString(number.String())
	f, _ := amount.Div(decimal.Precision(digits)).Float64()
	return f
}


func TestFloat(t *testing.T) {
	number, _ := new(big.Int).SetString("13363474503370876427871", 10)

	t.Log(decimalFloat64(number, 24))
	t.Log(params.Float64(number, params.Fixidity1))
}