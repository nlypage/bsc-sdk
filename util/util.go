package util

import (
	"math"
	"math/big"
)

func BnbBigIntToFloat(num *big.Int) float64 {
	float, _ := num.Float64()
	return float / (10 * math.Pow(10, 17))
}

func FloatToBnbInt64(num float64) int64 {
	return int64(num * (10 * math.Pow(10, 17)))
}
